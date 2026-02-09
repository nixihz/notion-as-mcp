package notion

import (
	"bytes"
	"fmt"
	"strings"
)

// MarkdownConverter converts a Page to Markdown.
type MarkdownConverter struct {
	Page *PageContent
	Buf  *bytes.Buffer
}

// NewMarkdownConverter creates a new Markdown converter.
func NewMarkdownConverter(pageContent *PageContent) *MarkdownConverter {
	return &MarkdownConverter{
		Page: pageContent,
		Buf:  &bytes.Buffer{},
	}
}

// WriteString writes a string to the buffer.
func (c *MarkdownConverter) WriteString(s string) {
	c.Buf.WriteString(s)
}

// Newline writes a newline to the buffer.
func (c *MarkdownConverter) Newline() {
	c.Buf.WriteString("\n\n")
}

// Eol writes end-of-line to the buffer.
func (c *MarkdownConverter) Eol() {
	d := c.Buf.Bytes()
	n := len(d)
	if n > 0 && d[n-1] != '\n' {
		c.Buf.WriteByte('\n')
	}
}

// RenderRichText renders rich text with formatting.
func (c *MarkdownConverter) RenderRichText(richTexts []RichText) string {
	var sb strings.Builder
	for _, rt := range richTexts {
		text := rt.PlainText
		if text == "" {
			text = rt.Text.Content
		}

		// Apply formatting based on annotations
		if rt.Annotations.Bold {
			text = "**" + text + "**"
		}
		if rt.Annotations.Italic {
			text = "*" + text + "*"
		}
		if rt.Annotations.Strikethrough {
			text = "~~" + text + "~~"
		}
		if rt.Annotations.Code {
			text = "`" + text + "`"
		}

		// Handle links
		if rt.Text.Link != nil && rt.Text.Link.URL != "" {
			text = fmt.Sprintf("[%s](%s)", text, rt.Text.Link.URL)
		} else if rt.Href != nil && *rt.Href != "" {
			text = fmt.Sprintf("[%s](%s)", text, *rt.Href)
		}

		sb.WriteString(text)
	}
	return sb.String()
}

// RenderParagraph renders a paragraph block.
func (c *MarkdownConverter) RenderParagraph(block Block) {
	var richTexts []RichText
	// Try to get from Paragraph field first
	if block.Paragraph != nil {
		richTexts = block.Paragraph.RichText
	} else {
		// Fallback to Content field
		richTexts = c.extractRichTexts(block.Content)
	}
	if len(richTexts) == 0 {
		return
	}
	text := c.RenderRichText(richTexts)
	c.WriteString(text)
	c.Newline()
}

// RenderHeading renders a heading block.
func (c *MarkdownConverter) RenderHeading(block Block, level int) {
	richTexts := c.extractRichTexts(block.Content)
	if len(richTexts) == 0 {
		return
	}
	text := c.RenderRichText(richTexts)
	if text == "" {
		return
	}
	prefix := strings.Repeat("#", level) + " "
	c.WriteString(prefix + strings.TrimSpace(text))
	c.Newline()
}

// RenderBulletedList renders a bulleted list item.
func (c *MarkdownConverter) RenderBulletedList(block Block) {
	richTexts := c.extractRichTexts(block.Content)
	if len(richTexts) == 0 {
		return
	}
	text := c.RenderRichText(richTexts)
	if text == "" {
		return
	}
	c.WriteString("- " + text)
	c.Eol()
}

// RenderNumberedList renders a numbered list item.
func (c *MarkdownConverter) RenderNumberedList(block Block, index int) {
	richTexts := c.extractRichTexts(block.Content)
	if len(richTexts) == 0 {
		return
	}
	text := c.RenderRichText(richTexts)
	if text == "" {
		return
	}
	c.WriteString(fmt.Sprintf("%d. %s", index, text))
	c.Eol()
}

