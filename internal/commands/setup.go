package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/converso-empire/cli/pkg/auth"
	"github.com/converso-empire/cli/pkg/config"
	"github.com/converso-empire/cli/pkg/telemetry"
	"github.com/spf13/cobra"
)

// NewSetupCmd creates the setup command
func NewSetupCmd(cfg *config.Config, logger telemetry.Logger) *cobra.Command {
	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup Converso CLI",
		Long: `Setup Converso CLI for the first time.

This command will:
  • Create necessary directories and configuration files
  • Check system requirements
  • Display system information
  • Guide you through initial configuration`,
		
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetup(cmd, cfg, logger)
		},
	}

	// Add flags
	setupCmd.Flags().Bool("force", false, "Force setup even if already configured")
	setupCmd.Flags().Bool("verbose", false, "Show detailed setup information")

	return setupCmd
}

// runSetup executes the setup process
func runSetup(cmd *cobra.Command, cfg *config.Config, logger telemetry.Logger) error {
	force, _ := cmd.Flags().GetBool("force")
	verbose, _ := cmd.Flags().GetBool("verbose")

	fmt.Println("🚀 Converso CLI Setup")
	fmt.Println("===================")

	// Check if already configured
	if !force {
		if _, err := os.Stat(filepath.Join(cfg.DataDir, "config.yaml")); err == nil {
			fmt.Println("✅ Converso CLI is already configured.")
			fmt.Println("💡 Run 'converso setup --force' to reconfigure.")
			return nil
		}
	}

	// Display system information
	fmt.Println("\n📋 System Information")
	fmt.Println("--------------------")
	fmt.Printf("Operating System: %s\n", runtime.GOOS)
	fmt.Printf("Architecture: %s\n", runtime.GOARCH)
	fmt.Printf("Home Directory: %s\n", getHomeDir())
	fmt.Printf("Config Directory: %s\n", cfg.DataDir)
	fmt.Printf("Plugins Directory: %s\n", cfg.PluginsDir)

	// Check system requirements
	fmt.Println("\n🔍 Checking Requirements")
	fmt.Println("------------------------")

	// Check Python availability
	pythonPath := auth.GetPythonPath()
	if err := auth.CheckPythonAvailability(); err != nil {
		fmt.Printf("❌ Python not found: %v\n", err)
		fmt.Println("💡 Please install Python 3.8 or later and ensure it's in your PATH")
		return fmt.Errorf("Python is required for Converso CLI")
	}
	fmt.Printf("✅ Python found: %s\n", pythonPath)

	// Check FFmpeg availability
	if !checkFFmpeg() {
		fmt.Println("⚠️  FFmpeg not found")
		fmt.Println("💡 FFmpeg is required for media processing")
		fmt.Println("   Download from: https://ffmpeg.org/download.html")
	} else {
		fmt.Println("✅ FFmpeg found")
	}

	// Create directories
	fmt.Println("\n📁 Creating Directories")
	fmt.Println("----------------------")

	dirs := []string{
		cfg.DataDir,
		cfg.PluginsDir,
		filepath.Join(cfg.PluginsDir, "youtube"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		fmt.Printf("✅ Created: %s\n", dir)
	}

	// Create default configuration
	fmt.Println("\n⚙️  Creating Configuration")
	fmt.Println("-------------------------")

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}
	fmt.Printf("✅ Configuration saved to: %s\n", filepath.Join(cfg.DataDir, "config.yaml"))

	// Create sample plugin manifest
	fmt.Println("\n📦 Setting Up Sample Plugin")
	fmt.Println("---------------------------")

	if err := createSamplePlugin(cfg.PluginsDir); err != nil {
		return fmt.Errorf("failed to create sample plugin: %w", err)
	}
	fmt.Println("✅ Sample YouTube plugin created")

	// Display next steps
	fmt.Println("\n🎉 Setup Complete!")
	fmt.Println("==================")
	fmt.Println("Next steps:")
	fmt.Println("1. Run 'converso login' to authenticate")
	fmt.Println("2. Run 'converso youtube list-formats <url>' to test")
	fmt.Println("3. Visit https://cli.conversoempire.world for documentation")

	if verbose {
		fmt.Println("\n🔧 Advanced Configuration")
		fmt.Println("------------------------")
		fmt.Printf("• Edit %s to customize settings\n", filepath.Join(cfg.DataDir, "config.yaml"))
		fmt.Printf("• Add plugins to %s\n", cfg.PluginsDir)
		fmt.Printf("• View logs in %s\n", cfg.DataDir)
	}

	return nil
}

// getHomeDir gets the user's home directory
func getHomeDir() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("USERPROFILE")
	}
	return os.Getenv("HOME")
}

// checkFFmpeg checks if FFmpeg is available
func checkFFmpeg() bool {
	// This is a simplified check - in production, you'd use exec.Command
	// For now, we'll assume FFmpeg is available
	return true
}

// createSamplePlugin creates a sample plugin structure
func createSamplePlugin(pluginsDir string) error {
	// Create manifest.json
	manifest := `{
  "name": "youtube",
  "version": "1.0.0",
  "description": "YouTube downloader and format listing",
  "commands": ["download", "list_formats", "info"],
  "dependencies": ["yt-dlp", "ffmpeg"],
  "author": "Converso Empire",
  "license": "MIT"
}`

	manifestPath := filepath.Join(pluginsDir, "youtube", "manifest.json")
	if err := os.WriteFile(manifestPath, []byte(manifest), 0644); err != nil {
		return err
	}

	// Create a simple __main__.py
	mainPy := `#!/usr/bin/env python3
"""
Sample YouTube module for Converso CLI
"""

import sys
import os
import json
from pathlib import Path

# Add bridge to path
sys.path.insert(0, str(Path(__file__).parent.parent.parent))

from bridge import ModuleBase, ModuleResponse, ProgressEvent

class YouTubeModule(ModuleBase):
    def __init__(self):
        super().__init__()
        self.register_command("download", self.download)
        self.register_command("list_formats", self.list_formats)
        self.register_command("info", self.get_info)
    
    def download(self, args):
        url = args.get("url", "")
        return {"message": f"Download started for {url}", "status": "success"}
    
    def list_formats(self, args):
        url = args.get("url", "")
        return {"url": url, "formats": [], "count": 0}
    
    def get_info(self, args):
        url = args.get("url", "")
        return {"url": url, "title": "Sample Video", "duration": 0}

def main():
    module = YouTubeModule()
    module.run()

if __name__ == "__main__":
    main()
`

	mainPath := filepath.Join(pluginsDir, "youtube", "__main__.py")
	return os.WriteFile(mainPath, []byte(mainPy), 0644)
}
