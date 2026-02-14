// Package server provides the MCP server implementation.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/samber/lo"

	"github.com/nixihz/notion-as-mcp/internal/cache"
	"github.com/nixihz/notion-as-mcp/internal/config"
	"github.com/nixihz/notion-as-mcp/internal/logger"
	"github.com/nixihz/notion-as-mcp/internal/notion"
	"github.com/nixihz/notion-as-mcp/internal/tools"
)

// Page type constants
const (
	pageTypePrompt   = "prompt"
	pageTypeResource = "resource"
	pageTypeTool     = "tool"
)

// Server represents the MCP server.
type Server struct {
	cfg      *config.Config
	client   *notion.Client
	cache    cache.Cache
	mcpCache *cache.MCPCache
	logger   *slog.Logger
	impl     *mcp.Implementation
	executor *tools.Executor
	toolReg  *tools.Registry
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

	// Initialize MCP cache manager
	mcpCacheManager := cache.NewMCPCache(cacheStore, log)

	srv := &Server{
		cfg:      cfg,
		client:   client,
		cache:    cacheStore,
		mcpCache: mcpCacheManager,
		logger:   log,
		impl: &mcp.Implementation{
			Name:    "notion-as-mcp",
			Version: "1.0.0",
		},
		executor: tools.NewExecutor(cfg.ExecTimeout, cfg.ExecLanguages),
		toolReg:  tools.NewRegistry(),
	}

	return srv, nil
}

// Start starts the MCP server with the configured transport.
func (s *Server) Start(ctx context.Context) error {
	// Warm cache on startup
	s.warmCache(ctx)

	// Start periodic refresh in background
	s.startPeriodicRefresh(ctx)

	// Get all pages - try cache first, then fallback to Notion
	allPages := s.getAllPagesWithCache(ctx)

	if s.cfg.TransportType == "streamable" {
		return s.startStreamable(ctx, allPages)
	}
	return s.startStdio(ctx, allPages)
}

// getAllPagesWithCache tries to get pages from cache first, falls back to Notion.
func (s *Server) getAllPagesWithCache(ctx context.Context) []notion.Page {
	// Try cache first
	cachedData, err := s.mcpCache.Get(ctx, cache.CacheKeyResources)
	if err == nil && cachedData != nil {
		var pages []notion.Page
		if json.Unmarshal(cachedData, &pages) == nil {
			s.logger.Info("using cached pages for resources")
			return pages
		}
	}

	// Cache miss or error, fetch from Notion
	s.logger.Info("fetching pages from Notion (cache miss)")
	pages, err := s.client.GetAllPages(ctx)
	if err != nil {
		s.logger.Warn("failed to query pages", slog.String("error", err.Error()))
		return nil
	}
	return pages
}

// warmCache fetches and caches all pages on startup.
func (s *Server) warmCache(ctx context.Context) {
	// Warm resources cache
	err := s.mcpCache.Warm(ctx, cache.CacheKeyResources, func(ctx context.Context) ([]byte, error) {
		pages, err := s.client.GetAllPages(ctx)
		if err != nil {
			return nil, err
		}
		// Filter only resource pages
		var resourcePages []notion.Page
		for _, p := range pages {
			pageType := notion.GetTypeFromProperties(p.Properties, s.cfg.NotionTypeField)
			if pageType == pageTypeResource {
				resourcePages = append(resourcePages, p)
			}
		}
		return s.serializePages(resourcePages)
	})
	if err != nil {
		s.logger.Warn("failed to warm resources cache", slog.String("error", err.Error()))
	}

	// Warm prompts cache
	err = s.mcpCache.Warm(ctx, cache.CacheKeyPrompts, func(ctx context.Context) ([]byte, error) {
		pages, err := s.client.GetAllPages(ctx)
		if err != nil {
			return nil, err
		}
		// Filter only prompt pages
		var promptPages []notion.Page
		for _, p := range pages {
			pageType := notion.GetTypeFromProperties(p.Properties, s.cfg.NotionTypeField)
			if pageType == pageTypePrompt {
				promptPages = append(promptPages, p)
			}
		}
		return s.serializePages(promptPages)
	})
	if err != nil {
		s.logger.Warn("failed to warm prompts cache", slog.String("error", err.Error()))
	}
}

// startPeriodicRefresh starts background goroutines to periodically refresh caches.
func (s *Server) startPeriodicRefresh(ctx context.Context) {
	// Periodic refresh for resources
	s.mcpCache.StartPeriodicRefresh(ctx, cache.CacheKeyResources, s.cfg.CacheRefreshInterval, func(ctx context.Context) ([]byte, error) {
		pages, err := s.client.GetAllPages(ctx)
		if err != nil {
			return nil, err
		}
		var resourcePages []notion.Page
		for _, p := range pages {
			pageType := notion.GetTypeFromProperties(p.Properties, s.cfg.NotionTypeField)
			if pageType == pageTypeResource {
				resourcePages = append(resourcePages, p)
			}
		}
		return s.serializePages(resourcePages)
	})

	// Periodic refresh for prompts
	s.mcpCache.StartPeriodicRefresh(ctx, cache.CacheKeyPrompts, s.cfg.CacheRefreshInterval, func(ctx context.Context) ([]byte, error) {
		pages, err := s.client.GetAllPages(ctx)
		if err != nil {
			return nil, err
		}
		var promptPages []notion.Page
		for _, p := range pages {
			pageType := notion.GetTypeFromProperties(p.Properties, s.cfg.NotionTypeField)
			if pageType == pageTypePrompt {
				promptPages = append(promptPages, p)
			}
		}
		return s.serializePages(promptPages)
	})
}

