package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	App      AppConfig
}

type ServerConfig struct {
	Port         string
	Host         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	Type             string // "memory", "postgres", "redis"
	ConnectionString string
}

type AppConfig struct {
	BaseURL         string
	Environment     string
	LogLevel        string
	ShortCodeLength int
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
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
			Type:             getEnv("DB_TYPE", "postgres"),
			ConnectionString: getEnv("DB_CONNECTION_STRING", "postgres://user:password@localhost:5432/shortener"),
		},
		App: AppConfig{
			BaseURL:         getEnv("BASE_URL", fmt.Sprintf("http://%s:%s", getEnv("HOST", "0.0.0.0"), getEnv("PORT", "3000"))),
			Environment:     getEnv("ENVIRONMENT", "development"),
			LogLevel:        getEnv("LOG_LEVEL", "info"),
			ShortCodeLength: getIntEnv("SHORT_CODE_LENGTH", 6),
		},
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// LoadForTest loads configuration from .env.test file for testing
func LoadForTest() (*Config, error) {
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
			Type:             getEnv("DB_TYPE", "memory"),
			ConnectionString: getEnv("DB_CONNECTION_STRING", "postgres://user:password@localhost:5432/shortener"),
		},
		App: AppConfig{
			BaseURL:         getEnv("BASE_URL", "http://localhost:3000"),
			Environment:     getEnv("ENVIRONMENT", "development"),
			LogLevel:        getEnv("LOG_LEVEL", "info"),
			ShortCodeLength: getIntEnv("SHORT_CODE_LENGTH", 6),
		},
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

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

func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}

func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

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

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
		log.Printf("Invalid boolean value for %s, using default: %t", key, defaultValue)
	}
	return defaultValue
}
