package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/converso-empire/cli/pkg/config"
	"github.com/converso-empire/cli/pkg/telemetry"
	"github.com/google/uuid"
	"github.com/shirou/gopsutil/v3/host"
)

// OAuth2Client handles OAuth2 authentication with device flow
type OAuth2Client struct {
	config     *config.Config
	httpClient *http.Client
	logger     telemetry.Logger
}

// NewOAuth2Client creates a new OAuth2 client
func NewOAuth2Client(cfg *config.Config, logger telemetry.Logger) *OAuth2Client {
	return &OAuth2Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// DeviceAuthFlow performs the OAuth2 device authorization flow
func (c *OAuth2Client) DeviceAuthFlow() (*AuthTokens, error) {
	c.logger.Info("Starting device authorization flow")

	// Get device information
	deviceInfo, err := c.getDeviceInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get device info: %w", err)
	}

	// Request device code
	deviceAuthResp, err := c.requestDeviceCode(deviceInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to request device code: %w", err)
	}

	// Display verification instructions
	c.displayVerificationInstructions(deviceAuthResp)

	// Poll for tokens
	tokens, err := c.pollForTokens(deviceAuthResp)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain tokens: %w", err)
	}

	// Register device with backend
	deviceResp, err := c.registerDevice(deviceInfo, tokens)
	if err != nil {
		return nil, fmt.Errorf("failed to register device: %w", err)
	}

	// Update tokens with device information
	tokens.DeviceID = deviceResp.DeviceID
	tokens.DeviceToken = deviceResp.DeviceToken

	c.logger.Info("Device authorization completed successfully")
	return tokens, nil
}

// RefreshTokens refreshes the access token using the refresh token
func (c *OAuth2Client) RefreshTokens(tokens *AuthTokens) (*AuthTokens, error) {
	c.logger.Info("Refreshing tokens")

	data := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": tokens.RefreshToken,
		"client_id":     c.config.ClientID,
	}

	resp, err := c.makeTokenRequest(data)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh tokens: %w", err)
	}

	// Update tokens
	tokens.AccessToken = resp.AccessToken
	tokens.RefreshToken = resp.RefreshToken
	tokens.ExpiresAt = time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second)
	tokens.TokenType = resp.TokenType
	tokens.Scope = resp.Scope

	c.logger.Info("Tokens refreshed successfully")
	return tokens, nil
}

// requestDeviceCode requests a device code from the authorization server
func (c *OAuth2Client) requestDeviceCode(deviceInfo *Device) (*DeviceAuthResponse, error) {
	data := map[string]string{
		"client_id": c.config.ClientID,
		"scope":     "openid profile email",
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.config.AuthURL+"/device/code", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("authorization server returned status %d", resp.StatusCode)
	}

	var deviceAuthResp DeviceAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&deviceAuthResp); err != nil {
		return nil, err
	}

	return &deviceAuthResp, nil
}

// pollForTokens polls the token endpoint until tokens are available
func (c *OAuth2Client) pollForTokens(deviceAuthResp *DeviceAuthResponse) (*AuthTokens, error) {
	ticker := time.NewTicker(time.Duration(deviceAuthResp.Interval) * time.Second)
	defer ticker.Stop()

	timeout := time.After(time.Duration(deviceAuthResp.ExpiresIn) * time.Second)

	for {
		select {
		case <-ticker.C:
			data := map[string]string{
				"grant_type":    "urn:ietf:params:oauth:grant-type:device_code",
				"client_id":     c.config.ClientID,
				"device_code":   deviceAuthResp.DeviceCode,
				"client_secret": c.config.ClientSecret,
			}

			resp, err := c.makeTokenRequest(data)
			if err != nil {
				if errors.Is(err, ErrAuthorizationPending) {
					continue
				}
				return nil, err
			}

			// Create AuthTokens
			tokens := &AuthTokens{
				AccessToken:  resp.AccessToken,
				RefreshToken: resp.RefreshToken,
				ExpiresAt:    time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second),
				TokenType:    resp.TokenType,
				Scope:        resp.Scope,
			}

			return tokens, nil

		case <-timeout:
			return nil, errors.New("device authorization timed out")
		}
	}
}

// registerDevice registers the device with the backend API
func (c *OAuth2Client) registerDevice(deviceInfo *Device, tokens *AuthTokens) (*RegisterDeviceResponse, error) {
	reqData := RegisterDeviceRequest{
		DeviceID:   uuid.New().String(),
		DeviceName: deviceInfo.Name,
		OS:         deviceInfo.OS,
		Arch:       deviceInfo.Architecture,
		Version:    deviceInfo.Version,
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.config.APIEndpoint+"/api/v1/devices/register", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("device registration failed: %s", string(body))
	}

	var deviceResp RegisterDeviceResponse
	if err := json.NewDecoder(resp.Body).Decode(&deviceResp); err != nil {
		return nil, err
	}

	return &deviceResp, nil
}

// makeTokenRequest makes a request to the token endpoint
func (c *OAuth2Client) makeTokenRequest(data map[string]string) (*TokenResponse, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.config.TokenURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrAuthorizationPending
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed with status %d", resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// displayVerificationInstructions displays the verification instructions to the user
func (c *OAuth2Client) displayVerificationInstructions(deviceAuthResp *DeviceAuthResponse) {
	fmt.Println()
	fmt.Println("🔑 Authentication Required")
	fmt.Println("=========================")
	fmt.Printf("1. Open your browser and go to: %s\n", deviceAuthResp.VerificationURI)
	fmt.Printf("2. Enter the following code: %s\n", deviceAuthResp.UserCode)
	fmt.Println("3. Complete the authentication process")
	fmt.Println()
	fmt.Println("Waiting for authentication...")
	fmt.Println()
}

// getDeviceInfo gets information about the current device
func (c *OAuth2Client) getDeviceInfo() (*Device, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	hostInfo, err := host.Info()
	if err != nil {
		return nil, err
	}

	device := &Device{
		ID:           uuid.New().String(),
		Name:         hostname,
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		Version:      hostInfo.PlatformVersion,
		CreatedAt:    time.Now(),
		LastSeen:     time.Now(),
	}

	return device, nil
}

// ErrAuthorizationPending is returned when authorization is still pending
var ErrAuthorizationPending = errors.New("authorization pending")

// OpenBrowser opens the default web browser to the specified URL
func OpenBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}
