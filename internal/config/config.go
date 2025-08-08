package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	App      AppConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port         string
	Host         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Type string // "memory", "postgres", "redis"
	// Add other database-specific configs as needed
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	BaseURL         string
	Environment     string
	LogLevel        string
	ShortCodeLength int
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config := &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "3000"),
			Host:         getEnv("HOST", "0.0.0.0"),
			ReadTimeout:  getDurationEnv("READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getDurationEnv("IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			Type: getEnv("DB_TYPE", "memory"),
		},
		App: AppConfig{
			BaseURL:         getRequiredEnv("BASE_URL"),
			Environment:     getEnv("ENVIRONMENT", "development"),
			LogLevel:        getEnv("LOG_LEVEL", "info"),
			ShortCodeLength: getIntEnv("SHORT_CODE_LENGTH", 6),
		},
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// LoadForTest loads configuration from .env.test file for testing
func LoadForTest() (*Config, error) {
	// Load .env.test file if it exists
	if err := godotenv.Load(".env.test"); err != nil {
		log.Println("No .env.test file found, using environment variables")
	}

	config := &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "3000"),
			Host:         getEnv("HOST", "0.0.0.0"),
			ReadTimeout:  getDurationEnv("READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getDurationEnv("IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			Type: getEnv("DB_TYPE", "memory"),
		},
		App: AppConfig{
			BaseURL:         getRequiredEnv("BASE_URL"),
			Environment:     getEnv("ENVIRONMENT", "development"),
			LogLevel:        getEnv("LOG_LEVEL", "info"),
			ShortCodeLength: getIntEnv("SHORT_CODE_LENGTH", 6),
		},
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.App.BaseURL == "" {
		return fmt.Errorf("BASE_URL is required")
	}

	if c.Server.Port == "" {
		return fmt.Errorf("PORT is required")
	}

	if c.App.ShortCodeLength < 3 || c.App.ShortCodeLength > 20 {
		return fmt.Errorf("SHORT_CODE_LENGTH must be between 3 and 20")
	}

	return nil
}

// GetServerAddress returns the full server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}

// IsDevelopment returns true if the environment is development
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// IsProduction returns true if the environment is production
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

// IsTest returns true if the environment is test
func (c *Config) IsTest() bool {
	return c.App.Environment == "test"
}

// Helper functions for environment variable handling

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getRequiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		// In test environment, we want to return empty string and let validation handle it
		// This prevents log.Fatalf from terminating tests
		return ""
	}
	return value
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Invalid integer value for %s, using default: %d", key, defaultValue)
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
		log.Printf("Invalid duration value for %s, using default: %v", key, defaultValue)
	}
	return defaultValue
}

// GetBoolEnv gets a boolean environment variable
func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
		log.Printf("Invalid boolean value for %s, using default: %t", key, defaultValue)
	}
	return defaultValue
}
