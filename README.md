# Notion as MCP Server

Dynamically generate MCP Prompts/Resources/Tools from Notion databases.

A server that uses Notion databases as MCP (Model Context Protocol) data sources, supporting three MCP primitives: prompts, resources, and tools.

## Overview

Notion-as-MCP transforms your Notion databases into a powerful MCP server, allowing you to:

- **Manage prompts** directly in Notion and use them in MCP clients
- **Store resources** as documentation or reference materials
- **Create executable tools** by writing code blocks in Notion pages

All content is automatically synced from your Notion database, with intelligent caching to minimize API calls.

## Features

- **Prompts**: Extract and serve prompt-type entries from Notion
- **Resources**: Extract and serve resource-type entries from Notion
- **Tools**: Extract tool-type entries from Notion and execute code blocks
- **Two-layer caching**: Memory cache (5 minutes) + file cache (1 hour) for optimal performance
- **Code execution**: Supports bash, python, and javascript with configurable allowlists
- **Type filtering**: Automatically distinguishes types through configurable database fields
- **Rate limiting**: Built-in exponential backoff for Notion API rate limits

## Quick Start

### Prerequisites

- Go 1.24+ (for building from source)
- A Notion account with API access
- A Notion database configured with a `Type` field

### Installation

1. **Clone the repository**:
   ```bash
   git clone https://github.com/nixihz/notion-as-mcp.git
   cd notion-as-mcp
   ```

2. **Build from source**:
   ```bash
   go build -o notion-mcp main.go
   ```

3. **Configure environment**:
   ```bash
   cp .env.example .env
   # Edit .env with your Notion credentials
   ```

4. **Run the server**:
   ```bash
   ./notion-mcp serve
   ```

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `NOTION_API_KEY` | Notion Integration Token | - | ✅ |
| `NOTION_DATABASE_ID` | Notion Database ID | - | ✅ |
| `NOTION_TYPE_FIELD` | Type field name in database | `Type` | ❌ |
| `CACHE_TTL` | Cache time-to-live | `5m` | ❌ |
| `CACHE_DIR` | Cache directory path | `~/.cache/notion-mcp` | ❌ |
| `LOG_LEVEL` | Logging level (debug/info/warn/error) | `info` | ❌ |
| `EXEC_TIMEOUT` | Code execution timeout | `30s` | ❌ |
| `EXEC_LANGUAGES` | Comma-separated allowed languages | `bash,python,js` | ❌ |

### Setting Up Notion

1. **Create a Notion Integration**:
   - Go to https://www.notion.so/my-integrations
   - Create a new integration
   - Copy the Integration Token

2. **Prepare Your Database**:
   - Create or select a Notion database
   - Add a `Select` property named `Type` (or your custom name)
   - Add options: `prompt`, `resource`, `tool`
   - Share the database with your integration

3. **Get Database ID**:
   - Open your database in Notion
   - Copy the ID from the URL (the part after the last `/` and before `?`)

## Notion Database Structure

### Required Properties

Your Notion database must have:

1. **Name property**: The title of each entry (standard Notion property)
2. **Type property**: A `Select` type field with these options:
   - `prompt` - MCP prompt entries
   - `resource` - MCP resource entries
   - `tool` - MCP tool entries (must contain code blocks)

### Example Database

| Name | Type |
|------|------|
| Code Review Prompt | prompt |
| API Documentation | resource |
| Git Commit Tool | tool |

### Entry Formats

#### Prompt Entry
Simply add text content to the page. The entire page content will be used as the prompt.

#### Resource Entry
Add any documentation or reference material. The content will be served as a resource.

#### Tool Entry
Tool entries must contain a code block with executable code:

```markdown
# Tool Name

Description of what this tool does...

```bash
#!/bin/bash
# Your code here
echo "Hello from Notion!"
```
```

The code block language determines the execution environment (bash, python, or javascript).

## Usage

### Claude Desktop Integration

Add the server to your Claude Desktop configuration at `~/.config/claude-desktop/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "notion": {
      "command": "/absolute/path/to/notion-mcp",
      "args": ["serve"]
    }
  }
}
```

Restart Claude Desktop to load the server.

### MCP Protocol Endpoints

The server implements the following MCP endpoints:

- **`prompts/list`** - List all available prompts
- **`prompts/get`** - Get a specific prompt by name
- **`resources/list`** - List all available resources
- **`resources/read`** - Read resource content by URI
- **`tools/list`** - List all available tools
- **`tools/call`** - Execute a tool with parameters

## Project Structure

```
notion-mcp/
├── cmd/
│   ├── root.go          # Cobra root command
│   └── serve.go         # serve subcommand
├── internal/
│   ├── cache/           # Cache implementation
│   │   ├── cache.go     # Cache interface
│   │   ├── memory.go    # Memory cache
│   │   ├── file.go      # File cache
│   │   └── layered.go   # Two-layer cache
│   ├── config/          # Configuration loading
│   │   └── config.go
│   ├── logger/          # Logging
│   │   └── logger.go
│   ├── notion/          # Notion API client
│   │   ├── client.go    # API client
│   │   ├── models.go    # Data models
│   │   ├── parser.go    # Content parser
│   │   └── markdown.go  # Markdown conversion
│   ├── server/          # MCP server
│   │   ├── server.go    # Server main logic
│   ├── tools/           # Tool execution
│   │   ├── executor.go  # Code executor
│   │   └── registry.go  # Tool registry
│   └── transport/       # Transport layer
│       └── stdio.go     # stdio transport
├── main.go              # Entry point
├── LICENSE              # MIT License
└── .env.example        # Example configuration
```

## Development

### Running Tests

```bash
go test ./...
```

### Code Linting

```bash
golangci-lint run
```

### Development Mode

```bash
go run main.go serve
```

## Security Considerations

- **Language allowlist**: Only configured languages can be executed
- **Timeout limits**: Code execution is limited to 30 seconds by default
- **Isolated execution**: Consider running in a sandboxed environment for production
- **API key security**: Never commit your `.env` file or expose API keys

## Troubleshooting

### Common Issues

**Server won't start**
- Verify `NOTION_API_KEY` and `NOTION_DATABASE_ID` are set correctly
- Check that your Notion integration has access to the database
- Review logs with `LOG_LEVEL=debug`

**Tools not appearing**
- Ensure database entries have `Type` set to `tool`
- Verify entries contain code blocks
- Check that code block language is in `EXEC_LANGUAGES`

**Rate limiting errors**
- Increase `CACHE_TTL` to reduce API calls
- The server automatically retries with exponential backoff

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT

See [LICENSE](LICENSE) for details.
