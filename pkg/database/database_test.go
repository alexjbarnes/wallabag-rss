package database_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
	"wallabag-rss-tool/pkg/database"
)

func TestInitDB(t *testing.T) {
	tests := []struct {
		setup     func() (string, func())
		checkFunc func(t *testing.T, db *sql.DB)
		name      string
		wantErr   bool
	}{
		{
			name: "InitDB with missing schema file",
			setup: func() (string, func()) {
				// No setup needed - schema file doesn't exist in test directory
				return "", func() {}
			},
			wantErr: true,
		},
		{
			name: "InitDB with default path - should fail in test environment",
			setup: func() (string, func()) {
				return "", func() {}
			},
			wantErr: true, // Expected to fail in test environment
		},
		{
			name: "InitDBWithPath creates new database successfully",
			setup: func() (string, func()) {
				tempDir, _ := os.MkdirTemp("", "wallabag_test_")
				dbPath := filepath.Join(tempDir, "test.db")
				
				// Change to project root where schema.sql exists
				originalDir, _ := os.Getwd()
				os.Chdir("../../")
				
				return dbPath, func() {
					os.Chdir(originalDir)
					os.RemoveAll(tempDir)
				}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, db *sql.DB) {
				t.Helper()
				// Verify tables were created
				var count int
				err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('feeds', 'articles', 'settings')").Scan(&count)
				assert.NoError(t, err)
				assert.Equal(t, 3, count)
				
				// Verify default setting was inserted
				var value string
				err = db.QueryRow("SELECT value FROM settings WHERE key='default_poll_interval_minutes'").Scan(&value)
				assert.NoError(t, err)
				assert.Equal(t, "1440", value)
			},
		},
		{
			name: "InitDBWithPath with invalid path",
			setup: func() (string, func()) {
				return "/non/existent/directory/test.db", func() {}
			},
			wantErr: true,
		},
		{
			name: "InitDBWithPath with existing database",
			setup: func() (string, func()) {
				tempDir, _ := os.MkdirTemp("", "wallabag_existing_")
				dbPath := filepath.Join(tempDir, "existing.db")
				
				// Create database file first
				file, _ := os.Create(dbPath)
				file.Close()
				
				// Change to project root
				originalDir, _ := os.Getwd()
				os.Chdir("../../")
				
				return dbPath, func() {
					os.Chdir(originalDir)
					os.RemoveAll(tempDir)
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbPath, cleanup := tt.setup()
			defer cleanup()

			var db *sql.DB
			var err error
			
			if dbPath == "" {
				db, err = database.InitDB()
			} else {
				db, err = database.InitDBWithPath(dbPath)
			}

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, db)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, db)
				defer db.Close()
				
				if tt.checkFunc != nil {
					tt.checkFunc(t, db)
				}
			}
		})
	}
}

