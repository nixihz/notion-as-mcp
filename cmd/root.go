// Package cmd provides CLI commands for the Notion MCP server.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// Root returns the root command.
func Root() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notion-as-mcp",
		Short: "Notion MCP Server - A Model Context Protocol server for Notion",
		Long: `Notion MCP Server is a CLI tool that provides MCP (Model Context Protocol)
access to Notion databases, exposing prompts, resources, and tools based on
type fields in your Notion database.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil // Skip config loading for version/help commands
		},
	}

	cmd.AddCommand(serveCmd())
	cmd.AddCommand(versionCmd())
	cmd.AddCommand(completionCmd())

	return cmd
}

// versionCmd returns the version command.
func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Notion MCP Server v1.0.0")
		},
	}
}

// completionCmd returns the completion command.
func completionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish]",
		Short: "Generate shell completion script",
		Long: `Generate shell completion script for notion-as-mcp.

To load completions:

Bash:

  $ source <(notion-as-mcp completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ notion-as-mcp completion bash > /etc/bash_completion.d/notion-as-mcp
  # macOS:
  $ notion-as-mcp completion bash > /usr/local/etc/bash_completion.d/notion-as-mcp

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ notion-as-mcp completion zsh > "${fpath[1]}/_notion-as-mcp"

  # You will need to start a new shell for this setup to take effect.

Fish:

  $ notion-as-mcp completion fish | source

  # To load completions for each session, execute once:
  $ notion-as-mcp completion fish > ~/.config/fish/completions/notion-as-mcp.fish
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			}
			return nil
		},
	}
}
