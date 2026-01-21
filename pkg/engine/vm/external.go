package vm

import "fmt"

// ExternalCallHandler adalah interface untuk memanggil slot eksternal dari VM.
// Implementasi konkret akan disediakan oleh engine package.
//
// OWNERSHIP: VM does NOT own this interface, only borrows it.
// THREAD-SAFETY: Implementation must be thread-safe if VM is used concurrently.
type ExternalCallHandler interface {
	// Call executes an external slot with given arguments.
	// Returns the result value and any error that occurred.
	//
	// PRECONDITION: slotName must be non-empty
	// POSTCONDITION: If error is non-nil, result is undefined
	Call(slotName string, args map[string]interface{}) (interface{}, error)
}

// ScopeInterface adalah abstraksi untuk variable storage.
// Ini memungkinkan VM untuk get/set variables tanpa dependency ke engine.Scope.
//
// OWNERSHIP: VM does NOT own this interface, only borrows it.
// THREAD-SAFETY: Implementation must handle concurrent access if needed.
type ScopeInterface interface {
	// Get retrieves a variable by key.
	// Returns (value, true) if found, (nil, false) if not found.
	Get(key string) (interface{}, bool)

	// Set stores a variable with given key and value.
	// Overwrites existing value if key already exists.
	Set(key string, val interface{})
}

// NoOpExternalHandler is a stub implementation that returns errors for all calls.
// Useful for testing VM in isolation without engine dependencies.
type NoOpExternalHandler struct{}

func (h *NoOpExternalHandler) Call(slotName string, args map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("external calls not supported (slot: %s)", slotName)
}

// MemoryScope is a simple in-memory implementation of ScopeInterface.
// Useful for testing VM in isolation.
//
// THREAD-SAFETY: NOT thread-safe. Use mutex if concurrent access needed.
type MemoryScope struct {
	vars map[string]interface{}
}

// NewMemoryScope creates a new in-memory scope.
func NewMemoryScope() *MemoryScope {
	return &MemoryScope{
		vars: make(map[string]interface{}),
	}
}

func (s *MemoryScope) Get(key string) (interface{}, bool) {
	val, ok := s.vars[key]
	return val, ok
}

func (s *MemoryScope) Set(key string, val interface{}) {
	s.vars[key] = val
}
