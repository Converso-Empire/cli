package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/converso-empire/cli/pkg/auth"
	"github.com/converso-empire/cli/pkg/bridge"
	"github.com/converso-empire/cli/pkg/config"
	"github.com/converso-empire/cli/pkg/plugin"
	"github.com/converso-empire/cli/pkg/telemetry"
	"github.com/spf13/cobra"
)

// NewYouTubeCmd creates the YouTube command
func NewYouTubeCmd(cfg *config.Config, logger telemetry.Logger) *cobra.Command {
	youtubeCmd := &cobra.Command{
		Use:   "youtube",
		Short: "YouTube module commands",
		Long:  "Access YouTube downloader and format listing functionality",
	}

	// Download command
	downloadCmd := &cobra.Command{
		Use:   "download <url>",
		Short: "Download YouTube video or audio",
		Long: `Download YouTube videos or extract audio with various options.

Examples:
  converso youtube download https://youtube.com/watch?v=example
  converso youtube download https://youtube.com/watch?v=example --mode audio
  converso youtube download https://youtube.com/watch?v=example --output-dir ./downloads`,
		
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runYouTubeDownload(cmd, args, cfg, logger)
		},
	}

	// Add flags
	downloadCmd.Flags().String("mode", "best", "Download mode: audio, video, merge, progressive")
	downloadCmd.Flags().String("format-id", "", "Specific format ID to download")
	downloadCmd.Flags().String("container", "mp4", "Output container format")
	downloadCmd.Flags().String("output-dir", "", "Output directory (default: ~/Downloads/Converso_YT)")
	downloadCmd.Flags().Bool("list-formats", false, "List available formats before downloading")

	youtubeCmd.AddCommand(downloadCmd)

	// List formats command
	listCmd := &cobra.Command{
		Use:   "list-formats <url>",
		Short: "List available formats for a YouTube URL",
		Long: `List all available formats for a YouTube video with details about
video quality, audio quality, and file sizes.

Example:
  converso youtube list-formats https://youtube.com/watch?v=example`,
		
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runYouTubeListFormats(cmd, args, cfg, logger)
		},
	}

	youtubeCmd.AddCommand(listCmd)

	// Info command
	infoCmd := &cobra.Command{
		Use:   "info <url>",
		Short: "Get video information",
		Long: `Get detailed information about a YouTube video including title,
uploader, duration, view count, and other metadata.

Example:
  converso youtube info https://youtube.com/watch?v=example`,
		
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runYouTubeInfo(cmd, args, cfg, logger)
		},
	}

	youtubeCmd.AddCommand(infoCmd)

	return youtubeCmd
}

// runYouTubeDownload executes the YouTube download command
func runYouTubeDownload(cmd *cobra.Command, args []string, cfg *config.Config, logger telemetry.Logger) error {
	url := args[0]
	
	// Get command flags
	mode, _ := cmd.Flags().GetString("mode")
	formatID, _ := cmd.Flags().GetString("format-id")
	container, _ := cmd.Flags().GetString("container")
	outputDir, _ := cmd.Flags().GetString("list-formats")
	listFormats, _ := cmd.Flags().GetBool("list-formats")

	// Validate mode
	validModes := map[string]bool{
		"audio": true, "video": true, "merge": true, "progressive": true, "best": true,
	}
	if !validModes[mode] {
		return fmt.Errorf("invalid mode: %s. Valid modes: audio, video, merge, progressive, best", mode)
	}

	// Set default output directory
	if outputDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		outputDir = filepath.Join(homeDir, "Downloads", "Converso_YT")
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// List formats if requested
	if listFormats {
		if err := runYouTubeListFormats(cmd, args, cfg, logger); err != nil {
			return err
		}
		fmt.Println()
	}

	// Load authentication
	authManager := auth.NewAuthManager(auth.NewFileStorage(cfg, logger), logger)
	tokens, err := authManager.storage.RetrieveTokens()
	if err != nil {
		return fmt.Errorf("authentication required. Run 'converso login' first: %w", err)
	}

	// Initialize plugin system
	bridge := bridge.NewJSONBridge(bridge.GetPythonPath(), cfg.PluginsDir, logger)
	registry := plugin.NewPluginRegistry(cfg, logger, bridge)
	
	if err := registry.LoadPlugins(); err != nil {
		return fmt.Errorf("failed to load plugins: %w", err)
	}

	// Check if YouTube module is available
	moduleInfo, err := registry.GetModuleInfo("youtube")
	if err != nil {
		return fmt.Errorf("YouTube module not found: %w", err)
	}

	logger.Info("Starting YouTube download",
		"url", url,
		"mode", mode,
		"container", container,
		"output_dir", outputDir,
		"module_version", moduleInfo.Manifest.Version,
	)

	// Prepare arguments
	argsMap := map[string]interface{}{
		"url":         url,
		"mode":        mode,
		"format_id":   formatID,
		"container":   container,
		"output_dir":  outputDir,
	}

	// Execute with progress tracking
	progressChan := make(chan *bridge.ProgressEvent, 100)
	
	go func() {
		for progress := range progressChan {
			printProgress(progress)
		}
	}()

	resp, err := registry.ExecuteCommandWithProgress("youtube", "download", argsMap, tokens, progressChan)
	close(progressChan)

	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("download failed: %s", resp.Error)
	}

	// Print results
	if result, ok := resp.Data["file_path"].(string); ok {
		fmt.Printf("\n✅ Download completed successfully!\n")
		fmt.Printf("📁 File: %s\n", result)
		
		if fileSize, ok := resp.Data["file_size"].(string); ok {
			fmt.Printf("📊 Size: %s\n", fileSize)
		}
		
		if duration, ok := resp.Data["duration"].(string); ok {
			fmt.Printf("⏱️  Duration: %s\n", duration)
		}
		
		fmt.Printf("📍 Output directory: %s\n", outputDir)
	}

	return nil
}

