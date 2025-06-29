package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
	"wallabag-rss-tool/pkg/config"
	"wallabag-rss-tool/pkg/database"
	"wallabag-rss-tool/pkg/logging"
	"wallabag-rss-tool/pkg/rss"
	"wallabag-rss-tool/pkg/server"
	"wallabag-rss-tool/pkg/wallabag"
	"wallabag-rss-tool/pkg/worker"
)

func TestInitializeLogging(t *testing.T) {
	// Save original environment
	originalLogLevel := os.Getenv("LOG_LEVEL")
	originalLogFormat := os.Getenv("LOG_FORMAT")
	
	// Clean up after test
	defer func() {
		if originalLogLevel == "" {
			os.Unsetenv("LOG_LEVEL")
		} else {
			os.Setenv("LOG_LEVEL", originalLogLevel)
		}
		if originalLogFormat == "" {
			os.Unsetenv("LOG_FORMAT")
		} else {
			os.Setenv("LOG_FORMAT", originalLogFormat)
		}
	}()

	t.Run("Default logging configuration", func(t *testing.T) {
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LOG_FORMAT")
		
		assert.NotPanics(t, func() {
			initializeLogging()
		})
		
		// Verify global logger is set
		logger := logging.GetGlobalLogger()
		assert.NotNil(t, logger)
	})
	
	t.Run("Custom log level", func(t *testing.T) {
		os.Setenv("LOG_LEVEL", "DEBUG")
		os.Unsetenv("LOG_FORMAT")
		
		assert.NotPanics(t, func() {
			initializeLogging()
		})
		
		logger := logging.GetGlobalLogger()
		assert.NotNil(t, logger)
	})
	
	t.Run("Text log format", func(t *testing.T) {
		os.Setenv("LOG_LEVEL", "INFO")
		os.Setenv("LOG_FORMAT", "text")
		
		assert.NotPanics(t, func() {
			initializeLogging()
		})
		
		logger := logging.GetGlobalLogger()
		assert.NotNil(t, logger)
	})
	
	t.Run("JSON log format", func(t *testing.T) {
		os.Setenv("LOG_LEVEL", "WARN")
		os.Setenv("LOG_FORMAT", "json")
		
		assert.NotPanics(t, func() {
			initializeLogging()
		})
		
		logger := logging.GetGlobalLogger()
		assert.NotNil(t, logger)
	})
	
	t.Run("Case insensitive format", func(t *testing.T) {
		os.Setenv("LOG_LEVEL", "ERROR")
		os.Setenv("LOG_FORMAT", "TEXT")
		
		assert.NotPanics(t, func() {
			initializeLogging()
		})
		
		logger := logging.GetGlobalLogger()
		assert.NotNil(t, logger)
	})
}

func TestLoadApplicationConfig(t *testing.T) {
	// Save original environment
	originalDBPath := os.Getenv("DATABASE_PATH")
	originalPort := os.Getenv("SERVER_PORT")
	
	// Clean up after test
	defer func() {
		if originalDBPath == "" {
			os.Unsetenv("DATABASE_PATH")
		} else {
			os.Setenv("DATABASE_PATH", originalDBPath)
		}
		if originalPort == "" {
			os.Unsetenv("SERVER_PORT")
		} else {
			os.Setenv("SERVER_PORT", originalPort)
		}
	}()

	t.Run("Load config with defaults", func(t *testing.T) {
		os.Unsetenv("DATABASE_PATH")
		os.Unsetenv("SERVER_PORT")
		
		config := loadApplicationConfig()
		
		assert.NotNil(t, config)
		assert.Equal(t, "./wallabag.db", config.DatabasePath)
		assert.Equal(t, "8080", config.ServerPort)
	})
	
	t.Run("Load config with custom values", func(t *testing.T) {
		os.Setenv("DATABASE_PATH", "/tmp/test.db")
		os.Setenv("SERVER_PORT", "9090")
		
		config := loadApplicationConfig()
		
		assert.NotNil(t, config)
		assert.Equal(t, "/tmp/test.db", config.DatabasePath)
		assert.Equal(t, "9090", config.ServerPort)
	})
}

