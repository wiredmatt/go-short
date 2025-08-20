package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoad_DefaultValues(t *testing.T) {
	// Clear environment variables to test defaults
	os.Unsetenv("BASE_URL")
	os.Unsetenv("PORT")
	os.Unsetenv("HOST")
	os.Unsetenv("ENVIRONMENT")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("SHORT_CODE_LENGTH")
	os.Unsetenv("READ_TIMEOUT")
	os.Unsetenv("WRITE_TIMEOUT")
	os.Unsetenv("IDLE_TIMEOUT")
	os.Unsetenv("DB_TYPE")
	os.Unsetenv("DB_CONNECTION_STRING")

	// Set required environment variable
	os.Setenv("BASE_URL", "https://short.url")

	cfg, err := LoadForTest()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Test default values
	assert.Equal(t, "3000", cfg.Server.Port)
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 30*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 30*time.Second, cfg.Server.WriteTimeout)
	assert.Equal(t, 60*time.Second, cfg.Server.IdleTimeout)
	assert.Equal(t, "memory", cfg.Database.Type)
	assert.Equal(t, "https://short.url", cfg.App.BaseURL)
	assert.Equal(t, "development", cfg.App.Environment)
	assert.Equal(t, "info", cfg.App.LogLevel)
	assert.Equal(t, 6, cfg.App.ShortCodeLength)
}

func TestLoad_CustomValues(t *testing.T) {
	// Set custom environment variables
	os.Setenv("BASE_URL", "https://custom.url")
	os.Setenv("PORT", "8080")
	os.Setenv("HOST", "localhost")
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("SHORT_CODE_LENGTH", "8")
	os.Setenv("READ_TIMEOUT", "60s")
	os.Setenv("WRITE_TIMEOUT", "60s")
	os.Setenv("IDLE_TIMEOUT", "120s")
	os.Setenv("DB_TYPE", "postgres")
	os.Setenv("DB_CONNECTION_STRING", "postgres://user:password@localhost:5432/shorten")

	cfg, err := LoadForTest()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Test custom values
	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, "localhost", cfg.Server.Host)
	assert.Equal(t, 60*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 60*time.Second, cfg.Server.WriteTimeout)
	assert.Equal(t, 120*time.Second, cfg.Server.IdleTimeout)
	assert.Equal(t, "postgres", cfg.Database.Type)
	assert.Equal(t, "postgres://user:password@localhost:5432/shorten", cfg.Database.ConnectionString)
	assert.Equal(t, "https://custom.url", cfg.App.BaseURL)
	assert.Equal(t, "production", cfg.App.Environment)
	assert.Equal(t, "debug", cfg.App.LogLevel)
	assert.Equal(t, 8, cfg.App.ShortCodeLength)
}

func TestLoad_InvalidInteger(t *testing.T) {
	// Set invalid integer value
	os.Setenv("BASE_URL", "https://short.url")
	os.Setenv("SHORT_CODE_LENGTH", "invalid")

	cfg, err := LoadForTest()

	assert.NoError(t, err)                      // Should use default value
	assert.Equal(t, 6, cfg.App.ShortCodeLength) // Default value
}

func TestLoad_InvalidDuration(t *testing.T) {
	// Set invalid duration value
	os.Setenv("BASE_URL", "https://short.url")
	os.Setenv("READ_TIMEOUT", "invalid")

	cfg, err := LoadForTest()

	assert.NoError(t, err)                                  // Should use default value
	assert.Equal(t, 30*time.Second, cfg.Server.ReadTimeout) // Default value
}

func TestValidate_Success(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port: "3000",
		},
		App: AppConfig{
			BaseURL:         "https://short.url",
			ShortCodeLength: 6,
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestValidate_MissingBaseURL(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port: "3000",
		},
		App: AppConfig{
			BaseURL:         "",
			ShortCodeLength: 6,
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "BASE_URL is required")
}

func TestValidate_MissingPort(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port: "",
		},
		App: AppConfig{
			BaseURL:         "https://short.url",
			ShortCodeLength: 6,
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PORT is required")
}

func TestValidate_InvalidShortCodeLength(t *testing.T) {
	tests := []struct {
		name     string
		length   int
		expected bool
	}{
		{"too short", 2, false},
		{"minimum valid", 3, true},
		{"valid", 6, true},
		{"maximum valid", 20, true},
		{"too long", 21, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Server: ServerConfig{
					Port: "3000",
				},
				App: AppConfig{
					BaseURL:         "https://short.url",
					ShortCodeLength: tt.length,
				},
			}

			err := cfg.Validate()
			if tt.expected {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "SHORT_CODE_LENGTH must be between 3 and 20")
			}
		})
	}
}

func TestGetServerAddress(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: "8080",
		},
	}

	address := cfg.GetServerAddress()
	assert.Equal(t, "localhost:8080", address)
}

func TestEnvironmentHelpers(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		isDev       bool
		isProd      bool
		isTest      bool
	}{
		{"development", "development", true, false, false},
		{"production", "production", false, true, false},
		{"test", "test", false, false, true},
		{"staging", "staging", false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				App: AppConfig{
					Environment: tt.environment,
				},
			}

			assert.Equal(t, tt.isDev, cfg.IsDevelopment())
			assert.Equal(t, tt.isProd, cfg.IsProduction())
			assert.Equal(t, tt.isTest, cfg.IsTest())
		})
	}
}
