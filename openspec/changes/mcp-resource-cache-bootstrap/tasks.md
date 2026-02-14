## Implementation Tasks

### 1. Extend cache package

- [x] 1.1 Add `Warm(ctx context.Context, key string, fetcher func() ([]byte, error)) error` method to cache interface
- [x] 1.2 Add `StartPeriodicRefresh(ctx context.Context, interval time.Duration, key string, fetcher func() ([]byte, error))` method
- [x] 1.3 Add content hash comparison utility in `internal/cache/`

### 2. Create cache manager for MCP data

- [x] 2.1 Create `internal/cache/mcp_cache.go` for managing MCP resources/prompts cache
- [x] 2.2 Implement cache key constants (e.g., "mcp:resources", "mcp:prompts")
- [x] 2.3 Implement `Warm()` method that fetches from Notion and stores in cache
- [x] 2.4 Implement hash-based change detection

### 3. Modify server initialization

- [x] 3.1 Add cache warm call in `Server.Start()` before registering handlers
- [x] 3.2 Add periodic refresh goroutine with configurable interval (default 5m)
- [x] 3.3 Add graceful shutdown for refresh goroutine via context

### 4. Add configuration

- [x] 4.1 Add `CacheRefreshInterval` to config struct
- [x] 4.2 Add environment variable `CACHE_REFRESH_INTERVAL` (default 5m)
- [x] 4.3 Pass interval to cache manager

### 5. Update cache reading logic

- [x] 5.1 Modify `registerResources()` to read from cache first
- [x] 5.2 Modify `registerPrompts()` to read from cache first
- [x] 5.3 Fall back to Notion API if cache miss

### 6. Testing

- [ ] 6.1 Add unit tests for cache warm functionality
- [ ] 6.2 Add integration test for periodic refresh
- [ ] 6.3 Verify graceful shutdown works correctly

### 7. Documentation

- [ ] 7.1 Update README with new configuration options
- [ ] 7.2 Add logging for cache operations
