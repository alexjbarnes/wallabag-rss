// Package database provides SQLite database initialization, schema management and path validation.
package database

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
	"wallabag-rss-tool/pkg/logging"
)

const schemaPath = "./db/schema.sql"

// InitDB initializes the SQLite database and applies migrations.
func InitDB() (*sql.DB, error) {
	return InitDBWithPath("./wallabag.db")
}

// InitDBWithPath initializes the SQLite database with a custom path and applies migrations.
func InitDBWithPath(dbPath string) (*sql.DB, error) {
	// Validate and sanitize database path
	if err := validateDatabasePath(dbPath); err != nil {
		return nil, fmt.Errorf("invalid database path: %w", err)
	}

	// Check if the database file exists, if not, create it.
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		logging.Info("Database file not found, creating new database", "db_path", dbPath)
		file, err := os.Create(dbPath)
		if err != nil {
			return nil, fmt.Errorf("os.Create failed for database file: %w", err)
		}
		if err := file.Close(); err != nil {
			return nil, fmt.Errorf("failed to close database file: %w", err)
		}
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("sql.Open failed for database: %w", err)
	}

	if err = applySchema(db); err != nil {
		return nil, fmt.Errorf("applySchema failed: %w", err)
	}

	logging.Info("Database initialized successfully", "db_path", dbPath)

	return db, nil
}

// applySchema reads the schema.sql file and executes its contents.
func applySchema(db *sql.DB) error {
	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("os.ReadFile failed for schema file %s: %w", schemaPath, err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		return fmt.Errorf("db.Exec failed for schema: %w", err)
	}

	return nil
}

// CloseDB closes the database connection.
func CloseDB(db *sql.DB) {
	if db != nil {
		if err := db.Close(); err != nil {
			logging.Error("Failed to close database connection", "error", err)
		} else {
			logging.Info("Database connection closed")
		}
	}
}

// validateDatabasePath validates and sanitizes the database file path
func validateDatabasePath(dbPath string) error {
	if dbPath == "" {
		return errors.New("database path cannot be empty")
	}

	// Clean the path to resolve any ".." or other potentially dangerous elements
	cleanPath := filepath.Clean(dbPath)

	// Prevent absolute paths to system directories
	if filepath.IsAbs(cleanPath) {
		// Allow only specific absolute paths (e.g., in user's home or /tmp)
		allowedPrefixes := []string{"/tmp/", "/var/tmp/"}
		allowed := false
		for _, prefix := range allowedPrefixes {
			if strings.HasPrefix(cleanPath, prefix) {
				allowed = true
				break
			}
		}
		if !allowed {
			return errors.New("absolute database paths not allowed except in /tmp")
		}
	}

	// Prevent path traversal attacks
	if strings.Contains(cleanPath, "..") {
		return errors.New("path traversal not allowed")
	}

	// Ensure it's a .db file
	if !strings.HasSuffix(cleanPath, ".db") {
		return errors.New("database file must have .db extension")
	}

	// Ensure the directory exists or can be created
	dir := filepath.Dir(cleanPath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("cannot create database directory: %w", err)
		}
	}

	return nil
}
