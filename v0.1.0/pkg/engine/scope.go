package engine

import (
	"strings"
	"sync"
)

type Scope struct {
	mu     sync.RWMutex
	vars   map[string]interface{}
	parent *Scope
}

func NewScope(parent *Scope) *Scope {
	return &Scope{
		vars:   make(map[string]interface{}),
		parent: parent,
	}
}

// Set menyimpan variabel dengan aman (Thread-Safe)
func (s *Scope) Set(key string, val interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vars[key] = val
}

// Get mengambil variabel dengan dukungan Dot Notation (user.id, form.judul)
func (s *Scope) Get(key string) (interface{}, bool) {
	s.mu.RLock()

	// 1. Cek Direct Key in current scope
	if val, ok := s.vars[key]; ok {
		s.mu.RUnlock()
		return val, true
	}

	// Store parent reference before unlocking
	parent := s.parent
	s.mu.RUnlock()

	// 2. Check Parent Scope (after releasing lock to avoid deadlock)
	if parent != nil {
		if val, ok := parent.Get(key); ok {
			return val, true
		}
	}

	// 3. Check Nested Key (Deep Navigation Support)
	if strings.Contains(key, ".") {
		s.mu.RLock()
		parts := strings.Split(key, ".")

		// Start with root (check current scope first, then parent)
		currentVal, ok := s.vars[parts[0]]
		parentRef := s.parent
		s.mu.RUnlock()

		if !ok && parentRef != nil {
			currentVal, ok = parentRef.Get(parts[0])
		}
		if !ok || currentVal == nil {
			return nil, false // Root not found or nil
		}

		// Traverse path
		for i := 1; i < len(parts); i++ {
			part := parts[i]

			// Check if currentVal is a Map
			if mapVal, ok := currentVal.(map[string]interface{}); ok {
				if val, exists := mapVal[part]; exists {
					currentVal = val
				} else {
					return nil, false // Property not found
				}
			} else {
				// Current Value is NOT a map, so we can't go deeper.
				// This acts as Safe Navigation: return nil, false (Not Found)
				return nil, false
			}

			// Safe Navigation check: If currentVal becomes nil during traversal, stop
			if currentVal == nil {
				return nil, true // Found explicitly as nil? Or just nil.
			}
		}

		return currentVal, true
	}

	return nil, false
}

// ToMap mengonversi scope menjadi map standar (berguna untuk template rendering)
func (s *Scope) ToMap() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m := make(map[string]interface{})
	for k, v := range s.vars {
		m[k] = v
	}
	return m
}

// Reset clears all variables from the scope.
// This is used by the object pool to safely reuse Scope instances
// without data leakage between requests.
func (s *Scope) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear map efficiently by deleting all keys
	for k := range s.vars {
		delete(s.vars, k)
	}
}

// [BARU] Clone membuat salinan scope baru (Deep Copy Level 1)
// Ini yang dicari oleh router.go
func (s *Scope) Clone() *Scope {
	s.mu.RLock()
	defer s.mu.RUnlock()

	newScope := NewScope(nil)

	// Salin semua variabel dari scope lama ke scope baru
	for k, v := range s.vars {
		newScope.vars[k] = v
	}

	return newScope
}
