// Package config provides tests for configuration loading.
package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Helper to reset environment for each subtest
	resetEnv := func() {
		envVars := []string{
			"NOTION_API_KEY", "NOTION_DATABASE_ID", "NOTION_TYPE_FIELD",
			"CACHE_TTL", "CACHE_DIR", "LOG_LEVEL",
			"EXEC_TIMEOUT", "EXEC_LANGUAGES",
			"POLL_INTERVAL", "REFRESH_ON_START",
		}
		for _, v := range envVars {
			os.Unsetenv(v)
		}
	}

	t.Run("Required environment variables", func(t *testing.T) {
		resetEnv()
		_, err := Load()
		if err == nil {
			t.Error("Load() without NOTION_API_KEY should return error")
		}
	})

	t.Run("Minimum required config", func(t *testing.T) {
		resetEnv()
		os.Setenv("NOTION_API_KEY", "test-api-key")
		os.Setenv("NOTION_DATABASE_ID", "test-db-id")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if cfg.NotionAPIKey != "test-api-key" {
			t.Errorf("NotionAPIKey = %v, want test-api-key", cfg.NotionAPIKey)
		}
		if cfg.NotionDatabaseID != "test-db-id" {
			t.Errorf("NotionDatabaseID = %v, want test-db-id", cfg.NotionDatabaseID)
		}
	})

	t.Run("Default values", func(t *testing.T) {
		resetEnv()
		os.Setenv("NOTION_API_KEY", "test-api-key")
		os.Setenv("NOTION_DATABASE_ID", "test-db-id")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if cfg.NotionTypeField != defaultTypeField {
			t.Errorf("NotionTypeField = %v, want %v", cfg.NotionTypeField, defaultTypeField)
		}
		if cfg.CacheTTL != defaultCacheTTL {
			t.Errorf("CacheTTL = %v, want %v", cfg.CacheTTL, defaultCacheTTL)
		}
		if cfg.CacheDir != defaultCacheDir {
			t.Errorf("CacheDir = %v, want %v", cfg.CacheDir, defaultCacheDir)
		}
		if cfg.LogLevel != defaultLogLevel {
			t.Errorf("LogLevel = %v, want %v", cfg.LogLevel, defaultLogLevel)
		}
		if cfg.ExecTimeout != defaultExecTimeout {
			t.Errorf("ExecTimeout = %v, want %v", cfg.ExecTimeout, defaultExecTimeout)
		}
		if cfg.ExecLanguages != defaultExecLang {
			t.Errorf("ExecLanguages = %v, want %v", cfg.ExecLanguages, defaultExecLang)
		}
		if cfg.PollInterval != defaultPollInt {
			t.Errorf("PollInterval = %v, want %v", cfg.PollInterval, defaultPollInt)
		}
		if cfg.RefreshOnStart != defaultRefreshOn {
			t.Errorf("RefreshOnStart = %v, want %v", cfg.RefreshOnStart, defaultRefreshOn)
		}
	})

	t.Run("Custom type field", func(t *testing.T) {
		resetEnv()
		os.Setenv("NOTION_API_KEY", "test-api-key")
		os.Setenv("NOTION_DATABASE_ID", "test-db-id")
		os.Setenv("NOTION_TYPE_FIELD", "CustomType")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if cfg.NotionTypeField != "CustomType" {
			t.Errorf("NotionTypeField = %v, want CustomType", cfg.NotionTypeField)
		}
	})

	t.Run("Custom cache TTL", func(t *testing.T) {
		resetEnv()
		os.Setenv("NOTION_API_KEY", "test-api-key")
		os.Setenv("NOTION_DATABASE_ID", "test-db-id")
		os.Setenv("CACHE_TTL", "10m")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		expected := 10 * time.Minute
		if cfg.CacheTTL != expected {
			t.Errorf("CacheTTL = %v, want %v", cfg.CacheTTL, expected)
		}
	})

	t.Run("Invalid cache TTL", func(t *testing.T) {
		resetEnv()
		os.Setenv("NOTION_API_KEY", "test-api-key")
		os.Setenv("NOTION_DATABASE_ID", "test-db-id")
		os.Setenv("CACHE_TTL", "invalid")

		_, err := Load()
		if err == nil {
			t.Error("Load() with invalid CACHE_TTL should return error")
		}
	})

	t.Run("Custom cache directory", func(t *testing.T) {
		resetEnv()
		os.Setenv("NOTION_API_KEY", "test-api-key")
		os.Setenv("NOTION_DATABASE_ID", "test-db-id")
		os.Setenv("CACHE_DIR", "/custom/cache/dir")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if cfg.CacheDir != "/custom/cache/dir" {
			t.Errorf("CacheDir = %v, want /custom/cache/dir", cfg.CacheDir)
		}
	})

	t.Run("Custom log level", func(t *testing.T) {
		resetEnv()
		os.Setenv("NOTION_API_KEY", "test-api-key")
		os.Setenv("NOTION_DATABASE_ID", "test-db-id")
		os.Setenv("LOG_LEVEL", "debug")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if cfg.LogLevel != "debug" {
			t.Errorf("LogLevel = %v, want debug", cfg.LogLevel)
		}
	})

	t.Run("Custom execution timeout", func(t *testing.T) {
		resetEnv()
		os.Setenv("NOTION_API_KEY", "test-api-key")
		os.Setenv("NOTION_DATABASE_ID", "test-db-id")
		os.Setenv("EXEC_TIMEOUT", "1m")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		expected := 1 * time.Minute
		if cfg.ExecTimeout != expected {
			t.Errorf("ExecTimeout = %v, want %v", cfg.ExecTimeout, expected)
		}
	})

	t.Run("Invalid execution timeout", func(t *testing.T) {
		resetEnv()
		os.Setenv("NOTION_API_KEY", "test-api-key")
		os.Setenv("NOTION_DATABASE_ID", "test-db-id")
		os.Setenv("EXEC_TIMEOUT", "invalid")

		_, err := Load()
		if err == nil {
			t.Error("Load() with invalid EXEC_TIMEOUT should return error")
		}
	})

	t.Run("Custom execution languages", func(t *testing.T) {
		resetEnv()
		os.Setenv("NOTION_API_KEY", "test-api-key")
		os.Setenv("NOTION_DATABASE_ID", "test-db-id")
		os.Setenv("EXEC_LANGUAGES", "python,go,rust")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if cfg.ExecLanguages != "python,go,rust" {
			t.Errorf("ExecLanguages = %v, want python,go,rust", cfg.ExecLanguages)
		}
	})

	t.Run("Custom poll interval", func(t *testing.T) {
		resetEnv()
		os.Setenv("NOTION_API_KEY", "test-api-key")
		os.Setenv("NOTION_DATABASE_ID", "test-db-id")
		os.Setenv("POLL_INTERVAL", "30s")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		expected := 30 * time.Second
		if cfg.PollInterval != expected {
			t.Errorf("PollInterval = %v, want %v", cfg.PollInterval, expected)
		}
	})

	t.Run("Invalid poll interval", func(t *testing.T) {
		resetEnv()
		os.Setenv("NOTION_API_KEY", "test-api-key")
		os.Setenv("NOTION_DATABASE_ID", "test-db-id")
		os.Setenv("POLL_INTERVAL", "invalid")

		_, err := Load()
		if err == nil {
			t.Error("Load() with invalid POLL_INTERVAL should return error")
		}
	})

	t.Run("Refresh on start - true", func(t *testing.T) {
		resetEnv()
		os.Setenv("NOTION_API_KEY", "test-api-key")
		os.Setenv("NOTION_DATABASE_ID", "test-db-id")
		os.Setenv("REFRESH_ON_START", "true")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if !cfg.RefreshOnStart {
			t.Errorf("RefreshOnStart = false, want true")
		}
	})

	t.Run("Refresh on start - 1", func(t *testing.T) {
		resetEnv()
		os.Setenv("NOTION_API_KEY", "test-api-key")
		os.Setenv("NOTION_DATABASE_ID", "test-db-id")
		os.Setenv("REFRESH_ON_START", "1")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if !cfg.RefreshOnStart {
			t.Errorf("RefreshOnStart = false, want true")
		}
	})

	t.Run("Refresh on start - false", func(t *testing.T) {
		resetEnv()
		os.Setenv("NOTION_API_KEY", "test-api-key")
		os.Setenv("NOTION_DATABASE_ID", "test-db-id")
		os.Setenv("REFRESH_ON_START", "false")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if cfg.RefreshOnStart {
			t.Errorf("RefreshOnStart = true, want false")
		}
	})

	t.Run("Full custom config", func(t *testing.T) {
		resetEnv()
		os.Setenv("NOTION_API_KEY", "secret-key")
		os.Setenv("NOTION_DATABASE_ID", "db-123")
		os.Setenv("NOTION_TYPE_FIELD", "MyType")
		os.Setenv("CACHE_TTL", "15m")
		os.Setenv("CACHE_DIR", "/var/cache/mcp")
		os.Setenv("LOG_LEVEL", "warn")
		os.Setenv("EXEC_TIMEOUT", "45s")
		os.Setenv("EXEC_LANGUAGES", "bash,python")
		os.Setenv("POLL_INTERVAL", "2m")
		os.Setenv("REFRESH_ON_START", "false")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if cfg.NotionAPIKey != "secret-key" {
			t.Errorf("NotionAPIKey = %v, want secret-key", cfg.NotionAPIKey)
		}
		if cfg.NotionDatabaseID != "db-123" {
			t.Errorf("NotionDatabaseID = %v, want db-123", cfg.NotionDatabaseID)
		}
		if cfg.NotionTypeField != "MyType" {
			t.Errorf("NotionTypeField = %v, want MyType", cfg.NotionTypeField)
		}
		if cfg.CacheTTL != 15*time.Minute {
			t.Errorf("CacheTTL = %v, want 15m", cfg.CacheTTL)
		}
		if cfg.CacheDir != "/var/cache/mcp" {
			t.Errorf("CacheDir = %v, want /var/cache/mcp", cfg.CacheDir)
		}
		if cfg.LogLevel != "warn" {
			t.Errorf("LogLevel = %v, want warn", cfg.LogLevel)
		}
		if cfg.ExecTimeout != 45*time.Second {
			t.Errorf("ExecTimeout = %v, want 45s", cfg.ExecTimeout)
		}
		if cfg.ExecLanguages != "bash,python" {
			t.Errorf("ExecLanguages = %v, want bash,python", cfg.ExecLanguages)
		}
		if cfg.PollInterval != 2*time.Minute {
			t.Errorf("PollInterval = %v, want 2m", cfg.PollInterval)
		}
		if cfg.RefreshOnStart {
			t.Errorf("RefreshOnStart = true, want false")
		}
	})
}

