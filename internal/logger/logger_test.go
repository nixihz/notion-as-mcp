// Package logger provides tests for structured logging.
package logger

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/nixihz/notion-as-mcp/internal/config"
)

func TestLevelMap(t *testing.T) {
	tests := []struct {
		name  string
		level string
		want  slog.Level
	}{
		{"debug level", "debug", slog.LevelDebug},
		{"info level", "info", slog.LevelInfo},
		{"warn level", "warn", slog.LevelWarn},
		{"error level", "error", slog.LevelError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := levelMap[tt.level]
			if !ok {
				t.Errorf("levelMap[%q] not found", tt.level)
			}
			if got != tt.want {
				t.Errorf("levelMap[%q] = %v, want %v", tt.level, got, tt.want)
			}
		})
	}
}

func TestInit(t *testing.T) {
	t.Run("First initialization", func(t *testing.T) {
		// Reset the once for testing
		once = *new(sync.Once)

		cfg := &config.Config{
			LogLevel: "debug",
		}

		err := Init(cfg)
		if err != nil {
			t.Fatalf("Init() failed: %v", err)
		}

		if defaultLogger == nil {
			t.Error("Init() did not set defaultLogger")
		}
	})

	t.Run("Init with info level", func(t *testing.T) {
		once = *new(sync.Once)

		cfg := &config.Config{
			LogLevel: "info",
		}

		err := Init(cfg)
		if err != nil {
			t.Fatalf("Init() failed: %v", err)
		}

		if defaultLogger == nil {
			t.Error("Init() did not set defaultLogger")
		}
	})

	t.Run("Init with warn level", func(t *testing.T) {
		once = *new(sync.Once)

		cfg := &config.Config{
			LogLevel: "warn",
		}

		err := Init(cfg)
		if err != nil {
			t.Fatalf("Init() failed: %v", err)
		}
	})

	t.Run("Init with error level", func(t *testing.T) {
		once = *new(sync.Once)

		cfg := &config.Config{
			LogLevel: "error",
		}

		err := Init(cfg)
		if err != nil {
			t.Fatalf("Init() failed: %v", err)
		}
	})
}

func TestGet(t *testing.T) {
	t.Run("Get after init", func(t *testing.T) {
		once = *new(sync.Once)

		cfg := &config.Config{LogLevel: "info"}
		Init(cfg)

		logger := Get()
		if logger == nil {
			t.Error("Get() returned nil")
		}
	})

	t.Run("Get before init", func(t *testing.T) {
		// Save and restore
		origLogger := defaultLogger
		defaultLogger = nil
		defer func() { defaultLogger = origLogger }()

		logger := Get()
		if logger != nil {
			// This is actually OK - it might return the default slog logger
		}
	})
}

func TestWith(t *testing.T) {
	once = *new(sync.Once)
	cfg := &config.Config{LogLevel: "info"}
	Init(cfg)

	t.Run("With attributes", func(t *testing.T) {
		logger := With("key1", "value1", "key2", 123)
		if logger == nil {
			t.Error("With() returned nil")
		}
	})

	t.Run("With before init", func(t *testing.T) {
		origLogger := defaultLogger
		defaultLogger = nil
		defer func() { defaultLogger = origLogger }()

		logger := With("key", "value")
		if logger == nil {
			t.Error("With() returned nil even before init")
		}
	})
}

func TestLogFunctions(t *testing.T) {
	once = *new(sync.Once)

	// Create a buffer to capture output
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	defaultLogger = slog.New(handler)

	t.Run("Debug", func(t *testing.T) {
		buf.Reset()
		Debug("debug message", "key", "value")

		output := buf.String()
		if !strings.Contains(output, "debug message") {
			t.Errorf("Debug() output = %q, should contain 'debug message'", output)
		}
	})

	t.Run("Info", func(t *testing.T) {
		buf.Reset()
		Info("info message", "key", "value")

		output := buf.String()
		if !strings.Contains(output, "info message") {
			t.Errorf("Info() output = %q, should contain 'info message'", output)
		}
	})

	t.Run("Warn", func(t *testing.T) {
		buf.Reset()
		Warn("warn message", "key", "value")

		output := buf.String()
		if !strings.Contains(output, "warn message") {
			t.Errorf("Warn() output = %q, should contain 'warn message'", output)
		}
	})

	t.Run("Error", func(t *testing.T) {
		buf.Reset()
		Error("error message", "key", "value")

		output := buf.String()
		if !strings.Contains(output, "error message") {
			t.Errorf("Error() output = %q, should contain 'error message'", output)
		}
	})

	t.Run("Err", func(t *testing.T) {
		buf.Reset()
		testErr := os.ErrInvalid
		Err("operation failed", testErr, "context", "test")

		output := buf.String()
		if !strings.Contains(output, "operation failed") {
			t.Errorf("Err() output = %q, should contain 'operation failed'", output)
		}
		if !strings.Contains(output, "error") {
			t.Errorf("Err() output = %q, should contain 'error' key", output)
		}
	})
}

