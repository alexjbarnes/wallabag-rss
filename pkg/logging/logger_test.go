package logging_test

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"wallabag-rss-tool/pkg/logging"
)

func TestNewLoggers(t *testing.T) {
	tests := []struct {
		createFunc  func() *logging.SlogLogger
		checkLogger func(t *testing.T, logger *logging.SlogLogger)
		name        string
	}{
		{
			name: "NewSlogLogger",
			createFunc: func() *logging.SlogLogger {
				slogInstance := slog.New(slog.NewJSONHandler(os.Stdout, nil))
				return logging.NewSlogLogger(slogInstance)
			},
			checkLogger: func(t *testing.T, logger *logging.SlogLogger) {
				t.Helper()
				assert.NotNil(t, logger)
				// Cannot access unexported logger field from external test package
			},
		},
		{
			name: "NewDefaultLogger",
			createFunc: logging.NewDefaultLogger,
			checkLogger: func(t *testing.T, logger *logging.SlogLogger) {
				t.Helper()
				assert.NotNil(t, logger)
				// Cannot access unexported logger field from external test package
			},
		},
		{
			name: "NewDevelopmentLogger",
			createFunc: logging.NewDevelopmentLogger,
			checkLogger: func(t *testing.T, logger *logging.SlogLogger) {
				t.Helper()
				assert.NotNil(t, logger)
				// Cannot access unexported logger field from external test package
			},
		},
		{
			name: "NewLoggerWithLevel JSON format",
			createFunc: func() *logging.SlogLogger {
				return logging.NewLoggerWithLevel("INFO", true)
			},
			checkLogger: func(t *testing.T, logger *logging.SlogLogger) {
				t.Helper()
				assert.NotNil(t, logger)
				// Cannot access unexported logger field from external test package
			},
		},
		{
			name: "NewLoggerWithLevel text format",
			createFunc: func() *logging.SlogLogger {
				return logging.NewLoggerWithLevel("DEBUG", false)
			},
			checkLogger: func(t *testing.T, logger *logging.SlogLogger) {
				t.Helper()
				assert.NotNil(t, logger)
				// Cannot access unexported logger field from external test package
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := tt.createFunc()
			tt.checkLogger(t, logger)
		})
	}
}

