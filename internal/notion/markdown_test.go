package notion

import (
	"testing"
)

func TestMarkdownConverter_NewMarkdownConverter(t *testing.T) {
	pageContent := &PageContent{
		Page: Page{ID: "test-page-id"},
		Blocks: []Block{
			{
				Type: BlockTypeParagraph,
				Content: Paragraph{
					RichText: []RichText{
						{PlainText: "Hello"},
					},
				},
			},
		},
		Text: "Hello",
	}

	converter := NewMarkdownConverter(pageContent)

	if converter.Page != pageContent {
		t.Error("converter.Page should be set to the input pageContent")
	}
	if converter.Buf == nil {
		t.Error("converter.Buf should be initialized")
	}
}

func TestMarkdownConverter_WriteString(t *testing.T) {
	converter := NewMarkdownConverter(&PageContent{})
	converter.WriteString("test")
	converter.WriteString(" string")

	result := converter.Buf.String()
	if result != "test string" {
		t.Errorf("WriteString() = %q, want %q", result, "test string")
	}
}

func TestMarkdownConverter_Newline(t *testing.T) {
	converter := NewMarkdownConverter(&PageContent{})
	converter.WriteString("text")
	converter.Newline()

	result := converter.Buf.String()
	expected := "text\n\n"
	if result != expected {
		t.Errorf("Newline() = %q, want %q", result, expected)
	}
}

func TestMarkdownConverter_Eol(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*MarkdownConverter)
		expected string
	}{
		{
			name:     "adds newline when no trailing newline",
			setup:    func(c *MarkdownConverter) { c.WriteString("text") },
			expected: "text\n",
		},
		{
			name:     "does not add newline when already has trailing newline",
			setup:    func(c *MarkdownConverter) { c.WriteString("text\n") },
			expected: "text\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMarkdownConverter(&PageContent{})
			tt.setup(converter)
			converter.Eol()
			if converter.Buf.String() != tt.expected {
				t.Errorf("Eol() = %q, want %q", converter.Buf.String(), tt.expected)
			}
		})
	}
}

