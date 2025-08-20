package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Redis        RedisConfig
	Docker       DockerConfig
	Minio        MinioConfig
	ServiceProxy ServiceProxyConfig
	App          AppConfig
	Capabilities CapabilityConfig
	ImageScan    ImageScanConfig
}

type RedisConfig struct {
	URL           string
	RequestCh     string
	ResponseCh    string
}

type DockerConfig struct {
	Host           string
	APIVersion     string
	WorkDir        string
	NetworkName    string
	CleanupTimeout time.Duration
}

type MinioConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
}

type ServiceProxyConfig struct {
	Port               int
	DataAbstractorURL  string
	AIAbstractorURL    string
}

type AppConfig struct {
	LogLevel string
	Port     int
}

type CapabilityConfig struct {
	RefreshInterval time.Duration
	Enabled         bool
}

type ImageScanConfig struct {
	Enabled         bool
	ScanInterval    time.Duration
	KnownImages     []string
}

func Load() (*Config, error) {
	godotenv.Load()

	port := 8082
	if portStr := os.Getenv("PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	proxyPort := 9000
	if portStr := os.Getenv("SERVICE_PROXY_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			proxyPort = p
		}
	}

	cleanupTimeoutSec := 300 // 5 minutes default
	if timeoutStr := os.Getenv("CLEANUP_TIMEOUT_SEC"); timeoutStr != "" {
		if t, err := strconv.Atoi(timeoutStr); err == nil {
			cleanupTimeoutSec = t
		}
	}

	useSSL := false
	if sslStr := os.Getenv("MINIO_USE_SSL"); sslStr == "true" {
		useSSL = true
	}

	refreshInterval := 5 * time.Minute
	if intervalStr := os.Getenv("CAPABILITY_REFRESH_INTERVAL"); intervalStr != "" {
		if interval, err := time.ParseDuration(intervalStr); err == nil {
			refreshInterval = interval
		}
	}

	imageScanInterval := 30 * time.Minute
	if intervalStr := os.Getenv("IMAGE_SCAN_INTERVAL"); intervalStr != "" {
		if interval, err := time.ParseDuration(intervalStr); err == nil {
			imageScanInterval = interval
		}
	}

	// Parse known images from environment variable (comma-separated)
	knownImages := []string{
		"python:3.9-slim",
		"node:16-alpine",
		"tensorflow/tensorflow:latest-py3",
		"pytorch/pytorch:latest",
		"data-processor:latest",
		"etl-worker:latest",
	}
	if imagesStr := os.Getenv("KNOWN_WORKER_IMAGES"); imagesStr != "" {
		knownImages = strings.Split(imagesStr, ",")
		// Trim whitespace from each image name
		for i, img := range knownImages {
			knownImages[i] = strings.TrimSpace(img)
		}
	}

	config := &Config{
		Redis: RedisConfig{
			URL:        getEnv("REDIS_URL", "redis://localhost:6379"),
			RequestCh:  getEnv("EXEC_REQUEST_CHANNEL", "exec-requests"),
			ResponseCh: getEnv("EXEC_RESPONSE_CHANNEL", "exec-responses"),
		},
		Docker: DockerConfig{
			Host:           getEnv("DOCKER_HOST", "unix:///var/run/docker.sock"),
			APIVersion:     getEnv("DOCKER_API_VERSION", "1.41"),
			WorkDir:        getEnv("EXEC_WORK_DIR", "/tmp/exec-agent"),
			NetworkName:    getEnv("DOCKER_NETWORK", "smart_data_abstractor_default"),
			CleanupTimeout: time.Duration(cleanupTimeoutSec) * time.Second,
		},
		Minio: MinioConfig{
			Endpoint:        getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKeyID:     getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretAccessKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
			UseSSL:          useSSL,
			BucketName:      getEnv("MINIO_BUCKET", "exec-data"),
		},
		ServiceProxy: ServiceProxyConfig{
			Port:              proxyPort,
			DataAbstractorURL: getEnv("DATA_ABSTRACTOR_URL", "http://data-abstractor:8080"),
			AIAbstractorURL:   getEnv("AI_ABSTRACTOR_URL", "http://ai-abstractor:8081"),
		},
		App: AppConfig{
			LogLevel: getEnv("LOG_LEVEL", "info"),
			Port:     port,
		},
		Capabilities: CapabilityConfig{
			RefreshInterval: refreshInterval,
			Enabled:         getBoolEnv("CAPABILITY_ANNOUNCEMENTS_ENABLED", true),
		},
		ImageScan: ImageScanConfig{
			Enabled:      getBoolEnv("IMAGE_SCAN_ENABLED", true),
			ScanInterval: imageScanInterval,
			KnownImages:  knownImages,
		},
	}

	logrus.WithFields(logrus.Fields{
		"redis_url":         config.Redis.URL,
		"docker_host":       config.Docker.Host,
		"minio_endpoint":    config.Minio.Endpoint,
		"work_dir":          config.Docker.WorkDir,
		"request_channel":   config.Redis.RequestCh,
		"response_channel":  config.Redis.ResponseCh,
		"service_proxy_port": config.ServiceProxy.Port,
	}).Info("Exec Agent configuration loaded")

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}