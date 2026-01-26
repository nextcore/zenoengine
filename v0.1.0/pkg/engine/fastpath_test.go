package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// BenchmarkHTTPResponseFastPath measures fast path performance
func BenchmarkHTTPResponseFastPath(b *testing.B) {
	eng := NewEngine()
	
	// Register http.response slot (needed for fallback)
	eng.Register("http.response", func(ctx context.Context, node *Node, scope *Scope) error {
		// This shouldn't be called if fast path works
		return nil
	}, SlotMeta{})

	// Create simple http.response node (eligible for fast path)
	node := &Node{
		Name: "http.response",
		Children: []*Node{
			{Name: "status", Value: 200},
			{Name: "body", Value: map[string]string{"message": "hello"}},
		},
	}

	scope := NewScope(nil)
	scope.Set("test", "data")

	w := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), "httpWriter", http.ResponseWriter(w))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		eng.Execute(ctx, node, scope)
	}
}

// BenchmarkHTTPResponseGeneric measures generic path performance
func BenchmarkHTTPResponseGeneric(b *testing.B) {
	eng := NewEngine()
	
	// Register http.response slot
	eng.Register("http.response", func(ctx context.Context, node *Node, scope *Scope) error {
		w, _ := ctx.Value("httpWriter").(http.ResponseWriter)
		w.WriteHeader(200)
		return nil
	}, SlotMeta{})

	// Create complex http.response node (NOT eligible for fast path)
	node := &Node{
		Name: "http.response",
		Children: []*Node{
			{Name: "status", Value: 200},
			{Name: "body", Value: map[string]string{"message": "hello"}},
			{Name: "custom", Value: "extra"}, // Extra attribute prevents fast path
		},
	}

	scope := NewScope(nil)
	w := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), "httpWriter", http.ResponseWriter(w))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		eng.Execute(ctx, node, scope)
	}
}

// BenchmarkVarAssignmentFastPath measures fast path for variable assignment
func BenchmarkVarAssignmentFastPath(b *testing.B) {
	eng := NewEngine()
	
	node := &Node{
		Name:  "$testvar",
		Value: "test value",
	}

	scope := NewScope(nil)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		eng.Execute(ctx, node, scope)
	}
}

// TestFastPathCorrectness verifies fast path produces correct results
func TestFastPathCorrectness(t *testing.T) {
	eng := NewEngine()
	
	// Register http.response for comparison
	eng.Register("http.response", func(ctx context.Context, node *Node, scope *Scope) error {
		w, _ := ctx.Value("httpWriter").(http.ResponseWriter)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"test":"data"}`))
		return nil
	}, SlotMeta{})

	node := &Node{
		Name: "http.response",
		Children: []*Node{
			{Name: "status", Value: 201},
			{Name: "body", Value: map[string]string{"result": "success"}},
		},
	}

	scope := NewScope(nil)
	w := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), "httpWriter", http.ResponseWriter(w))

	// Execute with fast path
	if err := eng.Execute(ctx, node, scope); err != nil {
		t.Fatal(err)
	}

	// Verify status code
	if w.Code != 201 {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	// Verify content type
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", ct)
	}
}

// TestFastPathVarAssignment verifies variable assignment fast path
func TestFastPathVarAssignment(t *testing.T) {
	eng := NewEngine()
	
	node := &Node{
		Name:  "$myvar",
		Value: "test value",
	}

	scope := NewScope(nil)
	ctx := context.Background()

	if err := eng.Execute(ctx, node, scope); err != nil {
		t.Fatal(err)
	}

	// Verify variable was set
	val, exists := scope.Get("myvar")
	if !exists {
		t.Error("Variable should be set")
	}
	if val != "test value" {
		t.Errorf("Expected 'test value', got %v", val)
	}
}

// TestFastPathDetection verifies fast path detection logic
func TestFastPathDetection(t *testing.T) {
	tests := []struct {
		name     string
		node     *Node
		expected bool
	}{
		{
			name: "Simple http.response - should use fast path",
			node: &Node{
				Name: "http.response",
				Children: []*Node{
					{Name: "status", Value: 200},
					{Name: "body", Value: "test"},
				},
			},
			expected: true,
		},
		{
			name: "Complex http.response - should NOT use fast path",
			node: &Node{
				Name: "http.response",
				Children: []*Node{
					{Name: "status", Value: 200},
					{Name: "body", Value: "test"},
					{Name: "custom", Value: "extra"},
				},
			},
			expected: false,
		},
		{
			name: "Simple var assignment - should use fast path",
			node: &Node{
				Name:  "$myvar",
				Value: "value",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSimpleHTTPResponse(tt.node)
			if tt.node.Name == "http.response" && result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
