// Package main is the entry point for the wallabag-rss-tool application that monitors RSS feeds and sends articles to Wallabag.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"wallabag-rss-tool/pkg/config"
	"wallabag-rss-tool/pkg/database"
	"wallabag-rss-tool/pkg/logging"
	"wallabag-rss-tool/pkg/rss"
	"wallabag-rss-tool/pkg/server"
	"wallabag-rss-tool/pkg/wallabag"
	"wallabag-rss-tool/pkg/worker"
)

func main() {
	config.LoadEnvFile()
	initializeLogging()

	appConfig := loadApplicationConfig()
	db := initializeDatabase(appConfig.DatabasePath)
	defer database.CloseDB(db)

	wallabagConfig := loadWallabagConfig(db)
	wallabagClient := createWallabagClient(wallabagConfig)

	runApplication(db, wallabagClient, appConfig.ServerPort)
}

// initializeLogging sets up structured logging based on LOG_LEVEL and LOG_FORMAT environment variables
func initializeLogging() {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "INFO" // Default to INFO level
	}
	
	logFormat := os.Getenv("LOG_FORMAT")
	useJSON := true // Default to JSON format
	if strings.EqualFold(logFormat, "text") {
		useJSON = false
	}
	
	logger := logging.NewLoggerWithLevel(logLevel, useJSON)
	logging.SetGlobalLogger(logger)
	logging.Info("Starting Wallabag RSS Tool", "log_level", logLevel, "log_format", func() string {
		if useJSON { return "json" } else { return "text" }
	}())
}

// loadApplicationConfig loads and validates application configuration
func loadApplicationConfig() *config.AppConfig {
	appConfig, err := config.LoadAppConfig()
	if err != nil {
		logging.Error("Failed to load application configuration", "error", err)
		os.Exit(1)
	}

	return appConfig
}

// initializeDatabase sets up the database connection
func initializeDatabase(databasePath string) *sql.DB {
	db, err := database.InitDBWithPath(databasePath)
	if err != nil {
		logging.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}

	return db
}

// loadWallabagConfig loads and validates Wallabag configuration
func loadWallabagConfig(db *sql.DB) *config.WallabagConfig {
	wallabagConfig, err := config.LoadWallabagConfig()
	if err != nil {
		logging.Error("Failed to load Wallabag configuration",
			"error", err,
			"message", "Please ensure all required environment variables are set")
		database.CloseDB(db)
		os.Exit(1)
	}

	logging.Info("Loaded Wallabag configuration", "base_url", wallabagConfig.BaseURL)

	return wallabagConfig
}

// createWallabagClient creates and authenticates Wallabag client
func createWallabagClient(wallabagConfig *config.WallabagConfig) *wallabag.Client {
	wallabagClient := wallabag.NewClient(
		wallabagConfig.BaseURL,
		wallabagConfig.ClientID,
		wallabagConfig.ClientSecret,
		wallabagConfig.Username,
		wallabagConfig.Password,
	)

	if err := wallabagClient.Authenticate(context.Background()); err != nil {
		logging.Warn("Initial Wallabag authentication failed",
			"error", err,
			"message", "Please check your environment variables")
	} else {
		logging.Info("Successfully authenticated with Wallabag")
	}

	return wallabagClient
}

// runApplication initializes and runs the main application components
func runApplication(db *sql.DB, wallabagClient *wallabag.Client, port string) {
	store := database.NewSQLStore(db)
	rssProcessor := rss.NewProcessor()

	worker := worker.NewWorker(store, rssProcessor, wallabagClient)
	worker.Start()
	defer worker.Stop()

	server := server.NewServer(store, wallabagClient, worker)
	logging.Info("Starting web server", "port", port)

	if err := server.Start(port); err != nil {
		logging.Error("Web server failed to start", "error", err, "port", port)
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		worker.Stop()
		os.Exit(1) //nolint:gocritic // Explicit cleanup before exit is required
	}
}
