// Package logging provides structured logging functionality using slog with additional convenience methods.
package logging

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"sync"
)

// Logger defines the interface for structured logging
type Logger interface {
	// Core logging methods
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)

	// Context-aware logging
	DebugContext(ctx context.Context, msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)

	// Create a logger with additional attributes
	With(args ...any) Logger
}

// SlogLogger implements Logger using the standard library slog
type SlogLogger struct {
	logger *slog.Logger
}

// NewSlogLogger creates a new SlogLogger
func NewSlogLogger(logger *slog.Logger) *SlogLogger {
	return &SlogLogger{logger: logger}
}

// NewDefaultLogger creates a logger with default JSON configuration
func NewDefaultLogger() *SlogLogger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	return NewSlogLogger(logger)
}

// NewDevelopmentLogger creates a logger with text output for development
func NewDevelopmentLogger() *SlogLogger {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	return NewSlogLogger(logger)
}

// NewLoggerWithLevel creates a logger with the specified level and format
func NewLoggerWithLevel(level string, useJSON bool) *SlogLogger {
	slogLevel := ParseLevelString(level)
	
	var handler slog.Handler
	if useJSON {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slogLevel,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slogLevel,
		})
	}

	logger := slog.New(handler)

	return NewSlogLogger(logger)
}

// ParseLevelString converts a string level to slog.Level
func ParseLevelString(level string) slog.Level {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo // Default to INFO if unknown
	}
}

// Debug logs a debug message
func (l *SlogLogger) Debug(msg string, args ...any) {
	l.logger.DebugContext(context.Background(), msg, args...)
}

// Info logs an info message
func (l *SlogLogger) Info(msg string, args ...any) {
	l.logger.InfoContext(context.Background(), msg, args...)
}

// Warn logs a warning message
func (l *SlogLogger) Warn(msg string, args ...any) {
	l.logger.WarnContext(context.Background(), msg, args...)
}

// Error logs an error message
func (l *SlogLogger) Error(msg string, args ...any) {
	l.logger.ErrorContext(context.Background(), msg, args...)
}

// DebugContext logs a debug message with context
func (l *SlogLogger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.logger.DebugContext(ctx, msg, args...)
}

// InfoContext logs an info message with context
func (l *SlogLogger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.logger.InfoContext(ctx, msg, args...)
}

// WarnContext logs a warning message with context
func (l *SlogLogger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.logger.WarnContext(ctx, msg, args...)
}

// ErrorContext logs an error message with context
func (l *SlogLogger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.logger.ErrorContext(ctx, msg, args...)
}

// With returns a new logger with additional attributes
func (l *SlogLogger) With(args ...any) Logger {
	return &SlogLogger{logger: l.logger.With(args...)}
}

// Global logger instance with thread-safe access
var (
	globalLogger Logger = NewDefaultLogger()
	loggerMutex  sync.RWMutex
)

// SetGlobalLogger sets the global logger instance (thread-safe)
func SetGlobalLogger(logger Logger) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	globalLogger = logger
}

// GetGlobalLogger returns the global logger instance (thread-safe)
func GetGlobalLogger() Logger {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()

	return globalLogger
}

// Debug logs a debug message using the global logger.
func Debug(msg string, args ...any) {
	GetGlobalLogger().Debug(msg, args...)
}

func Info(msg string, args ...any) {
	GetGlobalLogger().Info(msg, args...)
}

func Warn(msg string, args ...any) {
	GetGlobalLogger().Warn(msg, args...)
}

func Error(msg string, args ...any) {
	GetGlobalLogger().Error(msg, args...)
}

func DebugContext(ctx context.Context, msg string, args ...any) {
	GetGlobalLogger().DebugContext(ctx, msg, args...)
}

func InfoContext(ctx context.Context, msg string, args ...any) {
	GetGlobalLogger().InfoContext(ctx, msg, args...)
}

func WarnContext(ctx context.Context, msg string, args ...any) {
	GetGlobalLogger().WarnContext(ctx, msg, args...)
}

func ErrorContext(ctx context.Context, msg string, args ...any) {
	GetGlobalLogger().ErrorContext(ctx, msg, args...)
}

func With(args ...any) Logger {
	return GetGlobalLogger().With(args...)
}