// RenderCode renders a code block.
func (c *MarkdownConverter) RenderCode(block Block) {
	codeBlock, ok := block.Content.(CodeBlock)
	if !ok {
		// Try to parse from map
		if contentMap, ok := block.Content.(map[string]any); ok {
			codeBlock = c.parseCodeBlockFromMap(contentMap)
		} else {
			return
		}
	}

	// Extract code text: prefer RichText (Notion API field), fallback to Code
	richTexts := codeBlock.RichText
	if len(richTexts) == 0 {
		richTexts = codeBlock.Code
	}
	var codeText strings.Builder
	for _, rt := range richTexts {
		codeText.WriteString(rt.PlainText)
	}

	language := codeBlock.Language
	if language == "" {
		language = "text"
	}

	c.WriteString("```" + language)
	c.Eol()
	c.WriteString(codeText.String())
	c.Eol()
	c.WriteString("```")
	c.Newline()
}

// RenderQuote renders a quote block.
func (c *MarkdownConverter) RenderQuote(block Block) {
	richTexts := c.extractRichTexts(block.Content)
	if len(richTexts) == 0 {
		return
	}
	text := c.RenderRichText(richTexts)
	if text == "" {
		return
	}
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	for _, line := range lines {
		if line != "" {
			c.WriteString("> " + line)
			c.Eol()
		}
	}
	c.Newline()
}

// RenderDivider renders a divider block.
func (c *MarkdownConverter) RenderDivider(block Block) {
	c.WriteString("---")
	c.Newline()
}

// RenderToDo renders a to_do block as a markdown checkbox.
func (c *MarkdownConverter) RenderToDo(block Block) {
	checked := false
	if contentMap, ok := block.Content.(map[string]any); ok {
		checked = getMapBool(contentMap, "checked")
	}
	richTexts := c.extractRichTexts(block.Content)
	if len(richTexts) == 0 {
		return
	}
	text := c.RenderRichText(richTexts)
	if text == "" {
		return
	}
	if checked {
		c.WriteString("- [x] " + text)
	} else {
		c.WriteString("- [ ] " + text)
	}
	c.Eol()
}

// RenderCallout renders a callout block.
func (c *MarkdownConverter) RenderCallout(block Block) {
	richTexts := c.extractRichTexts(block.Content)
	if len(richTexts) == 0 {
		return
	}
	text := c.RenderRichText(richTexts)
	if text == "" {
		return
	}
	c.WriteString("> 💡 " + text)
	c.Newline()
}

// RenderImage renders an image block.
func (c *MarkdownConverter) RenderImage(block Block) {
	// Extract image URL from content
	if contentMap, ok := block.Content.(map[string]any); ok {
		if file, ok := contentMap["file"].(map[string]any); ok {
			if url, ok := file["url"].(string); ok {
				caption := ""
				if captionArr, ok := contentMap["caption"].([]any); ok && len(captionArr) > 0 {
					if captionMap, ok := captionArr[0].(map[string]any); ok {
						if plainText, ok := captionMap["plain_text"].(string); ok {
							caption = plainText
						}
					}
				}
				if caption != "" {
					c.WriteString(fmt.Sprintf("![%s](%s)", caption, url))
				} else {
					c.WriteString(fmt.Sprintf("![](%s)", url))
				}
				c.Newline()
			}
		}
	}
}

