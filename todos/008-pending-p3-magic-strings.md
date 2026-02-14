---
status: pending
priority: p3
issue_id: "008"
tags: [code-quality, magic-string, maintainability]
dependencies: []
---

## Problem Statement

**What is broken/missing and why it matters**

代码中多处使用硬编码的魔法字符串，如 `"resource"`, `"prompt"`, `"tool"`, `"p_"`, `"r_"` 等，应该提取为常量以提高可维护性。

## Findings

### 位置
- **文件**: `internal/server/server.go`

### 证据
```go
if pageType == "resource" {  // 多处重复
if pageType == "prompt" {   // 多处重复
if pageType == "tool" {     // 多处重复
promptName = "p_" + promptName
resourceName = "r_" + resourceName
```

## Proposed Solutions

### Solution 1: 提取为包级常量
**Pros:**
- 提高可读性和可维护性
- 便于统一修改

**Effort:** Small | **Risk:** Low

```go
const (
    PageTypeResource = "resource"
    PageTypePrompt   = "prompt"
    PageTypeTool     = "tool"
)
```

## Recommended Action

<!-- Filled during triage -->
[ ] 待定

## Acceptance Criteria

- [ ] 提取 page type 为常量
- [ ] 提取 name prefix 为常量

## Work Log

### 2026-02-14 - 代码审查发现

**By:** Claude Code

**Actions:**
- 执行模式识别审查，发现魔法字符串问题

**Learnings:**
- 魔法字符串应该尽量避免
