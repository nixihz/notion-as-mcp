package server

import (
	"testing"

	"github.com/nixihz/notion-as-mcp/internal/notion"
)

func TestSanitizeToolName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "lowercase letters",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "uppercase converted to lowercase",
			input:    "HELLO",
			expected: "hello",
		},
		{
			name:     "spaces converted to underscores",
			input:    "hello world",
			expected: "hello_world",
		},
		{
			name:     "multiple spaces",
			input:    "hello  world",
			expected: "hello__world",
		},
		{
			name:  "special characters",
			input: "hello@world#test",
			// Note: Current code removes special chars instead of converting to underscores
			expected: "helloworldtest",
		},
		{
			name:     "alphanumeric",
			input:    "test123",
			expected: "test123",
		},
		{
			name:     "starts with number",
			input:    "123test",
			expected: "p_123test",
		},
		{
			name:     "starts with underscore",
			input:    "_test",
			expected: "p__test",
		},
		{
			name:     "starts with hyphen",
			input:    "-test",
			expected: "p_-test",
		},
		{
			name:     "mixed case with spaces",
			input:    "Hello World Test",
			expected: "hello_world_test",
		},
		{
			name:     "tabs converted to underscores",
			input:    "hello\tworld",
			expected: "hello_world",
		},
		{
			name:     "keeps underscores and hyphens",
			input:    "hello_world-test",
			expected: "hello_world-test",
		},
		{
			name:  "only special characters",
			input: "@#$%",
			// Note: Current code returns empty string when all chars are removed
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: "",
		},
		{
			name:     "already valid name",
			input:    "my_tool-name123",
			expected: "my_tool-name123",
		},
		{
			name:     "Chinese characters",
			input:    "测试工具",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeToolName(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeToolName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetPageTitle(t *testing.T) {
	tests := []struct {
		name     string
		page     notion.Page
		expected string
	}{
		{
			name:     "empty properties returns page ID",
			page:     notion.Page{ID: "test-page-id"},
			expected: "test-page-id",
		},
		{
			name: "page with Name title",
			page: notion.Page{
				ID: "page-123",
				Properties: map[string]notion.Property{
					"Name": {
						Type: notion.PropertyTypeTitle,
						Title: []notion.Title{
							{PlainText: "My Page Title"},
						},
					},
				},
			},
			expected: "My Page Title",
		},
		{
			name: "page with empty title returns page ID",
			page: notion.Page{
				ID: "page-456",
				Properties: map[string]notion.Property{
					"Name": {
						Type:  notion.PropertyTypeTitle,
						Title: []notion.Title{},
					},
				},
			},
			expected: "page-456",
		},
		{
			name: "page with no Name property returns page ID",
			page: notion.Page{
				ID: "page-789",
				Properties: map[string]notion.Property{
					"Description": {
						Type:     notion.PropertyTypeRichText,
						RichText: []notion.RichText{{PlainText: "Description"}},
					},
				},
			},
			expected: "page-789",
		},
		{
			name: "page with Title containing empty string",
			page: notion.Page{
				ID: "page-empty",
				Properties: map[string]notion.Property{
					"Name": {
						Type: notion.PropertyTypeTitle,
						Title: []notion.Title{
							{PlainText: ""},
						},
					},
				},
			},
			expected: "page-empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPageTitle(tt.page)
			if result != tt.expected {
				t.Errorf("getPageTitle() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetPageDescription(t *testing.T) {
	tests := []struct {
		name     string
		page     notion.Page
		expected string
	}{
		{
			name:     "empty properties returns empty",
			page:     notion.Page{ID: "test-page-id"},
			expected: "",
		},
		{
			name: "page with Description",
			page: notion.Page{
				ID: "page-123",
				Properties: map[string]notion.Property{
					"Description": {
						Type: notion.PropertyTypeRichText,
						RichText: []notion.RichText{
							{PlainText: "This is a description"},
						},
					},
				},
			},
			expected: "This is a description",
		},
		{
			name: "page with empty Description returns empty",
			page: notion.Page{
				ID: "page-456",
				Properties: map[string]notion.Property{
					"Description": {
						Type:     notion.PropertyTypeRichText,
						RichText: []notion.RichText{},
					},
				},
			},
			expected: "",
		},
		{
			name: "page with no Description property returns empty",
			page: notion.Page{
				ID: "page-789",
				Properties: map[string]notion.Property{
					"Name": {
						Type: notion.PropertyTypeTitle,
						Title: []notion.Title{
							{PlainText: "Page Title"},
						},
					},
				},
			},
			expected: "",
		},
		{
			name: "multiple rich text items returns first",
			page: notion.Page{
				ID: "page-multi",
				Properties: map[string]notion.Property{
					"Description": {
						Type: notion.PropertyTypeRichText,
						RichText: []notion.RichText{
							{PlainText: "First"},
							{PlainText: "Second"},
						},
					},
				},
			},
			expected: "First",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPageDescription(tt.page)
			if result != tt.expected {
				t.Errorf("getPageDescription() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractCodeString(t *testing.T) {
	tests := []struct {
		name      string
		richTexts []notion.RichText
		expected  string
	}{
		{
			name:      "empty slice",
			richTexts: []notion.RichText{},
			expected:  "",
		},
		{
			name: "single rich text",
			richTexts: []notion.RichText{
				{PlainText: "console.log('hello')"},
			},
			expected: "console.log('hello')",
		},
		{
			name: "multiple rich texts",
			richTexts: []notion.RichText{
				{PlainText: "def "},
				{PlainText: "hello():\n"},
				{PlainText: "    print('world')"},
			},
			expected: "def hello():\n    print('world')",
		},
		{
			name: "rich texts with empty strings",
			richTexts: []notion.RichText{
				{PlainText: "code"},
				{PlainText: ""},
				{PlainText: "more"},
			},
			expected: "codemore",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractCodeString(tt.richTexts)
			if result != tt.expected {
				t.Errorf("extractCodeString() = %q, want %q", result, tt.expected)
			}
		})
	}
}
