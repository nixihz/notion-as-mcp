# Notion as MCP Server

[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![MCP](https://img.shields.io/badge/MCP-Compatible-blue)](https://modelcontextprotocol.io)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Turn your Notion databases into a dynamic MCP (Model Context Protocol) server — manage **prompts** and **resources** directly in Notion and use them in any MCP client.

## Features

- **Prompts** — Serve prompt templates from Notion pages
- **Resources** — Serve documentation and reference materials
- **Two-layer caching** — Memory (5 min) + file (1 hour) for minimal API calls
- **Dual transport** — Streamable HTTP (SSE) for remote, stdio for local
- **Docker ready** — Multi-stage build with non-root user
- **Auto-sync** — Configurable polling interval to detect Notion changes

## Roadmap

- **Tools** — Execute code blocks (bash, python, js) defined in Notion pages (code scaffolding in place, not yet wired up)

## Quick Start

### Prerequisites

- Go 1.24+ (for building from source), or Docker
- A [Notion Integration](https://www.notion.so/my-integrations) with API token
- A Notion database with a `Type` select property (`prompt` / `resource`)

### Install

**go install** (recommended):

```bash
go install github.com/nixihz/notion-as-mcp@latest
```

**Build from source**:

```bash
git clone https://github.com/nixihz/notion-as-mcp.git
cd notion-as-mcp
go build -o notion-as-mcp main.go
```

### Configure & Run

```bash
cp .env.example .env
# Edit .env with your NOTION_API_KEY and NOTION_DATABASE_ID

# Streamable HTTP (default, port 3100)
notion-as-mcp serve

# stdio mode (for Claude Desktop / local clients)
notion-as-mcp serve --transport stdio
```

## Configuration

All configuration via environment variables or `.env` file:

| Variable | Description | Default |
|----------|-------------|---------|
| `NOTION_API_KEY` | Notion Integration Token | **(required)** |
| `NOTION_DATABASE_ID` | Notion Database ID | **(required)** |
| `NOTION_TYPE_FIELD` | Type property name in database | `Type` |
| `TRANSPORT_TYPE` | `streamable` or `stdio` | `streamable` |
| `SERVER_HOST` | Listen address (streamable mode) | `0.0.0.0` |
| `SERVER_PORT` | Listen port (streamable mode) | `3100` |
| `CACHE_TTL` | Cache time-to-live | `5m` |
| `CACHE_DIR` | Cache directory path | `~/.cache/notion-as-mcp` |
| `POLL_INTERVAL` | Notion change polling interval (`0` to disable) | `60s` |
| `REFRESH_ON_START` | Refresh data on server start | `true` |
| `EXEC_TIMEOUT` | Code execution timeout (planned) | `30s` |
| `EXEC_LANGUAGES` | Allowed languages, comma-separated (planned) | `bash,python,js` |
| `LOG_LEVEL` | `debug` / `info` / `warn` / `error` | `info` |

CLI flags (`--host`, `--port`, `--transport`) override environment variables.

## Setting Up Notion

1. **Create Integration** — Go to [My Integrations](https://www.notion.so/my-integrations), create one, and copy the token.

2. **Prepare Database** — Add these properties:
   - `Type` — Select property with options: `prompt`, `resource`
   - `Description` — Text property (optional but recommended)

3. **Share Database** — Invite your integration to the database via the "..." menu → "Connections".

4. **Get Database ID** — From the database URL: `https://notion.so/{workspace}/{DATABASE_ID}?v=...`

### Example Database

| Name | Type | Description |
|------|------|-------------|
| Code Review Prompt | prompt | Prompt for reviewing code quality |
| API Documentation | resource | Complete API reference |

### Entry Content

- **Prompt**: Page content becomes the prompt template
- **Resource**: Page content served as documentation

## MCP Client Integration

### Claude Desktop

Add to `~/.config/claude-desktop/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "notion": {
      "command": "/path/to/notion-as-mcp",
      "args": ["serve", "--transport", "stdio"],
      "env": {
        "NOTION_API_KEY": "ntn_xxx",
        "NOTION_DATABASE_ID": "your-database-id"
      }
    }
  }
}
```

### Remote (Streamable HTTP)

Start the server, then connect your MCP client to `http://host:3100/mcp`.

## Docker

```bash
# Build
docker build -t notion-as-mcp .

# Run (streamable, default)
docker run -d --name notion-as-mcp \
  -p 3100:3100 \
  -e NOTION_API_KEY=ntn_xxx \
  -e NOTION_DATABASE_ID=your-db-id \
  notion-as-mcp

# Run (stdio)
docker run --rm -i \
  -e NOTION_API_KEY=ntn_xxx \
  -e NOTION_DATABASE_ID=your-db-id \
  -e TRANSPORT_TYPE=stdio \
  notion-as-mcp
```

## Development

Uses [Task](https://taskfile.dev) for common operations:

```bash
task serve          # Run with streamable transport (default)
task serve:stdio    # Run with stdio transport
task build          # Build binary
task docker:build   # Build Docker image
task docker:run     # Run Docker container
```

Or directly:

```bash
go run main.go serve          # Dev run
go test ./...                 # Tests
golangci-lint run             # Lint
```

## MCP Endpoints

| Endpoint | Description |
|----------|-------------|
| `prompts/list` | List available prompts |
| `prompts/get` | Get prompt content by name |
| `resources/list` | List available resources |
| `resources/read` | Read resource content by URI |

## Project Structure

```
notion-as-mcp/
├── cmd/
│   ├── root.go              # Cobra root command
│   └── serve.go             # serve subcommand
├── internal/
│   ├── cache/               # Memory + file two-layer cache
│   ├── config/              # Configuration loading
│   ├── logger/              # slog-based logging
│   ├── notion/              # Notion API client & parser
│   ├── server/              # MCP server implementation
│   └── tools/               # Code execution engine
├── main.go
├── Dockerfile
├── Taskfile.yml
└── .env.example
```

## Security

- **API keys**: Never commit `.env` — use environment variables in production
- **Code execution** (planned): Will be sandboxed by language allowlist and timeout
- **Docker**: Runs as non-root user with minimal Alpine image

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Server won't start | Verify `NOTION_API_KEY` and `NOTION_DATABASE_ID`; check integration has database access |
| Entries not appearing | Ensure `Type` is set to `prompt`/`resource`; verify `NOTION_TYPE_FIELD` matches |
| Stale data | Lower `POLL_INTERVAL` or restart server; check `CACHE_TTL` |

## Contributing

Contributions welcome! Please submit a Pull Request.

## License

[MIT](LICENSE)
