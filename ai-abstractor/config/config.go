package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Redis        RedisConfig
	OpenAI       OpenAIConfig
	Anthropic    AnthropicConfig
	App          AppConfig
	Capabilities CapabilityConfig
}

type RedisConfig struct {
	URL           string
	RequestCh     string
	ResponseCh    string
}

type OpenAIConfig struct {
	APIKey      string
	BaseURL     string
	Model       string
	MaxTokens   int
	Temperature float32
}

type AnthropicConfig struct {
	APIKey      string
	BaseURL     string
	Model       string
	MaxTokens   int
}

type AppConfig struct {
	LogLevel string
	Port     int
}

type CapabilityConfig struct {
	RefreshInterval time.Duration
	Enabled         bool
}

func Load() (*Config, error) {
	godotenv.Load()

	port := 8081
	if portStr := os.Getenv("PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	maxTokens := 4000
	if tokensStr := os.Getenv("OPENAI_MAX_TOKENS"); tokensStr != "" {
		if t, err := strconv.Atoi(tokensStr); err == nil {
			maxTokens = t
		}
	}

	anthropicMaxTokens := 4000
	if tokensStr := os.Getenv("ANTHROPIC_MAX_TOKENS"); tokensStr != "" {
		if t, err := strconv.Atoi(tokensStr); err == nil {
			anthropicMaxTokens = t
		}
	}

	temperature := float32(0.7)
	if tempStr := os.Getenv("OPENAI_TEMPERATURE"); tempStr != "" {
		if t, err := strconv.ParseFloat(tempStr, 32); err == nil {
			temperature = float32(t)
		}
	}

	refreshInterval := 5 * time.Minute
	if intervalStr := os.Getenv("CAPABILITY_REFRESH_INTERVAL"); intervalStr != "" {
		if interval, err := time.ParseDuration(intervalStr); err == nil {
			refreshInterval = interval
		}
	}

	config := &Config{
		Redis: RedisConfig{
			URL:        getEnv("REDIS_URL", "redis://localhost:6379"),
			RequestCh:  getEnv("AI_REQUEST_CHANNEL", "ai-requests"),
			ResponseCh: getEnv("AI_RESPONSE_CHANNEL", "ai-responses"),
		},
		OpenAI: OpenAIConfig{
			APIKey:      getEnv("OPENAI_API_KEY", ""),
			BaseURL:     getEnv("OPENAI_BASE_URL", ""),
			Model:       getEnv("OPENAI_MODEL", "gpt-4"),
			MaxTokens:   maxTokens,
			Temperature: temperature,
		},
		Anthropic: AnthropicConfig{
			APIKey:    getEnv("ANTHROPIC_API_KEY", ""),
			BaseURL:   getEnv("ANTHROPIC_BASE_URL", "https://api.anthropic.com"),
			Model:     getEnv("ANTHROPIC_MODEL", "claude-3-sonnet-20240229"),
			MaxTokens: anthropicMaxTokens,
		},
		App: AppConfig{
			LogLevel: getEnv("LOG_LEVEL", "info"),
			Port:     port,
		},
		Capabilities: CapabilityConfig{
			RefreshInterval: refreshInterval,
			Enabled:         getBoolEnv("CAPABILITY_ANNOUNCEMENTS_ENABLED", true),
		},
	}

	logrus.WithFields(logrus.Fields{
		"redis_url":        config.Redis.URL,
		"openai_model":     config.OpenAI.Model,
		"anthropic_model":  config.Anthropic.Model,
		"request_channel":  config.Redis.RequestCh,
		"response_channel": config.Redis.ResponseCh,
	}).Info("AI Abstractor configuration loaded")

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