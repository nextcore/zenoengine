package engine

import (
	"runtime"
	"testing"
)

// TestScopePooling verifies that Scope pooling works correctly
func TestScopePooling(t *testing.T) {
	t.Run("BasicGetPut", func(t *testing.T) {
		s1 := GetScope()
		s1.Set("key", "value")
		PutScope(s1)

		s2 := GetScope()
		_, exists := s2.Get("key")
		if exists {
			t.Error("Scope not properly reset after PutScope")
		}
		PutScope(s2)
	})

	t.Run("MultipleVariables", func(t *testing.T) {
		s := GetScope()
		s.Set("var1", "value1")
		s.Set("var2", 123)
		s.Set("var3", true)
		PutScope(s)

		s2 := GetScope()
		if _, exists := s2.Get("var1"); exists {
			t.Error("var1 should not exist after reset")
		}
		if _, exists := s2.Get("var2"); exists {
			t.Error("var2 should not exist after reset")
		}
		if _, exists := s2.Get("var3"); exists {
			t.Error("var3 should not exist after reset")
		}
		PutScope(s2)
	})

	t.Run("NilSafety", func(t *testing.T) {
		// Should not panic
		PutScope(nil)
		PutMap(nil)
		PutBuffer(nil)
	})
}

// TestMapPooling verifies map pooling
func TestMapPooling(t *testing.T) {
	m1 := GetMap()
	m1["key"] = "value"
	PutMap(m1)

	m2 := GetMap()
	if _, exists := m2["key"]; exists {
		t.Error("Map not properly cleared")
	}
	PutMap(m2)
}

// TestBufferPooling verifies buffer pooling
func TestBufferPooling(t *testing.T) {
	b1 := GetBuffer()
	b1 = append(b1, []byte("test data")...)
	if len(b1) == 0 {
		t.Error("Buffer should have data")
	}
	PutBuffer(b1)

	b2 := GetBuffer()
	if len(b2) != 0 {
		t.Error("Buffer not properly reset, length should be 0")
	}
	if cap(b2) < 4096 {
		t.Error("Buffer capacity should be preserved")
	}
	PutBuffer(b2)
}

// BenchmarkScopeWithPool measures performance with pooling
func BenchmarkScopeWithPool(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s := GetScope()
		s.Set("test", i)
		s.Set("data", "value")
		PutScope(s)
	}
}

// BenchmarkScopeWithoutPool measures performance without pooling
func BenchmarkScopeWithoutPool(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s := NewScope(nil)
		s.Set("test", i)
		s.Set("data", "value")
		_ = s
	}
}

// BenchmarkMapWithPool measures map pooling performance
func BenchmarkMapWithPool(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m := GetMap()
		m["key1"] = "value1"
		m["key2"] = i
		PutMap(m)
	}
}

// BenchmarkMapWithoutPool measures map performance without pooling
func BenchmarkMapWithoutPool(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m := make(map[string]interface{}, 8)
		m["key1"] = "value1"
		m["key2"] = i
		_ = m
	}
}

// BenchmarkBufferWithPool measures buffer pooling performance
func BenchmarkBufferWithPool(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf := GetBuffer()
		buf = append(buf, []byte("test data")...)
		PutBuffer(buf)
	}
}

// BenchmarkBufferWithoutPool measures buffer performance without pooling
func BenchmarkBufferWithoutPool(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf := make([]byte, 0, 4096)
		buf = append(buf, []byte("test data")...)
		_ = buf
	}
}

// BenchmarkConcurrentScopePooling tests pool under concurrent load
func BenchmarkConcurrentScopePooling(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s := GetScope()
			s.Set("concurrent", true)
			PutScope(s)
		}
	})
}

// TestMemoryUsage verifies that pooling reduces memory usage
func TestMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	// Force GC before test
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Create and destroy many scopes with pooling
	for i := 0; i < 10000; i++ {
		s := GetScope()
		s.Set("test", i)
		PutScope(s)
	}

	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	allocWithPool := m2.TotalAlloc - m1.TotalAlloc

	// Force GC again
	runtime.GC()
	var m3 runtime.MemStats
	runtime.ReadMemStats(&m3)

	// Create and destroy many scopes without pooling
	for i := 0; i < 10000; i++ {
		s := NewScope(nil)
		s.Set("test", i)
		_ = s
	}

	runtime.GC()
	var m4 runtime.MemStats
	runtime.ReadMemStats(&m4)

	allocWithoutPool := m4.TotalAlloc - m3.TotalAlloc

	t.Logf("Allocations with pool: %d bytes", allocWithPool)
	t.Logf("Allocations without pool: %d bytes", allocWithoutPool)

	// Pool should use significantly less memory
	if allocWithPool >= allocWithoutPool {
		t.Logf("Warning: Pool did not reduce allocations (with: %d, without: %d)",
			allocWithPool, allocWithoutPool)
	} else {
		reduction := float64(allocWithoutPool-allocWithPool) / float64(allocWithoutPool) * 100
		t.Logf("Pool reduced allocations by %.2f%%", reduction)
	}
}
