package wasm

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// PluginManifest represents the plugin configuration
type PluginManifest struct {
	Name        string                 `yaml:"name"`
	Version     string                 `yaml:"version"`
	Author      string                 `yaml:"author,omitempty"`
	Description string                 `yaml:"description,omitempty"`
	License     string                 `yaml:"license,omitempty"`
	Homepage    string                 `yaml:"homepage,omitempty"`
	Binary      string                 `yaml:"binary"` // WASM file name
	Requires    map[string]string      `yaml:"requires,omitempty"`
	Permissions Permissions            `yaml:"permissions,omitempty"`
	Config      map[string]ConfigParam `yaml:"config,omitempty"`
}

// Permissions defines what a plugin can access
type Permissions struct {
	Network    []string `yaml:"network,omitempty"`    // Allowed URLs/patterns
	Env        []string `yaml:"env,omitempty"`        // Allowed env vars
	Scope      []string `yaml:"scope,omitempty"`      // read, write
	Filesystem []string `yaml:"filesystem,omitempty"` // Allowed paths
	Database   []string `yaml:"database,omitempty"`   // Allowed databases
}

// ConfigParam defines a configuration parameter
type ConfigParam struct {
	Type        string `yaml:"type"`
	Required    bool   `yaml:"required"`
	Env         string `yaml:"env,omitempty"`
	Default     string `yaml:"default,omitempty"`
	Description string `yaml:"description,omitempty"`
}

// PluginManager manages loading and lifecycle of WASM plugins
type PluginManager struct {
	runtime       *Runtime
	hostFunctions *HostFunctions
	plugins       map[string]*LoadedPlugin
	pluginDir     string
	mu            sync.RWMutex // Protects plugins map for hot reload
}

// LoadedPlugin represents a loaded WASM plugin
type LoadedPlugin struct {
	Manifest *PluginManifest
	Plugin   *WASMPlugin
	Slots    []SlotDefinition
	Path     string // Path to plugin directory for reload
}

// NewPluginManager creates a new plugin manager
func NewPluginManager(ctx context.Context, pluginDir string) (*PluginManager, error) {
	// Create WASM runtime
	runtime, err := NewRuntime(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create WASM runtime: %w", err)
	}

	// Create host functions
	hostFunctions := NewHostFunctions(runtime)

	// Register host functions
	if err := hostFunctions.RegisterHostFunctions(ctx, "env"); err != nil {
		return nil, fmt.Errorf("failed to register host functions: %w", err)
	}

	return &PluginManager{
		runtime:       runtime,
		hostFunctions: hostFunctions,
		plugins:       make(map[string]*LoadedPlugin),
		pluginDir:     pluginDir,
	}, nil
}

// LoadPlugin loads a single plugin from a directory
func (pm *PluginManager) LoadPlugin(pluginPath string) error {
	// Read manifest
	manifestPath := filepath.Join(pluginPath, "manifest.yaml")
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest PluginManifest
	if err := yaml.Unmarshal(manifestBytes, &manifest); err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Validate manifest
	if manifest.Name == "" {
		return fmt.Errorf("plugin name is required")
	}
	if manifest.Binary == "" {
		return fmt.Errorf("plugin binary is required")
	}

	// Check if already loaded (with read lock)
	pm.mu.RLock()
	_, exists := pm.plugins[manifest.Name]
	pm.mu.RUnlock()
	if exists {
		return fmt.Errorf("plugin %s already loaded", manifest.Name)
	}

	// Load WASM module
	wasmPath := filepath.Join(pluginPath, manifest.Binary)
	if err := pm.runtime.LoadModule(manifest.Name, wasmPath); err != nil {
		return fmt.Errorf("failed to load WASM module: %w", err)
	}

	// Create plugin wrapper
	plugin := NewWASMPlugin(pm.runtime, manifest.Name)

	// Get plugin metadata
	metadata, err := plugin.GetMetadata(context.Background())
	if err != nil {
		pm.runtime.UnloadModule(manifest.Name)
		return fmt.Errorf("failed to get plugin metadata: %w", err)
	}

	slog.Info("Plugin metadata", "name", metadata.Name, "version", metadata.Version)

	// Get plugin slots
	slots, err := plugin.GetSlots(context.Background())
	if err != nil {
		pm.runtime.UnloadModule(manifest.Name)
		return fmt.Errorf("failed to get plugin slots: %w", err)
	}

	// Store loaded plugin (with write lock)
	pm.mu.Lock()
	pm.plugins[manifest.Name] = &LoadedPlugin{
		Manifest: &manifest,
		Plugin:   plugin,
		Slots:    slots,
		Path:     pluginPath,
	}
	pm.mu.Unlock()

	slog.Info("✅ Plugin loaded",
		"name", manifest.Name,
		"version", manifest.Version,
		"slots", len(slots))

	return nil
}

