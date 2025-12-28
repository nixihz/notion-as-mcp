// Package cmd provides CLI commands for the Notion MCP server.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/nixihz/notion-as-mcp/internal/config"
	"github.com/nixihz/notion-as-mcp/internal/logger"
	"github.com/nixihz/notion-as-mcp/internal/server"
)

// serveCmd returns the serve command.
func serveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the MCP server",
		Long: `Start the Notion MCP server.

The server will listen for MCP protocol messages over stdio
and communicate with Notion to provide prompts, resources,
and tools based on your Notion database.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load and validate configuration
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("validate config: %w", err)
			}

			// Create server (initializes logger internally)
			srv, err := server.NewServer(cfg)
			if err != nil {
				return fmt.Errorf("create server: %w", err)
			}
			defer srv.Stop()

			// Setup context with cancellation
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Handle shutdown signals
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			// Start server and wait for completion or signal
			errChan := make(chan error, 1)
			go func() {
				errChan <- srv.Start(ctx)
			}()

			select {
			case sig := <-sigChan:
				logger.Get().Info("received signal", slog.String("signal", sig.String()))
				cancel()
				// Wait for server to stop gracefully
				if err := <-errChan; err != nil && !errors.Is(err, context.Canceled) {
					return fmt.Errorf("server error: %w", err)
				}
			case err := <-errChan:
				if err != nil {
					return fmt.Errorf("server error: %w", err)
				}
			}

			return nil
		},
	}

	return cmd
}
