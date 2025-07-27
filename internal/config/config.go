package config

import (
	"os"
	"time"
)

// Config holds the application configuration
type Config struct {
	HTTPPort         string
	GRPCPort         string
	DatabasePath     string
	UpdateInterval   time.Duration
	MaxMindLicenseKey string
	LogLevel         string
	EnableTLS        bool
	DemoMode         bool
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() *Config {
	return &Config{
		HTTPPort:         getEnv("HTTP_PORT", "8080"),
		GRPCPort:         getEnv("GRPC_PORT", "9090"),
		DatabasePath:     getEnv("DB_PATH", "./data/GeoLite2-Country.mmdb"),
		UpdateInterval:   getDurationEnv("UPDATE_INTERVAL", 24*time.Hour),
		MaxMindLicenseKey: getEnv("MAXMIND_LICENSE_KEY", ""),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
		EnableTLS:        getBoolEnv("ENABLE_TLS", false),
		DemoMode:         getBoolEnv("DEMO_MODE", false),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "true" || value == "1" {
		return true
	}
	if value == "false" || value == "0" {
		return false
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}