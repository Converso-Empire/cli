package bridge

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/converso-empire/cli/pkg/telemetry"
)

// JSONBridge implements JSON-based IPC communication with Python modules
type JSONBridge struct {
	pythonPath string
	modulesDir string
	logger     telemetry.Logger
	mu         sync.RWMutex
	processes  map[string]*exec.Cmd
}

// NewJSONBridge creates a new JSON IPC bridge
func NewJSONBridge(pythonPath, modulesDir string, logger telemetry.Logger) *JSONBridge {
	return &JSONBridge{
		pythonPath: pythonPath,
		modulesDir: modulesDir,
		logger:     logger,
		processes:  make(map[string]*exec.Cmd),
	}
}

// Execute executes a command on a Python module
func (b *JSONBridge) Execute(ctx context.Context, module string, req *ModuleRequest) (*ModuleResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	b.logger.Info("Executing module command",
		"module", module,
		"command", req.Command,
		"timeout", req.Timeout,
	)

	// Find the module
	modulePath, err := b.findModule(module)
	if err != nil {
		return nil, ErrModuleNotFound(fmt.Sprintf("module %s not found: %v", module, err))
	}

	// Launch Python subprocess
	cmd, stdout, stderr, err := b.launchPythonProcess(modulePath)
	if err != nil {
		return nil, fmt.Errorf("failed to launch Python process: %w", err)
	}

	// Store process reference
	processID := fmt.Sprintf("%s-%d", module, time.Now().UnixNano())
	b.mu.Lock()
	b.processes[processID] = cmd
	b.mu.Unlock()

	defer func() {
		b.mu.Lock()
		delete(b.processes, processID)
		b.mu.Unlock()
		cmd.Process.Kill()
		cmd.Wait()
	}()

	// Set up context with timeout
	ctx, cancel := context.WithTimeout(ctx, time.Duration(req.Timeout)*time.Second)
	defer cancel()

	// Send request to Python module
	if err := b.sendRequest(stdout, req); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Read response from Python module
	resp, err := b.readResponse(ctx, stderr)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if err := resp.Validate(); err != nil {
		return nil, err
	}

	b.logger.Info("Module command completed successfully",
		"module", module,
		"command", req.Command,
		"success", resp.Success,
	)

	return resp, nil
}

// ExecuteWithProgress executes a command with progress tracking
func (b *JSONBridge) ExecuteWithProgress(ctx context.Context, module string, req *ModuleRequest, progressChan chan<- *ProgressEvent) (*ModuleResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	b.logger.Info("Executing module command with progress",
		"module", module,
		"command", req.Command,
		"timeout", req.Timeout,
	)

	// Find the module
	modulePath, err := b.findModule(module)
	if err != nil {
		return nil, ErrModuleNotFound(fmt.Sprintf("module %s not found: %v", module, err))
	}

	// Launch Python subprocess
	cmd, stdout, stderr, err := b.launchPythonProcess(modulePath)
	if err != nil {
		return nil, fmt.Errorf("failed to launch Python process: %w", err)
	}

	// Store process reference
	processID := fmt.Sprintf("%s-%d", module, time.Now().UnixNano())
	b.mu.Lock()
	b.processes[processID] = cmd
	b.mu.Unlock()

	defer func() {
		b.mu.Lock()
		delete(b.processes, processID)
		b.mu.Unlock()
		cmd.Process.Kill()
		cmd.Wait()
	}()

	// Set up context with timeout
	ctx, cancel := context.WithTimeout(ctx, time.Duration(req.Timeout)*time.Second)
	defer cancel()

	// Send request to Python module
	if err := b.sendRequest(stdout, req); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Read response with progress tracking
	resp, err := b.readResponseWithProgress(ctx, stderr, progressChan)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if err := resp.Validate(); err != nil {
		return nil, err
	}

	b.logger.Info("Module command completed successfully",
		"module", module,
		"command", req.Command,
		"success", resp.Success,
	)

	return resp, nil
}

// findModule finds the path to a Python module
func (b *JSONBridge) findModule(module string) (string, error) {
	// Look for module in the modules directory
	modulePath := fmt.Sprintf("%s/%s/__main__.py", b.modulesDir, module)
	
	// Check if the module file exists
	if _, err := os.Stat(modulePath); os.IsNotExist(err) {
		return "", fmt.Errorf("module file not found: %s", modulePath)
	}

	return modulePath, nil
}

// launchPythonProcess launches a Python subprocess for a module
func (b *JSONBridge) launchPythonProcess(modulePath string) (*exec.Cmd, io.WriteCloser, io.ReadCloser, error) {
	// Construct Python command
	cmd := exec.Command(b.pythonPath, modulePath)

	// Set up pipes for communication
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, nil, err
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		return nil, nil, nil, err
	}

	return cmd, stdin, stderr, nil
}

// sendRequest sends a request to the Python module
func (b *JSONBridge) sendRequest(stdin io.WriteCloser, req *ModuleRequest) error {
	data, err := req.ToJSON()
	if err != nil {
		return err
	}

	// Write request to stdin
	_, err = stdin.Write(data)
	if err != nil {
		return err
	}

	// Write newline to signal end of request
	_, err = stdin.Write([]byte("\n"))
	if err != nil {
		return err
	}

	// Close stdin to signal end of input
	return stdin.Close()
}

// readResponse reads a response from the Python module
func (b *JSONBridge) readResponse(ctx context.Context, stderr io.ReadCloser) (*ModuleResponse, error) {
	reader := bufio.NewReader(stderr)
	
	select {
	case <-ctx.Done():
		return nil, ErrModuleTimeout("module execution timed out")
	default:
		// Read response line
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, ErrModuleError("module process ended unexpectedly")
			}
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		// Parse response
		resp, err := ModuleResponseFromJSON([]byte(strings.TrimSpace(line)))
		if err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return resp, nil
	}
}

// readResponseWithProgress reads a response with progress tracking
func (b *JSONBridge) readResponseWithProgress(ctx context.Context, stderr io.ReadCloser, progressChan chan<- *ProgressEvent) (*ModuleResponse, error) {
	reader := bufio.NewReader(stderr)
	
	for {
		select {
		case <-ctx.Done():
			return nil, ErrModuleTimeout("module execution timed out")
		default:
			// Read line
			line, err := reader.ReadString('\n')
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil, ErrModuleError("module process ended unexpectedly")
				}
				return nil, fmt.Errorf("failed to read response: %w", err)
			}

			// Try to parse as progress event first
			progress, err := ProgressEventFromJSON([]byte(strings.TrimSpace(line)))
			if err == nil {
				// Validate progress event
				if err := progress.Validate(); err == nil {
					progress.Timestamp = time.Now()
					progressChan <- progress
					continue
				}
			}

			// Try to parse as response
			resp, err := ModuleResponseFromJSON([]byte(strings.TrimSpace(line)))
			if err != nil {
				return nil, fmt.Errorf("failed to parse response: %w", err)
			}

			return resp, nil
		}
	}
}

// GetPythonPath returns the path to the Python interpreter
func GetPythonPath() string {
	// Try common Python paths
	paths := []string{
		"python3",
		"python",
		"python.exe",
	}

	for _, path := range paths {
		if _, err := exec.LookPath(path); err == nil {
			return path
		}
	}

	// Try platform-specific paths
	if runtime.GOOS == "windows" {
		return "python.exe"
	}
	return "python3"
}

// CheckPythonAvailability checks if Python is available
func CheckPythonAvailability() error {
	pythonPath := GetPythonPath()
	cmd := exec.Command(pythonPath, "--version")
	return cmd.Run()
}
