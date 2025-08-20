package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"exec-agent/capabilities"
	"exec-agent/clients"
	"exec-agent/config"
	"exec-agent/handlers"

	"github.com/sirupsen/logrus"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load configuration")
	}

	level, err := logrus.ParseLevel(cfg.App.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.JSONFormatter{})

	logrus.Info("Starting Exec Agent service")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// Initialize Docker client
	dockerClient, err := clients.NewDockerClient(
		cfg.Docker.Host,
		cfg.Docker.APIVersion,
		cfg.Docker.WorkDir,
		cfg.Docker.NetworkName,
		cfg.Docker.CleanupTimeout,
	)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize Docker client")
	}
	defer func() {
		if err := dockerClient.Close(); err != nil {
			logrus.WithError(err).Error("Failed to close Docker client")
		}
	}()

	// Initialize Minio client
	minioClient, err := clients.NewMinioClient(
		cfg.Minio.Endpoint,
		cfg.Minio.AccessKeyID,
		cfg.Minio.SecretAccessKey,
		cfg.Minio.BucketName,
		cfg.Minio.UseSSL,
	)
	if err != nil {
		logrus.WithError(err).Warn("Failed to initialize Minio client - blob storage features will be disabled")
		minioClient = nil
	}

	// Initialize Redis client
	redisClient, err := clients.NewRedisClient(cfg.Redis.URL, cfg.Redis.RequestCh, cfg.Redis.ResponseCh)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to Redis")
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			logrus.WithError(err).Error("Failed to close Redis client")
		}
	}()

	// Initialize service proxy
	serviceProxy := handlers.NewServiceProxy(
		cfg.ServiceProxy.Port,
		cfg.ServiceProxy.DataAbstractorURL,
		cfg.ServiceProxy.AIAbstractorURL,
		redisClient.PublishToChannel,
	)

	// Start service proxy server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := serviceProxy.Start(ctx, cfg.ServiceProxy.Port); err != nil {
			logrus.WithError(err).Error("Service proxy server failed")
		}
	}()

	// Initialize execution handler
	executionHandler := handlers.NewExecutionHandler(dockerClient, minioClient, serviceProxy)

	// Initialize image scanner if enabled
	var imageScanner *capabilities.ImageScanner
	var enhancedCapabilities *capabilities.EnhancedExecCapabilities
	if cfg.ImageScan.Enabled {
		imageScanner = capabilities.NewImageScanner(
			cfg.ImageScan.KnownImages,
			cfg.ImageScan.ScanInterval,
		)
		
		// Start periodic scanning
		imageScanner.StartPeriodicScan(ctx)
		
		// Create enhanced capabilities with image scanner
		enhancedCapabilities = capabilities.NewEnhancedExecCapabilities(imageScanner)
		
		logrus.WithFields(logrus.Fields{
			"known_images":   len(cfg.ImageScan.KnownImages),
			"scan_interval": cfg.ImageScan.ScanInterval,
		}).Info("Image scanner initialized")
	}

	// Initialize capability manager if enabled
	var capabilityManager interface {
		Start(context.Context) error
		Stop()
	}
	if cfg.Capabilities.Enabled {
		if enhancedCapabilities != nil {
			// Use dynamic capability manager that handles image changes
			dynamicManager := capabilities.NewDynamicCapabilityManager(
				"exec-agent",
				enhancedCapabilities,
				redisClient.GetClient(),
				cfg.Capabilities.RefreshInterval,
				cfg.ImageScan.ScanInterval,
			)
			capabilityManager = dynamicManager
			
			logrus.WithFields(logrus.Fields{
				"image_refresh_interval": cfg.ImageScan.ScanInterval,
				"known_images":          len(cfg.ImageScan.KnownImages),
			}).Info("Dynamic capability manager configured with image scanning")
		} else {
			// Use basic capability manager
			basicManager := capabilities.NewCapabilityManager(
				"exec-agent",
				capabilities.GetExecAgentCapabilities(),
				redisClient.GetClient(),
				cfg.Capabilities.RefreshInterval,
			)
			capabilityManager = basicManager
		}

		// Start capability manager
		if err := capabilityManager.Start(ctx); err != nil {
			logrus.WithError(err).Error("Failed to start capability manager")
		} else {
			logrus.WithFields(logrus.Fields{
				"image_scanning": cfg.ImageScan.Enabled,
			}).Info("Capability manager started successfully")
		}
	}

	// Start Redis message listener
	wg.Add(1)
	go func() {
		defer wg.Done()
		
		messageHandler := func(data []byte) []byte {
			return executionHandler.HandleRequest(ctx, data)
		}
		
		if err := redisClient.Listen(ctx, messageHandler); err != nil && err != context.Canceled {
			logrus.WithError(err).Error("Redis listener failed")
		}
	}()

	// Setup graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	logrus.WithFields(logrus.Fields{
		"request_channel":    cfg.Redis.RequestCh,
		"response_channel":   cfg.Redis.ResponseCh,
		"docker_host":        cfg.Docker.Host,
		"work_dir":           cfg.Docker.WorkDir,
		"service_proxy_port": cfg.ServiceProxy.Port,
		"minio_enabled":      minioClient != nil,
	}).Info("Exec Agent service is running. Press Ctrl+C to stop.")

	<-sigCh
	logrus.Info("Shutdown signal received")

	// Stop capability manager if running
	if capabilityManager != nil {
		capabilityManager.Stop()
		logrus.Info("Capability manager stopped")
	}

	cancel()
	wg.Wait()

	logrus.Info("Exec Agent service stopped")
}