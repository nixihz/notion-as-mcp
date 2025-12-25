// Package server provides the MCP server implementation.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nixihz/notion-as-mcp/internal/cache"
	"github.com/nixihz/notion-as-mcp/internal/config"
	"github.com/nixihz/notion-as-mcp/internal/logger"
	"github.com/nixihz/notion-as-mcp/internal/notion"
	"github.com/nixihz/notion-as-mcp/internal/tools"
	"github.com/nixihz/notion-as-mcp/internal/transport"
)

// Server represents the MCP server.
type Server struct {
	cfg       *config.Config
	client    *notion.Client
	cache     cache.Cache
	logger    *slog.Logger
	impl      *mcp.Implementation
	executor  *tools.Executor
	toolReg   *tools.Registry
}

// NewServer creates a new MCP server.
func NewServer(cfg *config.Config) (*Server, error) {
	// Initialize logger
	if err := logger.Init(cfg); err != nil {
		return nil, fmt.Errorf("init logger: %w", err)
	}

	log := logger.Get()

	// Initialize cache
	cacheStore, err := cache.NewCache(
		cache.WithTTL(cfg.CacheTTL),
		cache.WithDir(cfg.CacheDir),
	)
	if err != nil {
		return nil, fmt.Errorf("init cache: %w", err)
	}

	// Create Notion client
	client := notion.NewClient(
		cfg.NotionAPIKey,
		cfg.NotionDatabaseID,
		cfg.NotionTypeField,
	)

	srv := &Server{
		cfg:      cfg,
		client:   client,
		cache:    cacheStore,
		logger:   log,
		impl: &mcp.Implementation{
			Name:    "notion-mcp",
			Version: "1.0.0",
		},
		executor: tools.NewExecutor(cfg.ExecTimeout, cfg.ExecLanguages),
		toolReg:  tools.NewRegistry(),
	}

	return srv, nil
}

// Start starts the MCP server with stdio transport.
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("starting Notion MCP server",
		slog.String("database_id", s.cfg.NotionDatabaseID),
		slog.String("type_field", s.cfg.NotionTypeField),
	)

	// Create stdio transport
	stdioTransport := transport.NewStdioTransport()

	// Create server
	server := mcp.NewServer(s.impl, nil)

	// Register handlers
	s.registerPrompts(server)
	s.registerResources(server)
	s.registerTools(server)

	// Accept connection with transport
	session, err := server.Connect(ctx, stdioTransport, nil)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	s.logger.Info("Notion MCP server started")

	// Wait for session to end
	return session.Wait()
}

// Stop stops the MCP server.
func (s *Server) Stop() error {
	if s.cache != nil {
		s.cache.Close()
	}
	return nil
}

// registerPrompts registers prompt handlers.
func (s *Server) registerPrompts(server *mcp.Server) {
	promptHandler := NewPromptHandler(s)
	server.AddPrompt(&mcp.Prompt{
		Name:        "notion-prompts",
		Description: "List all prompts from Notion database",
	}, promptHandler.GetPrompt)
}

// registerResources registers resource handlers.
func (s *Server) registerResources(server *mcp.Server) {
	resourceHandler := NewResourceHandler(s)
	server.AddResource(&mcp.Resource{
		URI:         "notion://root",
		Name:        "Notion Resources",
		Description: "Root resource for Notion database",
	}, resourceHandler.ReadResource)
}

// registerTools registers tool handlers.
func (s *Server) registerTools(server *mcp.Server) {
	// Get tools from Notion and register them
	pages, err := s.client.QueryDatabase(context.Background(), notion.NewTypeFilter(s.cfg.NotionTypeField, "tool"))
	if err != nil {
		s.logger.Warn("failed to query tools", slog.String("error", err.Error()))
		return
	}

	for _, page := range pages {
		toolName := sanitizeToolName(getPageTitle(page))
		toolDesc := fmt.Sprintf("Tool from Notion: %s", getPageTitle(page))

		server.AddTool(&mcp.Tool{
			Name:        toolName,
			Description: toolDesc,
		}, s.createToolHandler(page))
	}
}

// createToolHandler creates a handler for a specific tool.
func (s *Server) createToolHandler(page notion.Page) mcp.ToolHandler {
	return func(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Get page content
		content, err := s.client.GetPageContent(ctx, page.ID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error fetching content: %v", err)},
				},
				IsError: true,
			}, nil
		}

		// If no code block, return the text content
		if !content.HasCode {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: content.Text},
				},
			}, nil
		}

		// Extract code string from RichText
		codeStr := extractCodeString(content.Code.Code)
		language := content.Code.Language

		// Execute the code
		result, err := s.executor.Execute(ctx, language, codeStr)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Execution error: %v", err)},
				},
				IsError: true,
			}, nil
		}

		// Format output
		output := fmt.Sprintf("Language: %s\nExit Code: %d\nOutput:\n%s", language, result.ExitCode, result.Output)
		if result.Error != "" {
			output += fmt.Sprintf("\nError: %s", result.Error)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: output},
			},
		}, nil
	}
}

// extractCodeString extracts the code string from RichText array.
func extractCodeString(richTexts []notion.RichText) string {
	var sb strings.Builder
	for _, rt := range richTexts {
		sb.WriteString(rt.PlainText)
	}
	return sb.String()
}

// getPageTitle extracts the title from a page.
func getPageTitle(page notion.Page) string {
	if title, ok := page.Properties["Name"]; ok {
		if title.Value != nil {
			if v, ok := title.Value.(string); ok {
				return v
			}
		}
	}
	return page.ID
}

// sanitizeToolName converts a page title to a valid tool name.
func sanitizeToolName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)
	// Replace spaces and special chars with underscores
	var result strings.Builder
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' || c == '-' {
			result.WriteRune(c)
		} else if c == ' ' || c == '\t' {
			result.WriteRune('_')
		}
	}
	return result.String()
}