func TestMarkdownConverter_RenderRichText(t *testing.T) {
	tests := []struct {
		name      string
		richTexts []RichText
		expected  string
	}{
		{
			name:      "empty rich text",
			richTexts: []RichText{},
			expected:  "",
		},
		{
			name: "plain text",
			richTexts: []RichText{
				{PlainText: "Hello World"},
			},
			expected: "Hello World",
		},
		{
			name: "bold text",
			richTexts: []RichText{
				{PlainText: "bold", Annotations: Annotations{Bold: true}},
			},
			expected: "**bold**",
		},
		{
			name: "italic text",
			richTexts: []RichText{
				{PlainText: "italic", Annotations: Annotations{Italic: true}},
			},
			expected: "*italic*",
		},
		{
			name: "strikethrough text",
			richTexts: []RichText{
				{PlainText: "strike", Annotations: Annotations{Strikethrough: true}},
			},
			expected: "~~strike~~",
		},
		{
			name: "code text",
			richTexts: []RichText{
				{PlainText: "code", Annotations: Annotations{Code: true}},
			},
			expected: "`code`",
		},
		{
			name: "link with text object",
			richTexts: []RichText{
				{PlainText: "link", Text: Text{Link: &Link{URL: "https://example.com"}}},
			},
			expected: "[link](https://example.com)",
		},
		{
			name: "link with href",
			richTexts: []RichText{
				{Href: func(s string) *string { return &s }("https://example.com"), PlainText: "href link"},
			},
			expected: "[href link](https://example.com)",
		},
		{
			name: "multiple formatting",
			richTexts: []RichText{
				{PlainText: "bold italic", Annotations: Annotations{Bold: true, Italic: true}},
			},
			// Note: Code processes annotations sequentially, so Bold+Italic produces ***
			expected: "***bold italic***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMarkdownConverter(&PageContent{})
			result := converter.RenderRichText(tt.richTexts)
			if result != tt.expected {
				t.Errorf("RenderRichText() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestMarkdownConverter_RenderParagraph(t *testing.T) {
	converter := NewMarkdownConverter(&PageContent{})

	block := Block{
		Type: BlockTypeParagraph,
		Content: Paragraph{
			RichText: []RichText{
				{PlainText: "Paragraph text"},
			},
		},
	}
	converter.RenderParagraph(block)

	result := converter.Buf.String()
	if result != "Paragraph text\n\n" {
		t.Errorf("RenderParagraph() = %q, want %q", result, "Paragraph text\n\n")
	}
}

func TestMarkdownConverter_RenderHeading(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		level    int
		expected string
	}{
		{"heading 1", "Heading 1", 1, "# Heading 1\n\n"},
		{"heading 2", "Heading 2", 2, "## Heading 2\n\n"},
		{"heading 3", "Heading 3", 3, "### Heading 3\n\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMarkdownConverter(&PageContent{})
			block := Block{
				Type: BlockTypeHeading1,
				Content: map[string]any{
					"rich_text": []any{
						map[string]any{"plain_text": tt.content},
					},
				},
			}
			converter.RenderHeading(block, tt.level)
			if converter.Buf.String() != tt.expected {
				t.Errorf("RenderHeading() = %q, want %q", converter.Buf.String(), tt.expected)
			}
		})
	}
}

func TestMarkdownConverter_RenderBulletedList(t *testing.T) {
	converter := NewMarkdownConverter(&PageContent{})
	block := Block{
		Type: BlockTypeBulletedListItem,
		Content: map[string]any{
			"rich_text": []any{
				map[string]any{"plain_text": "List item"},
			},
		},
	}
	converter.RenderBulletedList(block)

	result := converter.Buf.String()
	if result != "- List item\n" {
		t.Errorf("RenderBulletedList() = %q, want %q", result, "- List item\n")
	}
}

func TestMarkdownConverter_RenderNumberedList(t *testing.T) {
	converter := NewMarkdownConverter(&PageContent{})
	block := Block{
		Type: BlockTypeNumberedListItem,
		Content: map[string]any{
			"rich_text": []any{
				map[string]any{"plain_text": "Numbered item"},
			},
		},
	}
	converter.RenderNumberedList(block, 1)

	result := converter.Buf.String()
	if result != "1. Numbered item\n" {
		t.Errorf("RenderNumberedList() = %q, want %q", result, "1. Numbered item\n")
	}
}

func TestMarkdownConverter_RenderCode(t *testing.T) {
	tests := []struct {
		name     string
		block    Block
		expected string
	}{
		{
			name: "code block with CodeBlock type",
			block: Block{
				Type: BlockTypeCode,
				Content: CodeBlock{
					Language: "python",
					Code: []RichText{
						{PlainText: "print('hello')"},
					},
				},
			},
			expected: "```python\nprint('hello')\n```\n\n",
		},
		{
			name: "code block with map content",
			block: Block{
				Type: BlockTypeCode,
				Content: map[string]any{
					"language": "javascript",
					"rich_text": []any{
						map[string]any{"plain_text": "console.log('test')"},
					},
				},
			},
			expected: "```javascript\nconsole.log('test')\n```\n\n",
		},
		{
			name: "code block with empty language defaults to text",
			block: Block{
				Type: BlockTypeCode,
				Content: CodeBlock{
					Language: "",
					Code:     []RichText{{PlainText: "code"}},
				},
			},
			expected: "```text\ncode\n```\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMarkdownConverter(&PageContent{})
			converter.RenderCode(tt.block)
			if converter.Buf.String() != tt.expected {
				t.Errorf("RenderCode() = %q, want %q", converter.Buf.String(), tt.expected)
			}
		})
	}
}

func TestMarkdownConverter_RenderQuote(t *testing.T) {
	converter := NewMarkdownConverter(&PageContent{})
	block := Block{
		Type: BlockTypeQuote,
		Content: map[string]any{
			"rich_text": []any{
				map[string]any{"plain_text": "A quote"},
			},
		},
	}
	converter.RenderQuote(block)

	result := converter.Buf.String()
	// Note: Quote renders with an extra newline after Eol + Newline
	if result != "> A quote\n\n\n" {
		t.Errorf("RenderQuote() = %q, want %q", result, "> A quote\n\n\n")
	}
}

func TestMarkdownConverter_RenderDivider(t *testing.T) {
	converter := NewMarkdownConverter(&PageContent{})
	block := Block{Type: BlockTypeDivider}
	converter.RenderDivider(block)

	result := converter.Buf.String()
	if result != "---\n\n" {
		t.Errorf("RenderDivider() = %q, want %q", result, "---\n\n")
	}
}

func TestMarkdownConverter_RenderToDo(t *testing.T) {
	tests := []struct {
		name     string
		block    Block
		expected string
	}{
		{
			name: "unchecked todo",
			block: Block{
				Type: BlockTypeToDo,
				Content: map[string]any{
					"checked": false,
					"rich_text": []any{
						map[string]any{"plain_text": "Task"},
					},
				},
			},
			expected: "- [ ] Task\n",
		},
		{
			name: "checked todo",
			block: Block{
				Type: BlockTypeToDo,
				Content: map[string]any{
					"checked": true,
					"rich_text": []any{
						map[string]any{"plain_text": "Done"},
					},
				},
			},
			expected: "- [x] Done\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMarkdownConverter(&PageContent{})
			converter.RenderToDo(tt.block)
			if converter.Buf.String() != tt.expected {
				t.Errorf("RenderToDo() = %q, want %q", converter.Buf.String(), tt.expected)
			}
		})
	}
}

func TestMarkdownConverter_RenderCallout(t *testing.T) {
	converter := NewMarkdownConverter(&PageContent{})
	block := Block{
		Type: BlockTypeCallout,
		Content: map[string]any{
			"rich_text": []any{
				map[string]any{"plain_text": "A tip"},
			},
		},
	}
	converter.RenderCallout(block)

	result := converter.Buf.String()
	if result != "> 💡 A tip\n\n" {
		t.Errorf("RenderCallout() = %q, want %q", result, "> 💡 A tip\n\n")
	}
}

