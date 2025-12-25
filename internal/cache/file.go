// Package cache provides caching functionality for the Notion MCP server.
package cache

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// fileCache implements a file-based cache.
type fileCache struct {
	dir       string
	defaultTTL time.Duration
}

// NewFileCache creates a new file-based cache.
func NewFileCache(opts ...CacheOption) (Cache, error) {
	fc := &fileCache{
		defaultTTL: 1 * time.Hour,
	}
	for _, opt := range opts {
		opt(&cacheOptions{})
	}

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(fc.dir, 0755); err != nil {
		return nil, err
	}

	return fc, nil
}

// Get retrieves a value from the cache.
func (fc *fileCache) Get(ctx context.Context, key string) ([]byte, error) {
	path := fc.cachePath(key)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var item fileCacheItem
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, err
	}

	// Check expiration
	if time.Now().After(item.ExpiresAt) {
		os.Remove(path)
		return nil, nil
	}

	return item.Value, nil
}

// Set stores a value in the cache.
func (fc *fileCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	path := fc.cachePath(key)

	item := fileCacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}

	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Delete removes a value from the cache.
func (fc *fileCache) Delete(ctx context.Context, key string) error {
	path := fc.cachePath(key)
	os.Remove(path)
	return nil
}

// Has returns true if the key exists and is not expired.
func (fc *fileCache) Has(ctx context.Context, key string) (bool, error) {
	path := fc.cachePath(key)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	var item fileCacheItem
	if err := json.Unmarshal(data, &item); err != nil {
		return false, err
	}

	if time.Now().After(item.ExpiresAt) {
		os.Remove(path)
		return false, nil
	}

	return true, nil
}

// Clear removes all cached values.
func (fc *fileCache) Clear(ctx context.Context) error {
	entries, err := os.ReadDir(fc.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			os.Remove(filepath.Join(fc.dir, entry.Name()))
		}
	}

	return nil
}

// Close cleans up resources.
func (fc *fileCache) Close() error {
	return nil
}

// cachePath generates the file path for a cache key.
func (fc *fileCache) cachePath(key string) string {
	// Sanitize key for file system
	safeKey := filepath.Base(key)
	return filepath.Join(fc.dir, safeKey+".cache")
}

// fileCacheItem represents a cached item.
type fileCacheItem struct {
	Value     []byte    `json:"value"`
	ExpiresAt time.Time `json:"expires_at"`
}
