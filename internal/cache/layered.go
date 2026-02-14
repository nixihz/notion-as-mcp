// Package cache provides caching functionality for the Notion MCP server.
package cache

import (
	"context"
	"time"
)

// layeredCache implements a two-layer cache (L1: memory, L2: file).
type layeredCache struct {
	l1 Cache // memory cache
	l2 Cache // file cache
}

// NewLayeredCache creates a new layered cache.
func NewLayeredCache(l1, l2 Cache) Cache {
	return &layeredCache{
		l1: l1,
		l2: l2,
	}
}

// Get retrieves a value from the cache, checking L1 first, then L2.
func (lc *layeredCache) Get(ctx context.Context, key string) ([]byte, error) {
	// Try L1 first
	value, err := lc.l1.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if value != nil {
		return value, nil
	}

	// Try L2
	value, err = lc.l2.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if value != nil {
		// Populate L1 for next time
		lc.l1.Set(ctx, key, value, 5*time.Minute)
	}

	return value, nil
}

// Set stores a value in both L1 and L2.
func (lc *layeredCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	// Set in both layers
	if err := lc.l1.Set(ctx, key, value, ttl); err != nil {
		return err
	}
	if err := lc.l2.Set(ctx, key, value, ttl); err != nil {
		// Log warning but don't fail - L2 is optional
		_ = err
	}
	return nil
}

// Delete removes a value from both layers.
func (lc *layeredCache) Delete(ctx context.Context, key string) error {
	lc.l1.Delete(ctx, key)
	lc.l2.Delete(ctx, key)
	return nil
}

// Has returns true if the key exists in either layer.
func (lc *layeredCache) Has(ctx context.Context, key string) (bool, error) {
	has, _ := lc.l1.Has(ctx, key)
	if has {
		return true, nil
	}
	return lc.l2.Has(ctx, key)
}

// Clear removes all values from both layers.
func (lc *layeredCache) Clear(ctx context.Context) error {
	lc.l1.Clear(ctx)
	lc.l2.Clear(ctx)
	return nil
}

// Close cleans up resources for both layers.
func (lc *layeredCache) Close() error {
	lc.l1.Close()
	lc.l2.Close()
	return nil
}