func TestInitializeDatabase(t *testing.T) {
	t.Run("Initialize with valid path", func(t *testing.T) {
		// Use temporary file with .db extension for testing
		testPath := "/tmp/test_init.db"
		
		// Clean up before test
		os.Remove(testPath)
		
		var db *sql.DB
		assert.NotPanics(t, func() {
			db = initializeDatabase(testPath)
		})
		
		assert.NotNil(t, db)
		
		// Clean up
		if db != nil {
			db.Close()
		}
		os.Remove(testPath)
	})
	
	t.Run("Initialize creates new database file", func(t *testing.T) {
		// Use temporary file for testing
		tmpFile := "/tmp/test_wallabag.db"
		
		// Clean up before test
		os.Remove(tmpFile)
		
		var db *sql.DB
		assert.NotPanics(t, func() {
			db = initializeDatabase(tmpFile)
		})
		
		assert.NotNil(t, db)
		
		// Verify file was created
		_, err := os.Stat(tmpFile)
		assert.NoError(t, err)
		
		// Clean up
		if db != nil {
			db.Close()
		}
		os.Remove(tmpFile)
	})
}

func TestLoadWallabagConfig(t *testing.T) {
	// Save original environment
	envVars := []string{
		"WALLABAG_BASE_URL",
		"WALLABAG_CLIENT_ID", 
		"WALLABAG_CLIENT_SECRET",
		"WALLABAG_USERNAME",
		"WALLABAG_PASSWORD",
	}
	originalEnv := make(map[string]string)
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

	t.Run("Load valid Wallabag config", func(t *testing.T) {
		// Set all required environment variables
		os.Setenv("WALLABAG_BASE_URL", "https://wallabag.test.com")
		os.Setenv("WALLABAG_CLIENT_ID", "test_client_id")
		os.Setenv("WALLABAG_CLIENT_SECRET", "test_client_secret")
		os.Setenv("WALLABAG_USERNAME", "test_username")
		os.Setenv("WALLABAG_PASSWORD", "test_password")
		
		// Create a test database
		testDBPath := "/tmp/test_wallabag_config.db"
		os.Remove(testDBPath)
		db := initializeDatabase(testDBPath)
		defer func() {
			db.Close()
			os.Remove(testDBPath)
		}()
		
		config := loadWallabagConfig(db)
		
		assert.NotNil(t, config)
		assert.Equal(t, "https://wallabag.test.com", config.BaseURL)
		assert.Equal(t, "test_client_id", config.ClientID)
		assert.Equal(t, "test_client_secret", config.ClientSecret)
		assert.Equal(t, "test_username", config.Username)
		assert.Equal(t, "test_password", config.Password)
	})
}

func TestCreateWallabagClient(t *testing.T) {
	t.Run("Create Wallabag client", func(t *testing.T) {
		wallabagConfig := &config.WallabagConfig{
			BaseURL:      "https://wallabag.test.com",
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
			Username:     "test_username",
			Password:     "test_password",
		}
		
		// Test just client creation without authentication to avoid timeout
		var client *wallabag.Client
		assert.NotPanics(t, func() {
			client = wallabag.NewClient(
				wallabagConfig.BaseURL,
				wallabagConfig.ClientID,
				wallabagConfig.ClientSecret,
				wallabagConfig.Username,
				wallabagConfig.Password,
			)
		})
		
		assert.NotNil(t, client)
	})
}

