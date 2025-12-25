# Notion as MCP Server

将 Notion 数据库作为 MCP (Model Context Protocol) 数据源的服务器，支持 prompts、resources 和 tools 三种 MCP 原语。

## 功能特性

- **Prompts**: 从 Notion 中提取 prompt 类型条目
- **Resources**: 从 Notion 中提取 resource 类型条目
- **Tools**: 从 Notion 中提取 tool 类型条目并执行代码块
- **双层缓存**: 内存缓存 (5分钟) + 文件缓存 (1小时)
- **代码执行**: 支持 bash、python、javascript
- **类型过滤**: 通过 Notion 数据库的 type 字段区分不同类型

## 安装

### 从源码构建

```bash
go build -o notion-mcp main.go
```

### 运行

```bash
./notion-mcp serve
```

## 配置

### 环境变量

| 变量 | 描述 | 默认值 |
|------|------|--------|
| `NOTION_API_KEY` | Notion Integration Token | 必需 |
| `NOTION_DATABASE_ID` | Notion Database ID | 必需 |
| `NOTION_TYPE_FIELD` | 类型字段名称 | `Type` |
| `CACHE_TTL` | 缓存 TTL | `5m` |
| `CACHE_DIR` | 缓存目录 | `~/.cache/notion-mcp` |
| `LOG_LEVEL` | 日志级别 | `info` |
| `EXEC_TIMEOUT` | 代码执行超时 | `30s` |
| `EXEC_LANGUAGES` | 允许执行的语言 | `bash,python,js` |

### .env 文件

```bash
cp .env.example .env
```

然后编辑 `.env` 文件填入你的配置。

## Notion 数据库要求

数据库需要包含：

1. **Name 属性**: 条目标题
2. **Type 属性**: Select 类型，可选值为：
   - `prompt` - MCP prompt
   - `resource` - MCP resource
   - `tool` - MCP tool (需要包含代码块)

### 示例数据库结构

| Name | Type |
|------|------|
| My Prompt | prompt |
| Documentation | resource |
| Run Script | tool |

### Tool 条目格式

Tool 类型条目需要包含一个代码块：

```markdown
# 工具名称

这是工具的描述...

```bash
echo "Hello, World!"
```
```

## 使用方法

### Claude Desktop 配置

添加到 `~/.config/claude-desktop/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "notion": {
      "command": "/path/to/notion-mcp",
      "args": ["serve"]
    }
  }
}
```

### MCP Protocol 支持

服务器实现以下 MCP 端点：

- `prompts/list` - 列出所有 prompts
- `prompts/get` - 获取指定 prompt
- `resources/list` - 列出所有 resources
- `resources/read` - 读取 resource 内容
- `tools/list` - 列出所有 tools
- `tools/call` - 调用 tool（执行代码）

## 项目结构

```
notion-mcp/
├── cmd/
│   ├── root.go          # Cobra root 命令
│   └── serve.go         # serve 子命令
├── internal/
│   ├── cache/           # 缓存实现
│   │   ├── cache.go     # 缓存接口
│   │   ├── memory.go    # 内存缓存
│   │   ├── file.go      # 文件缓存
│   │   └── layered.go   # 双层缓存
│   ├── config/          # 配置加载
│   │   └── config.go
│   ├── logger/          # 日志记录
│   │   └── logger.go
│   ├── notion/          # Notion API 客户端
│   │   ├── client.go    # API 客户端
│   │   ├── models.go    # 数据模型
│   │   └── parser.go    # 内容解析器
│   ├── server/          # MCP 服务器
│   │   ├── server.go    # 服务器主逻辑
│   │   ├── prompts.go   # Prompt 处理器
│   │   └── resources.go # Resource 处理器
│   ├── tools/           # 工具执行
│   │   ├── executor.go  # 代码执行器
│   │   └── registry.go  # 工具注册表
│   └── transport/       # 传输层
│       └── stdio.go     # stdio 传输
├── main.go              # 入口点
├── go.mod
└── .env.example
```

## 开发

### 运行测试

```bash
go test ./...
```

### 代码检查

```bash
golangci-lint run
```

## 安全注意事项

- 工具执行支持语言白名单配置
- 代码执行有超时限制 (默认 30 秒)
- 建议在隔离环境中运行代码执行

## 许可证

MIT
