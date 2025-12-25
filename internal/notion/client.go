// Package notion provides Notion API client and data models.
package notion

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is a Notion API client.
type Client struct {
	apiKey      string
	databaseID  string
	typeField   string
	httpClient  *http.Client
	baseURL     string
	apiVersion  string
}

// NewClient creates a new Notion API client.
func NewClient(apiKey, databaseID, typeField string) *Client {
	return &Client{
		apiKey:     apiKey,
		databaseID: databaseID,
		typeField:  typeField,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:    "https://api.notion.com/v1",
		apiVersion: "2022-06-28",
	}
}

// QueryDatabase queries a Notion database and returns pages.
func (c *Client) QueryDatabase(ctx context.Context, filter *DatabaseQueryFilter) ([]Page, error) {
	url := fmt.Sprintf("%s/databases/%s/query", c.baseURL, c.databaseID)

	var reqBody interface{} = nil
	if filter != nil {
		reqBody = filter
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal query: %w", err)
	}

	var pages []Page
	err = c.doRequest(ctx, "POST", url, bytes.NewReader(body), &pages)
	if err != nil {
		return nil, err
	}

	return pages, nil
}

// GetPage retrieves a single page by ID.
func (c *Client) GetPage(ctx context.Context, pageID string) (*Page, error) {
	url := fmt.Sprintf("%s/pages/%s", c.baseURL, pageID)

	var page Page
	err := c.doRequest(ctx, "GET", url, nil, &page)
	if err != nil {
		return nil, err
	}

	return &page, nil
}

// GetBlockChildren retrieves the children blocks of a page.
func (c *Client) GetBlockChildren(ctx context.Context, blockID string) ([]Block, error) {
	url := fmt.Sprintf("%s/blocks/%s/children", c.baseURL, blockID)

	type response struct {
		Results []Block `json:"results"`
	}

	var resp response
	err := c.doRequest(ctx, "GET", url, nil, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Results, nil
}

// GetPageContent retrieves a page with its content blocks.
func (c *Client) GetPageContent(ctx context.Context, pageID string) (*PageContent, error) {
	page, err := c.GetPage(ctx, pageID)
	if err != nil {
		return nil, err
	}

	blocks, err := c.GetBlockChildren(ctx, pageID)
	if err != nil {
		return nil, err
	}

	pc := &PageContent{
		Page:   *page,
		Blocks: blocks,
		Text:   ExtractText(blocks),
	}

	// Check for code block
	for _, block := range blocks {
		if block.Type == BlockTypeCode {
			pc.HasCode = true
			pc.Code = block.Content.(CodeBlock)
			break
		}
	}

	return pc, nil
}

// doRequest performs an HTTP request with retry logic.
func (c *Client) doRequest(ctx context.Context, method, url string, body io.Reader, response interface{}) error {
	maxRetries := 3
	backoff := time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, url, body)
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Notion-Version", c.apiVersion)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		// Handle rate limiting
		if resp.StatusCode == 429 {
			retryAfter := resp.Header.Get("Retry-After")
			waitTime := backoff
			if retryAfter != "" {
				waitDur, err := time.ParseDuration(retryAfter + "s")
				if err == nil {
					waitTime = waitDur
				}
			}
			time.Sleep(waitTime)
			backoff *= 2
			continue
		}

		if resp.StatusCode >= 400 {
			var errResp struct {
				Message string `json:"message"`
				Code    string `json:"code"`
			}
			json.NewDecoder(resp.Body).Decode(&errResp)
			return fmt.Errorf("notion API error: %s (%s)", errResp.Message, errResp.Code)
		}

		if response != nil {
			if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
				return fmt.Errorf("decode response: %w", err)
			}
		}

		return nil
	}

	return fmt.Errorf("max retries exceeded")
}

// DatabaseQueryFilter represents a filter for database queries.
type DatabaseQueryFilter struct {
	Filter Filter `json:"filter"`
}

// Filter represents a Notion filter.
type Filter struct {
	Property string      `json:"property"`
	Select   SelectFilter `json:"select"`
	Date     DateFilter   `json:"date"`
}

// SelectFilter filters by select property.
type SelectFilter struct {
	Equals string `json:"equals"`
}

// DateFilter filters by date property.
type DateFilter struct {
	After string `json:"after"`
}

// NewTypeFilter creates a filter for a specific type.
func NewTypeFilter(typeField, typeValue string) *DatabaseQueryFilter {
	return &DatabaseQueryFilter{
		Filter: Filter{
			Property: typeField,
			Select: SelectFilter{
				Equals: typeValue,
			},
		},
	}
}

// NewUpdatedSinceFilter creates a filter for pages updated after a time.
func NewUpdatedSinceFilter(since time.Time) *DatabaseQueryFilter {
	return &DatabaseQueryFilter{
		Filter: Filter{
			Property: "last_edited_time",
			Date: DateFilter{
				After: since.Format(time.RFC3339),
			},
		},
	}
}
