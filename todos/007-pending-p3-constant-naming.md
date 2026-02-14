---
status: pending
priority: p3
issue_id: "007"
tags: [code-quality, naming, convention]
dependencies: []
---

## Problem Statement

**What is broken/missing and why it matters**

Config.go 中常量命名不一致，部分使用缩写 (`Int`, `On`) 而非完整词，影响可读性。

## Findings

### 位置
- **文件**: `internal/config/config.go`

### 证据
| 常量名 | 问题 |
|--------|------|
| `defaultCacheRefreshInt` | 使用缩写 `Int` |
| `defaultPollInt` | 使用缩写 `Int` |
| `defaultRefreshOn` | 使用缩写 `On` |

## Proposed Solutions

### Solution 1: 统一常量命名风格
**Pros:**
- 提高可读性
- 统一风格

**Effort:** Small | **Risk:** Low

## Recommended Action

<!-- Filled during triage -->
[ ] 待定

## Acceptance Criteria

- [ ] 重命名为完整词 (defaultCacheRefreshInterval, defaultPollInterval, defaultRefreshEnabled)

## Work Log

### 2026-02-14 - 代码审查发现

**By:** Claude Code

**Actions:**
- 执行模式识别审查，发现命名不一致

**Learnings:**
- 命名一致性是代码规范的重要部分
