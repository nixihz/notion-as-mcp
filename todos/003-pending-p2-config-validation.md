---
status: pending
priority: p2
issue_id: "003"
tags: [security, validation, input]
dependencies: []
---

## Problem Statement

**What is broken/missing and why it matters**

Config 结构体缺少输入验证：端口范围未验证、缓存目录路径未验证、TransportType 未验证有效值，可能导致配置错误或安全风险。

## Findings

### 位置
- **文件**: `internal/config/config.go`
- **函数**: `Load()` (第60-175行)

### 证据

1. 端口范围未验证 (第161-167行):
```go
port, err := strconv.Atoi(sp)
// 缺少范围检查 (0-65535)
```

2. 缓存目录路径未验证 (第108-111行):
```go
if cdir := os.Getenv("CACHE_DIR"); cdir != "" {
    cfg.CacheDir = cdir  // 可能导致路径遍历
}
```

3. TransportType 未验证 (第169-172行):
```go
if tt := os.Getenv("TRANSPORT_TYPE"); tt != "" {
    cfg.TransportType = tt  // 缺少有效值验证
}
```

## Proposed Solutions

### Solution 1: 添加完整的输入验证
**Pros:**
- 防止无效配置导致运行时错误
- 提高安全性

**Cons:**
- 增加代码量

**Effort:** Small | **Risk:** Low

## Recommended Action

<!-- Filled during triage -->
[ ] 待定

## Acceptance Criteria

- [ ] 端口范围验证 (0-65535)
- [ ] 缓存目录绝对路径验证
- [ ] TransportType 有效值验证 ("streamable", "stdio")

## Work Log

### 2026-02-14 - 代码审查发现

**By:** Claude Code

**Actions:**
- 执行安全审查，发现输入验证不足

**Learnings:**
- 基础验证应该添加，防止配置错误
