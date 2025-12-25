// Package notion provides Notion API client and data models.
package notion

import "time"

// Page represents a Notion page/database entry.
type Page struct {
	ID             string                 `json:"id"`
	CreatedTime    time.Time              `json:"created_time"`
	LastEditedTime time.Time              `json:"last_edited_time"`
	Properties     map[string]Property    `json:"properties"`
	Content        []Block                `json:"content,omitempty"`
}

// Property represents a Notion property.
type Property struct {
	Name  string       `json:"name"`
	Type  PropertyType `json:"type"`
	Value any          `json:"value"`
}

// PropertyType represents the type of a Notion property.
type PropertyType string

const (
	PropertyTypeTitle       PropertyType = "title"
	PropertyTypeRichText    PropertyType = "rich_text"
	PropertyTypeSelect      PropertyType = "select"
	PropertyTypeMultiSelect PropertyType = "multi_select"
	PropertyTypeStatus      PropertyType = "status"
	PropertyTypeCheckbox    PropertyType = "checkbox"
	PropertyTypeDate        PropertyType = "date"
	PropertyTypeURL         PropertyType = "url"
	PropertyTypeEmail       PropertyType = "email"
	PropertyTypeNumber      PropertyType = "number"
)

// Block represents a Notion content block.
type Block struct {
	ID      string    `json:"id"`
	Type    BlockType `json:"type"`
	Content any       `json:"content"`
}

// BlockType represents the type of a Notion block.
type BlockType string

const (
	BlockTypeParagraph        BlockType = "paragraph"
	BlockTypeHeading1         BlockType = "heading_1"
	BlockTypeHeading2         BlockType = "heading_2"
	BlockTypeHeading3         BlockType = "heading_3"
	BlockTypeBulletedListItem BlockType = "bulleted_list_item"
	BlockTypeNumberedListItem BlockType = "numbered_list_item"
	BlockTypeCode             BlockType = "code"
	BlockTypeQuote            BlockType = "quote"
	BlockTypeDivider          BlockType = "divider"
	BlockTypeCallout          BlockType = "callout"
	BlockTypeImage            BlockType = "image"
)

// CodeBlock represents a code block content.
type CodeBlock struct {
	Language string     `json:"language"`
	Caption  []RichText `json:"caption"`
	Code     []RichText `json:"code"`
}

// RichText represents rich text in Notion.
type RichText struct {
	Type       string     `json:"type"`
	Content    string     `json:"content"`
	PlainText  string     `json:"plain_text"`
	Link       *Link      `json:"link,omitempty"`
	Annotations Annotations `json:"annotations,omitempty"`
}

// Link represents a hyperlink in rich text.
type Link struct {
	URL string `json:"url"`
}

// Annotations represents text formatting.
type Annotations struct {
	Bold          bool `json:"bold"`
	Italic        bool `json:"italic"`
	Strikethrough bool `json:"strikethrough"`
	Underline     bool `json:"underline"`
	Code          bool `json:"code"`
}

// PageContent represents a page with its content blocks.
type PageContent struct {
	Page    Page
	Blocks  []Block
	Text    string
	HasCode bool
	Code    CodeBlock
}
