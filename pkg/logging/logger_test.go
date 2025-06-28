package logging_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"wallabag-rss-tool/pkg/logging"
)

func TestMockLogger(t *testing.T) {
	t.Run("Basic logging", func(t *testing.T) {
		logger := logging.NewMockLogger()

		logger.Info("test message", "key", "value")
		logger.Error("error message", "error", "something went wrong")

		entries := logger.GetEntries()
		assert.Len(t, entries, 2)

		assert.Equal(t, "INFO", entries[0].Level)
		assert.Equal(t, "test message", entries[0].Message)
		assert.Contains(t, entries[0].Args, "key")
		assert.Contains(t, entries[0].Args, "value")

		assert.Equal(t, "ERROR", entries[1].Level)
		assert.Equal(t, "error message", entries[1].Message)
	})

	t.Run("With attributes", func(t *testing.T) {
		logger := logging.NewMockLogger()
		contextLogger := logger.With("component", "test", "version", "1.0")

		contextLogger.Info("test with context")

		entries := logger.GetEntries()
		assert.Len(t, entries, 1)

		entry := entries[0]
		assert.Equal(t, "INFO", entry.Level)
		assert.Equal(t, "test with context", entry.Message)
		assert.Contains(t, entry.Args, "component")
		assert.Contains(t, entry.Args, "test")
		assert.Contains(t, entry.Args, "version")
		assert.Contains(t, entry.Args, "1.0")
	})

	t.Run("Context logging", func(t *testing.T) {
		logger := logging.NewMockLogger()
		type ctxKey string
		ctx := context.WithValue(context.Background(), ctxKey("request_id"), "12345")

		logger.InfoContext(ctx, "context message", "user", "john")

		entries := logger.GetEntries()
		assert.Len(t, entries, 1)

		entry := entries[0]
		assert.Equal(t, "INFO", entry.Level)
		assert.Equal(t, "context message", entry.Message)
		assert.NotNil(t, entry.Context)
		assert.Equal(t, "12345", entry.Context.Value(ctxKey("request_id")))
	})

	t.Run("Filter by level", func(t *testing.T) {
		logger := logging.NewMockLogger()

		logger.Debug("debug message")
		logger.Info("info message")
		logger.Warn("warn message")
		logger.Error("error message")

		assert.Equal(t, 4, logger.Count())
		assert.Equal(t, 1, logger.CountByLevel("DEBUG"))
		assert.Equal(t, 1, logger.CountByLevel("INFO"))
		assert.Equal(t, 1, logger.CountByLevel("WARN"))
		assert.Equal(t, 1, logger.CountByLevel("ERROR"))

		errorEntries := logger.GetEntriesByLevel("ERROR")
		assert.Len(t, errorEntries, 1)
		assert.Equal(t, "error message", errorEntries[0].Message)
	})

	t.Run("HasEntry checks", func(t *testing.T) {
		logger := logging.NewMockLogger()

		logger.Info("user logged in", "user_id", 123, "ip", "192.168.1.1")
		logger.Error("failed to save", "error", "database timeout")

		assert.True(t, logger.HasEntry("INFO", "user logged in"))
		assert.True(t, logger.HasEntry("ERROR", "failed to save"))
		assert.False(t, logger.HasEntry("DEBUG", "user logged in"))
		assert.False(t, logger.HasEntry("INFO", "nonexistent message"))

		assert.True(t, logger.HasEntryWithArgs("INFO", "user logged in", "user_id", 123))
		assert.True(t, logger.HasEntryWithArgs("ERROR", "failed to save", "error", "database timeout"))
		assert.False(t, logger.HasEntryWithArgs("INFO", "user logged in", "user_id", 999))
	})

	t.Run("Clear entries", func(t *testing.T) {
		logger := logging.NewMockLogger()

		logger.Info("message 1")
		logger.Error("message 2")
		assert.Equal(t, 2, logger.Count())

		logger.Clear()
		assert.Equal(t, 0, logger.Count())
		assert.Empty(t, logger.GetEntries())
	})
}

func TestGlobalLogger(t *testing.T) {
	t.Run("Global logger functions", func(t *testing.T) {
		// Save original logger
		originalLogger := logging.GetGlobalLogger()
		defer logging.SetGlobalLogger(originalLogger)

		// Set mock logger as global
		mockLogger := logging.NewMockLogger()
		logging.SetGlobalLogger(mockLogger)

		// Use global functions
		logging.Info("global info message", "key", "value")
		logging.Error("global error message", "error", "test error")

		// Verify logs were captured
		entries := mockLogger.GetEntries()
		assert.Len(t, entries, 2)

		assert.Equal(t, "INFO", entries[0].Level)
		assert.Equal(t, "global info message", entries[0].Message)

		assert.Equal(t, "ERROR", entries[1].Level)
		assert.Equal(t, "global error message", entries[1].Message)
	})

	t.Run("Global With function", func(t *testing.T) {
		// Save original logger
		originalLogger := logging.GetGlobalLogger()
		defer logging.SetGlobalLogger(originalLogger)

		// Set mock logger as global
		mockLogger := logging.NewMockLogger()
		logging.SetGlobalLogger(mockLogger)

		// Use global With function
		contextLogger := logging.With("service", "test")
		contextLogger.Info("context message")

		// Verify logs were captured with context
		entries := mockLogger.GetEntries()
		assert.Len(t, entries, 1)

		entry := entries[0]
		assert.Equal(t, "INFO", entry.Level)
		assert.Equal(t, "context message", entry.Message)
		assert.Contains(t, entry.Args, "service")
		assert.Contains(t, entry.Args, "test")
	})
}
