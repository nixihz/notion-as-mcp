# Notion as MCP Server

Dynamically generate MCP Prompts/Resources from Notion databases.

A server that uses Notion databases as MCP (Model Context Protocol) data sources.

## Overview

Notion-as-MCP transforms your Notion databases into a powerful MCP server, allowing you to:

- **Manage prompts** directly in Notion and use them in MCP clients
- **Store resources** as documentation or reference materials

All content is automatically synced from your Notion database, with intelligent caching to minimize API calls.

## Features

- **Prompts**: Extract and serve prompt-type entries from Notion
- **Resources**: Extract and serve resource-type entries from Notion
- **Two-layer caching**: Memory cache (5 minutes) + file cache (1 hour) for optimal performance
- **Type filtering**: Automatically distinguishes types through configurable database fields

## Roadmap

Future features planned:

- **Tools**: Extract tool-type entries from Notion and execute code blocks (bash, python, javascript)
- **Code execution**: Configurable language allowlists and timeout limits
- **Rate limiting**: Built-in exponential backoff for Notion API rate limits

## Quick Start

### Prerequisites

- Go 1.24+ (for building from source)
- A Notion account with API access
- A Notion database configured with a `Type` field

### Installation

Choose one of the following installation methods:

#### Method 1: Using `go install` (Recommended)

Install directly from the repository:

```bash
go install github.com/nixihz/notion-as-mcp@latest
```

The binary will be installed to `$GOPATH/bin` (or `$HOME/go/bin` by default). Make sure this directory is in your `PATH`:

```bash
# Add to your ~/.bashrc, ~/.zshrc, or equivalent
export PATH=$PATH:$(go env GOPATH)/bin
```

#### Method 2: Build from Source

1. **Clone the repository**:
   ```bash
   git clone https://github.com/nixihz/notion-as-mcp.git
   cd notion-as-mcp
   ```

2. **Build the binary**:
   ```bash
   go build -o notion-mcp main.go
   ```

3. **(Optional) Install to system**:
   ```bash
   sudo mv notion-mcp /usr/local/bin/
   ```

### Configuration

1. **Configure environment variables**:
   ```bash
   # Create .env file in your working directory
   cp .env.example .env
   # Edit .env with your Notion credentials
   ```

2. **Run the server**:
   ```bash
   notion-mcp serve
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

### Setting Up Notion

1. **Create a Notion Integration**:
   - Go to https://www.notion.so/my-integrations
   - Create a new integration
   - Copy the Integration Token

2. **Prepare Your Database**:
   - Create or select a Notion database
   - Add a `Select` property named `Type` (or your custom name)
   - Add options: `prompt`, `resource`
   - Add a `Text` property named `Description` for entry descriptions
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
3. **Description property**: A `Text` type field for entry descriptions (optional but recommended)

### Example Database

| Name | Type | Description |
|------|------|-------------|
| Code Review Prompt | prompt | A helpful prompt for reviewing code quality and best practices |
| API Documentation | resource | Complete API reference documentation |

### Entry Formats

#### Prompt Entry
Simply add text content to the page. The entire page content will be used as the prompt.

#### Resource Entry
Add any documentation or reference material. The content will be served as a resource.

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
│   │   └── server.go    # Server main logic
│   ├── tools/           # Tool execution (planned)
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

- **API key security**: Never commit your `.env` file or expose API keys
- **Cache security**: Cache directory should be properly permissioned

## Troubleshooting

### Common Issues

**Server won't start**
- Verify `NOTION_API_KEY` and `NOTION_DATABASE_ID` are set correctly
- Check that your Notion integration has access to the database
- Review logs with `LOG_LEVEL=debug`

**Prompts/Resources not appearing**
- Ensure database entries have `Type` set to `prompt` or `resource`
- Check that the Type field name matches `NOTION_TYPE_FIELD` config

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT

See [LICENSE](LICENSE) for details.
