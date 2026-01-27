package engine

import (
	"context"
	"testing"
)

// BenchmarkSlotExecutionWithCache measures performance with inline caching
func BenchmarkSlotExecutionWithCache(b *testing.B) {
	eng := NewEngine()
	eng.Register("test.slot", func(ctx context.Context, node *Node, scope *Scope) error {
		scope.Set("result", "cached")
		return nil
	}, SlotMeta{Description: "Test slot"})

	node := &Node{Name: "test.slot"}
	scope := NewScope(nil)
	ctx := context.Background()

	// Pre-warm cache (first execution caches the handler)
	eng.Execute(ctx, node, scope)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		eng.Execute(ctx, node, scope)
	}
}

// BenchmarkSlotExecutionWithoutCache measures performance without caching
func BenchmarkSlotExecutionWithoutCache(b *testing.B) {
	eng := NewEngine()
	eng.Register("test.slot", func(ctx context.Context, node *Node, scope *Scope) error {
		scope.Set("result", "uncached")
		return nil
	}, SlotMeta{Description: "Test slot"})

	scope := NewScope(nil)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Create fresh node each time (no cache benefit)
		node := &Node{Name: "test.slot"}
		eng.Execute(ctx, node, scope)
	}
}

// BenchmarkComplexSlotWithCache tests caching with validation
func BenchmarkComplexSlotWithCache(b *testing.B) {
	eng := NewEngine()
	eng.Register("complex.slot", func(ctx context.Context, node *Node, scope *Scope) error {
		// Simulate some work
		for _, child := range node.Children {
			scope.Set(child.Name, child.Value)
		}
		return nil
	}, SlotMeta{
		Description: "Complex slot with validation",
		Inputs: map[string]InputMeta{
			"param1": {Required: true, Type: "string"},
			"param2": {Required: false, Type: "int"},
		},
	})

	node := &Node{
		Name: "complex.slot",
		Children: []*Node{
			{Name: "param1", Value: "test"},
			{Name: "param2", Value: 123},
		},
	}
	scope := NewScope(nil)
	ctx := context.Background()

	// Pre-warm cache
	eng.Execute(ctx, node, scope)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		eng.Execute(ctx, node, scope)
	}
}

// TestInlineCachingCorrectness verifies caching doesn't break functionality
func TestInlineCachingCorrectness(t *testing.T) {
	eng := NewEngine()
	
	callCount := 0
	eng.Register("counter.slot", func(ctx context.Context, node *Node, scope *Scope) error {
		callCount++
		scope.Set("count", callCount)
		return nil
	}, SlotMeta{})

	node := &Node{Name: "counter.slot"}
	scope := NewScope(nil)
	ctx := context.Background()

	// First execution - should cache handler
	if err := eng.Execute(ctx, node, scope); err != nil {
		t.Fatal(err)
	}
	if callCount != 1 {
		t.Errorf("Expected callCount=1, got %d", callCount)
	}

	// Second execution - should use cached handler
	if err := eng.Execute(ctx, node, scope); err != nil {
		t.Fatal(err)
	}
	if callCount != 2 {
		t.Errorf("Expected callCount=2, got %d", callCount)
	}

	// Verify handler is cached
	if node.cachedHandler == nil {
		t.Error("Handler should be cached after execution")
	}
}

// TestCacheIsolation verifies different nodes have independent caches
func TestCacheIsolation(t *testing.T) {
	eng := NewEngine()
	
	eng.Register("slot.a", func(ctx context.Context, node *Node, scope *Scope) error {
		scope.Set("type", "A")
		return nil
	}, SlotMeta{})
	
	eng.Register("slot.b", func(ctx context.Context, node *Node, scope *Scope) error {
		scope.Set("type", "B")
		return nil
	}, SlotMeta{})

	nodeA := &Node{Name: "slot.a"}
	nodeB := &Node{Name: "slot.b"}
	scope := NewScope(nil)
	ctx := context.Background()

	// Execute both
	eng.Execute(ctx, nodeA, scope)
	eng.Execute(ctx, nodeB, scope)

	// Verify separate caches
	if nodeA.cachedHandler == nil {
		t.Error("Node A should have cached handler")
	}
	if nodeB.cachedHandler == nil {
		t.Error("Node B should have cached handler")
	}

	// Verify they're different handlers
	val, _ := scope.Get("type")
	if val != "B" {
		t.Errorf("Expected type=B, got %v", val)
	}
}
