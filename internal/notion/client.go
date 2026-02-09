// Package notion provides Notion API client and data models.
package notion

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// Client is a Notion API client.
type Client struct {
	apiKey     string
	databaseID string
	typeField  string
	httpClient *http.Client
	baseURL    string
	apiVersion string
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

// QueryDatabase queries a Notion database and returns all pages.
// Handles pagination automatically.
func (c *Client) QueryDatabase(ctx context.Context) ([]Page, error) {
	url := fmt.Sprintf("%s/databases/%s/query", c.baseURL, c.databaseID)

	var allPages []Page
	var nextCursor *string

	for {
		// Build request body: empty object {} or with start_cursor for pagination
		reqBody := map[string]interface{}{}
		if nextCursor != nil {
			reqBody["start_cursor"] = *nextCursor
		}

		body, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("marshal query: %w", err)
		}

		type queryResponse struct {
			Results    []Page  `json:"results"`
			HasMore    bool    `json:"has_more"`
			NextCursor *string `json:"next_cursor"`
		}

		var resp queryResponse
		err = c.doRequest(ctx, "POST", url, bytes.NewReader(body), &resp)
		if err != nil {
			return nil, err
		}

		allPages = append(allPages, resp.Results...)

		// Stop if no more pages
		if !resp.HasMore {
			break
		}

		// Stop if next_cursor is nil (shouldn't happen if has_more is true, but safety check)
		if resp.NextCursor == nil {
			break
		}

		nextCursor = resp.NextCursor
	}

	return allPages, nil
}

// GetAllPages retrieves all pages from the database without filtering.
func (c *Client) GetAllPages(ctx context.Context) ([]Page, error) {
	return c.QueryDatabase(ctx)
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

// isRetryableError checks if the error is a transient network error worth retrying.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// Retry on broken pipe, connection reset, EOF
	for _, s := range []string{"broken pipe", "connection reset", "EOF", "use of closed network connection"} {
		if contains(errStr, s) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// doRequest performs an HTTP request with retry logic.
func (c *Client) doRequest(ctx context.Context, method, url string, body io.Reader, response interface{}) error {
	maxRetries := 3
	backoff := time.Second

	// Buffer the body so it can be re-read on retries
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = io.ReadAll(body)
		if err != nil {
			return fmt.Errorf("read request body: %w", err)
		}
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		var reqBody io.Reader
		if bodyBytes != nil {
			reqBody = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Notion-Version", c.apiVersion)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			// Retry on transient network errors (broken pipe, connection reset, etc.)
			if isRetryableError(err) && attempt < maxRetries-1 {
				slog.Warn("retrying request due to network error",
					"attempt", attempt+1,
					"error", err.Error(),
					"url", url,
				)
				time.Sleep(backoff)
				backoff *= 2
				continue
			}
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
		// Read response body for decoding
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read response body: %w", err)
		}
		// Debug log for API response (only in debug mode)
		slog.Debug("notion API response", "status", resp.StatusCode, "body_size", len(respBody))

		if response != nil {
			if err := json.Unmarshal(respBody, response); err != nil {
				return fmt.Errorf("decode response: %w", err)
			}
		}

		return nil
	}

	return fmt.Errorf("max retries exceeded")
}