func TestValidateDatabasePath(t *testing.T) {
	tests := []struct {
		name    string
		dbPath  string
		wantErr bool
	}{
		{
			name:    "empty path",
			dbPath:  "",
			wantErr: true,
		},
		{
			name:    "valid relative path",
			dbPath:  "./test.db",
			wantErr: false,
		},
		{
			name:    "valid absolute path in /tmp",
			dbPath:  "/tmp/test.db",
			wantErr: false,
		},
		{
			name:    "path with parent directory traversal",
			dbPath:  "../../../etc/passwd",
			wantErr: true, // validateDatabasePath rejects path traversal for security
		},
		{
			name:    "absolute path outside allowed directories",
			dbPath:  "/etc/passwd",
			wantErr: true,
		},
		{
			name:    "path with null byte",
			dbPath:  "test\x00.db",
			wantErr: false, // Current implementation doesn't check for null bytes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := database.ValidateDatabasePath(tt.dbPath)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestApplySchema(t *testing.T) {
	tests := []struct {
		setupDB   func() (*sql.DB, func())
		setupPath func() func()
		name      string
		wantErr   bool
	}{
		{
			name: "apply valid schema",
			setupDB: func() (*sql.DB, func()) {
				tempDir, _ := os.MkdirTemp("", "wallabag_schema_")
				dbPath := filepath.Join(tempDir, "test.db")
				db, _ := sql.Open("sqlite", dbPath)
				return db, func() {
					db.Close()
					os.RemoveAll(tempDir)
				}
			},
			setupPath: func() func() {
				originalDir, _ := os.Getwd()
				os.Chdir("../../")
				return func() { os.Chdir(originalDir) }
			},
			wantErr: false,
		},
		{
			name: "schema file not found",
			setupDB: func() (*sql.DB, func()) {
				tempDir, _ := os.MkdirTemp("", "wallabag_noschema_")
				dbPath := filepath.Join(tempDir, "test.db")
				db, _ := sql.Open("sqlite", dbPath)
				return db, func() {
					db.Close()
					os.RemoveAll(tempDir)
				}
			},
			setupPath: func() func() {
				// Stay in current directory where schema doesn't exist
				return func() {}
			},
			wantErr: true,
		},
		{
			name: "closed database",
			setupDB: func() (*sql.DB, func()) {
				tempDir, _ := os.MkdirTemp("", "wallabag_closed_")
				dbPath := filepath.Join(tempDir, "test.db")
				db, _ := sql.Open("sqlite", dbPath)
				db.Close() // Close before applying schema
				return db, func() {
					os.RemoveAll(tempDir)
				}
			},
			setupPath: func() func() {
				originalDir, _ := os.Getwd()
				os.Chdir("../../")
				return func() { os.Chdir(originalDir) }
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, cleanupDB := tt.setupDB()
			defer cleanupDB()
			
			cleanupPath := tt.setupPath()
			defer cleanupPath()

			err := database.ApplySchema(db)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCloseDB(t *testing.T) {
	tests := []struct {
		setupDB func() *sql.DB
		name    string
	}{
		{
			name: "close valid database connection",
			setupDB: func() *sql.DB {
				tempDir, _ := os.MkdirTemp("", "wallabag_close_")
				defer os.RemoveAll(tempDir)
				dbPath := filepath.Join(tempDir, "test.db")
				db, _ := sql.Open("sqlite", dbPath)
				return db
			},
		},
		{
			name: "close nil database",
			setupDB: func() *sql.DB {
				return nil
			},
		},
		{
			name: "close already closed database",
			setupDB: func() *sql.DB {
				tempDir, _ := os.MkdirTemp("", "wallabag_closed_")
				defer os.RemoveAll(tempDir)
				dbPath := filepath.Join(tempDir, "test.db")
				db, _ := sql.Open("sqlite", dbPath)
				db.Close()
				return db
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.setupDB()
			
			// Should not panic
			assert.NotPanics(t, func() {
				database.CloseDB(db)
			})
		})
	}
}

func TestDirectoryCreation(t *testing.T) {
	t.Run("InitDBWithPath creates nested directories", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "wallabag_nested_")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		dbPath := filepath.Join(tempDir, "level1", "level2", "level3", "test.db")
		
		// Change to project root
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir("../../")
		
		// InitDBWithPath should create all parent directories
		db, err := database.InitDBWithPath(dbPath)
		assert.NoError(t, err)
		assert.NotNil(t, db)
		defer db.Close()
		
		// Verify all directories were created
		assert.DirExists(t, filepath.Join(tempDir, "level1"))
		assert.DirExists(t, filepath.Join(tempDir, "level1", "level2"))
		assert.DirExists(t, filepath.Join(tempDir, "level1", "level2", "level3"))
		assert.FileExists(t, dbPath)
	})

	t.Run("InitDBWithPath with file as parent directory", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "wallabag_file_")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Create a file where we want to create a directory
		existingFile := filepath.Join(tempDir, "existing_file")
		err = os.WriteFile(existingFile, []byte("test"), 0644)
		require.NoError(t, err)

		// Try to create database "inside" a file
		dbPath := filepath.Join(existingFile, "impossible.db")
		
		db, err := database.InitDBWithPath(dbPath)
		assert.Error(t, err)
		assert.Nil(t, db)
	})
}