// Package server provides the MCP server implementation.
package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nixihz/notion-as-mcp/internal/notion"
)

// promptHandler handles MCP prompts list/get requests.
type promptHandler struct {
	server *Server
}

// NewPromptHandler creates a new prompt handler.
func NewPromptHandler(srv *Server) *promptHandler {
	return &promptHandler{server: srv}
}

// ListPrompts implements mcp.PromptHandler.
func (h *promptHandler) ListPrompts(ctx context.Context, req *mcp.ListPromptsRequest) (*mcp.ListPromptsResult, error) {
	// Query Notion database for prompt-type entries
	pages, err := h.server.client.QueryDatabase(ctx, notion.NewTypeFilter(h.server.cfg.NotionTypeField, "prompt"))
	if err != nil {
		return nil, err
	}

	var prompts []*mcp.Prompt
	for _, page := range pages {
		title := getPageTitle(page)
		prompts = append(prompts, &mcp.Prompt{
			Name:        sanitizeToolName(title),
			Title:       title,
			Description: title,
		})
	}

	return &mcp.ListPromptsResult{
		Prompts: prompts,
	}, nil
}

// GetPrompt implements mcp.PromptHandler.
func (h *promptHandler) GetPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	// Find the page by title
	pages, err := h.server.client.QueryDatabase(ctx, notion.NewTypeFilter(h.server.cfg.NotionTypeField, "prompt"))
	if err != nil {
		return nil, err
	}

	promptName := req.Params.Name
	for _, page := range pages {
		if sanitizeToolName(getPageTitle(page)) == promptName {
			// Get page content
			content, err := h.server.client.GetPageContent(ctx, page.ID)
			if err != nil {
				return nil, err
			}

			return &mcp.GetPromptResult{
				Description: getPageTitle(page),
				Messages: []*mcp.PromptMessage{
					{
						Role: "user",
						Content: &mcp.TextContent{
							Text: content.Text,
						},
					},
				},
			}, nil
		}
	}

	return nil, nil
}
