---
status: pending
priority: p2
issue_id: "005"
tags: [performance, memory, optimization]
dependencies: []
---

## Problem Statement

**What is broken/missing and why it matters**

`createToolHandler` 在服务启动时预获取所有工具的页面内容并保存在内存闭包中，如果有大量工具，会导致高内存占用且启动变慢。

## Findings

### 位置
- **文件**: `internal/server/server.go`
- **函数**: `createToolHandler` (第432-481行)

### 证据
```go
func (s *Server) createToolHandler(page notion.Page) mcp.ToolHandler {
    // 启动时就获取内容，存储在闭包中
    content, err := s.client.GetPageContent(context.Background(), page.ID)
    // ...
    return func(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        // 使用预获取的 content
    }
}
```

### 影响
- 50个工具 = 50次API调用 + 内存占用

## Proposed Solutions

### Solution 1: 延迟获取内容到实际调用时
**Pros:**
- 按需获取，减少内存占用
- 加快启动速度

**Cons:**
- 首次调用有延迟

**Effort:** Small | **Risk:** Low

## Recommended Action

<!-- Filled during triage -->
[ ] 待定

## Acceptance Criteria

- [ ] 移除启动时的预获取逻辑
- [ ] 改为在 handler 调用时获取内容

## Work Log

### 2026-02-14 - 代码审查发现

**By:** Claude Code

**Actions:**
- 执行性能审查，发现预获取导致的内存问题

**Learnings:**
- 延迟加载是优化内存占用的好方法