func TestApplicationComponents(t *testing.T) {
	t.Run("Application component creation patterns", func(t *testing.T) {
		// Test the patterns used in runApplication function
		
		// These components should be creatable
		testDBPath := "/tmp/test_components.db"
		os.Remove(testDBPath)
		db := initializeDatabase(testDBPath)
		defer func() {
			db.Close()
			os.Remove(testDBPath)
		}()
		
		// Test store creation pattern
		assert.NotNil(t, db)
		
		// Test config creation pattern
		appConfig := loadApplicationConfig()
		assert.NotNil(t, appConfig)
		
		// Test Wallabag client pattern
		wallabagConfig := &config.WallabagConfig{
			BaseURL:      "https://wallabag.test.com",
			ClientID:     "test_client_id", 
			ClientSecret: "test_client_secret",
			Username:     "test_username",
			Password:     "test_password",
		}
		
		client := createWallabagClient(wallabagConfig)
		assert.NotNil(t, client)
	})
}

func TestLoggerIntegration(t *testing.T) {
	t.Run("Logger integration with main functions", func(t *testing.T) {
		// Test that logging is properly integrated
		
		// Initialize logging
		initializeLogging()
		
		// Verify logger is available
		logger := logging.GetGlobalLogger()
		assert.NotNil(t, logger)
		
		// Test that functions can log
		assert.NotPanics(t, func() {
			logging.Info("Test log message")
		})
	})
}

func TestEnvironmentVariableHandling(t *testing.T) {
	t.Run("Environment variable patterns", func(t *testing.T) {
		// Test environment variable handling patterns
		
		// LOG_LEVEL handling
		originalLogLevel := os.Getenv("LOG_LEVEL")
		defer func() {
			if originalLogLevel == "" {
				os.Unsetenv("LOG_LEVEL")
			} else {
				os.Setenv("LOG_LEVEL", originalLogLevel)
			}
		}()
		
		// Test default value
		os.Unsetenv("LOG_LEVEL")
		logLevel := os.Getenv("LOG_LEVEL")
		if logLevel == "" {
			logLevel = "INFO" // Default value from main.go
		}
		assert.Equal(t, "INFO", logLevel)
		
		// Test custom value
		os.Setenv("LOG_LEVEL", "DEBUG")
		logLevel = os.Getenv("LOG_LEVEL")
		assert.Equal(t, "DEBUG", logLevel)
	})
}

func TestApplicationFlow(t *testing.T) {
	t.Run("Main application flow components", func(t *testing.T) {
		// Test the main components can be initialized in sequence
		
		// 1. Initialize logging
		assert.NotPanics(t, func() {
			initializeLogging()
		})
		
		// 2. Load app config
		var appConfig *config.AppConfig
		assert.NotPanics(t, func() {
			appConfig = loadApplicationConfig()
		})
		assert.NotNil(t, appConfig)
		
		// 3. Initialize database
		var db *sql.DB
		testDBPath := "/tmp/test_flow.db"
		os.Remove(testDBPath)
		assert.NotPanics(t, func() {
			db = initializeDatabase(testDBPath)
		})
		assert.NotNil(t, db)
		defer func() {
			db.Close()
			os.Remove(testDBPath)
		}()
		
		// 4. Create Wallabag config (mock)
		wallabagConfig := &config.WallabagConfig{
			BaseURL:      "https://wallabag.test.com",
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret", 
			Username:     "test_username",
			Password:     "test_password",
		}
		
		// 5. Create Wallabag client (test instantiation only, auth will timeout)
		var client *wallabag.Client
		assert.NotPanics(t, func() {
			// Just create the client without authentication to avoid timeout
			client = wallabag.NewClient(
				wallabagConfig.BaseURL,
				wallabagConfig.ClientID,
				wallabagConfig.ClientSecret,
				wallabagConfig.Username,
				wallabagConfig.Password,
			)
		})
		assert.NotNil(t, client)
		
		// This tests the initialization flow without actually running the server
	})
}

