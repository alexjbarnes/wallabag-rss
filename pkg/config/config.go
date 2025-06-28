// Package config handles loading and managing application configuration from environment variables.
package config

import (
	env "github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"wallabag-rss-tool/pkg/logging"
)

// WallabagConfig holds Wallabag API configuration.
//
//nolint:tagliatelle // Environment variable names use standard convention
type WallabagConfig struct {
	BaseURL      string `env:"WALLABAG_BASE_URL,required"`
	ClientID     string `env:"WALLABAG_CLIENT_ID,required"`
	ClientSecret string `env:"WALLABAG_CLIENT_SECRET,required"`
	Username     string `env:"WALLABAG_USERNAME,required"`
	Password     string `env:"WALLABAG_PASSWORD,required"`
}

// AppConfig holds application configuration.
//
//nolint:tagliatelle // Environment variable names use standard convention
type AppConfig struct {
	DatabasePath string `env:"DATABASE_PATH" envDefault:"./wallabag.db"`
	ServerPort   string `env:"SERVER_PORT"   envDefault:"8080"`
}

// LoadEnvFile loads environment variables from .env file if it exists.
// This should be called at application startup before loading config.
func LoadEnvFile() {
	if err := godotenv.Load(); err != nil {
		logging.Debug("No .env file found or error loading .env file", "error", err)
		logging.Info("Using system environment variables only")
	} else {
		logging.Info("Loaded environment variables from .env file")
	}
}

// LoadWallabagConfig loads Wallabag configuration from environment variables.
// All variables are required and the function will return an error if any are missing.
func LoadWallabagConfig() (*WallabagConfig, error) {
	var cfg WallabagConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// LoadAppConfig loads application configuration from environment variables.
func LoadAppConfig() (*AppConfig, error) {
	var cfg AppConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