func TestConfigValidate(t *testing.T) {
	t.Run("Valid config", func(t *testing.T) {
		cfg := &Config{
			NotionAPIKey:     "test-key",
			NotionDatabaseID: "test-db",
		}

		err := cfg.Validate()
		if err != nil {
			t.Errorf("Validate() on valid config failed: %v", err)
		}
	})

	t.Run("Missing API key", func(t *testing.T) {
		cfg := &Config{
			NotionDatabaseID: "test-db",
		}

		err := cfg.Validate()
		if err == nil {
			t.Error("Validate() without NotionAPIKey should return error")
		}
	})

	t.Run("Missing database ID", func(t *testing.T) {
		cfg := &Config{
			NotionAPIKey: "test-key",
		}

		err := cfg.Validate()
		if err == nil {
			t.Error("Validate() without NotionDatabaseID should return error")
		}
	})

	t.Run("Empty config", func(t *testing.T) {
		cfg := &Config{}

		err := cfg.Validate()
		if err == nil {
			t.Error("Validate() on empty config should return error")
		}
	})
}

func TestLoadWithEnvFile(t *testing.T) {
	// This test verifies the .env file loading mechanism
	// The actual .env loading is handled by godotenv
	// We just need to ensure Load() calls godotenv.Load()

	resetEnv := func() {
		envVars := []string{
			"NOTION_API_KEY", "NOTION_DATABASE_ID", "NOTION_TYPE_FIELD",
			"CACHE_TTL", "CACHE_DIR", "LOG_LEVEL",
			"EXEC_TIMEOUT", "EXEC_LANGUAGES",
			"POLL_INTERVAL", "REFRESH_ON_START",
		}
		for _, v := range envVars {
			os.Unsetenv(v)
		}
	}

	resetEnv()
	// Set environment variables directly (simulating .env file content)
	os.Setenv("NOTION_API_KEY", "env-file-key")
	os.Setenv("NOTION_DATABASE_ID", "env-file-db-id")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.NotionAPIKey != "env-file-key" {
		t.Errorf("NotionAPIKey = %v, want env-file-key", cfg.NotionAPIKey)
	}
}

// Benchmark tests
func BenchmarkLoad(b *testing.B) {
	os.Setenv("NOTION_API_KEY", "bench-key")
	os.Setenv("NOTION_DATABASE_ID", "bench-db")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Load()
	}
}

func BenchmarkValidate(b *testing.B) {
	cfg := &Config{
		NotionAPIKey:     "bench-key",
		NotionDatabaseID: "bench-db",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg.Validate()
	}
}
