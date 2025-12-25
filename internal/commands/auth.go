package commands

import (
	"fmt"
	"time"

	"github.com/converso-empire/cli/pkg/auth"
	"github.com/converso-empire/cli/pkg/config"
	"github.com/converso-empire/cli/pkg/telemetry"
	"github.com/spf13/cobra"
)

// NewLoginCmd creates the login command
func NewLoginCmd(cfg *config.Config, logger telemetry.Logger) *cobra.Command {
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Converso Empire",
		Long: `Authenticate with Converso Empire using OAuth2 device flow.

This command will:
  • Generate a device code and user code
  • Display instructions for completing authentication
  • Open your default browser automatically
  • Poll for authentication completion
  • Store authentication tokens securely`,
		
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogin(cmd, cfg, logger)
		},
	}

	// Add flags
	loginCmd.Flags().String("device-name", "", "Custom device name (default: hostname)")
	loginCmd.Flags().Bool("force", false, "Force re-authentication even if already logged in")

	return loginCmd
}

// NewLogoutCmd creates the logout command
func NewLogoutCmd(cfg *config.Config, logger telemetry.Logger) *cobra.Command {
	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout from Converso Empire",
		Long: `Logout from Converso Empire and clear stored authentication tokens.

This command will:
  • Clear all stored authentication tokens
  • Remove device registration
  • Require re-authentication for future commands`,
		
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogout(cmd, cfg, logger)
		},
	}

	// Add flags
	logoutCmd.Flags().Bool("force", false, "Force logout without confirmation")

	return logoutCmd
}

// runLogin executes the login process
func runLogin(cmd *cobra.Command, cfg *config.Config, logger telemetry.Logger) error {
	// Check if already authenticated
	authManager := auth.NewAuthManager(auth.NewFileStorage(cfg, logger), logger)
	if authManager.IsAuthenticated(cfg) {
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Println("✅ You are already logged in.")
			fmt.Println("💡 Run 'converso logout' to logout, or 'converso login --force' to re-authenticate.")
			return nil
		}
	}

	fmt.Println("🔑 Converso CLI Authentication")
	fmt.Println("==============================")

	// Create OAuth2 client
	oauthClient := auth.NewOAuth2Client(cfg, logger)

	// Get device name
	deviceName, _ := cmd.Flags().GetString("device-name")
	if deviceName == "" {
		deviceName = auth.GetDeviceName()
	}

	fmt.Printf("Device: %s\n", deviceName)
	fmt.Println()

	// Perform device authentication flow
	tokens, err := oauthClient.DeviceAuthFlow()
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Store tokens securely
	storage := auth.NewFileStorage(cfg, logger)
	if err := storage.StoreTokens(tokens); err != nil {
		return fmt.Errorf("failed to store tokens: %w", err)
	}

	// Display success message
	fmt.Println()
	fmt.Println("✅ Authentication successful!")
	fmt.Println()
	fmt.Println("Welcome to Converso CLI!")
	fmt.Println()
	fmt.Printf("Device: %s\n", deviceName)
	fmt.Printf("Device ID: %s\n", tokens.DeviceID)
	fmt.Printf("Expires: %s\n", tokens.ExpiresAt.Format("2006-01-02 15:04:05"))
	fmt.Println()
	fmt.Println("You can now use Converso CLI commands.")
	fmt.Println("Try: 'converso youtube list-formats <url>'")

	return nil
}

// runLogout executes the logout process
func runLogout(cmd *cobra.Command, cfg *config.Config, logger telemetry.Logger) error {
	// Check if authenticated
	authManager := auth.NewAuthManager(auth.NewFileStorage(cfg, logger), logger)
	if !authManager.IsAuthenticated(cfg) {
		fmt.Println("ℹ️  You are not currently logged in.")
		return nil
	}

	// Get confirmation
	force, _ := cmd.Flags().GetBool("force")
	if !force {
		fmt.Print("Are you sure you want to logout? This will clear all stored tokens [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" && response != "yes" {
			fmt.Println("Logout cancelled.")
			return nil
		}
	}

	// Clear authentication
	if err := authManager.ClearAuth(); err != nil {
		return fmt.Errorf("failed to clear authentication: %w", err)
	}

	fmt.Println("✅ Successfully logged out.")
	fmt.Println("💡 Run 'converso login' to authenticate again.")

	return nil
}

// NewStatusCmd creates the status command
func NewStatusCmd(cfg *config.Config, logger telemetry.Logger) *cobra.Command {
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		Long: `Show current authentication status and device information.

This command displays:
  • Authentication status
  • Device information
  • Token expiration
  • Account details`,
		
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, cfg, logger)
		},
	}

	return statusCmd
}

// runStatus shows authentication status
func runStatus(cmd *cobra.Command, cfg *config.Config, logger telemetry.Logger) error {
	authManager := auth.NewAuthManager(auth.NewFileStorage(cfg, logger), logger)
	status, err := authManager.GetAuthStatus(cfg)
	if err != nil {
		return fmt.Errorf("failed to get authentication status: %w", err)
	}

	fmt.Println("🔐 Authentication Status")
	fmt.Println("======================")

	if status.Authenticated {
		fmt.Println("✅ Authenticated")
		fmt.Printf("Device ID: %s\n", status.DeviceID)
		fmt.Printf("Username: %s\n", status.Username)
		if status.Email != "" {
			fmt.Printf("Email: %s\n", status.Email)
		}
		fmt.Printf("Expires: %s\n", status.ExpiresAt.Format("2006-01-02 15:04:05"))
		
		// Show time until expiration
		timeUntil := status.ExpiresAt.Sub(time.Now())
		if timeUntil > 0 {
			fmt.Printf("Time remaining: %s\n", formatDuration(timeUntil))
		} else {
			fmt.Println("⚠️  Token expired!")
		}
	} else {
		fmt.Println("❌ Not authenticated")
		fmt.Println("💡 Run 'converso login' to authenticate")
	}

	return nil
}

// Helper function to format duration
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return d.Round(time.Second).String()
	}
	if d < time.Hour {
		return d.Round(time.Minute).String()
	}
	return d.Round(time.Hour).String()
}
