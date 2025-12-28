// Package notion provides Notion API client and data models.
package notion

import (
	"encoding/json"
	"time"
)

// Page represents a Notion page/database entry.
type Page struct {
	ID             string              `json:"id"`
	CreatedTime    time.Time           `json:"created_time"`
	LastEditedTime time.Time           `json:"last_edited_time"`
	Properties     map[string]Property `json:"properties"`
	Content        []Block             `json:"content,omitempty"`
}

// Property represents a Notion property.
type Property struct {
	Name     string       `json:"name"`
	Type     PropertyType `json:"type"`
	Value    any          `json:"value"`
	Select   *Select      `json:"select"`
	Title    []Title      `json:"title"`
	RichText []RichText   `json:"rich_text"`
}

/*
*
"select": {
"id": "6b3883a3-56f2-4943-a2cc-761308de58ca",
"name": "resource",
"color": "orange"
}*
*/
type Select struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type Title struct {
	Type        string      `json:"type"`
	Text        Text        `json:"text"`
	Annotations Annotations `json:"annotations"`
	PlainText   string      `json:"plain_text"`
	Href        string      `json:"href"`
}

type Text struct {
	Content string `json:"content"`
	Link    *Link  `json:"link"`
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

type Block struct {
	Object         string     `json:"object"`
	ID             string     `json:"id"`
	Type           BlockType  `json:"type"`
	Content        any        `json:"-"` // Populated from type-specific fields during unmarshal
	Parent         *Parent    `json:"parent"`
	CreatedTime    time.Time  `json:"created_time"`
	LastEditedTime time.Time  `json:"last_edited_time"`
	CreatedBy      *User      `json:"created_by"`
	LastEditedBy   *User      `json:"last_edited_by"`
	HasChildren    bool       `json:"has_children"`
	Archived       bool       `json:"archived"`
	InTrash        bool       `json:"in_trash"`
	Paragraph      *Paragraph `json:"paragraph,omitempty"`
}

// UnmarshalJSON implements custom JSON unmarshaling to populate Content field.
func (b *Block) UnmarshalJSON(data []byte) error {
	// First, unmarshal into a map to get the type
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Unmarshal standard fields
	type Alias Block
	aux := (*Alias)(b)
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Populate Content based on type
	switch b.Type {
	case BlockTypeParagraph:
		if paraData, ok := raw["paragraph"]; ok {
			var para Paragraph
			if err := json.Unmarshal(paraData, &para); err == nil {
				b.Content = para
				b.Paragraph = &para
			} else {
				b.Content = paraData
			}
		}
	case BlockTypeCode:
		if codeData, ok := raw["code"]; ok {
			var codeBlock CodeBlock
			if err := json.Unmarshal(codeData, &codeBlock); err == nil {
				b.Content = codeBlock
			} else {
				b.Content = codeData
			}
		}
	default:
		// For other types, store the type-specific field as map
		if typeData, ok := raw[string(b.Type)]; ok {
			var typeContent map[string]any
			if err := json.Unmarshal(typeData, &typeContent); err == nil {
				b.Content = typeContent
			} else {
				b.Content = typeData
			}
		}
	}

	return nil
}

type Paragraph struct {
	RichText []RichText `json:"rich_text"`
	Color    string     `json:"color"`
}

type Parent struct {
	Type   string `json:"type"`
	PageID string `json:"page_id"`
}

type User struct {
	Object string `json:"object"`
	ID     string `json:"id"`
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
	RichText []RichText `json:"rich_text"`
}

// RichText represents rich text in Notion.
type RichText struct {
	Type        string      `json:"type"`
	Text        Text        `json:"text"`
	Annotations Annotations `json:"annotations"`
	PlainText   string      `json:"plain_text"`
	Href        *string     `json:"href"`
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
