package engine

import (
	"reflect"
	"unsafe"
)

// Arena provides high-performance linear memory allocation for request-scoped objects.
// It bypasses the Go GC for objects allocated within it by using a pre-allocated chunk.
type Arena struct {
	buf    []byte
	offset int
}

// NewArena creates a new Arena with the specified initial capacity.
func NewArena(capacity int) *Arena {
	return &Arena{
		buf:    make([]byte, capacity),
		offset: 0,
	}
}

// Reset resets the arena offset to zero, effectively "freeing" all allocated objects.
// This is an O(1) operation and does not involve the GC.
func (a *Arena) Reset() {
	a.offset = 0
}

// Alloc allocates n bytes of memory from the arena.
func (a *Arena) Alloc(n int) unsafe.Pointer {
	// Ensure alignment (8-byte alignment for 64-bit values)
	padding := (8 - (a.offset % 8)) % 8
	needed := n + padding

	if a.offset+needed > len(a.buf) {
		// Re-allocate/Grow if needed (simplified for now: just double)
		newBuf := make([]byte, len(a.buf)*2+needed)
		copy(newBuf, a.buf[:a.offset])
		a.buf = newBuf
	}

	a.offset += padding
	ptr := unsafe.Pointer(&a.buf[a.offset])
	a.offset += n
	return ptr
}

// AllocScope allocates a new Scope in the arena.
func (a *Arena) AllocScope(parent *Scope) *Scope {
	ptr := a.Alloc(int(unsafe.Sizeof(Scope{})))
	s := (*Scope)(ptr)

	// ZERO-INITIALIZE memory to prevent mutex corruption from reused arenas!
	*s = Scope{}

	// Initialize the Scope
	s.parent = parent
	// Note: We still use a regular map for variables for compatibility.
	// In Phase 3 (VM), we will move to a more arena-friendly storage.
	s.vars = a.AllocMap(16)

	return s
}

// AllocMap allocates a new map in the arena (wrapped as a standard map for compatibility).
// Warning: standard maps still have GC overhead. This is a bridge implementation.
func (a *Arena) AllocMap(capacity int) map[string]interface{} {
	return make(map[string]interface{}, capacity)
}

// AllocBuffer allocates a byte buffer in the arena.
func (a *Arena) AllocBuffer(capacity int) []byte {
	// We use the arena's underlying buffer slice directly
	needed := capacity
	if a.offset+needed > len(a.buf) {
		// Grow
		newBuf := make([]byte, len(a.buf)*2+needed)
		copy(newBuf, a.buf[:a.offset])
		a.buf = newBuf
	}

	res := a.buf[a.offset : a.offset+needed : a.offset+needed]
	a.offset += needed
	return res
}

// Stats returns information about the arena usage.
func (a *Arena) Stats() (used int, total int) {
	return a.offset, len(a.buf)
}

// CreateValuePointer creates a pointer to a value in the arena.
func (a *Arena) CreateValuePointer(v interface{}) unsafe.Pointer {
	typ := reflect.TypeOf(v)
	size := int(typ.Size())
	ptr := a.Alloc(size)

	// This is a bit tricky in pure Go without reflect.Value.Set
	// but for primitive types it works for proof-of-concept.
	return ptr
}
