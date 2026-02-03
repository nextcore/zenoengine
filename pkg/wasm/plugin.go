package wasm

import (
	"context"
	"encoding/json"
	"fmt"
)

// PluginMetadata contains information about a plugin
type PluginMetadata struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Author      string            `json:"author,omitempty"`
	Description string            `json:"description,omitempty"`
	License     string            `json:"license,omitempty"`
	Homepage    string            `json:"homepage,omitempty"`
	Config      map[string]string `json:"config,omitempty"`
}

// SlotDefinition defines a slot provided by a plugin
type SlotDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Example     string                 `json:"example,omitempty"`
	Inputs      map[string]InputMeta   `json:"inputs,omitempty"`
	Outputs     map[string]interface{} `json:"outputs,omitempty"`
}

// InputMeta describes an input parameter for a slot
type InputMeta struct {
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
	Description string      `json:"description,omitempty"`
}

// PluginRequest represents a request to execute a slot
type PluginRequest struct {
	SlotName   string                 `json:"slot_name"`
	Parameters map[string]interface{} `json:"parameters"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

// PluginResponse represents the result of slot execution
type PluginResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// HostRequest represents a request from plugin to host
type HostRequest struct {
	Function   string                 `json:"function"`
	Parameters map[string]interface{} `json:"parameters"`
}

// HostResponse represents host function response
type HostResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// Plugin interface that all plugins (WASM or Sidecar) must implement
type Plugin interface {
	// GetMetadata returns plugin metadata
	GetMetadata(ctx context.Context) (*PluginMetadata, error)

	// GetSlots returns all slots provided by this plugin
	GetSlots(ctx context.Context) ([]SlotDefinition, error)

	// Execute executes a slot with given parameters
	Execute(ctx context.Context, request *PluginRequest) (*PluginResponse, error)

	// Cleanup performs cleanup when plugin is unloaded
	Cleanup(ctx context.Context) error
}

// WASMPlugin wraps a WASM module to implement Plugin interface
type WASMPlugin struct {
	runtime    *Runtime
	moduleName string
}

// NewWASMPlugin creates a new WASM plugin wrapper
func NewWASMPlugin(runtime *Runtime, moduleName string) *WASMPlugin {
	return &WASMPlugin{
		runtime:    runtime,
		moduleName: moduleName,
	}
}

// GetMetadata calls plugin_init() and returns metadata
func (p *WASMPlugin) GetMetadata(ctx context.Context) (*PluginMetadata, error) {
	// Call plugin_init() which returns JSON pointer
	results, err := p.runtime.CallFunction(ctx, p.moduleName, "plugin_init")
	if err != nil {
		return nil, fmt.Errorf("failed to call plugin_init: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("plugin_init returned no results")
	}

	// Read JSON from memory
	ptr := uint32(results[0])

	// Read until null terminator or reasonable length
	// For now, assume max 4KB metadata
	jsonStr, err := p.readStringFromPtr(ptr, 4096)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	// Parse JSON
	var metadata PluginMetadata
	if err := json.Unmarshal([]byte(jsonStr), &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata JSON: %w", err)
	}

	return &metadata, nil
}

// GetSlots calls plugin_register_slots() and returns slot definitions
func (p *WASMPlugin) GetSlots(ctx context.Context) ([]SlotDefinition, error) {
	// Call plugin_register_slots() which returns JSON pointer
	results, err := p.runtime.CallFunction(ctx, p.moduleName, "plugin_register_slots")
	if err != nil {
		return nil, fmt.Errorf("failed to call plugin_register_slots: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("plugin_register_slots returned no results")
	}

	// Read JSON from memory
	ptr := uint32(results[0])
	jsonStr, err := p.readStringFromPtr(ptr, 8192) // Max 8KB for slots
	if err != nil {
		return nil, fmt.Errorf("failed to read slots: %w", err)
	}

	// Parse JSON
	var slots []SlotDefinition
	if err := json.Unmarshal([]byte(jsonStr), &slots); err != nil {
		return nil, fmt.Errorf("failed to parse slots JSON: %w", err)
	}

	return slots, nil
}

// Execute calls plugin_execute() with request parameters
func (p *WASMPlugin) Execute(ctx context.Context, request *PluginRequest) (*PluginResponse, error) {
	// Serialize request to JSON
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Write request JSON to WASM memory
	slotNamePtr, slotNameLen, err := p.runtime.WriteString(ctx, p.moduleName, request.SlotName)
	if err != nil {
		return nil, fmt.Errorf("failed to write slot name: %w", err)
	}

	paramsPtr, paramsLen, err := p.runtime.WriteString(ctx, p.moduleName, string(requestJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to write parameters: %w", err)
	}

	// Call plugin_execute(slot_name_ptr, slot_name_len, params_ptr, params_len)
	results, err := p.runtime.CallFunction(ctx, p.moduleName, "plugin_execute",
		uint64(slotNamePtr), uint64(slotNameLen),
		uint64(paramsPtr), uint64(paramsLen))
	if err != nil {
		return nil, fmt.Errorf("failed to call plugin_execute: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("plugin_execute returned no results")
	}

	// Read response JSON from memory
	responsePtr := uint32(results[0])
	responseJSON, err := p.readStringFromPtr(responsePtr, 16384) // Max 16KB response
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response JSON
	var response PluginResponse
	if err := json.Unmarshal([]byte(responseJSON), &response); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	return &response, nil
}

// Cleanup calls plugin_cleanup()
func (p *WASMPlugin) Cleanup(ctx context.Context) error {
	_, err := p.runtime.CallFunction(ctx, p.moduleName, "plugin_cleanup")
	if err != nil {
		return fmt.Errorf("failed to call plugin_cleanup: %w", err)
	}
	return nil
}

// readStringFromPtr reads a null-terminated string from WASM memory
func (p *WASMPlugin) readStringFromPtr(ptr uint32, maxLen uint32) (string, error) {
	memory, err := p.runtime.GetMemory(p.moduleName)
	if err != nil {
		return "", err
	}

	// Read bytes until null terminator or maxLen
	bytes, ok := memory.Read(ptr, maxLen)
	if !ok {
		return "", fmt.Errorf("failed to read memory at %d", ptr)
	}

	// Find null terminator
	length := uint32(0)
	for i, b := range bytes {
		if b == 0 {
			length = uint32(i)
			break
		}
	}

	if length == 0 {
		// No null terminator found, use maxLen
		length = maxLen
	}

	return string(bytes[:length]), nil
}
