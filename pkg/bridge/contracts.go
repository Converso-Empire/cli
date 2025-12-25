package bridge

import (
	"encoding/json"
	"time"
)

// ModuleRequest represents a request to a Python module
type ModuleRequest struct {
	Command     string                 `json:"command"`
	Args        map[string]interface{} `json:"args"`
	AuthToken   string                 `json:"auth_token"`
	DeviceToken string                 `json:"device_token"`
	Timeout     int                    `json:"timeout"`
}

// ModuleResponse represents a response from a Python module
type ModuleResponse struct {
	Success     bool                   `json:"success"`
	Data        map[string]interface{} `json:"data"`
	Error       string                 `json:"error"`
	Progress    *ProgressEvent         `json:"progress,omitempty"`
}

// ProgressEvent represents a progress update from a module
type ProgressEvent struct {
	Stage       string  `json:"stage"`
	Current     int64   `json:"current"`
	Total       int64   `json:"total"`
	Percentage  float64 `json:"percentage"`
	Message     string  `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
}

// ModuleManifest represents a Python module's manifest
type ModuleManifest struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Commands    []string `json:"commands"`
	Dependencies []string `json:"dependencies"`
	Author      string   `json:"author"`
	License     string   `json:"license"`
}

// ModuleInfo represents information about a loaded module
type ModuleInfo struct {
	Manifest  ModuleManifest `json:"manifest"`
	Path      string         `json:"path"`
	LoadedAt  time.Time      `json:"loaded_at"`
	Signature string         `json:"signature,omitempty"`
}

// Job represents a background job
type Job struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Module      string                 `json:"module"`
	Command     string                 `json:"command"`
	Args        map[string]interface{} `json:"args"`
	CreatedAt   time.Time              `json:"created_at"`
	ExpiresAt   time.Time              `json:"expires_at"`
	Priority    int                    `json:"priority"`
	Status      string                 `json:"status"`
	Progress    *ProgressEvent         `json:"progress,omitempty"`
	Result      *ModuleResponse        `json:"result,omitempty"`
}

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

// Validate checks if the request is valid
func (r *ModuleRequest) Validate() error {
	if r.Command == "" {
		return ErrInvalidRequest("command is required")
	}
	if r.Timeout <= 0 {
		r.Timeout = 300 // Default 5 minutes
	}
	return nil
}

// Validate checks if the response is valid
func (r *ModuleResponse) Validate() error {
	if !r.Success && r.Error == "" {
		return ErrInvalidResponse("error message required for failed response")
	}
	return nil
}

// Validate checks if the progress event is valid
func (p *ProgressEvent) Validate() error {
	if p.Stage == "" {
		return ErrInvalidProgress("stage is required")
	}
	if p.Total < 0 {
		return ErrInvalidProgress("total must be non-negative")
	}
	if p.Current < 0 {
		return ErrInvalidProgress("current must be non-negative")
	}
	if p.Current > p.Total {
		return ErrInvalidProgress("current cannot exceed total")
	}
	if p.Percentage < 0 || p.Percentage > 100 {
		return ErrInvalidProgress("percentage must be between 0 and 100")
	}
	return nil
}

// Error types for bridge operations
type BridgeError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *BridgeError) Error() string {
	return e.Message
}

var (
	ErrInvalidRequest  = func(msg string) *BridgeError { return &BridgeError{Code: "INVALID_REQUEST", Message: msg} }
	ErrInvalidResponse = func(msg string) *BridgeError { return &BridgeError{Code: "INVALID_RESPONSE", Message: msg} }
	ErrInvalidProgress = func(msg string) *BridgeError { return &BridgeError{Code: "INVALID_PROGRESS", Message: msg} }
	ErrModuleNotFound  = func(msg string) *BridgeError { return &BridgeError{Code: "MODULE_NOT_FOUND", Message: msg} }
	ErrModuleTimeout   = func(msg string) *BridgeError { return &BridgeError{Code: "MODULE_TIMEOUT", Message: msg} }
	ErrModuleError     = func(msg string) *BridgeError { return &BridgeError{Code: "MODULE_ERROR", Message: msg} }
)

// JSON serialization helpers
func (r *ModuleRequest) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

func (r *ModuleResponse) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

func (p *ProgressEvent) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

func (m *ModuleManifest) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

func (j *Job) ToJSON() ([]byte, error) {
	return json.Marshal(j)
}

// FromJSON deserializes JSON into a ModuleRequest
func ModuleRequestFromJSON(data []byte) (*ModuleRequest, error) {
	var req ModuleRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

// FromJSON deserializes JSON into a ModuleResponse
func ModuleResponseFromJSON(data []byte) (*ModuleResponse, error) {
	var resp ModuleResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// FromJSON deserializes JSON into a ProgressEvent
func ProgressEventFromJSON(data []byte) (*ProgressEvent, error) {
	var progress ProgressEvent
	if err := json.Unmarshal(data, &progress); err != nil {
		return nil, err
	}
	return &progress, nil
}

// FromJSON deserializes JSON into a ModuleManifest
func ModuleManifestFromJSON(data []byte) (*ModuleManifest, error) {
	var manifest ModuleManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}
	return &manifest, nil
}

// FromJSON deserializes JSON into a Job
func JobFromJSON(data []byte) (*Job, error) {
	var job Job
	if err := json.Unmarshal(data, &job); err != nil {
		return nil, err
	}
	return &job, nil
}
