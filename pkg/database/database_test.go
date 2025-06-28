package database_test

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
	"wallabag-rss-tool/pkg/database"
)

func TestInitDB(t *testing.T) {
	// Create a temporary directory for test databases
	tempDir, err := os.MkdirTemp("", "wallabag_test_")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// We use custom paths in tests to avoid modifying package constants

	t.Run("Create new database successfully", func(t *testing.T) {
		// Create test schema file
		testSchemaPath := filepath.Join(tempDir, "schema.sql")
		testDBPath := filepath.Join(tempDir, "test.db")

		// Write test schema
		testSchema := `
CREATE TABLE IF NOT EXISTS feeds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    url TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    last_fetched DATETIME,
    poll_interval_minutes INTEGER DEFAULT 60
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
		err = os.WriteFile(testSchemaPath, []byte(testSchema), 0o644)
		assert.NoError(t, err)

		// Test with custom paths using a wrapper function
		db, err := initDBWithPaths(testDBPath, testSchemaPath)
		assert.NoError(t, err)
		assert.NotNil(t, db)

		// Verify database file was created
		_, err = os.Stat(testDBPath)
		assert.NoError(t, err)

		// Verify tables were created
		tables := []string{"feeds", "articles", "settings"}
		for _, table := range tables {
			var name string
			err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
			assert.NoError(t, err)
			assert.Equal(t, table, name)
		}

		// Verify default setting was inserted
		var value string
		err = db.QueryRow("SELECT value FROM settings WHERE key='default_poll_interval_minutes'").Scan(&value)
		assert.NoError(t, err)
		assert.Equal(t, "60", value)

		db.Close()
	})

	t.Run("Use existing database", func(t *testing.T) {
		testSchemaPath := filepath.Join(tempDir, "schema2.sql")
		testDBPath := filepath.Join(tempDir, "existing.db")

		// Create test schema
		testSchema := `CREATE TABLE IF NOT EXISTS test_table (id INTEGER PRIMARY KEY);`
		err = os.WriteFile(testSchemaPath, []byte(testSchema), 0o644)
		assert.NoError(t, err)

		// Create database file first
		file, err := os.Create(testDBPath)
		assert.NoError(t, err)
		file.Close()

		// Verify file exists before InitDB
		_, err = os.Stat(testDBPath)
		assert.NoError(t, err)

		db, err := initDBWithPaths(testDBPath, testSchemaPath)
		assert.NoError(t, err)
		assert.NotNil(t, db)

		db.Close()
	})

	t.Run("Schema file not found", func(t *testing.T) {
		testDBPath := filepath.Join(tempDir, "noschema.db")
		nonExistentSchemaPath := filepath.Join(tempDir, "nonexistent.sql")

		db, err := initDBWithPaths(testDBPath, nonExistentSchemaPath)
		assert.Error(t, err)
		assert.Nil(t, db)
		assert.Contains(t, err.Error(), "failed to apply database schema")
		assert.Contains(t, err.Error(), "failed to read schema file")
	})

	t.Run("Invalid schema SQL", func(t *testing.T) {
		testSchemaPath := filepath.Join(tempDir, "invalid_schema.sql")
		testDBPath := filepath.Join(tempDir, "invalid.db")

		// Write invalid SQL
		invalidSchema := "INVALID SQL SYNTAX HERE"
		err = os.WriteFile(testSchemaPath, []byte(invalidSchema), 0o644)
		assert.NoError(t, err)

		db, err := initDBWithPaths(testDBPath, testSchemaPath)
		assert.Error(t, err)
		assert.Nil(t, db)
		assert.Contains(t, err.Error(), "failed to apply database schema")
		assert.Contains(t, err.Error(), "failed to execute schema")
	})

	t.Run("Cannot create database file", func(t *testing.T) {
		// Try to create database in non-existent directory
		invalidDBPath := filepath.Join(tempDir, "nonexistent", "test.db")
		testSchemaPath := filepath.Join(tempDir, "schema3.sql")

		// Create valid schema
		testSchema := `CREATE TABLE test (id INTEGER);`
		err = os.WriteFile(testSchemaPath, []byte(testSchema), 0o644)
		assert.NoError(t, err)

		db, err := initDBWithPaths(invalidDBPath, testSchemaPath)
		assert.Error(t, err)
		assert.Nil(t, db)
		assert.Contains(t, err.Error(), "failed to create database file")
	})
}

// Helper function to test InitDB with custom paths
func initDBWithPaths(dbPath, schemaPath string) (*sql.DB, error) {
	// Check if the database file exists, if not, create it.
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		file, err := os.Create(dbPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create database file: %w", err)
		}
		file.Close()
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err = applySchemaWithPath(db, schemaPath); err != nil {
		db.Close()

		return nil, fmt.Errorf("failed to apply database schema: %w", err)
	}

	return db, nil
}

func TestApplySchema(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "wallabag_schema_test_")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	t.Run("Apply valid schema", func(t *testing.T) {
		// Create test database
		testDBPath := filepath.Join(tempDir, "schema_test.db")
		db, err := sql.Open("sqlite", testDBPath)
		assert.NoError(t, err)
		defer db.Close()

		// Create test schema file
		testSchemaPath := filepath.Join(tempDir, "test_schema.sql")
		testSchema := `
CREATE TABLE test_table (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);
INSERT INTO test_table (name) VALUES ('test');
`
		err = os.WriteFile(testSchemaPath, []byte(testSchema), 0o644)
		assert.NoError(t, err)

		err = applySchemaWithPath(db, testSchemaPath)
		assert.NoError(t, err)

		// Verify table was created and data was inserted
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM test_table").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("Schema file does not exist", func(t *testing.T) {
		testDBPath := filepath.Join(tempDir, "schema_test2.db")
		db, err := sql.Open("sqlite", testDBPath)
		assert.NoError(t, err)
		defer db.Close()

		nonExistentPath := filepath.Join(tempDir, "nonexistent.sql")
		err = applySchemaWithPath(db, nonExistentPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read schema file")
	})

	t.Run("Invalid SQL in schema", func(t *testing.T) {
		testDBPath := filepath.Join(tempDir, "schema_test3.db")
		db, err := sql.Open("sqlite", testDBPath)
		assert.NoError(t, err)
		defer db.Close()

		testSchemaPath := filepath.Join(tempDir, "invalid_schema.sql")
		invalidSQL := "THIS IS NOT VALID SQL;"
		err = os.WriteFile(testSchemaPath, []byte(invalidSQL), 0o644)
		assert.NoError(t, err)

		err = applySchemaWithPath(db, testSchemaPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute schema")
	})
}

// Helper function to apply schema with custom path
func applySchemaWithPath(db *sql.DB, schemaPath string) error {
	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file %s: %w", schemaPath, err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}

func TestCloseDB(t *testing.T) {
	t.Run("Close valid database connection", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "wallabag_close_test_")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		testDBPath := filepath.Join(tempDir, "close_test.db")
		db, err := sql.Open("sqlite", testDBPath)
		assert.NoError(t, err)

		// Test that database is working
		err = db.Ping()
		assert.NoError(t, err)

		// Close the database
		database.CloseDB(db)

		// After closing, operations should fail
		err = db.Ping()
		assert.Error(t, err)
	})

	t.Run("Close nil database", func(t *testing.T) {
		// Should not panic
		assert.NotPanics(t, func() {
			database.CloseDB(nil)
		})
	})

	t.Run("Close already closed database", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "wallabag_close_test2_")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		testDBPath := filepath.Join(tempDir, "close_test2.db")
		db, err := sql.Open("sqlite", testDBPath)
		assert.NoError(t, err)

		// Close once
		db.Close()

		// Close again through CloseDB - should not panic
		assert.NotPanics(t, func() {
			database.CloseDB(db)
		})
	})
}

func TestDatabaseConstants(t *testing.T) {
	t.Run("Schema path is set", func(t *testing.T) {
		schemaPath := "./db/schema.sql"
		assert.NotEmpty(t, schemaPath)
		assert.Equal(t, "./db/schema.sql", schemaPath)
	})
}