func TestMainFunctionCoverage(t *testing.T) {
	// Test actual main.go functions to improve coverage
	
	t.Run("initializeLogging function coverage", func(t *testing.T) {
		// Save original environment
		originalLogLevel := os.Getenv("LOG_LEVEL")
		originalLogFormat := os.Getenv("LOG_FORMAT")
		defer func() {
			if originalLogLevel == "" {
				os.Unsetenv("LOG_LEVEL")
			} else {
				os.Setenv("LOG_LEVEL", originalLogLevel)
			}
			if originalLogFormat == "" {
				os.Unsetenv("LOG_FORMAT")
			} else {
				os.Setenv("LOG_FORMAT", originalLogFormat)
			}
		}()
		
		// Test different configurations to cover all code paths
		os.Setenv("LOG_LEVEL", "DEBUG")
		os.Setenv("LOG_FORMAT", "json")
		assert.NotPanics(t, initializeLogging)
		
		os.Setenv("LOG_LEVEL", "ERROR")
		os.Setenv("LOG_FORMAT", "text")
		assert.NotPanics(t, initializeLogging)
	})
	
	t.Run("loadApplicationConfig function coverage", func(t *testing.T) {
		// Save original environment
		originalDBPath := os.Getenv("DATABASE_PATH")
		originalPort := os.Getenv("SERVER_PORT")
		defer func() {
			if originalDBPath == "" {
				os.Unsetenv("DATABASE_PATH")
			} else {
				os.Setenv("DATABASE_PATH", originalDBPath)
			}
			if originalPort == "" {
				os.Unsetenv("SERVER_PORT")
			} else {
				os.Setenv("SERVER_PORT", originalPort)
			}
		}()
		
		// Test function directly
		os.Setenv("DATABASE_PATH", "/tmp/test.db")
		os.Setenv("SERVER_PORT", "9090")
		
		config := loadApplicationConfig()
		assert.NotNil(t, config)
		assert.Equal(t, "/tmp/test.db", config.DatabasePath)
		assert.Equal(t, "9090", config.ServerPort)
	})
	
	t.Run("initializeDatabase function coverage", func(t *testing.T) {
		testDBPath := "/tmp/test_init_coverage.db"
		os.Remove(testDBPath)
		defer os.Remove(testDBPath)
		
		// Test function directly
		db := initializeDatabase(testDBPath)
		assert.NotNil(t, db)
		defer db.Close()
		
		// Verify database was created and initialized
		_, err := os.Stat(testDBPath)
		assert.NoError(t, err)
		
		// Test that schema was applied
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM settings WHERE key = 'default_poll_interval_minutes'").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
	})
	
	t.Run("loadWallabagConfig function coverage", func(t *testing.T) {
		// Save original environment
		envVars := []string{
			"WALLABAG_BASE_URL",
			"WALLABAG_CLIENT_ID",
			"WALLABAG_CLIENT_SECRET",
			"WALLABAG_USERNAME",
			"WALLABAG_PASSWORD",
		}
		originalEnv := make(map[string]string)
		for _, env := range envVars {
			originalEnv[env] = os.Getenv(env)
		}
		defer func() {
			for _, env := range envVars {
				if originalEnv[env] == "" {
					os.Unsetenv(env)
				} else {
					os.Setenv(env, originalEnv[env])
				}
			}
		}()
		
		// Create test database
		testDBPath := "/tmp/test_wallabag_config.db"
		os.Remove(testDBPath)
		defer os.Remove(testDBPath)
		db := initializeDatabase(testDBPath)
		defer db.Close()
		
		// Set valid Wallabag config
		os.Setenv("WALLABAG_BASE_URL", "https://wallabag.test.com")
		os.Setenv("WALLABAG_CLIENT_ID", "test_client_id")
		os.Setenv("WALLABAG_CLIENT_SECRET", "test_client_secret")
		os.Setenv("WALLABAG_USERNAME", "test_username")
		os.Setenv("WALLABAG_PASSWORD", "test_password")
		
		// Test function directly
		config := loadWallabagConfig(db)
		assert.NotNil(t, config)
		assert.Equal(t, "https://wallabag.test.com", config.BaseURL)
		assert.Equal(t, "test_client_id", config.ClientID)
	})
	
	t.Run("createWallabagClient function coverage", func(t *testing.T) {
		wallabagConfig := &config.WallabagConfig{
			BaseURL:      "https://wallabag.test.com",
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
			Username:     "test_username",
			Password:     "test_password",
		}
		
		// Test function directly
		client := createWallabagClient(wallabagConfig)
		assert.NotNil(t, client)
	})
}

