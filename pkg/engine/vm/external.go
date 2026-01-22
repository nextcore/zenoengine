package vm

import "fmt"

// HostInterface defines the boundary between the VM and the host environment.
// This interface follows the Hexagonal Architecture pattern, allowing the VM
// to run in any environment that provides this interface (e.g. Go Engine, separate CLI, Rust, etc.).
//
// OWNERSHIP: VM does NOT own this interface, only borrows it.
// THREAD-SAFETY: Implementation must be thread-safe if VM is used concurrently.
type HostInterface interface {
	// Call executes an external slot/function.
	Call(slotName string, args map[string]interface{}) (interface{}, error)

	// Get retrieves a variable from the host environment.
	Get(key string) (interface{}, bool)

	// Set stores a variable in the host environment.
	Set(key string, val interface{})
}

// NoOpHost is a stub implementation for testing the VM in isolation.
type NoOpHost struct {
	vars map[string]interface{}
}

func NewNoOpHost() *NoOpHost {
	return &NoOpHost{
		vars: make(map[string]interface{}),
	}
}

func (h *NoOpHost) Call(slotName string, args map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("external calls not supported (slot: %s)", slotName)
}

func (h *NoOpHost) Get(key string) (interface{}, bool) {
	val, ok := h.vars[key]
	return val, ok
}

func (h *NoOpHost) Set(key string, val interface{}) {
	h.vars[key] = val
}
