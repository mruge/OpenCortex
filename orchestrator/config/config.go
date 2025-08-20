package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server       ServerConfig
	Redis        RedisConfig
	Services     ServicesConfig
	Orchestrator OrchestratorConfig
	Capabilities CapabilityConfig
}

type ServerConfig struct {
	Port            string
	Host            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type RedisConfig struct {
	Host         string
	Port         string
	Password     string
	Database     int
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
}

type ServicesConfig struct {
	DataService     ServiceConfig
	AIService       ServiceConfig
	ExecService     ServiceConfig
	MessageTimeout  time.Duration
	DefaultTimeout  time.Duration
}

type ServiceConfig struct {
	RequestChannel  string
	ResponseChannel string
	Timeout         time.Duration
}

type OrchestratorConfig struct {
	WorkspaceDir       string
	TemplatesDir       string
	MaxConcurrent      int
	DefaultTimeout     time.Duration
	ExecutionTTL       time.Duration
	CleanupInterval    time.Duration
	RecoveryEnabled    bool
	RecoveryInterval   time.Duration
}

type CapabilityConfig struct {
	RefreshInterval time.Duration
	Enabled         bool
}

func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:            getEnvOrDefault("ORCHESTRATOR_PORT", "8080"),
			Host:            getEnvOrDefault("ORCHESTRATOR_HOST", "0.0.0.0"),
			ReadTimeout:     getDurationOrDefault("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout:    getDurationOrDefault("SERVER_WRITE_TIMEOUT", 30*time.Second),
			ShutdownTimeout: getDurationOrDefault("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
		},
		Redis: RedisConfig{
			Host:         getEnvOrDefault("REDIS_HOST", "localhost"),
			Port:         getEnvOrDefault("REDIS_PORT", "6379"),
			Password:     getEnvOrDefault("REDIS_PASSWORD", ""),
			Database:     getIntOrDefault("REDIS_DATABASE", 0),
			PoolSize:     getIntOrDefault("REDIS_POOL_SIZE", 10),
			MinIdleConns: getIntOrDefault("REDIS_MIN_IDLE_CONNS", 5),
			MaxRetries:   getIntOrDefault("REDIS_MAX_RETRIES", 3),
		},
		Services: ServicesConfig{
			DataService: ServiceConfig{
				RequestChannel:  "data-requests",
				ResponseChannel: "data-responses",
				Timeout:         getDurationOrDefault("DATA_SERVICE_TIMEOUT", 60*time.Second),
			},
			AIService: ServiceConfig{
				RequestChannel:  "ai-requests",
				ResponseChannel: "ai-responses", 
				Timeout:         getDurationOrDefault("AI_SERVICE_TIMEOUT", 120*time.Second),
			},
			ExecService: ServiceConfig{
				RequestChannel:  "exec-requests",
				ResponseChannel: "exec-responses",
				Timeout:         getDurationOrDefault("EXEC_SERVICE_TIMEOUT", 300*time.Second),
			},
			MessageTimeout: getDurationOrDefault("MESSAGE_TIMEOUT", 300*time.Second),
			DefaultTimeout: getDurationOrDefault("DEFAULT_SERVICE_TIMEOUT", 60*time.Second),
		},
		Orchestrator: OrchestratorConfig{
			WorkspaceDir:     getEnvOrDefault("ORCHESTRATOR_WORKSPACE", "/tmp/orchestrator"),
			TemplatesDir:     getEnvOrDefault("ORCHESTRATOR_TEMPLATES", "./templates"),
			MaxConcurrent:    getIntOrDefault("MAX_CONCURRENT_WORKFLOWS", 10),
			DefaultTimeout:   getDurationOrDefault("DEFAULT_WORKFLOW_TIMEOUT", 3600*time.Second), // 1 hour
			ExecutionTTL:     getDurationOrDefault("EXECUTION_TTL", 24*time.Hour),
			CleanupInterval:  getDurationOrDefault("CLEANUP_INTERVAL", 1*time.Hour),
			RecoveryEnabled:  getBoolOrDefault("RECOVERY_ENABLED", true),
			RecoveryInterval: getDurationOrDefault("RECOVERY_INTERVAL", 5*time.Minute),
		},
		Capabilities: CapabilityConfig{
			RefreshInterval: getDurationOrDefault("CAPABILITY_REFRESH_INTERVAL", 5*time.Minute),
			Enabled:         getBoolOrDefault("CAPABILITY_ANNOUNCEMENTS_ENABLED", true),
		},
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}