# Research: Notion MCP Server

**Date**: 2025-12-25
**Branch**: `001-notion-mcp-tools`

## Technology Decisions

### 1. Language and Framework

| Component | Choice | Rationale |
|-----------|--------|-----------|
| Language | Go 1.25+ | 用户指定，现代标准库，优秀的并发支持 |
| MCP SDK | github.com/modelcontextprotocol/go-sdk | 官方 SDK，完整支持 MCP 协议 |
| CLI Framework | github.com/spf13/cobra | Go 生态标准 CLI 库，功能完善 |
| Config | github.com/joho/godotenv | 简单环境变量管理 |
| Logging | 标准库 slog | 结构化日志需求 |

### 2. MCP Server Implementation

**Decision**: 使用官方 go-sdk 创建 MCP 服务器，使用 stdio 传输

```go
// 服务器初始化模式
server := mcp.NewServer(&mcp.ServerOptions{
    Name:       "notion-mcp",
    Version:    "1.0.0",
    Transport:  stdio.NewStdioTransport(),
})

// 注册工具
server.RegisterTool(tool, handler)

// 启动
server.Serve()
```

### 3. Notion API Integration

**Decision**: 直接使用 Notion REST API，不使用第三方库

API 端点：
- 数据库查询: `POST /v1/databases/{id}/query`
- 页面内容: `GET /v1/blocks/{id}/children`
- 认证: Bearer Token (`NOTION_API_KEY`)

### 4. Tool Code Execution

**Decision**: 有限制的代码执行，支持常用脚本语言

| Language | Execution | Rationale |
|----------|-----------|-----------|
| shell/bash | 执行 | 功能强大，但需要参数清理 |
| python | 执行 | 广泛使用，脚本化友好 |
| JavaScript | Node.js | 需要检查是否可用 |

**安全考虑**：
- 仅执行单个代码块
- 可能需要超时限制
- 不支持文件 I/O（除非必要）
- 输出限制

### 5. Caching Strategy

**Decision**: 文件系统缓存 + 内存缓存双层架构

| Layer | TTL | Purpose |
|-------|-----|---------|
| 内存缓存 | 5 分钟 | 快速响应，减少 API 调用 |
| 文件缓存 | 1 小时 | 持久化，重启后仍有效 |

缓存键设计：
- `cache:{database_id}:{page_id}:{updated_time}`

### 6. Rate Limit Handling

**Decision**: 指数退避重试

- 限制：每秒 3 个请求（平均）
- 响应：HTTP 429 + `Retry-After` 头
- 策略：指数退避，最大 3 次重试

### 7. Change Detection Strategy

**Decision**: 定时轮询 + 重启刷新

| Capability | Status | Approach |
|------------|--------|----------|
| Native Webhooks | 不支持 | 无法使用 |
| last_edited_time | 支持 | 用于变更检测 |
| Polling | 支持 | 定时查询数据库 |
| 实时推送 | 不支持 | 使用轮询替代 |

**轮询策略**：

| Mode | Interval | Use Case |
|------|----------|----------|
| 定时刷新 | 60 秒 | 服务器运行时自动检测变更 |
| 重启刷新 | 启动时 | 服务器重启时强制刷新 |

**实现方式**：

```go
// 查询变更的过滤条件
POST /v1/databases/{id}/query
{
  "filter": {
    "property": "last_edited_time",
    "date": {
      "after": "2025-12-25T00:00:00.000Z"  // 上次查询时间
    }
  }
}
```

**配置选项**：

| Variable | Default | Description |
|----------|---------|-------------|
| `POLL_INTERVAL` | 60s | 轮询间隔（0 表示禁用） |
| `REFRESH_ON_START` | true | 启动时强制刷新 |

**注意**：由于 Notion API 速率限制，建议：
- 轮询间隔不小于 30 秒
- 启用缓存减少 API 调用
- 使用 `last_edited_time` 过滤减少数据量

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    MCP Client (stdio)                    │
└─────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────┐
│                  Notion MCP Server                       │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────────┐ │
│  │ Prompts │  │Resources│  │  Tools  │  │  CLI/Cobra  │ │
│  └─────────┘  └─────────┘  └─────────┘  └─────────────┘ │
│         │         │          │               │          │
│         └─────────┴──────────┴───────────────┘          │
│                           │                              │
│              ┌────────────▼────────────┐                 │
│              │   Cache Layer           │                 │
│              │  ┌─────────┬─────────┐  │                 │
│              │  │ In-Memory│ File   │  │                 │
│              │  └─────────┴─────────┘  │                 │
│              └─────────────────────────┘                 │
│                           │                              │
│              ┌────────────▼────────────┐                 │
│              │   Notion Client         │                 │
│              │  (REST API + Retry)     │                 │
│              └─────────────────────────┘                 │
│                           │                              │
│              ┌────────────▼────────────┐                 │
│              │   Change Detector       │                 │
│              │  (last_edited_time)     │                 │
│              │   Polling Worker        │                 │
│              └─────────────────────────┘                 │
└─────────────────────────────────────────────────────────┘
```

## Alternatives Considered

### 1. MCP SDK Alternatives

| Option | Rejected Because |
|--------|------------------|
| Python SDK | 用户指定 Go |
| Node.js SDK | 用户指定 Go |
| Custom implementation | 官方 SDK 更可靠 |

### 2. Notion API Client

| Option | Rejected Because |
|--------|------------------|
| notion-go | 第三方库，依赖复杂 |
| direct REST | 更简单，无额外依赖 |
| official SDK | 不存在官方 Go SDK |

### 3. Code Execution

| Option | Rejected Because |
|--------|------------------|
| Docker sandbox | 过于复杂，增加部署难度 |
| WebAssembly | 学习曲线陡峭 |
| Direct execution | 简单直接，用户可控 |

## Key References

- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)
- [MCP Specification](https://modelcontextprotocol.io/specification)
- [Notion API Documentation](https://developers.notion.com/)
