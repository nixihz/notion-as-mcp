// Package main provides the Notion MCP Server implementation.
//
// Notion-as-MCP is a Model Context Protocol server that uses Notion
// as a data source, exposing prompts, resources, and tools based on
// type fields in a Notion database.
package main

import (
	"os"

	"github.com/nixihz/notion-as-mcp/cmd"
)

func main() {
	if err := cmd.Root().Execute(); err != nil {
		os.Exit(1)
	}
}
