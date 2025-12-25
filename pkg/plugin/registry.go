package plugin

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/converso-empire/cli/pkg/auth"
	"github.com/converso-empire/cli/pkg/bridge"
	"github.com/converso-empire/cli/pkg/config"
	"github.com/converso-empire/cli/pkg/telemetry"
)

// PluginRegistry manages dynamic plugin loading and execution
type PluginRegistry struct {
	config     *config.Config
	logger     telemetry.Logger
	bridge     *bridge.JSONBridge
	modules    map[string]*ModuleInfo
	manifests  map[string]*bridge.ModuleManifest
	mu         sync.RWMutex
}

// ModuleInfo contains information about a loaded module
type ModuleInfo struct {
	Manifest  *bridge.ModuleManifest `json:"manifest"`
	Path      string                 `json:"path"`
	LoadedAt  time.Time              `json:"loaded_at"`
	Signature string                 `json:"signature,omitempty"`
}

// NewPluginRegistry creates a new plugin registry
func NewPluginRegistry(cfg *config.Config, logger telemetry.Logger, bridge *bridge.JSONBridge) *PluginRegistry {
	return &PluginRegistry{
		config:    cfg,
		logger:    logger,
		bridge:    bridge,
		modules:   make(map[string]*ModuleInfo),
		manifests: make(map[string]*bridge.ModuleManifest),
	}
}

// LoadPlugins scans for and loads available plugins
func (r *PluginRegistry) LoadPlugins() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.logger.Info("Loading plugins", "plugins_dir", r.config.PluginsDir)

	// Create plugins directory if it doesn't exist
	if err := os.MkdirAll(r.config.PluginsDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugins directory: %w", err)
	}

	// Scan for plugins
	entries, err := os.ReadDir(r.config.PluginsDir)
	if err != nil {
		return fmt.Errorf("failed to read plugins directory: %w", err)
	}

	loadedCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		moduleName := entry.Name()
		modulePath := filepath.Join(r.config.PluginsDir, moduleName)

		if err := r.loadModule(moduleName, modulePath); err != nil {
			r.logger.Warn("Failed to load module", "module", moduleName, "error", err)
			continue
		}

		loadedCount++
	}

	r.logger.Info("Plugins loaded successfully", "count", loadedCount)
	return nil
}

// loadModule loads a single module
func (r *PluginRegistry) loadModule(name, path string) error {
	// Check if module has a manifest
	manifestPath := filepath.Join(path, "manifest.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return fmt.Errorf("manifest.json not found")
	}

	// Read and validate manifest
	manifest, err := r.readManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	// Check if module has main file
	mainPath := filepath.Join(path, "__main__.py")
	if _, err := os.Stat(mainPath); os.IsNotExist(err) {
		return fmt.Errorf("__main__.py not found")
	}

	// Validate module
	if err := r.validateModule(manifest, path); err != nil {
		return fmt.Errorf("module validation failed: %w", err)
	}

	// Store module info
	moduleInfo := &ModuleInfo{
		Manifest: manifest,
		Path:     path,
		LoadedAt: time.Now(),
	}

	r.modules[name] = moduleInfo
	r.manifests[name] = manifest

	r.logger.Info("Module loaded", "name", name, "version", manifest.Version)
	return nil
}

// readManifest reads and parses a module manifest
func (r *PluginRegistry) readManifest(path string) (*bridge.ModuleManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	var manifest bridge.ModuleManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest JSON: %w", err)
	}

	// Validate manifest
	if err := r.validateManifest(&manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

// validateManifest validates a module manifest
func (r *PluginRegistry) validateManifest(manifest *bridge.ModuleManifest) error {
	if manifest.Name == "" {
		return fmt.Errorf("module name is required")
	}

	if manifest.Version == "" {
		return fmt.Errorf("module version is required")
	}

	if len(manifest.Commands) == 0 {
		return fmt.Errorf("module must define at least one command")
	}

	// Validate version format (semantic versioning)
	if !strings.Contains(manifest.Version, ".") {
		return fmt.Errorf("invalid version format, expected semantic versioning")
	}

	return nil
}

// validateModule validates a module's structure and dependencies
func (r *PluginRegistry) validateModule(manifest *bridge.ModuleManifest, path string) error {
	// Check required files
	requiredFiles := []string{
		"__main__.py",
		"manifest.json",
	}

	for _, file := range requiredFiles {
		filePath := filepath.Join(path, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("required file missing: %s", file)
		}
	}

	// Check Python syntax (basic validation)
	mainPath := filepath.Join(path, "__main__.py")
	if err := r.validatePythonSyntax(mainPath); err != nil {
		return fmt.Errorf("Python syntax error in __main__.py: %w", err)
	}

	// Check dependencies
	for _, dep := range manifest.Dependencies {
		if err := r.checkDependency(dep); err != nil {
			r.logger.Warn("Dependency check failed", "dependency", dep, "error", err)
			// Don't fail loading, just warn
		}
	}

	return nil
}

// validatePythonSyntax performs basic Python syntax validation
func (r *PluginRegistry) validatePythonSyntax(path string) error {
	// For now, we'll just check if the file is readable
	// In a production system, you might want to use python -m py_compile
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return fmt.Errorf("empty file")
	}

	return nil
}

// checkDependency checks if a Python dependency is available
func (r *PluginRegistry) checkDependency(dep string) error {
	// This is a simplified check - in production, you might want to
	// actually try importing the module
	if dep == "" {
		return fmt.Errorf("empty dependency name")
	}
	return nil
}

