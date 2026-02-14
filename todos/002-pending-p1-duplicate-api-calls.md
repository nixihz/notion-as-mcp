---
status: pending
priority: p1
issue_id: "002"
tags: [security, performance, api-call]
dependencies: []
---

## Problem Statement

**What is broken/missing and why it matters**

`warmCache` 和 `startPeriodicRefresh` 函数重复调用 `GetAllPages`，每次服务启动时发起 4 次 API 请求，当 Notion 数据库有大量页面时会显著延长启动时间并可能触发 API 速率限制。

## Findings

### 位置
- **文件**: `internal/server/server.go`
- **函数**: `warmCache` (第119-160行), `startPeriodicRefresh` (第162-195行)

### 证据
```go
// warmCache 中两次独立调用
err := s.mcpCache.Warm(ctx, cache.CacheKeyResources, func(ctx context.Context) ([]byte, error) {
    pages, err := s.client.GetAllPages(ctx)  // 调用1
    // ...
})

err = s.mcpCache.Warm(ctx, cache.CacheKeyPrompts, func(ctx context.Context) ([]byte, error) {
    pages, err := s.client.GetAllPages(ctx)  // 调用2 (重复!)
    // ...
})
```

### 影响
- 启动时 4+ 次 API 调用
- Notion API 限制 3 次/秒，成为瓶颈

## Proposed Solutions

### Solution 1: 统一获取一次页面，并行预热缓存
**Pros:**
- 减少 75% API 调用
- 提升启动性能

**Cons:**
- 需要修改缓存预热逻辑

**Effort:** Medium | **Risk:** Low

### Solution 2: 使用缓存层的批量获取接口
**Pros:**
- 架构更清晰
- 复用性好

**Cons:**
- 需要新增接口

**Effort:** Large | **Risk:** Low

## Recommended Action

<!-- Filled during triage -->
[ ] 待定

## Acceptance Criteria

- [ ] 启动时 API 调用次数从 4 次减少到 1 次
- [ ] 并行预热 resources 和 prompts 缓存

## Work Log

### 2026-02-14 - 代码审查发现

**By:** Claude Code

**Actions:**
- 执行性能审查，发现重复API调用问题

**Learnings:**
- 这是性能瓶颈，应优先优化
