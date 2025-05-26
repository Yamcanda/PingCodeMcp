package config

import (
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	ServerName    string
	ServerVersion string
	Port          string
	LogLevel      string
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		ServerName:    getEnv("SERVER_NAME", "pingcode-mcp-server"),
		ServerVersion: getEnv("SERVER_VERSION", "1.0.0"),
		Port:          getEnv("PORT", "8080"),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer with a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
