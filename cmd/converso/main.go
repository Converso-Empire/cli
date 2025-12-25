package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/converso-empire/cli/internal/commands"
	"github.com/converso-empire/cli/pkg/config"
	"github.com/converso-empire/cli/pkg/telemetry"
	"github.com/spf13/cobra"
)

var (
	// Build information (set during build)
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Initialize configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize telemetry
	logger := telemetry.NewLogger(cfg.Debug)

	// Create root command
	rootCmd := commands.NewRootCmd(version, commit, date, cfg, logger)

	// Execute command
	if err := rootCmd.Execute(); err != nil {
		logger.Error("Command failed", "error", err)
		os.Exit(1)
	}
}

// VersionInfo holds build-time version information
type VersionInfo struct {
	Version string
	Commit  string
	Date    string
	GoOS    string
	GoArch  string
}

// GetVersionInfo returns version information
func GetVersionInfo() VersionInfo {
	return VersionInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
		GoOS:    runtime.GOOS,
		GoArch:  runtime.GOARCH,
	}
}
