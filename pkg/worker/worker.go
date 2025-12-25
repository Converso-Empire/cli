package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/converso-empire/cli/pkg/auth"
	"github.com/converso-empire/cli/pkg/bridge"
	"github.com/converso-empire/cli/pkg/config"
	"github.com/converso-empire/cli/pkg/telemetry"
)

// Worker manages background tasks and job processing
type Worker struct {
	config     *config.Config
	logger     telemetry.Logger
	httpClient *http.Client
	authTokens *auth.AuthTokens
	jobQueue   chan *Job
	running    bool
	mu         sync.RWMutex
	wg         sync.WaitGroup
	stopCh     chan struct{}
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
	Progress    *bridge.ProgressEvent  `json:"progress,omitempty"`
	Result      *bridge.ModuleResponse `json:"result,omitempty"`
}

// JobStatus represents job status
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

// NewWorker creates a new background worker
func NewWorker(cfg *config.Config, logger telemetry.Logger) *Worker {
	return &Worker{
		config:     cfg,
		logger:     logger,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		jobQueue:   make(chan *Job, 100),
		stopCh:     make(chan struct{}),
	}
}

// Start starts the background worker
func (w *Worker) Start() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.running {
		return fmt.Errorf("worker is already running")
	}

	// Load authentication tokens
	tokens, err := w.loadAuthTokens()
	if err != nil {
		return fmt.Errorf("failed to load authentication tokens: %w", err)
	}
	w.authTokens = tokens

	w.running = true
	w.wg.Add(3)

	// Start job polling goroutine
	go w.pollJobs()

	// Start job processing goroutine
	go w.processJobs()

	// Start status reporting goroutine
	go w.reportStatus()

	w.logger.Info("Background worker started")
	return nil
}

// Stop stops the background worker
func (w *Worker) Stop() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return fmt.Errorf("worker is not running")
	}

	w.running = false
	close(w.stopCh)
	w.wg.Wait()

	w.logger.Info("Background worker stopped")
	return nil
}

// pollJobs polls the backend for new jobs
func (w *Worker) pollJobs() {
	defer w.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := w.fetchJobs(); err != nil {
				w.logger.Error("Failed to fetch jobs", "error", err)
			}
		case <-w.stopCh:
			return
		}
	}
}

// fetchJobs fetches jobs from the backend API
func (w *Worker) fetchJobs() error {
	if w.authTokens == nil || w.authTokens.IsExpired() {
		return fmt.Errorf("authentication required")
	}

	url := fmt.Sprintf("%s/api/v1/jobs/pending", w.config.APIEndpoint)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+w.authTokens.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch jobs: HTTP %d", resp.StatusCode)
	}

	var jobs []Job
	if err := json.NewDecoder(resp.Body).Decode(&jobs); err != nil {
		return err
	}

	// Add jobs to queue
	for _, job := range jobs {
		select {
		case w.jobQueue <- &job:
			w.logger.Info("Job added to queue", "job_id", job.ID, "module", job.Module)
		case <-w.stopCh:
			return nil
		default:
			w.logger.Warn("Job queue full, skipping job", "job_id", job.ID)
		}
	}

	return nil
}

// processJobs processes jobs from the queue
func (w *Worker) processJobs() {
	defer w.wg.Done()

	for {
		select {
		case job := <-w.jobQueue:
			w.processJob(job)
		case <-w.stopCh:
			return
		}
	}
}

// processJob processes a single job
func (w *Worker) processJob(job *Job) {
	w.logger.Info("Processing job", "job_id", job.ID, "module", job.Module, "command", job.Command)

	// Update job status
	job.Status = string(JobStatusRunning)
	job.Progress = &bridge.ProgressEvent{
		Stage:      "starting",
		Current:    0,
		Total:      100,
		Percentage: 0,
		Message:    "Starting job...",
		Timestamp:  time.Now(),
	}

	// Report job start
	if err := w.reportJobStatus(job); err != nil {
		w.logger.Error("Failed to report job start", "job_id", job.ID, "error", err)
	}

	// Execute job
	progressChan := make(chan *bridge.ProgressEvent, 100)
	
	go func() {
		for progress := range progressChan {
			job.Progress = progress
			w.reportJobProgress(job)
		}
	}()

	// Here you would integrate with the plugin system
	// For now, simulate job execution
	result, err := w.executeJob(job, progressChan)
	close(progressChan)

	if err != nil {
		job.Status = string(JobStatusFailed)
		job.Result = &bridge.ModuleResponse{
			Success: false,
			Data:    map[string]interface{}{},
			Error:   err.Error(),
		}
		w.logger.Error("Job failed", "job_id", job.ID, "error", err)
	} else {
		job.Status = string(JobStatusCompleted)
		job.Result = result
		w.logger.Info("Job completed", "job_id", job.ID)
	}

	// Report final status
	if err := w.reportJobStatus(job); err != nil {
		w.logger.Error("Failed to report job completion", "job_id", job.ID, "error", err)
	}
}

