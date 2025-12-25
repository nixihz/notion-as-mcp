// Package cmd provides CLI commands for the Notion MCP server.
package cmd

import (
	"context"
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
			// Load configuration
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			// Validate configuration
			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("validate config: %w", err)
			}

			// Initialize logger
			if err := logger.Init(cfg); err != nil {
				return fmt.Errorf("init logger: %w", err)
			}

			log := logger.Get()
			log.Info("starting Notion MCP server",
				slog.String("database_id", cfg.NotionDatabaseID),
				slog.String("type_field", cfg.NotionTypeField),
			)

			// Create server
			srv, err := server.NewServer(cfg)
			if err != nil {
				return fmt.Errorf("create server: %w", err)
			}

			// Setup context with cancellation
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Handle shutdown signals
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			// Start server in goroutine
			errChan := make(chan error, 1)
			go func() {
				errChan <- srv.Start(ctx)
			}()

			// Wait for shutdown signal or error
			select {
			case sig := <-sigChan:
				log.Info("received signal", slog.String("signal", sig.String()))
				cancel()
			case err := <-errChan:
				if err != nil {
					return fmt.Errorf("server error: %w", err)
				}
			}

			// Stop server
			srv.Stop()
			log.Info("server stopped")

			return nil
		},
	}

	return cmd
}