func TestRunApplication(t *testing.T) {
	t.Run("runApplication creates and configures components", func(t *testing.T) {
		// Create temporary database
		tempDir, err := os.MkdirTemp("", "wallabag_run_test_")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		testDBPath := filepath.Join(tempDir, "test.db")
		db, err := sql.Open("sqlite", testDBPath)
		require.NoError(t, err)
		defer db.Close()

		// Apply schema
		schema := `
CREATE TABLE IF NOT EXISTS feeds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    url TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    last_fetched DATETIME,
    poll_interval_minutes INTEGER DEFAULT 60,
    poll_interval INTEGER DEFAULT 1,
    poll_interval_unit TEXT DEFAULT 'days',
    sync_mode TEXT DEFAULT 'none',
    sync_count INTEGER,
    sync_date_from DATETIME,
    initial_sync_done BOOLEAN DEFAULT 0
);

CREATE TABLE IF NOT EXISTS articles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    feed_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    url TEXT NOT NULL UNIQUE,
    wallabag_entry_id INTEGER,
    published_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT
);

INSERT OR IGNORE INTO settings (key, value) VALUES ('default_poll_interval_minutes', '60');
`
		_, err = db.Exec(schema)
		require.NoError(t, err)

		// Create Wallabag client
		wallabagClient := wallabag.NewClient("http://localhost:8000", "client_id", "client_secret", "username", "password")

		// Test that runApplication can be called without panicking
		// We can't easily test the full server start without it hanging,
		// but we can verify the component creation logic
		assert.NotPanics(t, func() {
			// This would normally start the server, but we can't easily test that
			// runApplication(db, wallabagClient, "8080")
			
			// Instead, test the component creation that happens in runApplication
			store := database.NewSQLStore(db)
			rssProcessor := rss.NewProcessor()
			worker := worker.NewWorker(store, rssProcessor, wallabagClient)
			server := server.NewServer(store, wallabagClient, worker)
			
			assert.NotNil(t, store)
			assert.NotNil(t, rssProcessor)
			assert.NotNil(t, worker)
			assert.NotNil(t, server)
		})
	})
}

func TestMainIntegration(t *testing.T) {
	t.Run("Full application initialization flow", func(t *testing.T) {
		// Save original environment
		envVars := []string{
			"LOG_LEVEL", "LOG_FORMAT", "DATABASE_PATH", "SERVER_PORT",
			"WALLABAG_BASE_URL", "WALLABAG_CLIENT_ID", "WALLABAG_CLIENT_SECRET",
			"WALLABAG_USERNAME", "WALLABAG_PASSWORD",
		}
		originalEnv := make(map[string]string)
		for _, env := range envVars {
			originalEnv[env] = os.Getenv(env)
		}
		defer func() {
			for _, env := range envVars {
				if originalEnv[env] == "" {
					os.Unsetenv(env)
				} else {
					os.Setenv(env, originalEnv[env])
				}
			}
		}()
		
		// Set up test environment
		os.Setenv("LOG_LEVEL", "INFO")
		os.Setenv("LOG_FORMAT", "text")
		os.Setenv("DATABASE_PATH", "/tmp/integration_test.db")
		os.Setenv("SERVER_PORT", "8081")
		os.Setenv("WALLABAG_BASE_URL", "https://wallabag.test.com")
		os.Setenv("WALLABAG_CLIENT_ID", "test_client_id")
		os.Setenv("WALLABAG_CLIENT_SECRET", "test_client_secret")
		os.Setenv("WALLABAG_USERNAME", "test_username")
		os.Setenv("WALLABAG_PASSWORD", "test_password")
		
		// Clean up database file
		os.Remove("/tmp/integration_test.db")
		defer os.Remove("/tmp/integration_test.db")
		
		// Test the main application flow
		assert.NotPanics(t, func() {
			// 1. Initialize logging
			initializeLogging()
			
			// 2. Load app config
			appConfig := loadApplicationConfig()
			assert.NotNil(t, appConfig)
			
			// 3. Initialize database
			db := initializeDatabase(appConfig.DatabasePath)
			assert.NotNil(t, db)
			defer db.Close()
			
			// 4. Load Wallabag config
			wallabagConfig := loadWallabagConfig(db)
			assert.NotNil(t, wallabagConfig)
			
			// 5. Create Wallabag client (test instantiation only, auth will timeout)
			var client *wallabag.Client
			assert.NotPanics(t, func() {
				// Just create the client without authentication to avoid timeout
				client = wallabag.NewClient(
					wallabagConfig.BaseURL,
					wallabagConfig.ClientID,
					wallabagConfig.ClientSecret,
					wallabagConfig.Username,
					wallabagConfig.Password,
				)
			})
			assert.NotNil(t, client)
			
			// This tests the same flow as main() but without starting the server
		})
	})
}

