---
status: pending
priority: p2
issue_id: "004"
tags: [code-quality, duplication, dry]
dependencies: []
---

## Problem Statement

**What is broken/missing and why it matters**

Server.go 中存在多处代码重复：页面过滤逻辑重复 (4处)、注册逻辑重复 (prompts vs resources)、辅助函数重复。违反了 DRY 原则，修改时容易遗漏。

## Findings

### 位置
- **文件**: `internal/server/server.go`

### 证据

1. **页面过滤逻辑重复** (约12行重复 x 4处):
   - 第122-136行 (warmCache resources)
   - 第142-156行 (warmCache prompts)
   - 第165-178行 (periodic refresh resources)
   - 第181-194行 (periodic refresh prompts)

```go
for _, p := range pages {
    pageType := notion.GetTypeFromProperties(p.Properties, s.cfg.NotionTypeField)
    if pageType == "resource" {  // 或 "prompt"
        resourcePages = append(resourcePages, p)
    }
}
```

2. **注册逻辑重复** (第268-347行):
   - `registerPrompts` 和 `registerResources` 几乎相同
   - 空名称验证逻辑重复
   - 名称前缀处理逻辑重复

3. **辅助函数重复** (第493-511行):
   - `getPageTitle` 和 `getPageDescription` 结构相同

## Proposed Solutions

### Solution 1: 抽取页面过滤辅助方法
**Pros:**
- 消除重复代码
- 易于维护

**Effort:** Small | **Risk:** Low

```go
func (s *Server) filterPagesByType(pages []notion.Page, pageType string) []notion.Page {
    return lo.Filter(pages, func(p notion.Page, _ int) bool {
        return notion.GetTypeFromProperties(p.Properties, s.cfg.NotionTypeField) == pageType
    })
}
```

### Solution 2: 抽取通用注册方法
**Pros:**
- 大幅减少代码量

**Effort:** Medium | **Risk:** Medium

## Recommended Action

<!-- Filled during triage -->
[ ] 待定

## Acceptance Criteria

- [ ] 抽取 filterPagesByType 辅助方法
- [ ] 消除注册逻辑重复

## Work Log

### 2026-02-14 - 代码审查发现

**By:** Claude Code

**Actions:**
- 执行模式识别审查，发现代码重复问题

**Learnings:**
- DRY 原则是代码可维护性的基础