func TestMarkdownConverter_RenderImage(t *testing.T) {
	tests := []struct {
		name     string
		block    Block
		expected string
	}{
		{
			name: "image with caption",
			block: Block{
				Type: BlockTypeImage,
				Content: map[string]any{
					"file": map[string]any{
						"url": "https://example.com/image.png",
					},
					"caption": []any{
						map[string]any{"plain_text": "My Image"},
					},
				},
			},
			expected: "![My Image](https://example.com/image.png)\n\n",
		},
		{
			name: "image without caption",
			block: Block{
				Type: BlockTypeImage,
				Content: map[string]any{
					"file": map[string]any{
						"url": "https://example.com/image.png",
					},
				},
			},
			expected: "![](https://example.com/image.png)\n\n",
		},
		{
			name:     "empty content",
			block:    Block{Type: BlockTypeImage},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMarkdownConverter(&PageContent{})
			converter.RenderImage(tt.block)
			if converter.Buf.String() != tt.expected {
				t.Errorf("RenderImage() = %q, want %q", converter.Buf.String(), tt.expected)
			}
		})
	}
}

func TestMarkdownConverter_RenderBlock(t *testing.T) {
	converter := NewMarkdownConverter(&PageContent{})

	// Test paragraph
	paragraphBlock := Block{
		Type: BlockTypeParagraph,
		Content: Paragraph{
			RichText: []RichText{{PlainText: "Para"}},
		},
	}
	converter.RenderBlock(paragraphBlock, nil)
	if converter.Buf.Len() == 0 {
		t.Error("RenderBlock should write to buffer for paragraph")
	}
}

func TestMarkdownConverter_ToMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		page     *PageContent
		expected string
	}{
		{
			name:     "nil page",
			page:     nil,
			expected: "",
		},
		{
			name: "empty blocks",
			page: &PageContent{
				Blocks: []Block{},
			},
			expected: "",
		},
		{
			name: "single paragraph",
			page: &PageContent{
				Blocks: []Block{
					{
						Type: BlockTypeParagraph,
						Content: Paragraph{
							RichText: []RichText{{PlainText: "Hello"}},
						},
					},
				},
			},
			expected: "Hello",
		},
		{
			name: "multiple blocks",
			page: &PageContent{
				Blocks: []Block{
					{
						Type: BlockTypeHeading1,
						Content: map[string]any{
							"rich_text": []any{
								map[string]any{"plain_text": "Title"},
							},
						},
					},
					{
						Type: BlockTypeParagraph,
						Content: Paragraph{
							RichText: []RichText{{PlainText: "Content"}},
						},
					},
				},
			},
			expected: "# Title\n\nContent",
		},
		{
			name: "numbered list items",
			page: &PageContent{
				Blocks: []Block{
					{
						Type: BlockTypeNumberedListItem,
						Content: map[string]any{
							"rich_text": []any{
								map[string]any{"plain_text": "First"},
							},
						},
					},
					{
						Type: BlockTypeNumberedListItem,
						Content: map[string]any{
							"rich_text": []any{
								map[string]any{"plain_text": "Second"},
							},
						},
					},
				},
			},
			expected: "1. First\n2. Second",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMarkdownConverter(tt.page)
			result := converter.ToMarkdown()
			if result != tt.expected {
				t.Errorf("ToMarkdown() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestPageToMarkdown(t *testing.T) {
	pageContent := &PageContent{
		Blocks: []Block{
			{
				Type: BlockTypeParagraph,
				Content: Paragraph{
					RichText: []RichText{{PlainText: "Test"}},
				},
			},
		},
	}

	result := PageToMarkdown(pageContent)
	expected := "Test"
	if result != expected {
		t.Errorf("PageToMarkdown() = %q, want %q", result, expected)
	}
}

func TestMarkdownConverter_extractRichTexts(t *testing.T) {
	converter := NewMarkdownConverter(&PageContent{})

	tests := []struct {
		name     string
		content  any
		expected int
	}{
		{
			name:     "nil content",
			content:  nil,
			expected: 0,
		},
		{
			name:     "empty slice",
			content:  []RichText{},
			expected: 0,
		},
		{
			name: "rich text slice",
			content: []RichText{
				{PlainText: "One"},
				{PlainText: "Two"},
			},
			expected: 2,
		},
		{
			name: "paragraph type",
			content: Paragraph{
				RichText: []RichText{{PlainText: "Para"}},
			},
			expected: 1,
		},
		{
			name: "map with rich_text",
			content: map[string]any{
				"rich_text": []any{
					map[string]any{"plain_text": "Text"},
				},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.extractRichTexts(tt.content)
			if len(result) != tt.expected {
				t.Errorf("extractRichTexts() returned %d items, want %d", len(result), tt.expected)
			}
		})
	}
}
