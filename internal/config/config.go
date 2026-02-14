// Package config provides configuration loading for the Notion MCP server.
//
// It supports environment variables and .env file configuration.
package config

import (
	"fmt"
	"os"
	"strconv"
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
	CacheTTL             time.Duration `json:"cache_ttl"`
	CacheDir             string        `json:"cache_dir"`
	CacheRefreshInterval time.Duration `json:"cache_refresh_interval"`

	// Logging configuration
	LogLevel string `json:"log_level"`

	// Execution configuration
	ExecTimeout   time.Duration `json:"exec_timeout"`
	ExecLanguages string        `json:"exec_languages"`

	// Change detection configuration
	PollInterval   time.Duration `json:"poll_interval"`
	RefreshOnStart bool          `json:"refresh_on_start"`

	// Server configuration
	ServerHost    string `json:"server_host"`
	ServerPort    int    `json:"server_port"`
	TransportType string `json:"transport_type"`
}

// Default values.
const (
	defaultTypeField       = "Type"
	defaultCacheTTL        = 5 * time.Minute
	defaultCacheDir        = "~/.cache/notion-as-mcp"
	defaultCacheRefreshInt = 5 * time.Minute
	defaultLogLevel        = "info"
	defaultExecTimeout     = 30 * time.Second
	defaultExecLang        = "bash,python,js,javascript,ts,typescript"
	defaultPollInt         = 60 * time.Second
	defaultRefreshOn       = true
	defaultServerHost      = "0.0.0.0"
	defaultServerPort      = 3100
	defaultTransport       = "streamable"
)

// Load loads configuration from environment variables and .env file.
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	cfg := &Config{
		NotionTypeField:      defaultTypeField,
		CacheTTL:             defaultCacheTTL,
		CacheDir:             defaultCacheDir,
		CacheRefreshInterval: defaultCacheRefreshInt,
		LogLevel:             defaultLogLevel,
		ExecTimeout:          defaultExecTimeout,
		ExecLanguages:        defaultExecLang,
		PollInterval:         defaultPollInt,
		RefreshOnStart:       defaultRefreshOn,
		ServerHost:           defaultServerHost,
		ServerPort:           defaultServerPort,
		TransportType:        defaultTransport,
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

	// Optional: Cache refresh interval
	if cri := os.Getenv("CACHE_REFRESH_INTERVAL"); cri != "" {
		interval, err := time.ParseDuration(cri)
		if err != nil {
			return nil, fmt.Errorf("invalid CACHE_REFRESH_INTERVAL: %w", err)
		}
		cfg.CacheRefreshInterval = interval
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

	// Optional: Server host
	if sh := os.Getenv("SERVER_HOST"); sh != "" {
		cfg.ServerHost = sh
	}

	// Optional: Server port
	if sp := os.Getenv("SERVER_PORT"); sp != "" {
		port, err := strconv.Atoi(sp)
		if err != nil {
			return nil, fmt.Errorf("invalid SERVER_PORT: %w", err)
		}
		cfg.ServerPort = port
	}

	// Optional: Transport type
	if tt := os.Getenv("TRANSPORT_TYPE"); tt != "" {
		cfg.TransportType = tt
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