// ExecuteCommand executes a command on a loaded module
func (r *PluginRegistry) ExecuteCommand(module, command string, args map[string]interface{}, authTokens *auth.AuthTokens) (*bridge.ModuleResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check if module is loaded
	moduleInfo, exists := r.modules[module]
	if !exists {
		return nil, fmt.Errorf("module %s not found", module)
	}

	// Check if command is available
	manifest := moduleInfo.Manifest
	commandAvailable := false
	for _, cmd := range manifest.Commands {
		if cmd == command {
			commandAvailable = true
			break
		}
	}

	if !commandAvailable {
		return nil, fmt.Errorf("command %s not available in module %s", command, module)
	}

	// Create request
	req := &bridge.ModuleRequest{
		Command:     command,
		Args:        args,
		AuthToken:   authTokens.AccessToken,
		DeviceToken: authTokens.DeviceToken,
		Timeout:     300, // 5 minutes default
	}

	// Execute via bridge
	resp, err := r.bridge.Execute(context.Background(), module, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	return resp, nil
}

// ExecuteCommandWithProgress executes a command with progress tracking
func (r *PluginRegistry) ExecuteCommandWithProgress(module, command string, args map[string]interface{}, authTokens *auth.AuthTokens, progressChan chan<- *bridge.ProgressEvent) (*bridge.ModuleResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check if module is loaded
	moduleInfo, exists := r.modules[module]
	if !exists {
		return nil, fmt.Errorf("module %s not found", module)
	}

	// Check if command is available
	manifest := moduleInfo.Manifest
	commandAvailable := false
	for _, cmd := range manifest.Commands {
		if cmd == command {
			commandAvailable = true
			break
		}
	}

	if !commandAvailable {
		return nil, fmt.Errorf("command %s not available in module %s", command, module)
	}

	// Create request
	req := &bridge.ModuleRequest{
		Command:     command,
		Args:        args,
		AuthToken:   authTokens.AccessToken,
		DeviceToken: authTokens.DeviceToken,
		Timeout:     300, // 5 minutes default
	}

	// Execute via bridge with progress
	resp, err := r.bridge.ExecuteWithProgress(context.Background(), module, req, progressChan)
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	return resp, nil
}

// ListModules returns a list of loaded modules
func (r *PluginRegistry) ListModules() []*ModuleInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	modules := make([]*ModuleInfo, 0, len(r.modules))
	for _, module := range r.modules {
		modules = append(modules, module)
	}

	// Sort by name
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Manifest.Name < modules[j].Manifest.Name
	})

	return modules
}

// GetModuleInfo returns information about a specific module
func (r *PluginRegistry) GetModuleInfo(name string) (*ModuleInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	module, exists := r.modules[name]
	if !exists {
		return nil, fmt.Errorf("module %s not found", name)
	}

	return module, nil
}

// InstallModule installs a new module from a local path or URL
func (r *PluginRegistry) InstallModule(name, source string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if module already exists
	if _, exists := r.modules[name]; exists {
		return fmt.Errorf("module %s already exists", name)
	}

	// Create module directory
	modulePath := filepath.Join(r.config.PluginsDir, name)
	if err := os.MkdirAll(modulePath, 0755); err != nil {
		return fmt.Errorf("failed to create module directory: %w", err)
	}

	// Copy module files (simplified - in production, handle URLs, archives, etc.)
	if err := r.copyModuleFiles(source, modulePath); err != nil {
		return fmt.Errorf("failed to copy module files: %w", err)
	}

	// Load the new module
	if err := r.loadModule(name, modulePath); err != nil {
		// Clean up on failure
		os.RemoveAll(modulePath)
		return err
	}

	r.logger.Info("Module installed successfully", "name", name)
	return nil
}

// copyModuleFiles copies module files from source to destination
func (r *PluginRegistry) copyModuleFiles(source, destination string) error {
	// This is a simplified implementation
	// In production, you'd handle different source types (local paths, URLs, archives)
	
	// For now, assume source is a local directory
	if _, err := os.Stat(source); os.IsNotExist(err) {
		return fmt.Errorf("source directory does not exist: %s", source)
	}

	// Copy files recursively
	return filepath.Walk(source, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		// Create destination path
		destPath := filepath.Join(destination, relPath)

		// Copy file
		return r.copyFile(path, destPath)
	})
}

// copyFile copies a single file
func (r *PluginRegistry) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// UninstallModule removes a module
func (r *PluginRegistry) UninstallModule(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if module exists
	if _, exists := r.modules[name]; !exists {
		return fmt.Errorf("module %s not found", name)
	}

	// Remove from registry
	delete(r.modules, name)
	delete(r.manifests, name)

	// Remove directory
	modulePath := filepath.Join(r.config.PluginsDir, name)
	if err := os.RemoveAll(modulePath); err != nil {
		return fmt.Errorf("failed to remove module directory: %w", err)
	}

	r.logger.Info("Module uninstalled successfully", "name", name)
	return nil
}

// UpdateModule updates an existing module
func (r *PluginRegistry) UpdateModule(name, source string) error {
	// For now, uninstall and reinstall
	// In production, you'd implement proper update logic
	if err := r.UninstallModule(name); err != nil {
		return err
	}

	return r.InstallModule(name, source)
}