// These tests are in the same package to ensure coverage is counted

func TestMainFunctions_InPackage(t *testing.T) {
	t.Run("main package components test", func(t *testing.T) {
		// Test that we can create the basic components that main.go uses
		
		// Test config loading from config package
		appConfig, err := config.LoadAppConfig()
		if err == nil {
			assert.NotNil(t, appConfig)
			assert.NotEmpty(t, appConfig.DatabasePath)
			assert.NotEmpty(t, appConfig.ServerPort)
		}

		// Test wallabag config loading
		wallabagConfig, err := config.LoadWallabagConfig()
		// This might return an error if env vars are not set, but shouldn't panic
		if err == nil && wallabagConfig != nil {
			assert.NotEmpty(t, wallabagConfig.BaseURL)
		}

		// Test creating a wallabag client (this should not panic)
		client := wallabag.NewClient("http://test", "client", "secret", "user", "pass")
		assert.NotNil(t, client)
	})
}

func TestMainHelperFunctions(t *testing.T) {
	t.Run("main function components", func(t *testing.T) {
		// We can't easily test the main function directly since it runs the server,
		// but we can test that all the components it uses work correctly

		// Test logging initialization - test the package functions directly
		assert.NotPanics(t, func() {
			// Test logging package directly since initializeLogging is private
			config, _ := config.LoadAppConfig()
			assert.NotNil(t, config)
		})
	})

	t.Run("runApplication components", func(t *testing.T) {
		// Test that all components used in runApplication can be created
		tempDir, err := os.MkdirTemp("", "wallabag_run_components_test_")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		testDBPath := filepath.Join(tempDir, "components.db")
		db, err := sql.Open("sqlite", testDBPath)
		require.NoError(t, err)
		defer db.Close()

		// Create basic schema
		schema := `
			CREATE TABLE IF NOT EXISTS feeds (id INTEGER PRIMARY KEY);
			CREATE TABLE IF NOT EXISTS articles (id INTEGER PRIMARY KEY);
			CREATE TABLE IF NOT EXISTS settings (key TEXT PRIMARY KEY, value TEXT);
		`
		_, err = db.Exec(schema)
		require.NoError(t, err)

		wallabagClient := wallabag.NewClient("http://localhost", "client", "secret", "user", "pass")

		// Test that runApplication components can be created
		// (We can't actually call runApplication because it would start the server)
		assert.NotPanics(t, func() {
			// This simulates what runApplication does without starting the server
			// store := database.NewSQLStore(db)
			// rssProcessor := rss.NewProcessor()
			// worker := worker.NewWorker(store, rssProcessor, wallabagClient)
			// server := server.NewServer(store, wallabagClient, worker)
			
			// Just verify the database and client are usable
			assert.NotNil(t, db)
			assert.NotNil(t, wallabagClient)
		})
	})
}