// runYouTubeListFormats executes the list formats command
func runYouTubeListFormats(cmd *cobra.Command, args []string, cfg *config.Config, logger telemetry.Logger) error {
	url := args[0]

	// Load authentication
	authManager := auth.NewAuthManager(auth.NewFileStorage(cfg, logger), logger)
	tokens, err := authManager.storage.RetrieveTokens()
	if err != nil {
		return fmt.Errorf("authentication required. Run 'converso login' first: %w", err)
	}

	// Initialize plugin system
	bridge := bridge.NewJSONBridge(bridge.GetPythonPath(), cfg.PluginsDir, logger)
	registry := plugin.NewPluginRegistry(cfg, logger, bridge)
	
	if err := registry.LoadPlugins(); err != nil {
		return fmt.Errorf("failed to load plugins: %w", err)
	}

	// Check if YouTube module is available
	if _, err := registry.GetModuleInfo("youtube"); err != nil {
		return fmt.Errorf("YouTube module not found: %w", err)
	}

	logger.Info("Listing YouTube formats", "url", url)

	// Execute command
	resp, err := registry.ExecuteCommand("youtube", "list_formats", map[string]interface{}{"url": url}, tokens)
	if err != nil {
		return fmt.Errorf("failed to list formats: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to list formats: %s", resp.Error)
	}

	// Print results
	if formats, ok := resp.Data["formats"].([]interface{}); ok {
		fmt.Printf("\n📹 Available Formats for: %s\n", url)
		fmt.Println("=" + fmt.Sprintf("%s", url)[:len(url)-1] + "=")
		
		for i, format := range formats {
			if formatMap, ok := format.(map[string]interface{}); ok {
				printFormat(i, formatMap)
			}
		}
		
		if totalCount, ok := resp.Data["total_count"].(float64); ok {
			fmt.Printf("\n📋 Total formats available: %.0f\n", totalCount)
		}
	}

	return nil
}