// executeJob executes a job (placeholder implementation)
func (w *Worker) executeJob(job *Job, progressChan chan<- *bridge.ProgressEvent) (*bridge.ModuleResponse, error) {
	// Simulate job execution with progress
	stages := []struct {
		stage     string
		progress  int
		message   string
		duration  time.Duration
	}{
		{"preparing", 10, "Preparing job...", 1 * time.Second},
		{"executing", 70, "Executing job...", 3 * time.Second},
		{"finalizing", 90, "Finalizing...", 1 * time.Second},
		{"completed", 100, "Job completed!", 0},
	}

	for _, stage := range stages {
		// Send progress update
		progressChan <- &bridge.ProgressEvent{
			Stage:      stage.stage,
			Current:    int64(stage.progress),
			Total:      100,
			Percentage: float64(stage.progress),
			Message:    stage.message,
			Timestamp:  time.Now(),
		}

		// Simulate work
		if stage.duration > 0 {
			time.Sleep(stage.duration)
		}
	}

	// Return simulated result
	return &bridge.ModuleResponse{
		Success: true,
		Data: map[string]interface{}{
			"job_id":     job.ID,
			"module":     job.Module,
			"command":    job.Command,
			"completed":  true,
			"timestamp":  time.Now().Format(time.RFC3339),
			"result":     "Job executed successfully",
		},
	}, nil
}

// reportStatus reports worker status to backend
func (w *Worker) reportStatus() {
	defer w.wg.Done()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := w.reportWorkerStatus(); err != nil {
				w.logger.Error("Failed to report worker status", "error", err)
			}
		case <-w.stopCh:
			return
		}
	}
}

// reportWorkerStatus reports worker status to backend
func (w *Worker) reportWorkerStatus() error {
	if w.authTokens == nil || w.authTokens.IsExpired() {
		return fmt.Errorf("authentication required")
	}

	status := map[string]interface{}{
		"status":     "running",
		"queue_size": len(w.jobQueue),
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	data, err := json.Marshal(status)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/worker/status", w.config.APIEndpoint)
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+w.authTokens.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to report status: HTTP %d", resp.StatusCode)
	}

	return nil
}

// reportJobStatus reports job status to backend
func (w *Worker) reportJobStatus(job *Job) error {
	if w.authTokens == nil || w.authTokens.IsExpired() {
		return fmt.Errorf("authentication required")
	}

	url := fmt.Sprintf("%s/api/v1/jobs/%s/status", w.config.APIEndpoint, job.ID)
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+w.authTokens.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to report job status: HTTP %d", resp.StatusCode)
	}

	return nil
}

// reportJobProgress reports job progress to backend
func (w *Worker) reportJobProgress(job *Job) error {
	if w.authTokens == nil || w.authTokens.IsExpired() {
		return fmt.Errorf("authentication required")
	}

	url := fmt.Sprintf("%s/api/v1/jobs/%s/progress", w.config.APIEndpoint, job.ID)
	data, err := json.Marshal(job.Progress)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+w.authTokens.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to report job progress: HTTP %d", resp.StatusCode)
	}

	return nil
}

// loadAuthTokens loads authentication tokens from storage
func (w *Worker) loadAuthTokens() (*auth.AuthTokens, error) {
	// This would integrate with the auth storage system
	// For now, return a placeholder
	return &auth.AuthTokens{
		AccessToken: "placeholder",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}, nil
}

// IsRunning returns whether the worker is running
func (w *Worker) IsRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.running
}

// GetQueueSize returns the current queue size
func (w *Worker) GetQueueSize() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.jobQueue)
}
