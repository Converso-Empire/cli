package auth

import (
	"encoding/json"
	"time"
)

// Device represents a registered device
type Device struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	OS           string    `json:"os"`
	Architecture string    `json:"architecture"`
	Version      string    `json:"version"`
	CreatedAt    time.Time `json:"created_at"`
	LastSeen     time.Time `json:"last_seen"`
}

// AuthTokens represents authentication tokens
type AuthTokens struct {
	AccessToken     string    `json:"access_token"`
	RefreshToken    string    `json:"refresh_token"`
	ExpiresAt       time.Time `json:"expires_at"`
	TokenType       string    `json:"token_type"`
	Scope           string    `json:"scope"`
	DeviceID        string    `json:"device_id"`
	DeviceToken     string    `json:"device_token"`
}

// OAuth2Config represents OAuth2 configuration
type OAuth2Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	AuthURL      string `json:"auth_url"`
	TokenURL     string `json:"token_url"`
	RedirectURL  string `json:"redirect_url"`
}

// DeviceAuthResponse represents the response from device authorization endpoint
type DeviceAuthResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

// TokenResponse represents the OAuth2 token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// RegisterDeviceRequest represents the request to register a device
type RegisterDeviceRequest struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
	OS         string `json:"os"`
	Arch       string `json:"arch"`
	Version    string `json:"version"`
}

// RegisterDeviceResponse represents the response from device registration
type RegisterDeviceResponse struct {
	DeviceID    string `json:"device_id"`
	DeviceToken string `json:"device_token"`
}

// AuthStatus represents the current authentication status
type AuthStatus struct {
	Authenticated bool      `json:"authenticated"`
	DeviceID      string    `json:"device_id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	ExpiresAt     time.Time `json:"expires_at"`
}

// IsExpired checks if the tokens are expired
func (t *AuthTokens) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// NeedsRefresh checks if the tokens need to be refreshed
func (t *AuthTokens) NeedsRefresh() bool {
	// Refresh if expires within 5 minutes
	return time.Now().Add(5 * time.Minute).After(t.ExpiresAt)
}

// MarshalJSON implements custom JSON marshaling for AuthTokens
func (t *AuthTokens) MarshalJSON() ([]byte, error) {
	type Alias AuthTokens
	return json.Marshal(&struct {
		ExpiresAt string `json:"expires_at"`
		*Alias
	}{
		ExpiresAt: t.ExpiresAt.Format(time.RFC3339),
		Alias:     (*Alias)(t),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for AuthTokens
func (t *AuthTokens) UnmarshalJSON(data []byte) error {
	type Alias AuthTokens
	aux := &struct {
		ExpiresAt string `json:"expires_at"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	var err error
	t.ExpiresAt, err = time.Parse(time.RFC3339, aux.ExpiresAt)
	return err
}