func TestParseLevelString(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"DEBUG", slog.LevelDebug},
		{"debug", slog.LevelDebug},
		{"  Debug  ", slog.LevelDebug},
		{"INFO", slog.LevelInfo},
		{"info", slog.LevelInfo},
		{"WARN", slog.LevelWarn},
		{"warn", slog.LevelWarn},
		{"WARNING", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"ERROR", slog.LevelError},
		{"error", slog.LevelError},
		{"unknown", slog.LevelInfo},
		{"", slog.LevelInfo},
		{"invalid", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run("parse level "+tt.input, func(t *testing.T) {
			result := logging.ParseLevelString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSlogLogger_LoggingMethods(t *testing.T) {
	tests := []struct {
		logFunc   func(logger *logging.SlogLogger)
		checkFunc func(t *testing.T, output string)
		name      string
	}{
		{
			name: "Debug method",
			logFunc: func(logger *logging.SlogLogger) {
				logger.Debug("debug message", "key", "value")
			},
			checkFunc: func(t *testing.T, output string) {
				t.Helper()
				assert.Contains(t, output, "debug message")
				assert.Contains(t, output, "key=value")
			},
		},
		{
			name: "Info method",
			logFunc: func(logger *logging.SlogLogger) {
				logger.Info("info message", "key", "value")
			},
			checkFunc: func(t *testing.T, output string) {
				t.Helper()
				assert.Contains(t, output, "info message")
				assert.Contains(t, output, "key=value")
			},
		},
		{
			name: "Warn method",
			logFunc: func(logger *logging.SlogLogger) {
				logger.Warn("warn message", "key", "value")
			},
			checkFunc: func(t *testing.T, output string) {
				t.Helper()
				assert.Contains(t, output, "warn message")
				assert.Contains(t, output, "key=value")
			},
		},
		{
			name: "Error method",
			logFunc: func(logger *logging.SlogLogger) {
				logger.Error("error message", "key", "value")
			},
			checkFunc: func(t *testing.T, output string) {
				t.Helper()
				assert.Contains(t, output, "error message")
				assert.Contains(t, output, "key=value")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output strings.Builder
			handler := slog.NewTextHandler(&output, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})
			logger := logging.NewSlogLogger(slog.New(handler))

			tt.logFunc(logger)
			tt.checkFunc(t, output.String())
		})
	}
}

func TestSlogLogger_ContextMethods(t *testing.T) {
	ctx := context.Background()
	
	tests := []struct {
		logFunc   func(logger *logging.SlogLogger, ctx context.Context)
		checkFunc func(t *testing.T, output string)
		name      string
	}{
		{
			name: "DebugContext method",
			logFunc: func(logger *logging.SlogLogger, ctx context.Context) {
				logger.DebugContext(ctx, "debug context message", "key", "value")
			},
			checkFunc: func(t *testing.T, output string) {
				t.Helper()
				assert.Contains(t, output, "debug context message")
			},
		},
		{
			name: "InfoContext method",
			logFunc: func(logger *logging.SlogLogger, ctx context.Context) {
				logger.InfoContext(ctx, "info context message", "key", "value")
			},
			checkFunc: func(t *testing.T, output string) {
				t.Helper()
				assert.Contains(t, output, "info context message")
			},
		},
		{
			name: "WarnContext method",
			logFunc: func(logger *logging.SlogLogger, ctx context.Context) {
				logger.WarnContext(ctx, "warn context message", "key", "value")
			},
			checkFunc: func(t *testing.T, output string) {
				t.Helper()
				assert.Contains(t, output, "warn context message")
			},
		},
		{
			name: "ErrorContext method",
			logFunc: func(logger *logging.SlogLogger, ctx context.Context) {
				logger.ErrorContext(ctx, "error context message", "key", "value")
			},
			checkFunc: func(t *testing.T, output string) {
				t.Helper()
				assert.Contains(t, output, "error context message")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output strings.Builder
			handler := slog.NewTextHandler(&output, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})
			logger := logging.NewSlogLogger(slog.New(handler))

			tt.logFunc(logger, ctx)
			tt.checkFunc(t, output.String())
		})
	}
}

func TestSlogLogger_With(t *testing.T) {
	var output strings.Builder
	handler := slog.NewTextHandler(&output, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := logging.NewSlogLogger(slog.New(handler))

	t.Run("With method creates new logger with additional attributes", func(t *testing.T) {
		childLogger := logger.With("component", "test", "version", "1.0")
		assert.NotNil(t, childLogger)
		
		// Verify it's a different instance
		assert.NotEqual(t, logger, childLogger)
		
		// Test that the child logger includes the additional attributes
		output.Reset()
		childLogger.Info("test message")
		logOutput := output.String()
		assert.Contains(t, logOutput, "test message")
		assert.Contains(t, logOutput, "component=test")
		assert.Contains(t, logOutput, "version=1.0")
	})
}

func TestGlobalLogger(t *testing.T) {
	// Save original global logger
	originalLogger := logging.GetGlobalLogger()
	defer logging.SetGlobalLogger(originalLogger)

	t.Run("SetGlobalLogger and GetGlobalLogger", func(t *testing.T) {
		var output strings.Builder
		handler := slog.NewTextHandler(&output, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
		testLogger := logging.NewSlogLogger(slog.New(handler))

		logging.SetGlobalLogger(testLogger)
		retrievedLogger := logging.GetGlobalLogger()
		
		assert.Equal(t, testLogger, retrievedLogger)
	})

	t.Run("Global logger thread safety", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 10
		
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(_ int) {
				defer wg.Done()
				
				// Create a new logger for this goroutine
				var output strings.Builder
				handler := slog.NewTextHandler(&output, &slog.HandlerOptions{
					Level: slog.LevelDebug,
				})
				testLogger := logging.NewSlogLogger(slog.New(handler))
				
				// Set and get the global logger multiple times
				for j := 0; j < 5; j++ {
					logging.SetGlobalLogger(testLogger)
					_ = logging.GetGlobalLogger()
				}
			}(i)
		}
		
		wg.Wait()
		// Test passes if no race condition occurs
	})
}

func TestGlobalLoggingFunctions(t *testing.T) {
	// Save original global logger
	originalLogger := logging.GetGlobalLogger()
	defer logging.SetGlobalLogger(originalLogger)

	// Set up a test logger to capture output
	var output strings.Builder
	handler := slog.NewTextHandler(&output, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	testLogger := logging.NewSlogLogger(slog.New(handler))
	logging.SetGlobalLogger(testLogger)

	tests := []struct {
		logFunc   func()
		checkFunc func(t *testing.T, output string)
		name      string
	}{
		{
			name: "Global Debug function",
			logFunc: func() {
				logging.Debug("global debug message", "key", "value")
			},
			checkFunc: func(t *testing.T, output string) {
				t.Helper()
				assert.Contains(t, output, "global debug message")
			},
		},
		{
			name: "Global Info function",
			logFunc: func() {
				logging.Info("global info message", "key", "value")
			},
			checkFunc: func(t *testing.T, output string) {
				t.Helper()
				assert.Contains(t, output, "global info message")
			},
		},
		{
			name: "Global Warn function",
			logFunc: func() {
				logging.Warn("global warn message", "key", "value")
			},
			checkFunc: func(t *testing.T, output string) {
				t.Helper()
				assert.Contains(t, output, "global warn message")
			},
		},
		{
			name: "Global Error function",
			logFunc: func() {
				logging.Error("global error message", "key", "value")
			},
			checkFunc: func(t *testing.T, output string) {
				t.Helper()
				assert.Contains(t, output, "global error message")
			},
		},
		{
			name: "Global DebugContext function",
			logFunc: func() {
				logging.DebugContext(context.Background(), "global debug context", "key", "value")
			},
			checkFunc: func(t *testing.T, output string) {
				t.Helper()
				assert.Contains(t, output, "global debug context")
			},
		},
		{
			name: "Global InfoContext function",
			logFunc: func() {
				logging.InfoContext(context.Background(), "global info context", "key", "value")
			},
			checkFunc: func(t *testing.T, output string) {
				t.Helper()
				assert.Contains(t, output, "global info context")
			},
		},
		{
			name: "Global WarnContext function",
			logFunc: func() {
				logging.WarnContext(context.Background(), "global warn context", "key", "value")
			},
			checkFunc: func(t *testing.T, output string) {
				t.Helper()
				assert.Contains(t, output, "global warn context")
			},
		},
		{
			name: "Global ErrorContext function",
			logFunc: func() {
				logging.ErrorContext(context.Background(), "global error context", "key", "value")
			},
			checkFunc: func(t *testing.T, output string) {
				t.Helper()
				assert.Contains(t, output, "global error context")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output.Reset()
			tt.logFunc()
			tt.checkFunc(t, output.String())
		})
	}

	t.Run("Global With function", func(t *testing.T) {
		childLogger := logging.With("global", "component")
		assert.NotNil(t, childLogger)
		
		output.Reset()
		childLogger.Info("test with global logger")
		assert.Contains(t, output.String(), "test with global logger")
		assert.Contains(t, output.String(), "global=component")
	})
}

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

		logger.InfoContext(ctx, "context message", "action", "test")

		entries := logger.GetEntries()
		assert.Len(t, entries, 1)
		assert.Equal(t, "INFO", entries[0].Level)
		assert.Equal(t, "context message", entries[0].Message)
		assert.NotNil(t, entries[0].Context)
		assert.Equal(t, "12345", entries[0].Context.Value(ctxKey("request_id")))
	})

	t.Run("All log levels", func(t *testing.T) {
		logger := logging.NewMockLogger()
		ctx := context.Background()

		logger.Debug("debug", "level", 1)
		logger.Info("info", "level", 2)
		logger.Warn("warn", "level", 3)
		logger.Error("error", "level", 4)
		logger.DebugContext(ctx, "debug ctx", "level", 5)
		logger.InfoContext(ctx, "info ctx", "level", 6)
		logger.WarnContext(ctx, "warn ctx", "level", 7)
		logger.ErrorContext(ctx, "error ctx", "level", 8)

		entries := logger.GetEntries()
		assert.Len(t, entries, 8)

		expectedLevels := []string{"DEBUG", "INFO", "WARN", "ERROR", "DEBUG", "INFO", "WARN", "ERROR"}
		for i, level := range expectedLevels {
			assert.Equal(t, level, entries[i].Level)
		}
	})
}