package wasm

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"path/filepath"
	"sync"
)

// SidecarPlugin implements the Plugin interface by running a native process
type SidecarPlugin struct {
	binaryPath string
	workDir    string
	cmd        *exec.Cmd
	stdin      io.WriteCloser
	stdout     *bufio.Scanner
	mu         sync.Mutex
	initialized bool
}

// NewSidecarPlugin creates a new sidecar plugin wrapper
func NewSidecarPlugin(binary string, workDir string) *SidecarPlugin {
	return &SidecarPlugin{
		binaryPath: binary,
		workDir:    workDir,
	}
}

func (p *SidecarPlugin) start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.initialized {
		return nil
	}

	fullPath := filepath.Join(p.workDir, p.binaryPath)
	p.cmd = exec.CommandContext(ctx, fullPath)
	p.cmd.Dir = p.workDir

	var err error
	p.stdin, err = p.cmd.StdinPipe()
	if err != nil {
		return err
	}

	stdoutPipe, err := p.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	p.stdout = bufio.NewScanner(stdoutPipe)

	if err := p.cmd.Start(); err != nil {
		return err
	}

	p.initialized = true
	slog.Info("🚀 Sidecar process started", "path", fullPath)
	return nil
}

func (p *SidecarPlugin) GetMetadata(ctx context.Context) (*PluginMetadata, error) {
	if err := p.start(ctx); err != nil {
		return nil, err
	}

	resp, err := p.call(ctx, "plugin_init", nil)
	if err != nil {
		return nil, err
	}

	var metadata PluginMetadata
	dataJSON, _ := json.Marshal(resp.Data)
	if err := json.Unmarshal(dataJSON, &metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

func (p *SidecarPlugin) GetSlots(ctx context.Context) ([]SlotDefinition, error) {
	if err := p.start(ctx); err != nil {
		return nil, err
	}

	resp, err := p.call(ctx, "plugin_register_slots", nil)
	if err != nil {
		return nil, err
	}

	// In Sidecar, data might be a list directly or wrapped
	var slots []SlotDefinition
	dataJSON, _ := json.Marshal(resp.Data)

	// Handle both cases: { "slots": [...] } or [...]
	var wrapper struct {
		Slots []SlotDefinition `json:"slots"`
	}
	if err := json.Unmarshal(dataJSON, &wrapper); err == nil && len(wrapper.Slots) > 0 {
		slots = wrapper.Slots
	} else {
		if err := json.Unmarshal(dataJSON, &slots); err != nil {
			return nil, fmt.Errorf("failed to parse slots: %v", err)
		}
	}

	return slots, nil
}

func (p *SidecarPlugin) Execute(ctx context.Context, request *PluginRequest) (*PluginResponse, error) {
	if err := p.start(ctx); err != nil {
		return nil, err
	}
	return p.call(ctx, request.SlotName, request.Parameters)
}

func (p *SidecarPlugin) Cleanup(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.initialized || p.cmd == nil {
		return nil
	}

	if p.stdin != nil {
		p.stdin.Close()
	}

	if err := p.cmd.Process.Kill(); err != nil {
		return err
	}

	p.initialized = false
	return nil
}

func (p *SidecarPlugin) call(ctx context.Context, method string, params interface{}) (*PluginResponse, error) {
	// Sidecars use simple JSON-RPC over StdIn/StdOut
	req := map[string]interface{}{
		"slot_name":  method,
		"parameters": params,
	}

	reqJSON, _ := json.Marshal(req)

	p.mu.Lock()
	_, err := fmt.Fprintln(p.stdin, string(reqJSON))
	p.mu.Unlock()

	if err != nil {
		return nil, fmt.Errorf("failed to send request to sidecar: %w", err)
	}

	// Read response
	if p.stdout.Scan() {
		line := p.stdout.Bytes()
		var resp PluginResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			return nil, fmt.Errorf("invalid response from sidecar: %v", err)
		}
		return &resp, nil
	}

	if err := p.stdout.Err(); err != nil {
		return nil, fmt.Errorf("sidecar read error: %w", err)
	}

	return nil, fmt.Errorf("sidecar process exited unexpectedly")
}
