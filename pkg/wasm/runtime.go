package wasm

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// Runtime manages WASM module execution
type Runtime struct {
	runtime wazero.Runtime
	ctx     context.Context
	modules map[string]api.Module // Loaded WASM modules
}

// NewRuntime creates a new WASM runtime
func NewRuntime(ctx context.Context) (*Runtime, error) {
	// Create Wazero runtime with compilation cache
	r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().
		WithCompilationCache(wazero.NewCompilationCache()))

	// Instantiate WASI for filesystem/env access
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		return nil, fmt.Errorf("failed to instantiate WASI: %w", err)
	}

	return &Runtime{
		runtime: r,
		ctx:     ctx,
		modules: make(map[string]api.Module),
	}, nil
}

// LoadModule loads a WASM module from file
func (r *Runtime) LoadModule(name string, wasmPath string) error {
	// Read WASM file
	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		return fmt.Errorf("failed to read WASM file %s: %w", wasmPath, err)
	}

	// Compile module
	compiled, err := r.runtime.CompileModule(r.ctx, wasmBytes)
	if err != nil {
		return fmt.Errorf("failed to compile WASM module %s: %w", name, err)
	}

	// Instantiate module
	module, err := r.runtime.InstantiateModule(r.ctx, compiled, wazero.NewModuleConfig().WithName(name))
	if err != nil {
		return fmt.Errorf("failed to instantiate WASM module %s: %w", name, err)
	}

	r.modules[name] = module
	slog.Info("✅ WASM module loaded", "name", name, "path", wasmPath)

	return nil
}

// CallFunction calls an exported function from a loaded module
// Returns result and any error. Includes panic recovery.
func (r *Runtime) CallFunction(moduleName string, functionName string, params ...uint64) (results []uint64, err error) {
	// Panic recovery for WASM execution
	defer func() {
		if rec := recover(); rec != nil {
			stack := string(debug.Stack())
			slog.Error("🔥 PANIC in WASM execution",
				"module", moduleName,
				"function", functionName,
				"panic", rec,
				"stack", stack,
			)
			err = fmt.Errorf("WASM panic in %s.%s: %v", moduleName, functionName, rec)
		}
	}()

	// Get module
	module, exists := r.modules[moduleName]
	if !exists {
		return nil, fmt.Errorf("module %s not loaded", moduleName)
	}

	// Get function
	fn := module.ExportedFunction(functionName)
	if fn == nil {
		return nil, fmt.Errorf("function %s not found in module %s", functionName, moduleName)
	}

	// Call function
	results, err = fn.Call(r.ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to call %s.%s: %w", moduleName, functionName, err)
	}

	return results, nil
}

// GetMemory returns the memory of a module for reading/writing
func (r *Runtime) GetMemory(moduleName string) (api.Memory, error) {
	module, exists := r.modules[moduleName]
	if !exists {
		return nil, fmt.Errorf("module %s not loaded", moduleName)
	}

	memory := module.Memory()
	if memory == nil {
		return nil, fmt.Errorf("module %s has no exported memory", moduleName)
	}

	return memory, nil
}

// ReadString reads a string from WASM memory
func (r *Runtime) ReadString(moduleName string, ptr, length uint32) (string, error) {
	memory, err := r.GetMemory(moduleName)
	if err != nil {
		return "", err
	}

	bytes, ok := memory.Read(ptr, length)
	if !ok {
		return "", fmt.Errorf("failed to read memory at %d (length %d)", ptr, length)
	}

	return string(bytes), nil
}

// WriteString writes a string to WASM memory
// Returns pointer to the written string
func (r *Runtime) WriteString(moduleName string, data string) (ptr uint32, length uint32, err error) {
	memory, err := r.GetMemory(moduleName)
	if err != nil {
		return 0, 0, err
	}

	bytes := []byte(data)
	length = uint32(len(bytes))

	// Allocate memory in WASM module (assumes module exports "alloc" function)
	allocFn := r.modules[moduleName].ExportedFunction("alloc")
	if allocFn == nil {
		return 0, 0, fmt.Errorf("module %s does not export 'alloc' function", moduleName)
	}

	results, err := allocFn.Call(r.ctx, uint64(length))
	if err != nil {
		return 0, 0, fmt.Errorf("failed to allocate memory: %w", err)
	}

	ptr = uint32(results[0])

	// Write data to allocated memory
	if !memory.Write(ptr, bytes) {
		return 0, 0, fmt.Errorf("failed to write to memory at %d", ptr)
	}

	return ptr, length, nil
}

// UnloadModule unloads a WASM module and frees resources
func (r *Runtime) UnloadModule(name string) error {
	module, exists := r.modules[name]
	if !exists {
		return fmt.Errorf("module %s not loaded", name)
	}

	if err := module.Close(r.ctx); err != nil {
		return fmt.Errorf("failed to close module %s: %w", name, err)
	}

	delete(r.modules, name)
	slog.Info("WASM module unloaded", "name", name)

	return nil
}

// Close closes the runtime and all loaded modules
func (r *Runtime) Close() error {
	// Close all modules
	for name := range r.modules {
		if err := r.UnloadModule(name); err != nil {
			slog.Error("Failed to unload module", "name", name, "error", err)
		}
	}

	// Close runtime
	if err := r.runtime.Close(r.ctx); err != nil {
		return fmt.Errorf("failed to close WASM runtime: %w", err)
	}

	slog.Info("WASM runtime closed")
	return nil
}

// ListModules returns names of all loaded modules
func (r *Runtime) ListModules() []string {
	names := make([]string, 0, len(r.modules))
	for name := range r.modules {
		names = append(names, name)
	}
	return names
}