// serializePages serializes pages to JSON bytes.
func (s *Server) serializePages(pages []notion.Page) ([]byte, error) {
	return json.Marshal(pages)
}

// startStreamable starts the MCP server with streamable HTTP transport.
func (s *Server) startStreamable(ctx context.Context, allPages []notion.Page) error {
	server := mcp.NewServer(s.impl, nil)

	// Register handlers
	s.registerPrompts(server, allPages)
	s.registerResources(server, allPages)

	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return server
	}, nil)

	addr := fmt.Sprintf("%s:%d", s.cfg.ServerHost, s.cfg.ServerPort)
	s.logger.Info("starting Notion MCP server with streamable transport",
		slog.String("database_id", s.cfg.NotionDatabaseID),
		slog.String("type_field", s.cfg.NotionTypeField),
		slog.String("addr", addr),
	)

	// Start HTTP server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- http.ListenAndServe(addr, handler)
	}()

	// Wait for HTTP server error or context cancellation
	select {
	case err := <-serverErr:
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
		return ctx.Err()
	}
}

// startStdio starts the MCP server with stdio transport.
func (s *Server) startStdio(ctx context.Context, allPages []notion.Page) error {
	s.logger.Info("starting Notion MCP server with stdio transport",
		slog.String("database_id", s.cfg.NotionDatabaseID),
		slog.String("type_field", s.cfg.NotionTypeField),
	)

	server := mcp.NewServer(s.impl, nil)

	// Register handlers
	s.registerPrompts(server, allPages)
	s.registerResources(server, allPages)

	s.logger.Info("Notion MCP server started")

	// Use SDK built-in StdioTransport with server.Run
	return server.Run(ctx, &mcp.StdioTransport{})
}

// Stop stops the MCP server.
func (s *Server) Stop() error {
	// Stop periodic refresh
	if s.mcpCache != nil {
		s.mcpCache.StopAll()
	}
	if s.cache != nil {
		s.cache.Close()
	}
	return nil
}

// registerPrompts registers prompt handlers.
func (s *Server) registerPrompts(server *mcp.Server, allPages []notion.Page) {
	// Filter pages by type using functional programming
	promptPages := lo.Filter(allPages, func(page notion.Page, _ int) bool {
		pageType := notion.GetTypeFromProperties(page.Properties, s.cfg.NotionTypeField)
		return pageType == pageTypePrompt
	})

	// Register each prompt page
	lo.ForEach(promptPages, func(page notion.Page, _ int) {
		title := getPageTitle(page)
		promptName := sanitizeToolName(title)
		promptDesc := getPageDescription(page)

		// Validate prompt name (must match pattern: ^[a-z][a-z0-9_-]*$)
		if promptName == "" {
			s.logger.Warn("skipping prompt with empty name", slog.String("page_id", page.ID), slog.String("title", title))
			return
		}

		// Ensure name starts with lowercase letter
		if len(promptName) > 0 && (promptName[0] < 'a' || promptName[0] > 'z') {
			// Prepend 'p_' if name doesn't start with lowercase letter
			promptName = "p_" + promptName
		}

		s.logger.Info("registering prompt",
			"name", promptName,
			"title", title,
			"page_id", page.ID,
		)
		promptHandler := s.createPromptHandler(page)
		server.AddPrompt(&mcp.Prompt{
			Name:        promptName,
			Description: promptDesc,
		}, promptHandler)
	})

	s.logger.Info("registered prompts", slog.Int("count", len(promptPages)))
}

// registerResources registers resource handlers.
func (s *Server) registerResources(server *mcp.Server, allPages []notion.Page) {
	resourcePages := lo.Filter(allPages, func(page notion.Page, _ int) bool {
		pageType := notion.GetTypeFromProperties(page.Properties, s.cfg.NotionTypeField)
		return pageType == pageTypeResource
	})

	// Register each resource page
	lo.ForEach(resourcePages, func(page notion.Page, _ int) {
		title := getPageTitle(page)
		resourceName := sanitizeToolName(title)
		resourceDesc := getPageDescription(page)

		// Validate resource name (must match pattern: ^[a-z][a-z0-9_-]*$)
		if resourceName == "" {
			s.logger.Warn("skipping resource with empty name", slog.String("page_id", page.ID), slog.String("title", title))
			return
		}

		// Ensure name starts with lowercase letter
		if len(resourceName) > 0 && (resourceName[0] < 'a' || resourceName[0] > 'z') {
			// Prepend 'r_' if name doesn't start with lowercase letter
			resourceName = "r_" + resourceName
		}

		s.logger.Info("registering resource",
			"name", resourceName,
			"title", title,
			"page_id", page.ID,
		)
		resourceHandler := s.createResourceHandler(page)
		server.AddResource(&mcp.Resource{
			URI:         "file:///notion/" + page.ID,
			Name:        resourceName,
			Description: resourceDesc,
		}, resourceHandler)
	})

	s.logger.Info("registered resources", "count", len(resourcePages))
}