// LoadPluginsFromDir loads all plugins from a directory
func (pm *PluginManager) LoadPluginsFromDir(dir string) error {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		slog.Warn("Plugin directory does not exist", "dir", dir)
		return nil
	}

	// Find all plugin directories (containing manifest.yaml)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read plugin directory: %w", err)
	}

	loadedCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pluginPath := filepath.Join(dir, entry.Name())
		manifestPath := filepath.Join(pluginPath, "manifest.yaml")

		// Check if manifest exists
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			continue
		}

		// Load plugin
		if err := pm.LoadPlugin(pluginPath); err != nil {
			slog.Error("Failed to load plugin",
				"path", pluginPath,
				"error", err)
			// Continue loading other plugins
			continue
		}

		loadedCount++
	}

	slog.Info("🔌 Plugins loaded", "count", loadedCount, "dir", dir)

	return nil
}

// GetPlugin returns a loaded plugin by name
func (pm *PluginManager) GetPlugin(name string) (*LoadedPlugin, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	plugin, exists := pm.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not loaded", name)
	}
	return plugin, nil
}

// ExecuteSlot executes a plugin slot
func (pm *PluginManager) ExecuteSlot(ctx context.Context, pluginName, slotName string, params map[string]interface{}) (*PluginResponse, error) {
	plugin, err := pm.GetPlugin(pluginName)
	if err != nil {
		return nil, err
	}

	// Check if slot exists
	slotExists := false
	for _, slot := range plugin.Slots {
		if slot.Name == slotName {
			slotExists = true
			break
		}
	}
	if !slotExists {
		return nil, fmt.Errorf("slot %s not found in plugin %s", slotName, pluginName)
	}

	// Execute slot
	request := &PluginRequest{
		SlotName:   slotName,
		Parameters: params,
	}

	return plugin.Plugin.Execute(ctx, request)
}

// ListPlugins returns all loaded plugins
func (pm *PluginManager) ListPlugins() []*LoadedPlugin {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	plugins := make([]*LoadedPlugin, 0, len(pm.plugins))
	for _, p := range pm.plugins {
		plugins = append(plugins, p)
	}
	return plugins
}

// UnloadPlugin unloads a plugin
func (pm *PluginManager) UnloadPlugin(name string) error {
	plugin, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not loaded", name)
	}

	// Cleanup plugin
	if err := plugin.Plugin.Cleanup(context.Background()); err != nil {
		slog.Error("Failed to cleanup plugin", "name", name, "error", err)
	}

	// Unload WASM module
	if err := pm.runtime.UnloadModule(name); err != nil {
		return fmt.Errorf("failed to unload module: %w", err)
	}

	delete(pm.plugins, name)
	slog.Info("Plugin unloaded", "name", name)

	return nil
}

// UnloadAll unloads all plugins
func (pm *PluginManager) UnloadAll() error {
	for name := range pm.plugins {
		if err := pm.UnloadPlugin(name); err != nil {
			slog.Error("Failed to unload plugin", "name", name, "error", err)
		}
	}
	return nil
}

// Close closes the plugin manager and runtime
func (pm *PluginManager) Close() error {
	// Unload all plugins
	pm.UnloadAll()

	// Close runtime
	if err := pm.runtime.Close(); err != nil {
		return fmt.Errorf("failed to close runtime: %w", err)
	}

	return nil
}

