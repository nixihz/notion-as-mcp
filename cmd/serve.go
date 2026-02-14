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
	var (
		host      string
		port      int
		transport string
	)

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the MCP server",
		Long: `Start the Notion MCP server.

The server will listen for MCP protocol messages over stdio or streamable HTTP
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

			// Override config with CLI flags if provided
			if host != "" {
				cfg.ServerHost = host
			}
			if port != 0 {
				cfg.ServerPort = port
			}
			if transport != "" {
				cfg.TransportType = transport
			}

			// Create server (initializes logger internally)
			srv, err := server.NewServer(cfg)
			if err != nil {
				return fmt.Errorf("create server: %w", err)
			}
			defer func() { _ = srv.Stop() }()

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

	// Add flags
	cmd.Flags().StringVar(&host, "host", "", "Server host address (default: 0.0.0.0)")
	cmd.Flags().IntVarP(&port, "port", "p", 0, "Server port (default: 3100)")
	cmd.Flags().StringVarP(&transport, "transport", "t", "", "Transport type: streamable or stdio (default: streamable)")

	return cmd
}
