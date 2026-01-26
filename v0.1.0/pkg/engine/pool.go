package engine

import "sync"

// Global pool for Scope objects
var scopePool = sync.Pool{
	New: func() interface{} {
		return &Scope{
			vars: make(map[string]interface{}, 16), // Pre-allocate capacity for common use cases
		}
	},
}

// GetScope retrieves a Scope from the pool.
// The returned Scope is ready to use but may contain data from previous use.
// Always call PutScope when done to return it to the pool.
func GetScope() *Scope {
	s := scopePool.Get().(*Scope)
	return s
}

// PutScope returns a Scope to the pool after clearing its data.
// This should be called with defer immediately after GetScope().
// Example:
//   scope := engine.GetScope()
//   defer engine.PutScope(scope)
func PutScope(s *Scope) {
	if s == nil {
		return
	}
	s.Reset()
	scopePool.Put(s)
}

// Pool for form/query data maps
var mapPool = sync.Pool{
	New: func() interface{} {
		return make(map[string]interface{}, 8)
	},
}

// GetMap retrieves a map from the pool.
// Always call PutMap when done.
func GetMap() map[string]interface{} {
	return mapPool.Get().(map[string]interface{})
}

// PutMap returns a map to the pool after clearing its data.
func PutMap(m map[string]interface{}) {
	if m == nil {
		return
	}
	// Clear the map
	for k := range m {
		delete(m, k)
	}
	mapPool.Put(m)
}

// Pool for byte buffers (response buffering, string building)
var bufferPool = sync.Pool{
	New: func() interface{} {
		// Pre-allocate 4KB buffer (common page size)
		return make([]byte, 0, 4096)
	},
}

// GetBuffer retrieves a byte buffer from the pool.
// Always call PutBuffer when done.
func GetBuffer() []byte {
	return bufferPool.Get().([]byte)
}

// PutBuffer returns a byte buffer to the pool.
// The buffer's length is reset to 0 but capacity is preserved.
func PutBuffer(b []byte) {
	if b == nil {
		return
	}
	// Reset length to 0, keep capacity
	bufferPool.Put(b[:0])
}
