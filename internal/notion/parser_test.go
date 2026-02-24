package notion

import (
	"testing"
)

func TestExtractText(t *testing.T) {
	tests := []struct {
		name     string
		blocks   []Block
		expected string
	}{
		{
			name:     "empty blocks",
			blocks:   []Block{},
			expected: "",
		},
		{
			name: "paragraph block",
			blocks: []Block{
				{
					Type: BlockTypeParagraph,
					Content: map[string]any{
						"rich_text": []any{
							map[string]any{"plain_text": "Hello World"},
						},
					},
				},
			},
			expected: "Hello World",
		},
		{
			name: "heading blocks",
			blocks: []Block{
				{
					Type: BlockTypeHeading1,
					Content: map[string]any{
						"rich_text": []any{
							map[string]any{"plain_text": "Heading 1"},
						},
					},
				},
				{
					Type: BlockTypeHeading2,
					Content: map[string]any{
						"rich_text": []any{
							map[string]any{"plain_text": "Heading 2"},
						},
					},
				},
			},
			// Note: Current code doesn't add "#" prefix for headings in ExtractText
			expected: "Heading 1\nHeading 2",
		},
		{
			name: "bulleted list item",
			blocks: []Block{
				{
					Type: BlockTypeBulletedListItem,
					Content: map[string]any{
						"rich_text": []any{
							map[string]any{"plain_text": "List item"},
						},
					},
				},
			},
			expected: "- List item",
		},
		{
			name: "numbered list item",
			blocks: []Block{
				{
					Type: BlockTypeNumberedListItem,
					Content: map[string]any{
						"rich_text": []any{
							map[string]any{"plain_text": "Numbered item"},
						},
					},
				},
			},
			expected: "1. Numbered item",
		},
		{
			name: "code block",
			blocks: []Block{
				{
					Type: BlockTypeCode,
					Content: CodeBlock{
						Language: "python",
						Code: []RichText{
							{PlainText: "print('hello')"},
						},
					},
				},
			},
			expected: "```python\nprint('hello')\n```",
		},
		{
			name: "to-do unchecked",
			blocks: []Block{
				{
					Type: BlockTypeToDo,
					Content: map[string]any{
						"checked": false,
						"rich_text": []any{
							map[string]any{"plain_text": "Task"},
						},
					},
				},
			},
			expected: "- [ ] Task",
		},
		{
			name: "to-do checked",
			blocks: []Block{
				{
					Type: BlockTypeToDo,
					Content: map[string]any{
						"checked": true,
						"rich_text": []any{
							map[string]any{"plain_text": "Completed Task"},
						},
					},
				},
			},
			expected: "- [x] Completed Task",
		},
		{
			name: "quote block",
			blocks: []Block{
				{
					Type: BlockTypeQuote,
					Content: map[string]any{
						"rich_text": []any{
							map[string]any{"plain_text": "A quote"},
						},
					},
				},
			},
			expected: "> A quote",
		},
		{
			name: "divider block",
			blocks: []Block{
				{Type: BlockTypeDivider},
			},
			expected: "---",
		},
		{
			name: "callout block",
			blocks: []Block{
				{
					Type: BlockTypeCallout,
					Content: map[string]any{
						"rich_text": []any{
							map[string]any{"plain_text": "A tip"},
						},
					},
				},
			},
			expected: "💡 A tip",
		},
		{
			name: "multiple blocks",
			blocks: []Block{
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
					Content: map[string]any{
						"rich_text": []any{
							map[string]any{"plain_text": "Paragraph text"},
						},
					},
				},
			},
			// Note: Current code doesn't add "#" prefix for headings in ExtractText
			expected: "Title\nParagraph text",
		},
		{
			name: "unknown block type",
			blocks: []Block{
				{Type: "unknown"},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractText(tt.blocks)
			if result != tt.expected {
				t.Errorf("ExtractText() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetTypeFromProperties(t *testing.T) {
	tests := []struct {
		name       string
		properties map[string]Property
		typeField  string
		expected   string
	}{
		{
			name:       "empty properties",
			properties: map[string]Property{},
			typeField:  "Type",
			expected:   "",
		},
		{
			name: "matching select property",
			properties: map[string]Property{
				"Type": {
					Type: PropertyTypeSelect,
					Select: &Select{
						Name: "resource",
					},
				},
			},
			typeField: "Type",
			expected:  "resource",
		},
		{
			name: "different type field name",
			properties: map[string]Property{
				"Category": {
					Type: PropertyTypeSelect,
					Select: &Select{
						Name: "tool",
					},
				},
			},
			typeField: "Category",
			expected:  "tool",
		},
		{
			name: "type field not found",
			properties: map[string]Property{
				"Name": {
					Type: PropertyTypeTitle,
				},
			},
			typeField: "Type",
			expected:  "",
		},
		{
			name: "property is not select type",
			properties: map[string]Property{
				"Type": {
					Type: PropertyTypeTitle,
				},
			},
			typeField: "Type",
			expected:  "",
		},
		{
			name: "select is nil",
			properties: map[string]Property{
				"Type": {
					Type:   PropertyTypeSelect,
					Select: nil,
				},
			},
			typeField: "Type",
			expected:  "",
		},
		{
			name: "multiple properties",
			properties: map[string]Property{
				"Name": {
					Type: PropertyTypeTitle,
				},
				"Type": {
					Type: PropertyTypeSelect,
					Select: &Select{
						Name: "prompt",
					},
				},
				"Description": {
					Type: PropertyTypeRichText,
				},
			},
			typeField: "Type",
			expected:  "prompt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTypeFromProperties(tt.properties, tt.typeField)
			if result != tt.expected {
				t.Errorf("GetTypeFromProperties() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseCodeBlock(t *testing.T) {
	tests := []struct {
		name     string
		block    Block
		wantOk   bool
		wantCode CodeBlock
	}{
		{
			name: "valid code block",
			block: Block{
				Type: BlockTypeCode,
				Content: map[string]any{
					"language": "python",
					"rich_text": []any{
						map[string]any{
							"plain_text": "print('hello')",
						},
					},
				},
			},
			wantOk: true,
			wantCode: CodeBlock{
				Language: "python",
				Code: []RichText{
					{PlainText: "print('hello')"},
				},
			},
		},
		{
			name: "not a code block",
			block: Block{
				Type: BlockTypeParagraph,
			},
			wantOk:   false,
			wantCode: CodeBlock{},
		},
		{
			name: "content is not map",
			block: Block{
				Type:    BlockTypeCode,
				Content: "not a map",
			},
			wantOk:   false,
			wantCode: CodeBlock{},
		},
		{
			name: "empty content",
			block: Block{
				Type:    BlockTypeCode,
				Content: map[string]any{},
			},
			wantOk:   true,
			wantCode: CodeBlock{},
		},
		{
			name: "code block with text content",
			block: Block{
				Type: BlockTypeCode,
				Content: map[string]any{
					"language": "javascript",
					"rich_text": []any{
						map[string]any{
							"plain_text": "console.log('test')",
							"text": map[string]any{
								"content": "console.log('test')",
							},
						},
					},
				},
			},
			wantOk: true,
			wantCode: CodeBlock{
				Language: "javascript",
				Code: []RichText{
					{
						PlainText: "console.log('test')",
						Text:      Text{Content: "console.log('test')"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := ParseCodeBlock(tt.block)
			if ok != tt.wantOk {
				t.Errorf("ParseCodeBlock() ok = %v, want %v", ok, tt.wantOk)
			}
			if result.Language != tt.wantCode.Language {
				t.Errorf("ParseCodeBlock().Language = %q, want %q", result.Language, tt.wantCode.Language)
			}
			if len(result.Code) != len(tt.wantCode.Code) {
				t.Errorf("ParseCodeBlock().Code length = %d, want %d", len(result.Code), len(tt.wantCode.Code))
			}
		})
	}
}

func TestExtractRichText(t *testing.T) {
	tests := []struct {
		name     string
		content  any
		expected string
	}{
		{
			name:     "nil content",
			content:  nil,
			expected: "",
		},
		{
			name:     "empty slice",
			content:  []RichText{},
			expected: "",
		},
		{
			name: "rich text slice",
			content: []RichText{
				{PlainText: "Hello "},
				{PlainText: "World"},
			},
			expected: "Hello World",
		},
		{
			name: "map with rich_text array",
			content: map[string]any{
				"rich_text": []any{
					map[string]any{"plain_text": "Test "},
					map[string]any{"plain_text": "Text"},
				},
			},
			expected: "Test Text",
		},
		{
			name: "map without rich_text",
			content: map[string]any{
				"other": "value",
			},
			expected: "",
		},
		{
			name: "Paragraph type",
			content: Paragraph{
				RichText: []RichText{
					{PlainText: "Para"},
					{PlainText: "graph"},
				},
			},
			// Note: extractRichText doesn't handle Paragraph type as content
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractRichText(tt.content)
			if result != tt.expected {
				t.Errorf("extractRichText() = %q, want %q", result, tt.expected)
			}
		})
	}
}
