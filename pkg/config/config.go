package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Debug       bool   `mapstructure:"debug"`
	ConfigFile  string `mapstructure:"config_file"`
	APIEndpoint string `mapstructure:"api_endpoint"`
	AuthURL     string `mapstructure:"auth_url"`
	TokenURL    string `mapstructure:"token_url"`
	ClientID    string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	DeviceName  string `mapstructure:"device_name"`
	Concurrency int    `mapstructure:"concurrency"`
	PluginsDir  string `mapstructure:"plugins_dir"`
	DataDir     string `mapstructure:"data_dir"`
}

// Default configuration values
const (
	DefaultAPIEndpoint = "https://capi.conversoempire.world"
	DefaultAuthURL     = "https://clerk.conversoempire.world/oauth/authorize"
	DefaultTokenURL    = "https://clerk.conversoempire.world/oauth/token"
	DefaultClientID    = "converso-cli"
	DefaultConcurrency = 10
)

// Load loads the configuration from various sources
func Load() (*Config, error) {
	cfg := &Config{}

	// Set default values
	viper.SetDefault("debug", false)
	viper.SetDefault("api_endpoint", DefaultAPIEndpoint)
	viper.SetDefault("auth_url", DefaultAuthURL)
	viper.SetDefault("token_url", DefaultTokenURL)
	viper.SetDefault("client_id", DefaultClientID)
	viper.SetDefault("concurrency", DefaultConcurrency)

	// Set configuration file name and type
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Add configuration paths
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".converso")
	viper.AddConfigPath(configDir)
	viper.AddConfigPath(".")

	// Set environment variables
	viper.SetEnvPrefix("CONVERSO")
	viper.AutomaticEnv()

	// Read configuration file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, create default config
			if err := createDefaultConfig(configDir); err != nil {
				return nil, fmt.Errorf("failed to create default config: %w", err)
			}
			if err := viper.ReadInConfig(); err != nil {
				return nil, fmt.Errorf("failed to read config after creation: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal configuration
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Set computed paths
	cfg.DataDir = filepath.Join(configDir, "data")
	cfg.PluginsDir = filepath.Join(configDir, "plugins")

	return cfg, nil
}

// createDefaultConfig creates a default configuration file
func createDefaultConfig(configDir string) error {
	// Create config directory
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create default config content
	defaultConfig := []byte(`# Converso CLI Configuration
debug: false

# API Configuration
api_endpoint: "https://capi.conversoempire.world"
auth_url: "https://clerk.conversoempire.world/oauth/authorize"
token_url: "https://clerk.conversoempire.world/oauth/token"
client_id: "ssUkfqPfE4NC9TWz"

# Application Settings
concurrency: 10
device_name: "default"

# Paths (auto-generated)
# data_dir: "~/.converso/data"
# plugins_dir: "~/.converso/plugins"
`)

	configFile := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configFile, defaultConfig, 0644); err != nil {
		return fmt.Errorf("failed to write default config file: %w", err)
	}

	return nil
}

// Save saves the configuration to file
func (c *Config) Save() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".converso")
	configFile := filepath.Join(configDir, "config.yaml")

	// Set viper values
	viper.Set("debug", c.Debug)
	viper.Set("api_endpoint", c.APIEndpoint)
	viper.Set("auth_url", c.AuthURL)
	viper.Set("token_url", c.TokenURL)
	viper.Set("client_id", c.ClientID)
	viper.Set("concurrency", c.Concurrency)
	viper.Set("device_name", c.DeviceName)

	// Write to file
	if err := viper.WriteConfigAs(configFile); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
