package wasm

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestNewRuntime tests runtime creation
func TestNewRuntime(t *testing.T) {
	ctx := context.Background()
	runtime, err := NewRuntime(ctx)
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close()

	if runtime.runtime == nil {
		t.Error("Runtime is nil")
	}

	if runtime.modules == nil {
		t.Error("Modules map is nil")
	}
}

// TestLoadModule tests loading a WASM module
func TestLoadModule(t *testing.T) {
	// Skip if no test WASM file available
	testWASM := filepath.Join("testdata", "test.wasm")
	if _, err := os.Stat(testWASM); os.IsNotExist(err) {
		t.Skip("Test WASM file not found, skipping")
	}

	ctx := context.Background()
	runtime, err := NewRuntime(ctx)
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close()

	// Load module
	err = runtime.LoadModule("test", testWASM)
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Verify module is loaded
	modules := runtime.ListModules()
	if len(modules) != 1 {
		t.Errorf("Expected 1 module, got %d", len(modules))
	}

	if modules[0] != "test" {
		t.Errorf("Expected module name 'test', got '%s'", modules[0])
	}
}

// TestUnloadModule tests unloading a module
func TestUnloadModule(t *testing.T) {
	testWASM := filepath.Join("testdata", "test.wasm")
	if _, err := os.Stat(testWASM); os.IsNotExist(err) {
		t.Skip("Test WASM file not found, skipping")
	}

	ctx := context.Background()
	runtime, err := NewRuntime(ctx)
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close()

	// Load module
	runtime.LoadModule("test", testWASM)

	// Unload module
	err = runtime.UnloadModule("test")
	if err != nil {
		t.Fatalf("Failed to unload module: %v", err)
	}

	// Verify module is unloaded
	modules := runtime.ListModules()
	if len(modules) != 0 {
		t.Errorf("Expected 0 modules after unload, got %d", len(modules))
	}
}

// TestCallFunction tests calling an exported function
func TestCallFunction(t *testing.T) {
	testWASM := filepath.Join("testdata", "test.wasm")
	if _, err := os.Stat(testWASM); os.IsNotExist(err) {
		t.Skip("Test WASM file not found, skipping")
	}

	ctx := context.Background()
	runtime, err := NewRuntime(ctx)
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close()

	// Load module
	runtime.LoadModule("test", testWASM)

	// Call function (assumes test.wasm exports "add" function)
	results, err := runtime.CallFunction("test", "add", 5, 3)
	if err != nil {
		t.Fatalf("Failed to call function: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if results[0] != 8 {
		t.Errorf("Expected result 8, got %d", results[0])
	}
}

// TestCallFunctionPanicRecovery tests panic recovery
func TestCallFunctionPanicRecovery(t *testing.T) {
	ctx := context.Background()
	runtime, err := NewRuntime(ctx)
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close()

	// Try to call function on non-existent module (should not panic)
	_, err = runtime.CallFunction("nonexistent", "test")
	if err == nil {
		t.Error("Expected error for non-existent module, got nil")
	}
}

// TestMemoryOperations tests reading/writing WASM memory
func TestMemoryOperations(t *testing.T) {
	testWASM := filepath.Join("testdata", "test.wasm")
	if _, err := os.Stat(testWASM); os.IsNotExist(err) {
		t.Skip("Test WASM file not found, skipping")
	}

	ctx := context.Background()
	runtime, err := NewRuntime(ctx)
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close()

	// Load module
	runtime.LoadModule("test", testWASM)

	// Test WriteString (if module exports alloc)
	testStr := "Hello WASM"
	ptr, length, err := runtime.WriteString("test", testStr)
	if err != nil {
		// If alloc not available, skip this test
		t.Skipf("Module does not support WriteString: %v", err)
	}

	// Test ReadString
	readStr, err := runtime.ReadString("test", ptr, length)
	if err != nil {
		t.Fatalf("Failed to read string: %v", err)
	}

	if readStr != testStr {
		t.Errorf("Expected '%s', got '%s'", testStr, readStr)
	}
}

// BenchmarkModuleLoad benchmarks module loading
func BenchmarkModuleLoad(b *testing.B) {
	testWASM := filepath.Join("testdata", "test.wasm")
	if _, err := os.Stat(testWASM); os.IsNotExist(err) {
		b.Skip("Test WASM file not found, skipping")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runtime, _ := NewRuntime(ctx)
		runtime.LoadModule("test", testWASM)
		runtime.Close()
	}
}

// BenchmarkFunctionCall benchmarks function calls
func BenchmarkFunctionCall(b *testing.B) {
	testWASM := filepath.Join("testdata", "test.wasm")
	if _, err := os.Stat(testWASM); os.IsNotExist(err) {
		b.Skip("Test WASM file not found, skipping")
	}

	ctx := context.Background()
	runtime, _ := NewRuntime(ctx)
	defer runtime.Close()

	runtime.LoadModule("test", testWASM)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runtime.CallFunction("test", "add", 5, 3)
	}
}
