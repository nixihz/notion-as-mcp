// Package server provides the MCP server implementation.
package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nixihz/notion-as-mcp/internal/notion"
)

// resourceHandler handles MCP resources list/read requests.
type resourceHandler struct {
	server *Server
}

// NewResourceHandler creates a new resource handler.
func NewResourceHandler(srv *Server) *resourceHandler {
	return &resourceHandler{server: srv}
}

// ListResources implements mcp.ResourceHandler.
func (h *resourceHandler) ListResources(ctx context.Context, req *mcp.ListResourcesRequest) (*mcp.ListResourcesResult, error) {
	// Query Notion database for resource-type entries
	pages, err := h.server.client.QueryDatabase(ctx, notion.NewTypeFilter(h.server.cfg.NotionTypeField, "resource"))
	if err != nil {
		return nil, err
	}

	var resources []*mcp.Resource
	for _, page := range pages {
		pageID := page.ID
		title := getPageTitle(page)
		resources = append(resources, &mcp.Resource{
			URI:         "notion://resource/" + pageID,
			Name:        title,
			Description: title,
		})
	}

	return &mcp.ListResourcesResult{
		Resources: resources,
	}, nil
}

// ReadResource implements mcp.ResourceHandler.
func (h *resourceHandler) ReadResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	// Extract page ID from URI
	// URI format: notion://resource/{pageID}
	pageID := extractPageIDFromURI(req.Params.URI)
	if pageID == "" {
		return nil, mcp.ResourceNotFoundError(req.Params.URI)
	}

	// Get page content
	content, err := h.server.client.GetPageContent(ctx, pageID)
	if err != nil {
		return nil, err
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI: req.Params.URI,
				Text: content.Text,
			},
		},
	}, nil
}

// extractPageIDFromURI extracts the page ID from a resource URI.
func extractPageIDFromURI(uri string) string {
	prefix := "notion://resource/"
	if len(uri) > len(prefix) {
		return uri[len(prefix):]
	}
	return ""
}