// SetHostCallback sets a host function callback
func (pm *PluginManager) SetHostCallback(name string, callback interface{}) error {
	switch name {
	case "log":
		if cb, ok := callback.(func(context.Context, string, string)); ok {
			pm.hostFunctions.OnLog = cb
		} else {
			return fmt.Errorf("invalid callback type for log")
		}
	case "db_query":
		if cb, ok := callback.(func(context.Context, string, string, map[string]interface{}) ([]map[string]interface{}, error)); ok {
			pm.hostFunctions.OnDBQuery = cb
		} else {
			return fmt.Errorf("invalid callback type for db_query")
		}
	case "http_request":
		if cb, ok := callback.(func(context.Context, string, string, map[string]interface{}, map[string]interface{}) (map[string]interface{}, error)); ok {
			pm.hostFunctions.OnHTTPRequest = cb
		} else {
			return fmt.Errorf("invalid callback type for http_request")
		}
	case "scope_get":
		if cb, ok := callback.(func(context.Context, string) (interface{}, error)); ok {
			pm.hostFunctions.OnScopeGet = cb
		} else {
			return fmt.Errorf("invalid callback type for scope_get")
		}
	case "scope_set":
		if cb, ok := callback.(func(context.Context, string, interface{}) error); ok {
			pm.hostFunctions.OnScopeSet = cb
		} else {
			return fmt.Errorf("invalid callback type for scope_set")
		}
	case "file_read":
		if cb, ok := callback.(func(context.Context, string) (string, error)); ok {
			pm.hostFunctions.OnFileRead = cb
		} else {
			return fmt.Errorf("invalid callback type for file_read")
		}
	case "file_write":
		if cb, ok := callback.(func(context.Context, string, string) error); ok {
			pm.hostFunctions.OnFileWrite = cb
		} else {
			return fmt.Errorf("invalid callback type for file_write")
		}
	case "env_get":
		if cb, ok := callback.(func(context.Context, string) string); ok {
			pm.hostFunctions.OnEnvGet = cb
		} else {
			return fmt.Errorf("invalid callback type for env_get")
		}
	default:
		return fmt.Errorf("unknown host function: %s", name)
	}

	return nil
}

// CheckPermission checks if a plugin has permission for an operation
func (pm *PluginManager) CheckPermission(pluginName, permissionType, resource string) bool {
	plugin, err := pm.GetPlugin(pluginName)
	if err != nil {
		return false
	}

	switch permissionType {
	case "network":
		for _, pattern := range plugin.Manifest.Permissions.Network {
			if matchPattern(pattern, resource) {
				return true
			}
		}
	case "env":
		for _, envVar := range plugin.Manifest.Permissions.Env {
			if envVar == resource {
				return true
			}
		}
	case "filesystem":
		for _, path := range plugin.Manifest.Permissions.Filesystem {
			if strings.HasPrefix(resource, path) {
				return true
			}
		}
	case "database":
		for _, db := range plugin.Manifest.Permissions.Database {
			if db == resource {
				return true
			}
		}
	case "scope":
		for _, perm := range plugin.Manifest.Permissions.Scope {
			if perm == resource {
				return true
			}
		}
	}

	return false
}

// matchPattern matches a URL pattern (simple wildcard support)
func matchPattern(pattern, url string) bool {
	if pattern == "*" {
		return true
	}
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(url, prefix)
	}
	return pattern == url
}

// ReloadPlugin reloads a specific plugin by name
func (pm *PluginManager) ReloadPlugin(name string) error {
	slog.Info("Reloading plugin...", "name", name)

	// 1. Get plugin path and verify existence
	pm.mu.RLock()
	plugin, exists := pm.plugins[name]
	pm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	path := plugin.Path

	// 2. Unload existing plugin
	// Note: We remove from map first to allow LoadPlugin to succeed
	pm.mu.Lock()
	delete(pm.plugins, name)
	pm.mu.Unlock()

	// Unload from runtime
	if err := pm.runtime.UnloadModule(name); err != nil {
		slog.Warn("Failed to unload module during reload", "name", name, "error", err)
		// Continue anyway to try loading new version
	}

	// 3. Load plugin again
	if err := pm.LoadPlugin(path); err != nil {
		// If load fails, try to restore old plugin entry (best effort)
		pm.mu.Lock()
		pm.plugins[name] = plugin
		pm.mu.Unlock()
		return fmt.Errorf("failed to reload plugin: %w", err)
	}

	slog.Info("✅ Plugin reloaded successfully", "name", name)
	return nil
}

// ReloadAllPlugins reloads all currently loaded plugins
func (pm *PluginManager) ReloadAllPlugins() map[string]error {
	slog.Info("Reloading all plugins...")

	// Get list of plugins to reload
	pm.mu.RLock()
	plugins := make([]string, 0, len(pm.plugins))
	for name := range pm.plugins {
		plugins = append(plugins, name)
	}
	pm.mu.RUnlock()

	errors := make(map[string]error)
	for _, name := range plugins {
		if err := pm.ReloadPlugin(name); err != nil {
			slog.Error("Failed to reload plugin", "name", name, "error", err)
			errors[name] = err
		}
	}

	if len(errors) > 0 {
		slog.Warn("Reload completed with errors", "errors", len(errors))
	} else {
		slog.Info("✅ All plugins reloaded successfully")
	}

	return errors
}
