package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidatePlugin_Sidecar(t *testing.T) {
	// Create a temporary directory for the test plugin
	tmpDir, err := os.MkdirTemp("", "zeno-plugin-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a dummy sidecar binary (just an empty file)
	binaryName := "plugin.exe"
	binaryPath := filepath.Join(tmpDir, binaryName)
	if err := os.WriteFile(binaryPath, []byte("fake binary"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create manifest.yaml for a sidecar plugin
	manifestContent := `
name: test-sidecar
version: 1.0.0
type: sidecar
binary: plugin.exe
permissions:
  network: ["*"]
`
	manifestPath := filepath.Join(tmpDir, "manifest.yaml")
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Run validation
	var output bytes.Buffer
	errors, warnings := validatePlugin(tmpDir, &output)

	// Verify results
	if errors != 0 {
		t.Errorf("Expected 0 errors, got %d", errors)
	}
	// We expect 0 warnings because sidecar shouldn't warn about non-wasm extension
	// (Check for author/description/license warnings - we didn't include those in manifest so we might get some)
	// Actually, let's include optional fields to avoid those warnings
	manifestContentFull := `
name: test-sidecar
version: 1.0.0
type: sidecar
binary: plugin.exe
author: Tester
description: A test plugin
license: MIT
permissions:
  network: ["*"]
`
	if err := os.WriteFile(manifestPath, []byte(manifestContentFull), 0644); err != nil {
		t.Fatal(err)
	}

	output.Reset()
	errors, warnings = validatePlugin(tmpDir, &output)

	if errors != 0 {
		t.Errorf("Expected 0 errors, got %d", errors)
	}
	if warnings != 0 {
		t.Errorf("Expected 0 warnings, got %d", warnings)
	}

	outputStr := output.String()
	if strings.Contains(outputStr, "Binary doesn't have .wasm extension") {
		t.Error("Validation output contains unwanted warning about extension")
	}
}

func TestValidatePlugin_Wasm(t *testing.T) {
	// Create a temporary directory for the test plugin
	tmpDir, err := os.MkdirTemp("", "zeno-plugin-test-wasm")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a dummy wasm binary with wrong extension
	binaryName := "plugin.bin"
	binaryPath := filepath.Join(tmpDir, binaryName)
	if err := os.WriteFile(binaryPath, []byte("fake binary"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create manifest.yaml for a wasm plugin (default type)
	manifestContent := `
name: test-wasm
version: 1.0.0
binary: plugin.bin
author: Tester
description: A test plugin
license: MIT
`
	manifestPath := filepath.Join(tmpDir, "manifest.yaml")
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Run validation
	var output bytes.Buffer
	errors, warnings := validatePlugin(tmpDir, &output)

	if errors != 0 {
		t.Errorf("Expected 0 errors, got %d", errors)
	}
	if warnings != 1 {
		t.Errorf("Expected 1 warning for wrong extension, got %d", warnings)
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, "Binary doesn't have .wasm extension") {
		t.Error("Expected warning about extension missing")
	}
}
