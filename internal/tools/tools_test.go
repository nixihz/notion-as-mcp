// Package tools provides tests for tool execution and registry.
package tools

import (
	"context"
	"testing"
	"time"
)

func TestNewExecutor(t *testing.T) {
	t.Run("Default configuration", func(t *testing.T) {
		e := NewExecutor(30*time.Second, "bash,python,js")

		if e == nil {
			t.Fatal("NewExecutor() returned nil")
		}
		if e.timeout != 30*time.Second {
			t.Errorf("timeout = %v, want 30s", e.timeout)
		}
	})

	t.Run("Empty languages", func(t *testing.T) {
		e := NewExecutor(10*time.Second, "")

		if len(e.languages) != 0 {
			t.Errorf("languages = %v, want empty map", e.languages)
		}
	})

	t.Run("Multiple languages with spaces", func(t *testing.T) {
		e := NewExecutor(10*time.Second, "bash, python , js ,go")

		expectedLangs := []string{"bash", "python", "js", "go"}
		for _, lang := range expectedLangs {
			if !e.languages[lang] {
				t.Errorf("language %q not found in executor", lang)
			}
		}
	})
}

func TestExecutorIsLanguageAllowed(t *testing.T) {
	t.Run("Empty allow list allows all", func(t *testing.T) {
		e := NewExecutor(10*time.Second, "")

		if !e.isLanguageAllowed("bash") {
			t.Error("isLanguageAllowed() with empty list should return true")
		}
		if !e.isLanguageAllowed("python") {
			t.Error("isLanguageAllowed() with empty list should return true")
		}
	})

	t.Run("Specific allow list", func(t *testing.T) {
		e := NewExecutor(10*time.Second, "bash,python")

		if !e.isLanguageAllowed("bash") {
			t.Error("isLanguageAllowed() for bash should return true")
		}
		if !e.isLanguageAllowed("python") {
			t.Error("isLanguageAllowed() for python should return true")
		}
		if e.isLanguageAllowed("javascript") {
			t.Error("isLanguageAllowed() for javascript should return false")
		}
	})
}

func TestExecutorExecute(t *testing.T) {
	ctx := context.Background()

	t.Run("Bash execution", func(t *testing.T) {
		e := NewExecutor(5*time.Second, "bash")

		result, err := e.Execute(ctx, "bash", `echo "Hello, World!"`, nil)
		if err != nil {
			t.Fatalf("Execute() failed: %v", err)
		}

		if result.ExitCode != 0 {
			t.Errorf("ExitCode = %d, want 0", result.ExitCode)
		}
		if result.Output != "Hello, World!\n" {
			t.Errorf("Output = %q, want %q", result.Output, "Hello, World!\n")
		}
	})

	t.Run("Bash with error", func(t *testing.T) {
		e := NewExecutor(5*time.Second, "bash")

		result, err := e.Execute(ctx, "bash", "exit 42", nil)
		if err != nil {
			t.Fatalf("Execute() failed: %v", err)
		}

		if result.ExitCode != 42 {
			t.Errorf("ExitCode = %d, want 42", result.ExitCode)
		}
	})

	t.Run("Python execution", func(t *testing.T) {
		e := NewExecutor(5*time.Second, "python")

		result, err := e.Execute(ctx, "python", "print('Hello from Python')", nil)
		if err != nil {
			t.Fatalf("Execute() failed: %v", err)
		}

		if result.ExitCode != 0 {
			t.Errorf("ExitCode = %d, want 0", result.ExitCode)
		}
		if result.Output != "Hello from Python\n" {
			t.Errorf("Output = %q, want %q", result.Output, "Hello from Python\n")
		}
	})

	t.Run("JavaScript execution", func(t *testing.T) {
		e := NewExecutor(5*time.Second, "js")

		result, err := e.Execute(ctx, "js", "console.log('Hello from JS')", nil)
		if err != nil {
			t.Fatalf("Execute() failed: %v", err)
		}

		if result.ExitCode != 0 {
			t.Errorf("ExitCode = %d, want 0", result.ExitCode)
		}
		// Node.js adds a newline
		if result.Output != "Hello from JS\n" {
			t.Errorf("Output = %q, want %q", result.Output, "Hello from JS\n")
		}
	})

	t.Run("Language not allowed", func(t *testing.T) {
		e := NewExecutor(5*time.Second, "bash")

		_, err := e.Execute(ctx, "python", "print('test')", nil)
		if err == nil {
			t.Error("Execute() with disallowed language should return error")
		}
	})

	t.Run("Unsupported language", func(t *testing.T) {
		e := NewExecutor(5*time.Second, "ruby")

		_, err := e.Execute(ctx, "ruby", "puts 'test'", nil)
		if err == nil {
			t.Error("Execute() with unsupported language should return error")
		}
	})

	t.Run("Timeout", func(t *testing.T) {
		e := NewExecutor(100*time.Millisecond, "bash")

		// This should timeout
		result, err := e.Execute(ctx, "bash", "sleep 10", nil)
		if err != nil {
			// Timeout is expected
			if result.Error == "" {
				t.Errorf("Expected timeout error, got: %v", err)
			}
		}
	})

	t.Run("Sh alias for bash", func(t *testing.T) {
		e := NewExecutor(5*time.Second, "sh")

		result, err := e.Execute(ctx, "sh", `echo "sh test"`, nil)
		if err != nil {
			t.Fatalf("Execute() failed: %v", err)
		}

		if result.ExitCode != 0 {
			t.Errorf("ExitCode = %d, want 0", result.ExitCode)
		}
	})

	t.Run("Py alias for python", func(t *testing.T) {
		e := NewExecutor(5*time.Second, "py")

		result, err := e.Execute(ctx, "py", "print('py test')", nil)
		if err != nil {
			t.Fatalf("Execute() failed: %v", err)
		}

		if result.ExitCode != 0 {
			t.Errorf("ExitCode = %d, want 0", result.ExitCode)
		}
	})
}

