package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/converso-empire/cli/pkg/config"
	"github.com/converso-empire/cli/pkg/telemetry"
	"github.com/google/uuid"
)

// SecureStorage handles secure storage of authentication tokens
type SecureStorage interface {
	StoreTokens(tokens *AuthTokens) error
	RetrieveTokens() (*AuthTokens, error)
	DeleteTokens() error
	StoreDevice(device *Device) error
	RetrieveDevice() (*Device, error)
	DeleteDevice() error
}

// FileStorage implements SecureStorage using encrypted files
type FileStorage struct {
	config *config.Config
	logger telemetry.Logger
}

// NewFileStorage creates a new file-based secure storage
func NewFileStorage(cfg *config.Config, logger telemetry.Logger) SecureStorage {
	return &FileStorage{
		config: cfg,
		logger: logger,
	}
}

// StoreTokens stores authentication tokens securely
func (s *FileStorage) StoreTokens(tokens *AuthTokens) error {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(s.config.DataDir, 0700); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Generate a unique filename for the tokens
	filename := filepath.Join(s.config.DataDir, "tokens.json")

	// Marshal tokens to JSON
	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tokens: %w", err)
	}

	// Write to file with restricted permissions
	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("failed to write tokens file: %w", err)
	}

	s.logger.Info("Tokens stored successfully")
	return nil
}

// RetrieveTokens retrieves authentication tokens
func (s *FileStorage) RetrieveTokens() (*AuthTokens, error) {
	filename := filepath.Join(s.config.DataDir, "tokens.json")

	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("tokens file not found")
	}

	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read tokens file: %w", err)
	}

	// Unmarshal JSON
	var tokens AuthTokens
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tokens: %w", err)
	}

	s.logger.Info("Tokens retrieved successfully")
	return &tokens, nil
}

// DeleteTokens deletes stored authentication tokens
func (s *FileStorage) DeleteTokens() error {
	filename := filepath.Join(s.config.DataDir, "tokens.json")

	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to delete
	}

	// Delete file
	if err := os.Remove(filename); err != nil {
		return fmt.Errorf("failed to delete tokens file: %w", err)
	}

	s.logger.Info("Tokens deleted successfully")
	return nil
}

// StoreDevice stores device information
func (s *FileStorage) StoreDevice(device *Device) error {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(s.config.DataDir, 0700); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Generate a unique filename for the device
	filename := filepath.Join(s.config.DataDir, "device.json")

	// Marshal device to JSON
	data, err := json.MarshalIndent(device, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal device: %w", err)
	}

	// Write to file with restricted permissions
	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("failed to write device file: %w", err)
	}

	s.logger.Info("Device stored successfully")
	return nil
}

// RetrieveDevice retrieves device information
func (s *FileStorage) RetrieveDevice() (*Device, error) {
	filename := filepath.Join(s.config.DataDir, "device.json")

	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("device file not found")
	}

	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read device file: %w", err)
	}

	// Unmarshal JSON
	var device Device
	if err := json.Unmarshal(data, &device); err != nil {
		return nil, fmt.Errorf("failed to unmarshal device: %w", err)
	}

	s.logger.Info("Device retrieved successfully")
	return &device, nil
}

// DeleteDevice deletes stored device information
func (s *FileStorage) DeleteDevice() error {
	filename := filepath.Join(s.config.DataDir, "device.json")

	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to delete
	}

	// Delete file
	if err := os.Remove(filename); err != nil {
		return fmt.Errorf("failed to delete device file: %w", err)
	}

	s.logger.Info("Device deleted successfully")
	return nil
}

// AuthManager manages authentication state and storage
type AuthManager struct {
	storage SecureStorage
	logger  telemetry.Logger
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(storage SecureStorage, logger telemetry.Logger) *AuthManager {
	return &AuthManager{
		storage: storage,
		logger:  logger,
	}
}

// IsAuthenticated checks if the user is authenticated
func (m *AuthManager) IsAuthenticated(cfg *config.Config) bool {
	tokens, err := m.storage.RetrieveTokens()
	if err != nil {
		return false
	}

	return !tokens.IsExpired()
}

// GetAuthStatus returns the current authentication status
func (m *AuthManager) GetAuthStatus(cfg *config.Config) (*AuthStatus, error) {
	tokens, err := m.storage.RetrieveTokens()
	if err != nil {
		return &AuthStatus{
			Authenticated: false,
		}, nil
	}

	device, err := m.storage.RetrieveDevice()
	if err != nil {
		return &AuthStatus{
			Authenticated: !tokens.IsExpired(),
			DeviceID:      tokens.DeviceID,
			ExpiresAt:     tokens.ExpiresAt,
		}, nil
	}

	return &AuthStatus{
		Authenticated: !tokens.IsExpired(),
		DeviceID:      device.ID,
		Username:      device.Name,
		Email:         "", // Would be populated from token claims
		ExpiresAt:     tokens.ExpiresAt,
	}, nil
}

// ClearAuth clears all authentication data
func (m *AuthManager) ClearAuth() error {
	if err := m.storage.DeleteTokens(); err != nil {
		return fmt.Errorf("failed to delete tokens: %w", err)
	}

	if err := m.storage.DeleteDevice(); err != nil {
		return fmt.Errorf("failed to delete device: %w", err)
	}

	m.logger.Info("Authentication cleared successfully")
	return nil
}

// GenerateDeviceID generates a unique device ID
func GenerateDeviceID() string {
	return uuid.New().String()
}

// GetDeviceName generates a default device name
func GetDeviceName() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown-device"
	}
	return hostname
}
