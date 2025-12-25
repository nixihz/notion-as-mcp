# Quickstart: Notion MCP Server

**Branch**: `001-notion-mcp-tools`
**Date**: 2025-12-25

## Prerequisites

- Go 1.25+
- Notion Integration Token
- Notion Database with type field

## Setup

### 1. Clone and Install

```bash
git checkout 001-notion-mcp-tools
cd notion-as-mcp
go install ./cmd/notion-mcp
```

### 2. Configure Notion Integration

1. 访问 https://www.notion.so/my-integrations
2. 创建新的 Integration
3. 复制 Integration Token

### 3. Configure Database

1. 在 Notion 中创建或选择数据库
2. 添加 Select 类型字段，命名为 `Type`（或其他名称）
3. 添加选项：`prompt`, `resource`, `tool`
4. 在数据库页面右上角点击 `...` > `Connections` > 添加你的 Integration

### 4. Create Environment File

```bash
cp .env.example .env
```

编辑 `.env`:

```env
NOTION_API_KEY=secret_your_integration_token
NOTION_DATABASE_ID=database_id_from_url
NOTION_TYPE_FIELD=Type
CACHE_TTL=5m
LOG_LEVEL=info
```

### 5. Populate Database

在 Notion 数据库中添加条目：

**Prompt 类型**:
- Type = `prompt`
- 页面内容：提示词文本

**Resource 类型**:
- Type = `resource`
- 页面内容：资源文档

**Tool 类型**:
- Type = `tool`
- 页面内容：代码块

## Usage

### Start Server

```bash
notion-mcp serve
```

### Configure MCP Client

在 Claude Code 或其他 MCP 客户端中添加：

```json
{
  "command": "notion-mcp",
  "args": ["serve"],
  "env": {
    "NOTION_API_KEY": "secret_...",
    "NOTION_DATABASE_ID": "..."
  }
}
```

### Available Tools

从 Notion tool 类型条目中提取的工具将自动可用。

## Configuration Reference

| Variable | Default | Description |
|----------|---------|-------------|
| `NOTION_API_KEY` | - | Notion Integration Token (必需) |
| `NOTION_DATABASE_ID` | - | Database ID from URL (必需) |
| `NOTION_TYPE_FIELD` | `Type` | Property name for type classification |
| `CACHE_TTL` | `5m` | Cache time-to-live |
| `CACHE_DIR` | `~/.cache/notion-mcp` | Cache directory |
| `LOG_LEVEL` | `info` | Logging level: debug/info/warn/error |
| `EXEC_TIMEOUT` | `30s` | Tool execution timeout |
| `EXEC_LANGUAGES` | `bash,python,js` | Allowed execution languages |
| `POLL_INTERVAL` | `60s` | Change detection interval (0 to disable) |
| `REFRESH_ON_START` | `true` | Refresh data on server start |

## Troubleshooting

### Connection Failed

```bash
# 检查环境变量
notion-mcp doctor

# 查看详细日志
NOTION_API_KEY=... LOG_LEVEL=debug notion-mcp serve
```

### No Tools Available

1. 确认数据库中存在类型为 `tool` 的条目
2. 确认条目包含代码块
3. 检查日志中的解析错误

### Rate Limiting

如果看到 429 错误，增加 `CACHE_TTL` 或 `POLL_INTERVAL` 以减少 API 调用频率。

### Data Not Updating

如果 Notion 中的变更没有反映到工具列表：

1. 检查 `POLL_INTERVAL` 是否设置为正数（默认 60 秒）
2. 重启服务器：`REFRESH_ON_START=true notion-mcp serve`
3. 手动触发刷新（如果支持命令）

## Examples

### Add a Tool

在 Notion 中创建新页面：

1. Type = `tool`
2. 标题 = `Calculate`
3. 内容添加代码块：

```python
def calculate(a: float, b: float, operation: str) -> float:
    if operation == "add":
        return a + b
    elif operation == "subtract":
        return a - b
    elif operation == "multiply":
        return a * b
    elif operation == "divide":
        return a / b
    else:
        raise ValueError(f"Unknown operation: {operation}")

# 调用示例
result = calculate(10, 5, "add")
print(result)  # 输出: 15
```

工具将在服务器重启后自动加载。
