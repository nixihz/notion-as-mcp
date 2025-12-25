// Package cache provides caching functionality for the Notion MCP server.
package cache

import (
	"context"
	"time"
)

// Cache defines the interface for caching operations.
type Cache interface {
	// Get retrieves a value by key.
	Get(ctx context.Context, key string) ([]byte, error)
	// Set stores a value with the given TTL.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	// Delete removes a value by key.
	Delete(ctx context.Context, key string) error
	// Has returns true if the key exists.
	Has(ctx context.Context, key string) (bool, error)
	// Clear removes all cached values.
	Clear(ctx context.Context) error
	// Close cleans up resources.
	Close() error
}

// Stats holds cache statistics.
type Stats struct {
	Hits      int64 `json:"hits"`
	Misses    int64 `json:"misses"`
	Items     int   `json:"items"`
	BytesUsed int64 `json:"bytes_used"`
}

// CacheOption configures a cache.
type CacheOption func(*cacheOptions)

// WithTTL sets the default TTL for cache entries.
func WithTTL(ttl time.Duration) CacheOption {
	return func(o *cacheOptions) {
		o.DefaultTTL = ttl
	}
}

// WithDir sets the cache directory for file cache.
func WithDir(dir string) CacheOption {
	return func(o *cacheOptions) {
		o.Directory = dir
	}
}

type cacheOptions struct {
	DefaultTTL time.Duration
	Directory  string
}

// NewCache creates a new cache instance based on configuration.
// It creates a layered cache with memory cache as L1 and file cache as L2.
func NewCache(opts ...CacheOption) (Cache, error) {
	o := &cacheOptions{
		DefaultTTL: 5 * time.Minute,
		Directory:  "~/.cache/notion-mcp",
	}
	for _, opt := range opts {
		opt(o)
	}

	memoryCache, err := NewMemoryCache(WithTTL(o.DefaultTTL))
	if err != nil {
		return nil, err
	}

	fileCache, err := NewFileCache(WithDir(o.Directory), WithTTL(1*time.Hour))
	if err != nil {
		// If file cache fails, just use memory cache
		return memoryCache, nil
	}

	return NewLayeredCache(memoryCache, fileCache), nil
}
