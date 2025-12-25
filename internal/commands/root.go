package commands

import (
	"fmt"
	"os"

	"github.com/converso-empire/cli/pkg/auth"
	"github.com/converso-empire/cli/pkg/config"
	"github.com/converso-empire/cli/pkg/telemetry"
	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
type RootCmd struct {
	*cobra.Command
	cfg    *config.Config
	logger telemetry.Logger
}

// NewRootCmd creates a new root command
func NewRootCmd(version, commit, date string, cfg *config.Config, logger telemetry.Logger) *cobra.Command {
	root := &RootCmd{
		cfg:    cfg,
		logger: logger,
	}

	cmd := &cobra.Command{
		Use:   "converso",
		Short: "Converso CLI - Enterprise SaaS Command Line Interface",
		Long: `Converso CLI provides unified access to all Converso Empire services.

This is the official command-line interface for Converso Empire's enterprise
platform, designed for developers and system administrators.

Features:
  • OAuth2 authentication with device flow
  • Dynamic plugin system for extensibility
  • Background worker for long-running tasks
  • Enterprise-grade security and compliance
  • Cross-platform support (Linux, macOS, Windows)`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Check if command requires authentication
			if requiresAuth(cmd) {
				if !auth.IsAuthenticated(cfg) {
					return fmt.Errorf("authentication required. Run 'converso login' first")
				}
			}
			return nil
		},
	}

	// Add subcommands
	cmd.AddCommand(NewSetupCmd(cfg, logger))
	cmd.AddCommand(NewLoginCmd(cfg, logger))
	cmd.AddCommand(NewLogoutCmd(cfg, logger))
	cmd.AddCommand(NewYouTubeCmd(cfg, logger))
	cmd.AddCommand(NewVersionCmd(version, commit, date))

	// Global flags
	cmd.PersistentFlags().BoolVar(&cfg.Debug, "debug", false, "Enable debug logging")
	cmd.PersistentFlags().StringVar(&cfg.ConfigFile, "config", "", "Config file (default is $HOME/.converso/config.yaml)")

	return cmd
}

// requiresAuth checks if a command requires authentication
func requiresAuth(cmd *cobra.Command) bool {
	// Commands that don't require authentication
	noAuthCommands := map[string]bool{
		"setup":   true,
		"login":   true,
		"logout":  true,
		"version": true,
		"help":    true,
	}

	return !noAuthCommands[cmd.Name()]
}

// NewVersionCmd creates the version command
func NewVersionCmd(version, commit, date string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Long:  "Print the version number of Converso CLI",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Converso CLI v%s\n", version)
			fmt.Printf("Commit: %s\n", commit)
			fmt.Printf("Build Date: %s\n", date)
			fmt.Printf("Go Version: %s\n", os.Getenv("GOVERSION"))
			fmt.Printf("Platform: %s/%s\n", os.Getenv("GOOS"), os.Getenv("GOARCH"))
		},
	}
}