// runYouTubeInfo executes the info command
func runYouTubeInfo(cmd *cobra.Command, args []string, cfg *config.Config, logger telemetry.Logger) error {
	url := args[0]

	// Load authentication
	authManager := auth.NewAuthManager(auth.NewFileStorage(cfg, logger), logger)
	tokens, err := authManager.storage.RetrieveTokens()
	if err != nil {
		return fmt.Errorf("authentication required. Run 'converso login' first: %w", err)
	}

	// Initialize plugin system
	bridge := bridge.NewJSONBridge(bridge.GetPythonPath(), cfg.PluginsDir, logger)
	registry := plugin.NewPluginRegistry(cfg, logger, bridge)
	
	if err := registry.LoadPlugins(); err != nil {
		return fmt.Errorf("failed to load plugins: %w", err)
	}

	// Check if YouTube module is available
	if _, err := registry.GetModuleInfo("youtube"); err != nil {
		return fmt.Errorf("YouTube module not found: %w", err)
	}

	logger.Info("Getting YouTube video info", "url", url)

	// Execute command
	resp, err := registry.ExecuteCommand("youtube", "info", map[string]interface{}{"url": url}, tokens)
	if err != nil {
		return fmt.Errorf("failed to get video info: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to get video info: %s", resp.Error)
	}

	// Print results
	fmt.Printf("\n🎬 Video Information\n")
	fmt.Println("==================")
	
	if title, ok := resp.Data["title"].(string); ok {
		fmt.Printf("📺 Title: %s\n", title)
	}
	
	if uploader, ok := resp.Data["uploader"].(string); ok {
		fmt.Printf("👤 Uploader: %s\n", uploader)
	}
	
	if duration, ok := resp.Data["duration"].(float64); ok {
		fmt.Printf("⏱️  Duration: %s\n", formatDuration(int(duration)))
	}
	
	if viewCount, ok := resp.Data["view_count"].(float64); ok {
		fmt.Printf("👁️  Views: %s\n", formatNumber(int(viewCount)))
	}
	
	if uploadDate, ok := resp.Data["upload_date"].(string); ok {
		fmt.Printf("📅 Upload Date: %s\n", formatUploadDate(uploadDate))
	}
	
	if description, ok := resp.Data["description"].(string); ok && description != "" {
		fmt.Printf("📝 Description: %s\n", description)
	}

	return nil
}

// Helper functions for output formatting

func printProgress(progress *bridge.ProgressEvent) {
	percentage := int(progress.Percentage)
	barLength := 30
	filledLength := int(float64(barLength) * progress.Percentage / 100)

	bar := ""
	for i := 0; i < barLength; i++ {
		if i < filledLength {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	fmt.Printf("\r%s [%s] %d%% %s", progress.Stage, bar, percentage, progress.Message)
}

func printFormat(index int, format map[string]interface{}) {
	fmt.Printf("\n[%d] ", index)
	
	if formatID, ok := format["format_id"].(string); ok {
		fmt.Printf("ID: %s", formatID)
	}
	
	if ext, ok := format["ext"].(string); ok {
		fmt.Printf(" | Ext: %s", ext)
	}
	
	if vcodec, ok := format["vcodec"].(string); ok && vcodec != "none" {
		fmt.Printf(" | Video: %s", vcodec)
	}
	
	if acodec, ok := format["acodec"].(string); ok && acodec != "none" {
		fmt.Printf(" | Audio: %s", acodec)
	}
	
	if height, ok := format["height"].(float64); ok && height > 0 {
		fmt.Printf(" | %dp", int(height))
	}
	
	if fps, ok := format["fps"].(float64); ok && fps > 0 {
		fmt.Printf(" | %dfps", int(fps))
	}
	
	if abr, ok := format["abr"].(float64); ok && abr > 0 {
		fmt.Printf(" | %dkbps", int(abr))
	}
	
	if asr, ok := format["asr"].(float64); ok && asr > 0 {
		fmt.Printf(" | %dHz", int(asr))
	}
	
	if filesize, ok := format["filesize"].(float64); ok && filesize > 0 {
		fmt.Printf(" | %s", formatFileSize(int64(filesize)))
	}
	
	if formatNote, ok := format["format_note"].(string); ok && formatNote != "" {
		fmt.Printf(" | %s", formatNote)
	}
	
	fmt.Println()
}

func formatDuration(seconds int) string {
	if seconds <= 0 {
		return "Unknown"
	}
	
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60
	
	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, secs)
	}
	return fmt.Sprintf("%d:%02d", minutes, secs)
}

func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	
	// Simple number formatting
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	
	result := ""
	for i, char := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += ","
		}
		result += string(char)
	}
	
	return result
}

func formatUploadDate(dateStr string) string {
	if len(dateStr) != 8 {
		return dateStr
	}
	
	year := dateStr[:4]
	month := dateStr[4:6]
	day := dateStr[6:8]
	
	return fmt.Sprintf("%s-%s-%s", year, month, day)
}

func formatFileSize(bytes int64) string {
	if bytes <= 0 {
		return "0 B"
	}
	
	units := []string{"B", "KB", "MB", "GB", "TB"}
	i := 0
	size := float64(bytes)
	
	for size >= 1024 && i < len(units)-1 {
		size /= 1024
		i++
	}
	
	return fmt.Sprintf("%.1f %s", size, units[i])
}
