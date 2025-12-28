// Package logger provides structured logging using Go's slog package.
package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/nixihz/notion-as-mcp/internal/config"
)

var (
	// defaultLogger is the global logger instance.
	defaultLogger *slog.Logger
	// logFile is the current log file handle
	logFile *os.File
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

		// Create logs directory and log file
		logDir := "logs"
		if err := os.MkdirAll(logDir, 0755); err != nil {
			initErr = fmt.Errorf("failed to create logs directory: %w", err)
			return
		}

		// Generate log filename with current date (YYYYMMDD.log)
		currentDate := time.Now().Format("20060102")
		logFilePath := filepath.Join(logDir, fmt.Sprintf("%s.log", currentDate))

		// Open log file in append mode, create if not exists
		file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			initErr = fmt.Errorf("failed to open log file: %w", err)
			return
		}
		logFile = file

		// Use JSON handler for structured logging (output to file only)
		defaultLogger = slog.New(slog.NewJSONHandler(logFile, handlerOptions))
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

// Close closes the log file if it is open.
func Close() error {
	if logFile != nil {
		err := logFile.Close()
		logFile = nil
		return err
	}
	return nil
}
