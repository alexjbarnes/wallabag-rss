package logging

import (
	"context"
	"fmt"
	"sync"
)

// LogEntry represents a single log entry for testing
type LogEntry struct { //nolint:govet // Field order optimized for functionality over memory alignment
	Context context.Context //nolint:containedctx // Required for test verification
	Args    []any
	Level   string
	Message string
}

// MockLogger implements Logger interface for testing
type MockLogger struct {
	mu      *sync.RWMutex
	entries *[]LogEntry
	attrs   []any
}

// NewMockLogger creates a new mock logger for testing
func NewMockLogger() *MockLogger {
	entries := make([]LogEntry, 0)

	return &MockLogger{
		mu:      &sync.RWMutex{},
		entries: &entries,
		attrs:   make([]any, 0),
	}
}

// Debug logs a debug message
func (m *MockLogger) Debug(msg string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	*m.entries = append(*m.entries, LogEntry{
		Level:   "DEBUG",
		Message: msg,
		Args:    append(m.attrs, args...),
	})
}

// Info logs an info message
func (m *MockLogger) Info(msg string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	*m.entries = append(*m.entries, LogEntry{
		Level:   "INFO",
		Message: msg,
		Args:    append(m.attrs, args...),
	})
}

// Warn logs a warning message
func (m *MockLogger) Warn(msg string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	*m.entries = append(*m.entries, LogEntry{
		Level:   "WARN",
		Message: msg,
		Args:    append(m.attrs, args...),
	})
}

// Error logs an error message
func (m *MockLogger) Error(msg string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	*m.entries = append(*m.entries, LogEntry{
		Level:   "ERROR",
		Message: msg,
		Args:    append(m.attrs, args...),
	})
}

// DebugContext logs a debug message with context
func (m *MockLogger) DebugContext(ctx context.Context, msg string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	*m.entries = append(*m.entries, LogEntry{
		Level:   "DEBUG",
		Message: msg,
		Args:    append(m.attrs, args...),
		Context: ctx,
	})
}

// InfoContext logs an info message with context
func (m *MockLogger) InfoContext(ctx context.Context, msg string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	*m.entries = append(*m.entries, LogEntry{
		Level:   "INFO",
		Message: msg,
		Args:    append(m.attrs, args...),
		Context: ctx,
	})
}

// WarnContext logs a warning message with context
func (m *MockLogger) WarnContext(ctx context.Context, msg string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	*m.entries = append(*m.entries, LogEntry{
		Level:   "WARN",
		Message: msg,
		Args:    append(m.attrs, args...),
		Context: ctx,
	})
}

// ErrorContext logs an error message with context
func (m *MockLogger) ErrorContext(ctx context.Context, msg string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	*m.entries = append(*m.entries, LogEntry{
		Level:   "ERROR",
		Message: msg,
		Args:    append(m.attrs, args...),
		Context: ctx,
	})
}

// With returns a new logger with additional attributes
func (m *MockLogger) With(args ...any) Logger {
	return &MockLogger{
		mu:      m.mu,
		entries: m.entries,
		attrs:   append(append([]any{}, m.attrs...), args...),
	}
}

// GetEntries returns all logged entries (thread-safe)
func (m *MockLogger) GetEntries() []LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	entries := make([]LogEntry, len(*m.entries))
	copy(entries, *m.entries)

	return entries
}

// GetEntriesByLevel returns entries filtered by level
func (m *MockLogger) GetEntriesByLevel(level string) []LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var filtered []LogEntry
	for _, entry := range *m.entries {
		if entry.Level == level {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

// HasEntry checks if a log entry with the given message exists
func (m *MockLogger) HasEntry(level, message string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, entry := range *m.entries {
		if entry.Level == level && entry.Message == message {
			return true
		}
	}

	return false
}

// HasEntryWithArgs checks if a log entry exists with specific arguments
func (m *MockLogger) HasEntryWithArgs(level, message string, expectedArgs ...any) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, entry := range *m.entries {
		if m.entryMatches(entry, level, message, expectedArgs) {
			return true
		}
	}

	return false
}

// entryMatches checks if a log entry matches the given criteria
func (m *MockLogger) entryMatches(entry LogEntry, level, message string, expectedArgs []any) bool {
	if entry.Level != level || entry.Message != message {
		return false
	}

	if len(expectedArgs) == 0 {
		return true
	}

	return m.argsMatch(entry.Args, expectedArgs)
}

// argsMatch checks if the entry args contain the expected args
func (m *MockLogger) argsMatch(entryArgs, expectedArgs []any) bool {
	for i := 0; i < len(expectedArgs) && i < len(entryArgs); i += 2 {
		if m.argPairMatches(entryArgs, expectedArgs, i) {
			return true
		}
	}

	return false
}

// argPairMatches checks if a key-value pair matches
func (m *MockLogger) argPairMatches(entryArgs, expectedArgs []any, index int) bool {
	if index+1 >= len(expectedArgs) || index+1 >= len(entryArgs) {
		return false
	}

	return fmt.Sprintf("%v", entryArgs[index]) == fmt.Sprintf("%v", expectedArgs[index]) &&
		fmt.Sprintf("%v", entryArgs[index+1]) == fmt.Sprintf("%v", expectedArgs[index+1])
}

// Clear removes all logged entries
func (m *MockLogger) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	*m.entries = (*m.entries)[:0]
}

// Count returns the number of logged entries
func (m *MockLogger) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(*m.entries)
}

// CountByLevel returns the number of entries for a specific level
func (m *MockLogger) CountByLevel(level string) int {
	return len(m.GetEntriesByLevel(level))
}
