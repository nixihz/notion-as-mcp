// Package notion provides Notion API client and data models.
package notion

import (
	"strings"
)

// ExtractText extracts plain text from a list of blocks.
func ExtractText(blocks []Block) string {
	var sb strings.Builder

	for _, block := range blocks {
		sb.WriteString(extractBlockText(block))
		sb.WriteRune('\n')
	}

	return strings.TrimSpace(sb.String())
}

// extractBlockText extracts text from a single block.
func extractBlockText(block Block) string {
	switch block.Type {
	case BlockTypeParagraph:
		return extractRichText(block.Content)
	case BlockTypeHeading1, BlockTypeHeading2, BlockTypeHeading3:
		return extractRichText(block.Content)
	case BlockTypeBulletedListItem:
		return "- " + extractRichText(block.Content)
	case BlockTypeNumberedListItem:
		return "1. " + extractRichText(block.Content)
	case BlockTypeCode:
		codeBlock, ok := block.Content.(CodeBlock)
		if ok {
			return "```" + codeBlock.Language + "\n" + extractRichText(codeBlock.Code) + "\n```"
		}
	case BlockTypeQuote:
		return "> " + extractRichText(block.Content)
	case BlockTypeDivider:
		return "---"
	case BlockTypeCallout:
		return "💡 " + extractRichText(block.Content)
	}
	return ""
}

// extractRichText extracts text from rich text array.
func extractRichText(content any) string {
	var texts []RichText

	switch v := content.(type) {
	case []RichText:
		texts = v
	case map[string]any:
		if rt, ok := v["rich_text"].([]any); ok {
			for _, r := range rt {
				if m, ok := r.(map[string]any); ok {
					rt := RichText{
						PlainText: getMapString(m, "plain_text"),
					}
					if textMap, ok := m["text"].(map[string]any); ok {
						rt.Text = Text{
							Content: getMapString(textMap, "content"),
						}
					}
					texts = append(texts, rt)
				}
			}
		}
	}

	var sb strings.Builder
	for _, text := range texts {
		sb.WriteString(text.PlainText)
	}
	return sb.String()
}

// getMapString gets a string value from a map.
func getMapString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetTypeFromProperties extracts the type value from page properties.
func GetTypeFromProperties(properties map[string]Property, typeField string) string {
	for name, prop := range properties {
		if name == typeField {
			if prop.Type == PropertyTypeSelect {
				return prop.Select.Name
			}
		}
	}
	return ""
}

// ParseCodeBlock parses a code block from content.
func ParseCodeBlock(block Block) (CodeBlock, bool) {
	if block.Type != BlockTypeCode {
		return CodeBlock{}, false
	}

	content, ok := block.Content.(map[string]any)
	if !ok {
		return CodeBlock{}, false
	}

	lang := getMapString(content, "language")

	var richTexts []RichText
	if rt, ok := content["rich_text"].([]any); ok {
		for _, r := range rt {
			if m, ok := r.(map[string]any); ok {
				rt := RichText{
					PlainText: getMapString(m, "plain_text"),
				}
				if textMap, ok := m["text"].(map[string]any); ok {
					rt.Text = Text{
						Content: getMapString(textMap, "content"),
					}
				}
				richTexts = append(richTexts, rt)
			}
		}
	}

	return CodeBlock{
		Language: lang,
		Code:     richTexts,
	}, true
}
