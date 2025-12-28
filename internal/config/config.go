// Package config provides configuration loading for the Notion MCP server.
//
// It supports environment variables and .env file configuration.
package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the Notion MCP server.
type Config struct {
	// Notion API configuration
	NotionAPIKey     string `json:"notion_api_key"`
	NotionDatabaseID string `json:"notion_database_id"`
	NotionTypeField  string `json:"notion_type_field"`

	// Cache configuration
	CacheTTL time.Duration `json:"cache_ttl"`
	CacheDir string        `json:"cache_dir"`

	// Logging configuration
	LogLevel string `json:"log_level"`

	// Execution configuration
	ExecTimeout   time.Duration `json:"exec_timeout"`
	ExecLanguages string        `json:"exec_languages"`

	// Change detection configuration
	PollInterval   time.Duration `json:"poll_interval"`
	RefreshOnStart bool          `json:"refresh_on_start"`
}

// Default values.
const (
	defaultTypeField   = "Type"
	defaultCacheTTL    = 5 * time.Minute
	defaultCacheDir    = "~/.cache/notion-mcp"
	defaultLogLevel    = "info"
	defaultExecTimeout = 30 * time.Second
	defaultExecLang    = "bash,python,js,javascript,ts,typescript"
	defaultPollInt     = 60 * time.Second
	defaultRefreshOn   = true
)

// Load loads configuration from environment variables and .env file.
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	cfg := &Config{
		NotionTypeField: defaultTypeField,
		CacheTTL:        defaultCacheTTL,
		CacheDir:        defaultCacheDir,
		LogLevel:        defaultLogLevel,
		ExecTimeout:     defaultExecTimeout,
		ExecLanguages:   defaultExecLang,
		PollInterval:    defaultPollInt,
		RefreshOnStart:  defaultRefreshOn,
	}

	// Required: Notion API Key
	if key := os.Getenv("NOTION_API_KEY"); key != "" {
		cfg.NotionAPIKey = key
	} else {
		return nil, fmt.Errorf("NOTION_API_KEY is required")
	}

	// Required: Notion Database ID
	if dbID := os.Getenv("NOTION_DATABASE_ID"); dbID != "" {
		cfg.NotionDatabaseID = dbID
	} else {
		return nil, fmt.Errorf("NOTION_DATABASE_ID is required")
	}

	// Optional: Type field name
	if tf := os.Getenv("NOTION_TYPE_FIELD"); tf != "" {
		cfg.NotionTypeField = tf
	}

	// Optional: Cache TTL
	if cttl := os.Getenv("CACHE_TTL"); cttl != "" {
		ttl, err := time.ParseDuration(cttl)
		if err != nil {
			return nil, fmt.Errorf("invalid CACHE_TTL: %w", err)
		}
		cfg.CacheTTL = ttl
	}

	// Optional: Cache directory
	if cdir := os.Getenv("CACHE_DIR"); cdir != "" {
		cfg.CacheDir = cdir
	}

	// Optional: Log level
	if ll := os.Getenv("LOG_LEVEL"); ll != "" {
		cfg.LogLevel = ll
	}

	// Optional: Execution timeout
	if et := os.Getenv("EXEC_TIMEOUT"); et != "" {
		timeout, err := time.ParseDuration(et)
		if err != nil {
			return nil, fmt.Errorf("invalid EXEC_TIMEOUT: %w", err)
		}
		cfg.ExecTimeout = timeout
	}

	// Optional: Execution languages
	if el := os.Getenv("EXEC_LANGUAGES"); el != "" {
		cfg.ExecLanguages = el
	}

	// Optional: Poll interval
	if pi := os.Getenv("POLL_INTERVAL"); pi != "" {
		interval, err := time.ParseDuration(pi)
		if err != nil {
			return nil, fmt.Errorf("invalid POLL_INTERVAL: %w", err)
		}
		cfg.PollInterval = interval
	}

	// Optional: Refresh on start
	if ros := os.Getenv("REFRESH_ON_START"); ros != "" {
		cfg.RefreshOnStart = ros == "true" || ros == "1"
	}

	return cfg, nil
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.NotionAPIKey == "" {
		return fmt.Errorf("NOTION_API_KEY is required")
	}
	if c.NotionDatabaseID == "" {
		return fmt.Errorf("NOTION_DATABASE_ID is required")
	}
	return nil
}
