package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"ai-abstractor/capabilities"
	"ai-abstractor/clients"
	"ai-abstractor/config"
	"ai-abstractor/handlers"

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

	logrus.Info("Starting AI Abstractor service")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// Initialize OpenAI client if API key is provided
	var openAIClient *clients.OpenAIClient
	if cfg.OpenAI.APIKey != "" {
		openAIClient, err = clients.NewOpenAIClient(
			cfg.OpenAI.APIKey,
			cfg.OpenAI.BaseURL,
			cfg.OpenAI.Model,
			cfg.OpenAI.MaxTokens,
			cfg.OpenAI.Temperature,
		)
		if err != nil {
			logrus.WithError(err).Error("Failed to initialize OpenAI client")
		}
	} else {
		logrus.Info("OpenAI API key not provided, OpenAI client disabled")
	}

	// Initialize Anthropic client if API key is provided
	var anthropicClient *clients.AnthropicClient
	if cfg.Anthropic.APIKey != "" {
		anthropicClient, err = clients.NewAnthropicClient(
			cfg.Anthropic.APIKey,
			cfg.Anthropic.BaseURL,
			cfg.Anthropic.Model,
			cfg.Anthropic.MaxTokens,
		)
		if err != nil {
			logrus.WithError(err).Error("Failed to initialize Anthropic client")
		}
	} else {
		logrus.Info("Anthropic API key not provided, Anthropic client disabled")
	}

	// Check that at least one AI provider is configured
	if openAIClient == nil && anthropicClient == nil {
		logrus.Fatal("No AI providers configured. Please provide OPENAI_API_KEY or ANTHROPIC_API_KEY")
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

	// Initialize AI handler
	aiHandler := handlers.NewAIHandler(openAIClient, anthropicClient)

	// Initialize capability manager if enabled
	var capabilityManager *capabilities.CapabilityManager
	if cfg.Capabilities.Enabled {
		capabilityManager = capabilities.NewCapabilityManager(
			"ai-abstractor",
			capabilities.GetAIAbstractorCapabilities(),
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

	// Start Redis message listener
	wg.Add(1)
	go func() {
		defer wg.Done()
		
		messageHandler := func(data []byte) []byte {
			return aiHandler.HandleRequest(ctx, data)
		}
		
		if err := redisClient.Listen(ctx, messageHandler); err != nil && err != context.Canceled {
			logrus.WithError(err).Error("Redis listener failed")
		}
	}()

	// Setup graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	logrus.WithFields(logrus.Fields{
		"request_channel":  cfg.Redis.RequestCh,
		"response_channel": cfg.Redis.ResponseCh,
		"openai_enabled":   openAIClient != nil,
		"anthropic_enabled": anthropicClient != nil,
	}).Info("AI Abstractor service is running. Press Ctrl+C to stop.")

	<-sigCh
	logrus.Info("Shutdown signal received")

	// Stop capability manager if running
	if capabilityManager != nil {
		capabilityManager.Stop()
		logrus.Info("Capability manager stopped")
	}

	cancel()
	wg.Wait()

	logrus.Info("AI Abstractor service stopped")
}