// extractRichTexts extracts rich text array from block content.
func (c *MarkdownConverter) extractRichTexts(content any) []RichText {
	switch v := content.(type) {
	case []RichText:
		return v
	case Paragraph:
		return v.RichText
	case *Paragraph:
		if v != nil {
			return v.RichText
		}
	case map[string]any:
		if rt, ok := v["rich_text"].([]any); ok {
			var richTexts []RichText
			for _, r := range rt {
				if m, ok := r.(map[string]any); ok {
					rt := RichText{
						Type:      getMapString(m, "type"),
						PlainText: getMapString(m, "plain_text"),
					}
					// Parse text object
					if textMap, ok := m["text"].(map[string]any); ok {
						rt.Text = Text{
							Content: getMapString(textMap, "content"),
						}
						if linkMap, ok := textMap["link"].(map[string]any); ok {
							if url, ok := linkMap["url"].(string); ok {
								rt.Text.Link = &Link{URL: url}
							}
						}
					}
					// Parse annotations
					if ann, ok := m["annotations"].(map[string]any); ok {
						rt.Annotations = Annotations{
							Bold:          getMapBool(ann, "bold"),
							Italic:        getMapBool(ann, "italic"),
							Strikethrough: getMapBool(ann, "strikethrough"),
							Underline:     getMapBool(ann, "underline"),
							Code:          getMapBool(ann, "code"),
						}
					}
					// Parse href
					if href, ok := m["href"].(string); ok && href != "" {
						rt.Href = &href
					}
					richTexts = append(richTexts, rt)
				}
			}
			return richTexts
		}
	}
	return nil
}

// parseCodeBlockFromMap parses CodeBlock from map.
func (c *MarkdownConverter) parseCodeBlockFromMap(contentMap map[string]any) CodeBlock {
	codeBlock := CodeBlock{
		Language: getMapString(contentMap, "language"),
	}

	if rt, ok := contentMap["rich_text"].([]any); ok {
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
				codeBlock.Code = append(codeBlock.Code, rt)
			}
		}
	}

	return codeBlock
}

// getMapBool gets a bool value from a map.
func getMapBool(m map[string]any, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// RenderBlock renders a single block.
func (c *MarkdownConverter) RenderBlock(block Block, numberedListIndex *int) {
	switch block.Type {
	case BlockTypeParagraph:
		c.RenderParagraph(block)
	case BlockTypeHeading1:
		c.RenderHeading(block, 1)
	case BlockTypeHeading2:
		c.RenderHeading(block, 2)
	case BlockTypeHeading3:
		c.RenderHeading(block, 3)
	case BlockTypeBulletedListItem:
		c.RenderBulletedList(block)
	// NumberedListItem is handled in ToMarkdown
	case BlockTypeCode:
		c.RenderCode(block)
	case BlockTypeQuote:
		c.RenderQuote(block)
	case BlockTypeDivider:
		c.RenderDivider(block)
	case BlockTypeToDo:
		c.RenderToDo(block)
	case BlockTypeCallout:
		c.RenderCallout(block)
	case BlockTypeImage:
		c.RenderImage(block)
	default:
		// For unknown types, try to extract text
		richTexts := c.extractRichTexts(block.Content)
		if len(richTexts) > 0 {
			text := c.RenderRichText(richTexts)
			c.WriteString(text)
			c.Newline()
		}
	}
}

// ToMarkdown converts PageContent to Markdown string.
func (c *MarkdownConverter) ToMarkdown() string {
	if c.Page == nil {
		return ""
	}

	// Ensure Buf is initialized
	if c.Buf == nil {
		c.Buf = &bytes.Buffer{}
	}

	// Render all blocks
	var numberedListIndex int
	var inNumberedList bool
	for _, block := range c.Page.Blocks {
		if block.Type == BlockTypeNumberedListItem {
			if !inNumberedList {
				numberedListIndex = 1
				inNumberedList = true
			} else {
				numberedListIndex++
			}
			c.RenderNumberedList(block, numberedListIndex)
		} else {
			if inNumberedList {
				inNumberedList = false
				numberedListIndex = 0
			}
			c.RenderBlock(block, nil)
		}
	}

	result := c.Buf.String()
	// Trim trailing whitespace
	result = strings.TrimSpace(result)
	return result
}

// PageToMarkdown converts a PageContent to Markdown string.
func PageToMarkdown(pageContent *PageContent) string {
	converter := NewMarkdownConverter(pageContent)
	return converter.ToMarkdown()
}
