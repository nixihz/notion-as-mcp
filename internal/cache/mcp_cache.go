// Package cache provides caching functionality for the Notion MCP server.
package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"sync"
	"time"
)

// Cache key constants for MCP data
const (
	CacheKeyResources = "mcp:resources"
	CacheKeyPrompts   = "mcp:prompts"
)

// Fetcher is a function that fetches data to be cached.
type Fetcher func(ctx context.Context) ([]byte, error)

// MCPCache manages cached MCP resources and prompts.
type MCPCache struct {
	cache     Cache
	logger    *slog.Logger
	mu        sync.RWMutex
	stopChans map[string]chan struct{}
}

// NewMCPCache creates a new MCP cache manager.
func NewMCPCache(cache Cache, logger *slog.Logger) *MCPCache {
	return &MCPCache{
		cache:     cache,
		logger:    logger,
		stopChans: make(map[string]chan struct{}),
	}
}

// Warm fetches data and stores it in cache.
func (m *MCPCache) Warm(ctx context.Context, key string, fetcher Fetcher) error {
	m.logger.Info("warming cache", slog.String("key", key))

	data, err := fetcher(ctx)
	if err != nil {
		m.logger.Warn("failed to warm cache", slog.String("key", key), slog.String("error", err.Error()))
		return err
	}

	// Store with long TTL (1 hour for file cache)
	err = m.cache.Set(ctx, key, data, time.Hour)
	if err != nil {
		m.logger.Warn("failed to set cache", slog.String("key", key), slog.String("error", err.Error()))
		return err
	}

	m.logger.Info("cache warmed successfully", slog.String("key", key), slog.Int("size", len(data)))
	return nil
}

// StartPeriodicRefresh starts a background goroutine that periodically refreshes the cache.
func (m *MCPCache) StartPeriodicRefresh(ctx context.Context, key string, interval time.Duration, fetcher Fetcher) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop existing refresh if any
	if stopChan, ok := m.stopChans[key]; ok {
		close(stopChan)
		delete(m.stopChans, key)
	}

	stopChan := make(chan struct{})
	m.stopChans[key] = stopChan

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				m.logger.Info("stopping periodic refresh", slog.String("key", key))
				return
			case <-stopChan:
				m.logger.Info("periodic refresh stopped", slog.String("key", key))
				return
			case <-ticker.C:
				m.refreshOnce(ctx, key, fetcher)
			}
		}
	}()

	m.logger.Info("periodic refresh started", slog.String("key", key), slog.String("interval", interval.String()))
}

// refreshOnce fetches new data and updates cache only if content changed.
func (m *MCPCache) refreshOnce(ctx context.Context, key string, fetcher Fetcher) {
	m.logger.Debug("refreshing cache", slog.String("key", key))

	newData, err := fetcher(ctx)
	if err != nil {
		m.logger.Warn("failed to refresh cache", slog.String("key", key), slog.String("error", err.Error()))
		return
	}

	// Get existing cached data for comparison
	existingData, err := m.cache.Get(ctx, key)
	if err != nil || existingData == nil {
		// No existing data, just set the new one
		if err := m.cache.Set(ctx, key, newData, time.Hour); err != nil {
			m.logger.Warn("failed to set cache", slog.String("key", key), slog.String("error", err.Error()))
			return
		}
		m.logger.Info("cache updated (was empty)", slog.String("key", key))
		return
	}

	// Compare hashes
	newHash := HashContent(newData)
	existingHash := HashContent(existingData)

	if newHash == existingHash {
		m.logger.Debug("cache unchanged, skipping update", slog.String("key", key))
		return
	}

	// Content changed, update cache
	if err := m.cache.Set(ctx, key, newData, time.Hour); err != nil {
		m.logger.Warn("failed to update cache", slog.String("key", key), slog.String("error", err.Error()))
		return
	}

	m.logger.Info("cache updated", slog.String("key", key))
}

// Get retrieves cached data, returns nil if not found.
func (m *MCPCache) Get(ctx context.Context, key string) ([]byte, error) {
	return m.cache.Get(ctx, key)
}

// RefreshOnce triggers an immediate cache refresh for a given key.
func (m *MCPCache) RefreshOnce(ctx context.Context, key string, fetcher Fetcher) {
	m.refreshOnce(ctx, key, fetcher)
}

// StopPeriodicRefresh stops the periodic refresh for a key.
func (m *MCPCache) StopPeriodicRefresh(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if stopChan, ok := m.stopChans[key]; ok {
		close(stopChan)
		delete(m.stopChans, key)
	}
}

// StopAll stops all periodic refreshes.
func (m *MCPCache) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for key, stopChan := range m.stopChans {
		close(stopChan)
		delete(m.stopChans, key)
	}
}

// HashContent computes SHA256 hash of content.
func HashContent(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
