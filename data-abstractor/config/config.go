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
	Neo4j        Neo4jConfig
	MongoDB      MongoConfig
	Qdrant       QdrantConfig
	App          AppConfig
	Capabilities CapabilityConfig
}

type RedisConfig struct {
	URL     string
	Channel string
}

type Neo4jConfig struct {
	URL      string
	Username string
	Password string
}

type MongoConfig struct {
	URL      string
	Database string
}

type QdrantConfig struct {
	URL        string
	Collection string
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

	port := 8080
	if portStr := os.Getenv("PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
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
			URL:     getEnv("REDIS_URL", "redis://localhost:6379"),
			Channel: getEnv("REDIS_CHANNEL", "data-requests"),
		},
		Neo4j: Neo4jConfig{
			URL:      getEnv("NEO4J_URL", "bolt://localhost:7687"),
			Username: getEnv("NEO4J_USER", "neo4j"),
			Password: getEnv("NEO4J_PASSWORD", "password"),
		},
		MongoDB: MongoConfig{
			URL:      getEnv("MONGODB_URL", "mongodb://localhost:27017"),
			Database: getEnv("MONGODB_DATABASE", "enrichment"),
		},
		Qdrant: QdrantConfig{
			URL:        getEnv("QDRANT_URL", "http://localhost:6333"),
			Collection: getEnv("QDRANT_COLLECTION", "embeddings"),
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
		"redis_url":   config.Redis.URL,
		"neo4j_url":   config.Neo4j.URL,
		"mongodb_url": config.MongoDB.URL,
		"qdrant_url":  config.Qdrant.URL,
	}).Info("Configuration loaded")

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