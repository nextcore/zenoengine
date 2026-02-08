package wasm

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func compilePlugin(t *testing.T, pluginDir string) string {
	// Read binary name from manifest to know what to output
	manifestPath := filepath.Join(pluginDir, "manifest.yaml")
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("Failed to read manifest: %v", err)
	}

	binaryName := "test-plugin" // Default fallback
	for _, line := range strings.Split(string(manifestBytes), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "binary:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				binaryName = strings.TrimSpace(parts[1])
				// Remove comments if any
				if idx := strings.Index(binaryName, "#"); idx != -1 {
					binaryName = strings.TrimSpace(binaryName[:idx])
				}
			}
			break
		}
	}

	cmd := exec.Command("go", "build", "-o", binaryName, ".")
	cmd.Dir = pluginDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to compile plugin in %s: %v\nOutput: %s", pluginDir, err, out)
	}
	return binaryName
}

func TestSidecarLargePayload(t *testing.T) {
	// Setup plugin manager
	ctx := context.Background()
	cwd, _ := os.Getwd()
	pluginDir := filepath.Join(cwd, "../../examples/sidecar-plugins")

	pm, err := NewPluginManager(ctx, pluginDir)
	if err != nil {
		t.Fatalf("Failed to create plugin manager: %v", err)
	}
	defer pm.Close()

	// Load the large-payload-test plugin
	pluginPath := filepath.Join(pluginDir, "large-payload-test")
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		t.Skip("large-payload-test plugin not found, skipping")
	}

	// Compile the plugin
	binaryName := compilePlugin(t, pluginPath)
	defer os.Remove(filepath.Join(pluginPath, binaryName))

	if err := pm.LoadPlugin(pluginPath); err != nil {
		t.Fatalf("Failed to load plugin: %v", err)
	}

	// Execute slot
	pluginName := "large-payload-test"
	slotName := "test.large_payload"
	params := map[string]interface{}{}

	// Set a timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := pm.ExecuteSlot(ctx, pluginName, slotName, params)
	if err != nil {
		t.Fatalf("ExecuteSlot failed: %v", err)
	}

	if !resp.Success {
		t.Fatalf("Slot execution failed: %s", resp.Error)
	}

	payload, ok := resp.Data["payload"].(string)
	if !ok {
		t.Fatalf("Expected string payload, got %T", resp.Data["payload"])
	}

	if len(payload) < 100*1024 {
		t.Errorf("Expected payload > 100KB, got %d bytes", len(payload))
	}
}

func TestSidecarSecurity(t *testing.T) {
	// Setup plugin manager
	ctx := context.Background()
	cwd, _ := os.Getwd()
	pluginDir := filepath.Join(cwd, "../../examples/sidecar-plugins")

	pm, err := NewPluginManager(ctx, pluginDir)
	if err != nil {
		t.Fatalf("Failed to create plugin manager: %v", err)
	}
	defer pm.Close()

	// Load the large-payload-test plugin
	pluginPath := filepath.Join(pluginDir, "large-payload-test")
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		t.Skip("large-payload-test plugin not found, skipping")
	}

	// Compile the plugin
	binaryName := compilePlugin(t, pluginPath)
	defer os.Remove(filepath.Join(pluginPath, binaryName))

	// Register a mock callback for file_read
	callbackCalled := false
	pm.SetHostCallback("file_read", func(ctx context.Context, path string) (string, error) {
		callbackCalled = true
		return "secret data", nil
	})

	if err := pm.LoadPlugin(pluginPath); err != nil {
		t.Fatalf("Failed to load plugin: %v", err)
	}

	// Execute slot that tries to read a file
	pluginName := "large-payload-test"
	slotName := "test.forbidden_read"
	params := map[string]interface{}{}

	// Set a timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Execute slot
	resp, err := pm.ExecuteSlot(ctx, pluginName, slotName, params)
	if err != nil {
		t.Fatalf("ExecuteSlot failed: %v", err)
	}

	// The slot itself succeeds (it just sends a host_call), but we need to check if the host call was blocked
	if !resp.Success {
		t.Fatalf("Slot execution failed: %s", resp.Error)
	}

	// Wait a bit for async host call to process (since handleHostCall is goroutine)
	time.Sleep(100 * time.Millisecond)

	if callbackCalled {
		t.Error("Security breach: file_read callback was called despite missing permissions!")
	}
}

func TestOrthancConnectorCompilation(t *testing.T) {
	// Setup plugin manager
	ctx := context.Background()
	cwd, _ := os.Getwd()
	pluginDir := filepath.Join(cwd, "../../examples/sidecar-plugins")

	pm, err := NewPluginManager(ctx, pluginDir)
	if err != nil {
		t.Fatalf("Failed to create plugin manager: %v", err)
	}
	defer pm.Close()

	// Load the orthanc-connector plugin
	pluginPath := filepath.Join(pluginDir, "orthanc-connector")
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		t.Skip("orthanc-connector plugin not found, skipping")
	}

	// Compile the plugin
	binaryName := compilePlugin(t, pluginPath)
	defer os.Remove(filepath.Join(pluginPath, binaryName))

	// Load the plugin
	if err := pm.LoadPlugin(pluginPath); err != nil {
		t.Fatalf("Failed to load plugin: %v", err)
	}

	// Verify metadata
	plugin, err := pm.GetPlugin("orthanc-connector")
	if err != nil {
		t.Fatalf("Failed to get loaded plugin: %v", err)
	}

	if plugin.Manifest.Type != "sidecar" {
		t.Errorf("Expected sidecar type, got %s", plugin.Manifest.Type)
	}

	// Verify slots are registered
	foundSlot := false
	for _, slot := range plugin.Slots {
		if slot.Name == "pacs.search_patient" {
			foundSlot = true
			break
		}
	}
	if !foundSlot {
		t.Error("Slot pacs.search_patient not found in loaded plugin")
	}
}