func TestLogFunctionsBeforeInit(t *testing.T) {
	// Reset logger
	origLogger := defaultLogger
	defaultLogger = nil
	defer func() { defaultLogger = origLogger }()

	t.Run("Debug before init", func(t *testing.T) {
		// Should not panic
		Debug("debug message")
	})

	t.Run("Info before init", func(t *testing.T) {
		// Should not panic
		Info("info message")
	})

	t.Run("Warn before init", func(t *testing.T) {
		// Should not panic
		Warn("warn message")
	})

	t.Run("Error before init", func(t *testing.T) {
		// Should not panic
		Error("error message")
	})

	t.Run("Err before init", func(t *testing.T) {
		// Should not panic
		Err("operation failed", os.ErrInvalid)
	})
}

func TestSetOutput(t *testing.T) {
	once = *new(sync.Once)
	cfg := &config.Config{LogLevel: "info"}
	Init(cfg)

	t.Run("SetOutput to buffer", func(t *testing.T) {
		var buf bytes.Buffer

		SetOutput(&buf)

		// Log something
		Info("test message to buffer")

		output := buf.String()
		if !strings.Contains(output, "test message to buffer") {
			t.Errorf("Output = %q, should contain 'test message to buffer'", output)
		}
	})

	t.Run("SetOutput before init", func(t *testing.T) {
		origLogger := defaultLogger
		defaultLogger = nil
		defer func() { defaultLogger = origLogger }()

		var buf bytes.Buffer

		// Should not panic
		SetOutput(&buf)
	})
}

func TestInitIdempotent(t *testing.T) {
	once = *new(sync.Once)

	cfg1 := &config.Config{LogLevel: "debug"}
	cfg2 := &config.Config{LogLevel: "error"}

	// First init
	Init(cfg1)
	firstLogger := defaultLogger

	// Second init should be ignored
	Init(cfg2)
	secondLogger := defaultLogger

	// Logger should be the same instance
	if firstLogger != secondLogger {
		t.Error("Init() called twice created different loggers")
	}
}

func TestConcurrentLogging(t *testing.T) {
	once = *new(sync.Once)

	cfg := &config.Config{LogLevel: "info"}
	Init(cfg)

	done := make(chan bool)

	// Concurrent loggers
	for i := 0; i < 10; i++ {
		go func(idx int) {
			for j := 0; j < 100; j++ {
				Info("concurrent message", "worker", idx, "iteration", j)
			}
			done <- true
		}(i)
	}

	// Wait for completion
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic or deadlock
}

func TestStructuredLogging(t *testing.T) {
	once = *new(sync.Once)

	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	defaultLogger = slog.New(handler)

	t.Run("Log with structured attributes", func(t *testing.T) {
		buf.Reset()

		Info("user action",
			"user_id", "12345",
			"action", "login",
			"ip", "192.168.1.1",
			"success", true)

		output := buf.String()
		if !strings.Contains(output, "user_id") {
			t.Errorf("Output should contain user_id")
		}
		if !strings.Contains(output, "action") {
			t.Errorf("Output should contain action")
		}
	})

	t.Run("Log with complex values", func(t *testing.T) {
		buf.Reset()

		Debug("complex data",
			"numbers", []int{1, 2, 3},
			"nested", map[string]interface{}{"key": "value"})

		output := buf.String()
		// Should contain the field names
		if !strings.Contains(output, "numbers") {
			t.Errorf("Output should contain numbers field")
		}
	})
}

// Benchmark tests
func BenchmarkInfo(b *testing.B) {
	once = *new(sync.Once)
	cfg := &config.Config{LogLevel: "info"}
	Init(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Info("benchmark message", "key", "value")
	}
}

func BenchmarkDebug(b *testing.B) {
	once = *new(sync.Once)
	cfg := &config.Config{LogLevel: "debug"}
	Init(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Debug("benchmark message", "key", "value")
	}
}

func BenchmarkWith(b *testing.B) {
	once = *new(sync.Once)
	cfg := &config.Config{LogLevel: "info"}
	Init(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger := With("key", "value")
		logger.Info("message")
	}
}

func BenchmarkConcurrentLogging(b *testing.B) {
	once = *new(sync.Once)
	cfg := &config.Config{LogLevel: "info"}
	Init(cfg)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Info("concurrent message", "worker", "test")
		}
	})
}

func TestLogFileCreation(t *testing.T) {
	// Reset the once for testing
	once = *new(sync.Once)

	cfg := &config.Config{
		LogLevel: "info",
	}

	err := Init(cfg)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	if defaultLogger == nil {
		t.Error("Init() did not set defaultLogger")
	}

	// Verify log file was created at $HOME/.mcp/notion-as-mcp.log
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}
	expectedLogPath := filepath.Join(homeDir, ".mcp", "notion-as-mcp.log")

	if _, err := os.Stat(expectedLogPath); os.IsNotExist(err) {
		t.Errorf("Log file was not created at %s", expectedLogPath)
	}

	// Test Close function
	err = Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}

	// Verify we can call Close multiple times without error
	err = Close()
	if err != nil {
		t.Errorf("Close() called twice should not error: %v", err)
	}
}

func TestCloseBeforeInit(t *testing.T) {
	// Reset the once for testing
	once = *new(sync.Once)

	// Reset logFile
	origLogFile := logFile
	logFile = nil
	defer func() {
		logFile = origLogFile
	}()

	// Calling Close before Init should not panic
	err := Close()
	if err != nil {
		t.Errorf("Close() before Init() should not error: %v", err)
	}
}
