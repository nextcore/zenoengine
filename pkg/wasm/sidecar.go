package wasm

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// SidecarPlugin implements the Plugin interface by running a native process
type SidecarPlugin struct {
	binaryPath   string
	workDir      string
	cmd          *exec.Cmd
	stdin        io.WriteCloser
	stdout       *bufio.Scanner
	mu           sync.Mutex
	initialized  bool
	cancel       context.CancelFunc
	manager      *PluginManager
	pluginName   string
	pendingCalls map[string]chan *PluginResponse
	callID       uint64
}

// NewSidecarPlugin creates a new sidecar plugin wrapper
func NewSidecarPlugin(name string, binary string, workDir string, manager *PluginManager) *SidecarPlugin {
	return &SidecarPlugin{
		pluginName:   name,
		binaryPath:   binary,
		workDir:      workDir,
		manager:      manager,
		pendingCalls: make(map[string]chan *PluginResponse),
	}
}

func (p *SidecarPlugin) start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.initialized {
		return nil
	}

	fullPath := filepath.Join(p.workDir, p.binaryPath)

	// Create a long-lived context for the process
	procCtx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	p.cmd = exec.CommandContext(procCtx, fullPath)
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

	// Capture StdErr
	stderrPipe, err := p.cmd.StderrPipe()
	if err != nil {
		return err
	}
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			slog.Error("⚠️ [Sidecar StdErr]", "msg", scanner.Text(), "binary", p.binaryPath)
		}
	}()

	if err := p.cmd.Start(); err != nil {
		return err
	}

	// Start communication loop
	go p.commLoop(procCtx)

	// Watch for process exit
	go func() {
		err := p.cmd.Wait()
		p.mu.Lock()
		p.initialized = false
		// Clear pending calls
		for id, ch := range p.pendingCalls {
			ch <- &PluginResponse{Success: false, Error: "sidecar process exited"}
			delete(p.pendingCalls, id)
		}
		p.mu.Unlock()
		if err != nil {
			slog.Error("❌ Sidecar process exited with error", "error", err, "binary", p.binaryPath)
		} else {
			slog.Info("ℹ️ Sidecar process exited gracefully", "binary", p.binaryPath)
		}
	}()

	p.initialized = true
	slog.Info("🚀 Sidecar process started", "path", fullPath)
	return nil
}

func (p *SidecarPlugin) GetMetadata(ctx context.Context) (*PluginMetadata, error) {
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
	resp, err := p.call(ctx, "plugin_register_slots", nil)
	if err != nil {
		return nil, err
	}

	var slots []SlotDefinition
	dataJSON, _ := json.Marshal(resp.Data)

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
	return p.call(ctx, request.SlotName, request.Parameters)
}

func (p *SidecarPlugin) Cleanup(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cancel != nil {
		p.cancel()
	}

	if !p.initialized || p.cmd == nil {
		return nil
	}

	if p.stdin != nil {
		p.stdin.Close()
	}

	p.cmd.Process.Signal(os.Interrupt)
	time.AfterFunc(1*time.Second, func() {
		if p.cmd != nil && p.cmd.Process != nil {
			p.cmd.Process.Kill()
		}
	})

	p.initialized = false
	return nil
}

func (p *SidecarPlugin) call(ctx context.Context, method string, params interface{}) (*PluginResponse, error) {
	p.mu.Lock()
	if !p.initialized {
		p.mu.Unlock()
		if err := p.start(ctx); err != nil {
			return nil, err
		}
		p.mu.Lock()
	}

	p.callID++
	id := fmt.Sprintf("%d", p.callID)

	// Check if this is a "legacy" call (plugin_init, etc)
	isLegacy := (method == "plugin_init" || method == "plugin_register_slots")
	if isLegacy {
		id = "legacy"
	}

	req := map[string]interface{}{
		"type":       "guest_call",
		"id":         id,
		"slot_name":  method,
		"parameters": params,
	}

	reqJSON, _ := json.Marshal(req)
	ch := make(chan *PluginResponse, 1)
	p.pendingCalls[id] = ch

	_, err := fmt.Fprintln(p.stdin, string(reqJSON))
	p.mu.Unlock()

	if err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case resp := <-ch:
		return resp, nil
	}
}

func (p *SidecarPlugin) commLoop(ctx context.Context) {
	for p.stdout.Scan() {
		line := p.stdout.Bytes()

		// Try legacy parse first (for simple plugins that only output one JSON line per request)
		var legacyResp PluginResponse
		if err := json.Unmarshal(line, &legacyResp); err == nil && (legacyResp.Success || legacyResp.Error != "") {
			p.mu.Lock()
			if ch, ok := p.pendingCalls["legacy"]; ok {
				delete(p.pendingCalls, "legacy")
				ch <- &legacyResp
				p.mu.Unlock()
				continue
			}
			p.mu.Unlock()
		}

		var msg struct {
			Type     string          `json:"type"`
			ID       string          `json:"id"`
			Function string          `json:"function"`
			Params   json.RawMessage `json:"parameters"`
			Data     json.RawMessage `json:"data"`
			Error    string          `json:"error"`
			Success  bool            `json:"success"`
		}

		if err := json.Unmarshal(line, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "host_call":
			go p.handleHostCall(ctx, msg.ID, msg.Function, msg.Params)
		case "guest_response":
			p.mu.Lock()
			if ch, ok := p.pendingCalls[msg.ID]; ok {
				delete(p.pendingCalls, msg.ID)
				ch <- &PluginResponse{
					Success: msg.Success,
					Data:    decodeRaw(msg.Data),
					Error:   msg.Error,
				}
			}
			p.mu.Unlock()
		}
	}
}

func (p *SidecarPlugin) handleHostCall(ctx context.Context, id string, fn string, params json.RawMessage) {
	pName := p.pluginName
	// Note: We need scope in ctx. Sidecar execution usually has it.
	ctx = context.WithValue(ctx, "pluginName", pName)

	var result interface{}
	var err error

	hf := p.manager.hostFunctions
	pMap := make(map[string]interface{})
	json.Unmarshal(params, &pMap)

	switch fn {
	case "log":
		level, _ := pMap["level"].(string)
		message, _ := pMap["message"].(string)
		hf.OnLog(ctx, level, message)
	case "db_query":
		conn, _ := pMap["connection"].(string)
		sql, _ := pMap["sql"].(string)
		bind, _ := pMap["params"].(map[string]interface{})
		result, err = hf.OnDBQuery(ctx, conn, sql, bind)
	case "scope_get":
		key, _ := pMap["key"].(string)
		result, err = hf.OnScopeGet(ctx, key)
	case "scope_set":
		key, _ := pMap["key"].(string)
		val := pMap["value"]
		err = hf.OnScopeSet(ctx, key, val)
	default:
		err = fmt.Errorf("unknown host function: %s", fn)
	}

	response := map[string]interface{}{
		"type":    "host_response",
		"id":      id,
		"success": err == nil,
		"data":    result,
	}
	if err != nil {
		response["error"] = err.Error()
	}

	respJSON, _ := json.Marshal(response)
	p.mu.Lock()
	if p.initialized {
		fmt.Fprintln(p.stdin, string(respJSON))
	}
	p.mu.Unlock()
}

func decodeRaw(raw json.RawMessage) map[string]interface{} {
	var m map[string]interface{}
	json.Unmarshal(raw, &m)
	return m
}