// registerTools registers tool handlers.
func (s *Server) registerTools(server *mcp.Server, allPages []notion.Page) {
	// Filter pages by type
	toolPages := lo.Filter(allPages, func(page notion.Page, _ int) bool {
		pageType := notion.GetTypeFromProperties(page.Properties, s.cfg.NotionTypeField)
		return pageType == pageTypeTool
	})

	// Register each tool page
	lo.ForEach(toolPages, func(page notion.Page, _ int) {
		title := getPageTitle(page)
		toolName := sanitizeToolName(getPageTitle(page))
		toolDesc := getPageDescription(page)

		s.logger.Info("registering tool",
			"name", toolName,
			"title", title,
			"page_id", page.ID,
		)
		toolHandler := s.createToolHandler(page)
		if os.Getenv("ENV") == "development" || os.Getenv("GO_ENV") == "development" {
			result, err := toolHandler(context.Background(), nil)
			if err != nil {
				fmt.Print(result)
				s.logger.Warn("failed to create tool handler", "error", err.Error())
				return
			}
		}

		server.AddTool(&mcp.Tool{
			Name:        toolName,
			Description: toolDesc,
		}, toolHandler)
	})

	s.logger.Info("registered tools", slog.Int("count", len(toolPages)))
}

// createPromptHandler creates a handler for a specific prompt.
func (s *Server) createPromptHandler(page notion.Page) mcp.PromptHandler {
	return func(ctx context.Context, request *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		// Get page content
		content, err := s.client.GetPageContent(ctx, page.ID)
		if err != nil {
			return nil, fmt.Errorf("error fetching content: %w", err)
		}
		markdown := notion.PageToMarkdown(content)

		title := getPageTitle(page)
		return &mcp.GetPromptResult{
			Description: title,
			Messages: []*mcp.PromptMessage{
				{
					Role: "user",
					Content: &mcp.TextContent{
						Text: markdown,
					},
				},
			},
		}, nil
	}
}

// createResourceHandler creates a handler for a specific resource.
func (s *Server) createResourceHandler(page notion.Page) mcp.ResourceHandler {
	return func(ctx context.Context, request *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		// Get page content
		content, err := s.client.GetPageContent(ctx, page.ID)
		if err != nil {
			return nil, fmt.Errorf("error fetching content: %w", err)
		}
		markdown := notion.PageToMarkdown(content)
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:  "file:///resource/" + page.ID,
					Text: markdown,
				},
			},
		}, nil
	}
}

// createToolHandler creates a handler for a specific tool.
func (s *Server) createToolHandler(page notion.Page) mcp.ToolHandler {

	// Get page content
	content, err := s.client.GetPageContent(context.Background(), page.ID)
	if err != nil {
		s.logger.Warn("failed to fetch content", slog.String("error", err.Error()))
		return nil
	}

	// If no code block, return the text content
	if !content.HasCode {
		s.logger.Warn("no code block found", slog.String("page_id", page.ID))
		return nil
	}
	codeStr := extractCodeString(content.Code.RichText)
	language := content.Code.Language

	return func(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract code string from RichText

		input := "{ numberList: [ 1, 2, 3, 4, 5 ] }"
		if request != nil && request.Params != nil && request.Params.Arguments != nil {
			input = string(request.Params.Arguments)
		}

		// Execute the code
		result, err := s.executor.Execute(ctx, language, codeStr, input)
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
		if len(title.Title) > 0 {
			t := title.Title[0].PlainText
			if t != "" {
				return t
			}
		}
	}
	return page.ID
}
func getPageDescription(page notion.Page) string {
	if description, ok := page.Properties["Description"]; ok {
		if len(description.RichText) > 0 {
			return description.RichText[0].PlainText
		}
	}
	return ""
}

// sanitizeToolName converts a page title to a valid tool/prompt name.
// MCP requires: ^[a-z][a-z0-9_-]*$ (must start with lowercase letter)
func sanitizeToolName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return ""
	}

	// Replace spaces and special chars with underscores
	var result strings.Builder
	firstChar := true
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' || c == '-' {
			// If first char is a number, underscore, or hyphen, prepend 'p_'
			if firstChar && (c >= '0' && c <= '9' || c == '_' || c == '-') {
				result.WriteString("p_")
			}
			result.WriteRune(c)
			firstChar = false
		} else if c == ' ' || c == '\t' {
			if !firstChar {
				result.WriteRune('_')
			}
		}
	}

	sanitized := result.String()
	// Ensure it starts with a lowercase letter
	if sanitized == "" {
		return ""
	}
	if sanitized[0] < 'a' || sanitized[0] > 'z' {
		return "p_" + sanitized
	}

	return sanitized
}
