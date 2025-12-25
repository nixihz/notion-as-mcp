// Package cache provides caching functionality for the Notion MCP server.
package cache

import (
	"context"
	"sync"
	"time"
)

// memoryCache implements an in-memory cache using a map with RWMutex.
type memoryCache struct {
	mu       sync.RWMutex
	items    map[string]memoryItem
	stats    Stats
	maxSize  int
	maxBytes int64
}

type memoryItem struct {
	Value     []byte
	ExpiresAt time.Time
}

// NewMemoryCache creates a new in-memory cache.
func NewMemoryCache(opts ...CacheOption) (Cache, error) {
	m := &memoryCache{
		items:   make(map[string]memoryItem),
		stats:   Stats{},
		maxSize: 10000,
	}
	for _, opt := range opts {
		opt(&cacheOptions{})
	}
	return m, nil
}

// Get retrieves a value from the cache.
func (m *memoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	m.mu.RLock()
	item, ok := m.items[key]
	m.mu.RUnlock()

	if !ok {
		m.mu.Lock()
		m.stats.Misses++
		m.mu.Unlock()
		return nil, nil
	}

	// Check expiration
	if time.Now().After(item.ExpiresAt) {
		m.mu.Lock()
		delete(m.items, key)
		m.stats.Misses++
		m.mu.Unlock()
		return nil, nil
	}

	m.mu.Lock()
	m.stats.Hits++
	m.mu.Unlock()

	return item.Value, nil
}

// Set stores a value in the cache.
func (m *memoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Evict old items if at capacity
	if len(m.items) >= m.maxSize {
		m.evictOldest()
	}

	m.items[key] = memoryItem{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
	m.stats.Items = len(m.items)
	m.stats.BytesUsed += int64(len(value))

	return nil
}

// Delete removes a value from the cache.
func (m *memoryCache) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if item, ok := m.items[key]; ok {
		m.stats.BytesUsed -= int64(len(item.Value))
		delete(m.items, key)
		m.stats.Items = len(m.items)
	}

	return nil
}

// Has returns true if the key exists and is not expired.
func (m *memoryCache) Has(ctx context.Context, key string) (bool, error) {
	m.mu.RLock()
	item, ok := m.items[key]
	m.mu.RUnlock()

	if !ok {
		return false, nil
	}

	if time.Now().After(item.ExpiresAt) {
		m.mu.Lock()
		delete(m.items, key)
		m.mu.Unlock()
		return false, nil
	}

	return true, nil
}

// Clear removes all cached values.
func (m *memoryCache) Clear(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.items = make(map[string]memoryItem)
	m.stats = Stats{}
	return nil
}

// Close cleans up resources.
func (m *memoryCache) Close() error {
	return m.Clear(context.Background())
}

// evictOldest removes the oldest item from the cache.
func (m *memoryCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range m.items {
		if oldestKey == "" || item.ExpiresAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.ExpiresAt
		}
	}

	if oldestKey != "" {
		delete(m.items, oldestKey)
	}
}

// Stats returns cache statistics.
func (m *memoryCache) Stats() Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stats
}
