// Package transport provides MCP transport implementations.
package transport

import (
	"bufio"
	"context"
	"io"
	"os"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// StdioTransport implements mcp.Transport for stdio communication.
type StdioTransport struct{}

// NewStdioTransport creates a new stdio transport.
func NewStdioTransport() *StdioTransport {
	return &StdioTransport{}
}

// Connect implements mcp.Transport.
func (t *StdioTransport) Connect(ctx context.Context) (mcp.Connection, error) {
	return &stdioConnection{
		reader: bufio.NewReader(os.Stdin),
		writer: os.Stdout,
	}, nil
}

// stdioConnection implements mcp.Connection over stdio.
type stdioConnection struct {
	reader *bufio.Reader
	writer io.Writer
	closed bool
}

// Read implements mcp.Connection.
func (c *stdioConnection) Read(ctx context.Context) (jsonrpc.Message, error) {
	line, err := c.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	// Remove trailing newline
	if len(line) > 0 && line[len(line)-1] == '\n' {
		line = line[:len(line)-1]
	}
	if len(line) > 0 && line[len(line)-1] == '\r' {
		line = line[:len(line)-1]
	}

	return jsonrpc.DecodeMessage(line)
}

// Write implements mcp.Connection.
func (c *stdioConnection) Write(ctx context.Context, msg jsonrpc.Message) error {
	data, err := jsonrpc.EncodeMessage(msg)
	if err != nil {
		return err
	}

	_, err = c.writer.Write(data)
	if err != nil {
		return err
	}

	_, err = c.writer.Write([]byte("\n"))
	return err
}

// Close implements mcp.Connection.
func (c *stdioConnection) Close() error {
	c.closed = true
	return nil
}

// SessionID implements mcp.Connection.
func (c *stdioConnection) SessionID() string {
	return ""
}

// Ensure stdioConnection implements mcp.Connection at compile time.
var _ mcp.Connection = (*stdioConnection)(nil)
