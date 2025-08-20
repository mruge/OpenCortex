package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"data-abstractor/capabilities"
	"data-abstractor/clients"
	"data-abstractor/config"
	"data-abstractor/handlers"

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

	logrus.Info("Starting Data Abstractor service")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	neo4jClient, err := clients.NewNeo4jClient(cfg.Neo4j.URL, cfg.Neo4j.Username, cfg.Neo4j.Password)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to Neo4j")
	}
	defer func() {
		if err := neo4jClient.Close(); err != nil {
			logrus.WithError(err).Error("Failed to close Neo4j client")
		}
	}()

	mongoClient, err := clients.NewMongoClient(cfg.MongoDB.URL, cfg.MongoDB.Database)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to MongoDB")
	}
	defer func() {
		if err := mongoClient.Close(); err != nil {
			logrus.WithError(err).Error("Failed to close MongoDB client")
		}
	}()

	qdrantClient, err := clients.NewQdrantClient(cfg.Qdrant.URL, cfg.Qdrant.Collection)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to Qdrant")
	}
	defer func() {
		if err := qdrantClient.Close(); err != nil {
			logrus.WithError(err).Error("Failed to close Qdrant client")
		}
	}()

	redisClient, err := clients.NewRedisClient(cfg.Redis.URL, "data-requests", "data-responses")
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to Redis")
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			logrus.WithError(err).Error("Failed to close Redis client")
		}
	}()

	dataHandler := handlers.NewDataHandler(neo4jClient, mongoClient, qdrantClient)

	// Initialize capability manager if enabled
	var capabilityManager *capabilities.CapabilityManager
	if cfg.Capabilities.Enabled {
		capabilityManager = capabilities.NewCapabilityManager(
			"data-abstractor",
			capabilities.GetDataAbstractorCapabilities(),
			redisClient.GetClient(),
			cfg.Capabilities.RefreshInterval,
		)

		// Start capability manager
		if err := capabilityManager.Start(ctx); err != nil {
			logrus.WithError(err).Error("Failed to start capability manager")
		} else {
			logrus.Info("Capability manager started successfully")
		}
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		
		messageHandler := func(data []byte) []byte {
			return dataHandler.HandleRequest(ctx, data)
		}
		
		if err := redisClient.Listen(ctx, messageHandler); err != nil && err != context.Canceled {
			logrus.WithError(err).Error("Redis listener failed")
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	logrus.Info("Data Abstractor service is running. Press Ctrl+C to stop.")

	<-sigCh
	logrus.Info("Shutdown signal received")

	// Stop capability manager if running
	if capabilityManager != nil {
		capabilityManager.Stop()
		logrus.Info("Capability manager stopped")
	}

	cancel()
	wg.Wait()

	logrus.Info("Data Abstractor service stopped")
}