func TestRegistry(t *testing.T) {
	t.Run("NewRegistry", func(t *testing.T) {
		r := NewRegistry()

		if r == nil {
			t.Fatal("NewRegistry() returned nil")
		}
		if r.Count() != 0 {
			t.Errorf("Count() = %d, want 0", r.Count())
		}
	})

	t.Run("Add and Get", func(t *testing.T) {
		r := NewRegistry()

		tool := &Tool{
			ID:          "tool-1",
			Name:        "test-tool",
			Description: "A test tool",
			Language:    "bash",
			Code:        "echo test",
		}

		r.Add(tool)

		if r.Count() != 1 {
			t.Errorf("Count() = %d, want 1", r.Count())
		}

		got, ok := r.Get("test-tool")
		if !ok {
			t.Error("Get() returned false for existing tool")
		}
		if got.Name != "test-tool" {
			t.Errorf("Get() returned tool with name %q, want test-tool", got.Name)
		}
	})

	t.Run("Get missing tool", func(t *testing.T) {
		r := NewRegistry()

		_, ok := r.Get("missing-tool")
		if ok {
			t.Error("Get() returned true for missing tool")
		}
	})

	t.Run("List tools", func(t *testing.T) {
		r := NewRegistry()

		r.Add(&Tool{Name: "tool1", Language: "bash", Code: "code1"})
		r.Add(&Tool{Name: "tool2", Language: "python", Code: "code2"})
		r.Add(&Tool{Name: "tool3", Language: "js", Code: "code3"})

		tools := r.List()

		if len(tools) != 3 {
			t.Errorf("List() returned %d tools, want 3", len(tools))
		}
	})

	t.Run("Clear registry", func(t *testing.T) {
		r := NewRegistry()

		r.Add(&Tool{Name: "tool1", Language: "bash", Code: "code1"})
		r.Add(&Tool{Name: "tool2", Language: "python", Code: "code2"})

		r.Clear()

		if r.Count() != 0 {
			t.Errorf("Count() after Clear() = %d, want 0", r.Count())
		}
	})

	t.Run("Overwrite existing tool", func(t *testing.T) {
		r := NewRegistry()

		r.Add(&Tool{
			Name:        "my-tool",
			Description: "Original",
			Language:    "bash",
			Code:        "echo original",
		})

		r.Add(&Tool{
			Name:        "my-tool",
			Description: "Updated",
			Language:    "python",
			Code:        "print('updated')",
		})

		if r.Count() != 1 {
			t.Errorf("Count() = %d, want 1", r.Count())
		}

		tool, _ := r.Get("my-tool")
		if tool.Description != "Updated" {
			t.Errorf("Tool Description = %q, want Updated", tool.Description)
		}
		if tool.Language != "python" {
			t.Errorf("Tool Language = %q, want python", tool.Language)
		}
	})
}

func TestRegistryConcurrency(t *testing.T) {
	r := NewRegistry()

	// Test concurrent operations
	done := make(chan bool)

	// Concurrent writers
	for i := 0; i < 10; i++ {
		go func(idx int) {
			for j := 0; j < 100; j++ {
				r.Add(&Tool{
					Name:     "tool-" + string(rune(idx)),
					Language: "bash",
					Code:     "echo test",
				})
			}
			done <- true
		}(i)
	}

	// Concurrent readers
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				r.List()
				r.Count()
				r.Get("any-tool")
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 15; i++ {
		<-done
	}

	// Should not panic or deadlock
	if r.Count() < 0 {
		t.Error("Invalid count after concurrent operations")
	}
}

func TestToolStruct(t *testing.T) {
	tool := &Tool{
		ID:          "test-id",
		Name:        "Test Tool",
		Description: "A tool for testing",
		Language:    "bash",
		Code:        `echo "Hello, World!"`,
	}

	if tool.ID != "test-id" {
		t.Errorf("ID = %q, want test-id", tool.ID)
	}
	if tool.Name != "Test Tool" {
		t.Errorf("Name = %q, want Test Tool", tool.Name)
	}
	if tool.Description != "A tool for testing" {
		t.Errorf("Description = %q, want 'A tool for testing'", tool.Description)
	}
	if tool.Language != "bash" {
		t.Errorf("Language = %q, want bash", tool.Language)
	}
	if tool.Code != `echo "Hello, World!"` {
		t.Errorf("Code = %q, want 'echo Hello, World!'", tool.Code)
	}
}

// Benchmark tests
func BenchmarkExecutorExecuteBash(b *testing.B) {
	ctx := context.Background()
	e := NewExecutor(5*time.Second, "bash")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Execute(ctx, "bash", `echo "benchmark"`, nil)
	}
}

func BenchmarkRegistryAdd(b *testing.B) {
	r := NewRegistry()
	tool := &Tool{Name: "bench-tool", Language: "bash", Code: "code"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Add(tool)
	}
}

func BenchmarkRegistryGet(b *testing.B) {
	r := NewRegistry()
	r.Add(&Tool{Name: "bench-tool", Language: "bash", Code: "code"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Get("bench-tool")
	}
}

func BenchmarkRegistryList(b *testing.B) {
	r := NewRegistry()
	for i := 0; i < 100; i++ {
		r.Add(&Tool{Name: "tool-" + string(rune(i)), Language: "bash", Code: "code"})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.List()
	}
}
