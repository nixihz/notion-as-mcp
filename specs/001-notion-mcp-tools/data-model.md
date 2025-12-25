# Data Model: Notion MCP Server

**Branch**: `001-notion-mcp-tools`
**Date**: 2025-12-25

## Configuration

### Config

环境变量配置模型。

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `NOTION_API_KEY` | string | Yes | Notion Integration Token |
| `NOTION_DATABASE_ID` | string | Yes | Notion Database ID |
| `NOTION_TYPE_FIELD` | string | No | 类型字段名称，默认 "Type" |
| `CACHE_TTL` | duration | No | 缓存 TTL，默认 5m |
| `CACHE_DIR` | string | No | 缓存目录，默认 ~/.cache/notion-mcp |
| `LOG_LEVEL` | string | No | 日志级别，默认 info |
| `EXEC_TIMEOUT` | duration | No | 代码执行超时，默认 30s |
| `EXEC_LANGUAGES` | string | No | 允许执行的语言，逗号分隔 |
| `POLL_INTERVAL` | duration | No | 轮询间隔，默认 60s（0 禁用） |
| `REFRESH_ON_START` | bool | No | 启动时刷新，默认 true |

## Notion Domain

### Page

Notion 页面/条目。

| Field | Type | Description |
|-------|------|-------------|
| ID | string | 页面 UUID |
| CreatedTime | time.Time | 创建时间 |
| LastEditedTime | time.Time | 最后编辑时间 |
| Properties | map[string]Property | 属性映射 |
| Content | []Block | 内容块列表 |

### Property

Notion 属性，支持多种类型。

| Field | Type | Description |
|-------|------|-------------|
| Name | string | 属性名称 |
| Type | PropertyType | 属性类型 |
| Value | any | 属性值 |

**PropertyType**: `title`, `rich_text`, `select`, `multi_select`, `status`, `checkbox`, `date`, `url`, `email`, `number`

### Block

Notion 内容块。

| Field | Type | Description |
|-------|------|-------------|
| ID | string | 块 UUID |
| Type | BlockType | 块类型 |
| Content | any | 块内容 |

**BlockType**: `paragraph`, `heading_1`, `heading_2`, `heading_3`, `bulleted_list_item`, `numbered_list_item`, `code`, `quote`, `divider`, `callout`, `image`

### CodeBlock

代码块内容。

| Field | Type | Description |
|-------|------|-------------|
| Language | string | 编程语言 |
| Caption | []RichText | 标题 |
| Code | []RichText | 代码内容 |

### RichText

富文本。

| Field | Type | Description |
|-------|------|-------------|
| Type | string | "text" |
| Content | string | 文本内容 |
| PlainText | string | 纯文本 |
| Link | *Link | 链接 |
| Annotations | Annotations | 样式 |

## MCP Domain

### Prompt

MCP 提示词。

| Field | Type | Description |
|-------|------|-------------|
| URI | string | 格式: `notion://prompt/{page_id}` |
| Name | string | 提示词名称（页面标题） |
| Description | string | 描述（页面摘要） |
| Arguments | []PromptArgument | 参数列表 |
| Content | string | 提示词内容 |

### PromptArgument

提示词参数。

| Field | Type | Description |
|-------|------|-------------|
| Name | string | 参数名称 |
| Description | string | 参数描述 |
| Required | bool | 是否必需 |

### Resource

MCP 资源。

| Field | Type | Description |
|-------|------|-------------|
| URI | string | 格式: `notion://resource/{page_id}` |
| Name | string | 资源名称 |
| Description | string | 描述 |
| MIMEType | string | MIME 类型，默认 "text/plain" |
| Content | string | 资源内容 |

### Tool

MCP 工具。

| Field | Type | Description |
|-------|------|-------------|
| Name | string | 工具名称 |
| Description | string | 工具描述 |
| InputSchema | InputSchema | 输入参数模式 |
| Code | CodeBlock | 可执行代码 |
| Language | string | 代码语言 |

### InputSchema

JSON Schema 风格的输入模式。

| Field | Type | Description |
|-------|------|-------------|
| Type | string | "object" |
| Properties | map[string]PropertySchema | 属性 |
| Required | []string | 必需字段 |

### PropertySchema

属性 Schema。

| Field | Type | Description |
|-------|------|-------------|
| Type | string | 类型: string/number/boolean |
| Description | string | 描述 |
| Enum | []string | 枚举值 |

### ToolCall

工具调用请求。

| Field | Type | Description |
|-------|------|-------------|
| Name | string | 工具名称 |
| Arguments | map[string]any | 参数 |

### ToolResult

工具调用结果。

| Field | Type | Description |
|-------|------|-------------|
| Success | bool | 是否成功 |
| Output | string | 输出内容 |
| Error | string | 错误信息 |

## Change Detection Domain

### ChangeDetector

变更检测器状态。

| Field | Type | Description |
|-------|------|-------------|
| LastPollTime | time.Time | 上次轮询时间 |
| LastSeenVersions | map[string]time.Time | 页面 ID → 最后编辑时间 |
| PollInterval | duration | 轮询间隔 |
| Running | bool | 是否正在运行 |
| StopChan | chan struct{} | 停止信号 |

### ChangeEvent

变更事件。

| Field | Type | Description |
|-------|------|-------------|
| Type | ChangeType | 变更类型 |
| PageID | string | 页面 ID |
| Page | *Page | 页面数据（如果已加载） |
| Timestamp | time.Time | 变更时间 |

**ChangeType**: `created`, `updated`, `deleted`

## Cache Domain

### CacheEntry

缓存条目。

| Field | Type | Description |
|-------|------|-------------|
| Key | string | 缓存键 |
| Data | []byte | 序列化数据 |
| CreatedAt | time.Time | 创建时间 |
| ExpiresAt | time.Time | 过期时间 |
| Version | int | 数据版本 |

### CacheConfig

缓存配置。

| Field | Type | Description |
|-------|------|-------------|
| MemoryTTL | duration | 内存缓存 TTL |
| FileTTL | duration | 文件缓存 TTL |
| Dir | string | 缓存目录 |
| MaxSize | int64 | 最大缓存大小 |

## Logging Domain

### LogEntry

日志条目。

| Field | Type | Description |
|-------|------|-------------|
| Timestamp | time.Time | 时间戳 |
| Level | string | 级别: debug/info/warn/error |
| Message | string | 消息 |
| Fields | map[string]any | 上下文字段 |

## Entity Relationships

```
Config
  └── NOTION_API_KEY, NOTION_DATABASE_ID
         │
         ▼
    Notion Client
         │
         ├── Page ─── Properties ─── Property
         │                           └── select/type ──→ Prompt/Resource/Tool
         │
         └── Block ─── CodeBlock ───→ Tool.Code
                                  └── Text ───→ Prompt/Resource.Content
```

## State Transitions

### Page Lifecycle

```
Draft ──→ Published ──→ Archived
  │           │              │
  └── Query ──┴─── Query ───┘
  (仅返回 Published)
```

### Tool Execution State

```
Idle ──→ Validating ──→ Executing ──→ Completed/Failed
            │                │
            └── Error ───────┘
```

### Polling State

```
Stopped ──→ Starting ──→ Polling ──→ Detected Changes ──→ Updating Cache
                     │              │
                     └── Stop ──────┘
```

## Validation Rules

| Rule | Description |
|------|-------------|
| Page Title | 不能为空 |
| Tool Name | 必须符合 `[a-z][a-z0-9_-]*` |
| Code Language | 必须在允许列表中 |
| Cache Key | 不超过 256 字符 |
| Tool Arguments | 必须符合 InputSchema |
