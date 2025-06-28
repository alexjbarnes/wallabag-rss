package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"wallabag-rss-tool/pkg/config"
)

func TestWallabagConfigStruct(t *testing.T) {
	tests := []struct {
		name   string
		config config.WallabagConfig
	}{
		{
			name: "Complete config",
			config: config.WallabagConfig{
				BaseURL:      "https://wallabag.example.com",
				ClientID:     "client123",
				ClientSecret: "secret456",
				Username:     "testuser",
				Password:     "testpass",
			},
		},
		{
			name:   "Zero values",
			config: config.WallabagConfig{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.config

			assert.Equal(t, tt.config.BaseURL, config.BaseURL)
			assert.Equal(t, tt.config.ClientID, config.ClientID)
			assert.Equal(t, tt.config.ClientSecret, config.ClientSecret)
			assert.Equal(t, tt.config.Username, config.Username)
			assert.Equal(t, tt.config.Password, config.Password)
		})
	}
}

func TestLoadWallabagConfig(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"WALLABAG_BASE_URL",
		"WALLABAG_CLIENT_ID",
		"WALLABAG_CLIENT_SECRET",
		"WALLABAG_USERNAME",
		"WALLABAG_PASSWORD",
	}

	for _, env := range envVars {
		originalEnv[env] = os.Getenv(env)
	}

	// Clean up after test
	defer func() {
		for _, env := range envVars {
			if originalEnv[env] == "" {
				os.Unsetenv(env)
			} else {
				os.Setenv(env, originalEnv[env])
			}
		}
	}()

	t.Run("All environment variables set", func(t *testing.T) {
		// Set all required environment variables
		os.Setenv("WALLABAG_BASE_URL", "https://wallabag.test.com")
		os.Setenv("WALLABAG_CLIENT_ID", "test_client_id")
		os.Setenv("WALLABAG_CLIENT_SECRET", "test_client_secret")
		os.Setenv("WALLABAG_USERNAME", "test_username")
		os.Setenv("WALLABAG_PASSWORD", "test_password")

		config, err := config.LoadWallabagConfig()

		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "https://wallabag.test.com", config.BaseURL)
		assert.Equal(t, "test_client_id", config.ClientID)
		assert.Equal(t, "test_client_secret", config.ClientSecret)
		assert.Equal(t, "test_username", config.Username)
		assert.Equal(t, "test_password", config.Password)
	})

	t.Run("Missing WALLABAG_BASE_URL", func(t *testing.T) {
		// Unset one required variable
		os.Unsetenv("WALLABAG_BASE_URL")
		os.Setenv("WALLABAG_CLIENT_ID", "test_client_id")
		os.Setenv("WALLABAG_CLIENT_SECRET", "test_client_secret")
		os.Setenv("WALLABAG_USERNAME", "test_username")
		os.Setenv("WALLABAG_PASSWORD", "test_password")

		config, err := config.LoadWallabagConfig()

		assert.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "WALLABAG_BASE_URL")
	})

	t.Run("Missing WALLABAG_CLIENT_ID", func(t *testing.T) {
		os.Setenv("WALLABAG_BASE_URL", "https://wallabag.test.com")
		os.Unsetenv("WALLABAG_CLIENT_ID")
		os.Setenv("WALLABAG_CLIENT_SECRET", "test_client_secret")
		os.Setenv("WALLABAG_USERNAME", "test_username")
		os.Setenv("WALLABAG_PASSWORD", "test_password")

		config, err := config.LoadWallabagConfig()

		assert.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "WALLABAG_CLIENT_ID")
	})

	t.Run("Missing WALLABAG_CLIENT_SECRET", func(t *testing.T) {
		os.Setenv("WALLABAG_BASE_URL", "https://wallabag.test.com")
		os.Setenv("WALLABAG_CLIENT_ID", "test_client_id")
		os.Unsetenv("WALLABAG_CLIENT_SECRET")
		os.Setenv("WALLABAG_USERNAME", "test_username")
		os.Setenv("WALLABAG_PASSWORD", "test_password")

		config, err := config.LoadWallabagConfig()

		assert.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "WALLABAG_CLIENT_SECRET")
	})

	t.Run("Missing WALLABAG_USERNAME", func(t *testing.T) {
		os.Setenv("WALLABAG_BASE_URL", "https://wallabag.test.com")
		os.Setenv("WALLABAG_CLIENT_ID", "test_client_id")
		os.Setenv("WALLABAG_CLIENT_SECRET", "test_client_secret")
		os.Unsetenv("WALLABAG_USERNAME")
		os.Setenv("WALLABAG_PASSWORD", "test_password")

		config, err := config.LoadWallabagConfig()

		assert.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "WALLABAG_USERNAME")
	})

	t.Run("Missing WALLABAG_PASSWORD", func(t *testing.T) {
		os.Setenv("WALLABAG_BASE_URL", "https://wallabag.test.com")
		os.Setenv("WALLABAG_CLIENT_ID", "test_client_id")
		os.Setenv("WALLABAG_CLIENT_SECRET", "test_client_secret")
		os.Setenv("WALLABAG_USERNAME", "test_username")
		os.Unsetenv("WALLABAG_PASSWORD")

		config, err := config.LoadWallabagConfig()

		assert.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "WALLABAG_PASSWORD")
	})

	t.Run("All environment variables empty", func(t *testing.T) {
		// Unset all variables
		for _, env := range envVars {
			os.Unsetenv(env)
		}

		config, err := config.LoadWallabagConfig()

		assert.Error(t, err)
		assert.Nil(t, config)
		// With caarlos0/env, we just check that it's an error about required fields
		assert.Contains(t, err.Error(), "required")
	})

	t.Run("Empty string values", func(t *testing.T) {
		// Set variables to empty strings
		os.Setenv("WALLABAG_BASE_URL", "")
		os.Setenv("WALLABAG_CLIENT_ID", "")
		os.Setenv("WALLABAG_CLIENT_SECRET", "")
		os.Setenv("WALLABAG_USERNAME", "")
		os.Setenv("WALLABAG_PASSWORD", "")

		config, err := config.LoadWallabagConfig()

		// caarlos0/env considers empty strings as valid values for required fields
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "", config.BaseURL)
		assert.Equal(t, "", config.ClientID)
		assert.Equal(t, "", config.ClientSecret)
		assert.Equal(t, "", config.Username)
		assert.Equal(t, "", config.Password)
	})

	t.Run("Mixed empty and set values", func(t *testing.T) {
		os.Setenv("WALLABAG_BASE_URL", "https://wallabag.test.com")
		os.Setenv("WALLABAG_CLIENT_ID", "")
		os.Setenv("WALLABAG_CLIENT_SECRET", "test_secret")
		os.Setenv("WALLABAG_USERNAME", "")
		os.Setenv("WALLABAG_PASSWORD", "test_pass")

		config, err := config.LoadWallabagConfig()

		// caarlos0/env considers empty strings as valid
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "https://wallabag.test.com", config.BaseURL)
		assert.Equal(t, "", config.ClientID)
		assert.Equal(t, "test_secret", config.ClientSecret)
		assert.Equal(t, "", config.Username)
		assert.Equal(t, "test_pass", config.Password)
	})

	t.Run("Whitespace only values", func(t *testing.T) {
		os.Setenv("WALLABAG_BASE_URL", "   ")
		os.Setenv("WALLABAG_CLIENT_ID", "\t")
		os.Setenv("WALLABAG_CLIENT_SECRET", "\n")
		os.Setenv("WALLABAG_USERNAME", " \t\n ")
		os.Setenv("WALLABAG_PASSWORD", "  ")

		config, err := config.LoadWallabagConfig()

		// The current implementation doesn't trim whitespace,
		// so whitespace-only values are considered valid
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "   ", config.BaseURL)
		assert.Equal(t, "\t", config.ClientID)
		assert.Equal(t, "\n", config.ClientSecret)
		assert.Equal(t, " \t\n ", config.Username)
		assert.Equal(t, "  ", config.Password)
	})
}

func TestLoadEnvFile(t *testing.T) {
	t.Run("LoadEnvFile does not panic", func(t *testing.T) {
		// LoadEnvFile should not panic even if .env file doesn't exist
		assert.NotPanics(t, func() {
			config.LoadEnvFile()
		})
	})
}
