// Package logger provides structured logging using Go's slog package.
package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync"

	"github.com/nixihz/notion-as-mcp/internal/config"
)

var (
	// defaultLogger is the global logger instance.
	defaultLogger *slog.Logger
	// once ensures the default logger is initialized only once.
	once sync.Once
)

// Level strings to slog levels.
var levelMap = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

// Init initializes the global logger with the given configuration.
func Init(cfg *config.Config) error {
	var initErr error
	once.Do(func() {
		level := slog.LevelInfo
		if l, ok := levelMap[cfg.LogLevel]; ok {
			level = l
		}

		handlerOptions := &slog.HandlerOptions{
			Level: level,
		}

		// Use JSON handler for structured logging
		defaultLogger = slog.New(slog.NewJSONHandler(os.Stdout, handlerOptions))
		slog.SetDefault(defaultLogger)
	})
	return initErr
}

// Get returns the global logger instance.
func Get() *slog.Logger {
	return defaultLogger
}

// With returns a new logger with the given attributes.
func With(attrs ...any) *slog.Logger {
	if defaultLogger == nil {
		return slog.Default()
	}
	return defaultLogger.With(attrs...)
}

// Debug logs a debug message.
func Debug(msg string, args ...any) {
	if defaultLogger != nil {
		defaultLogger.Debug(msg, args...)
	}
}

// Info logs an info message.
func Info(msg string, args ...any) {
	if defaultLogger != nil {
		defaultLogger.Info(msg, args...)
	}
}

// Warn logs a warning message.
func Warn(msg string, args ...any) {
	if defaultLogger != nil {
		defaultLogger.Warn(msg, args...)
	}
}

// Error logs an error message.
func Error(msg string, args ...any) {
	if defaultLogger != nil {
		defaultLogger.Error(msg, args...)
	}
}

// Err logs an error with the given message and error.
func Err(msg string, err error, args ...any) {
	if defaultLogger != nil {
		defaultLogger.Error(msg, append(args, slog.String("error", err.Error()))...)
	}
}

// Context returns a context with the logger.
func Context(ctx context.Context) context.Context {
	return context.WithValue(ctx, "logger", defaultLogger)
}

// SetOutput redirects the logger output to the given writer.
// Note: In Go 1.25, we need to create a new handler with the new writer.
// This is a limitation of the current implementation.
func SetOutput(w io.Writer) {
	if defaultLogger != nil {
		// In Go 1.25, we cannot easily change the output of an existing handler.
		// The handler would need to be recreated. For now, we create a new logger.
		handlerOptions := &slog.HandlerOptions{}
		defaultLogger = slog.New(slog.NewJSONHandler(w, handlerOptions))
		slog.SetDefault(defaultLogger)
	}
}
