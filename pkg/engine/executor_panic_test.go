package engine

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// TestExecutePanicRecovery tests that panics are caught and converted to errors
func TestExecutePanicRecovery(t *testing.T) {
	eng := NewEngine()

	// Register a slot that intentionally panics
	eng.Register("panic.test", func(ctx context.Context, node *Node, scope *Scope) error {
		panic("intentional panic for testing")
	}, SlotMeta{})

	node := &Node{
		Name:     "panic.test",
		Filename: "test.zl",
		Line:     1,
		Col:      1,
	}

	scope := NewScope(nil)

	// Execute should NOT panic, but return an error
	err := eng.Execute(context.Background(), node, scope)

	// Verify error was returned
	if err == nil {
		t.Fatal("Expected error from panic, got nil")
	}

	// Verify error contains panic information
	errMsg := err.Error()
	if !strings.Contains(errMsg, "PANIC") {
		t.Errorf("Error should contain 'PANIC', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "intentional panic for testing") {
		t.Errorf("Error should contain panic message, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "test.zl") {
		t.Errorf("Error should contain filename, got: %s", errMsg)
	}
}

// TestExecuteNilPointerPanic tests recovery from nil pointer dereference
func TestExecuteNilPointerPanic(t *testing.T) {
	eng := NewEngine()

	// Register a slot that causes nil pointer dereference
	eng.Register("nil.test", func(ctx context.Context, node *Node, scope *Scope) error {
		var ptr *string
		_ = *ptr // This will panic
		return nil
	}, SlotMeta{})

	node := &Node{
		Name:     "nil.test",
		Filename: "test.zl",
		Line:     5,
		Col:      10,
	}

	scope := NewScope(nil)

	// Execute should catch the panic
	err := eng.Execute(context.Background(), node, scope)

	if err == nil {
		t.Fatal("Expected error from nil pointer panic, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "PANIC") {
		t.Errorf("Error should contain 'PANIC', got: %s", errMsg)
	}
}

// TestExecuteDivisionByZeroPanic tests recovery from division by zero
func TestExecuteDivisionByZeroPanic(t *testing.T) {
	eng := NewEngine()

	// Register a slot that causes division by zero
	eng.Register("div.test", func(ctx context.Context, node *Node, scope *Scope) error {
		x := 10
		y := 0
		_ = x / y // This will panic
		return nil
	}, SlotMeta{})

	node := &Node{
		Name:     "div.test",
		Filename: "test.zl",
		Line:     8,
		Col:      5,
	}

	scope := NewScope(nil)

	// Execute should catch the panic
	err := eng.Execute(context.Background(), node, scope)

	if err == nil {
		t.Fatal("Expected error from division by zero panic, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "PANIC") {
		t.Errorf("Error should contain 'PANIC', got: %s", errMsg)
	}
}

// TestExecuteNestedPanic tests panic in nested execution
func TestExecuteNestedPanic(t *testing.T) {
	eng := NewEngine()

	// Register outer slot that calls inner slot
	eng.Register("outer.test", func(ctx context.Context, node *Node, scope *Scope) error {
		innerNode := &Node{
			Name:     "inner.test",
			Filename: "test.zl",
			Line:     20,
			Col:      5,
		}
		return eng.Execute(ctx, innerNode, scope)
	}, SlotMeta{})

	// Register inner slot that panics
	eng.Register("inner.test", func(ctx context.Context, node *Node, scope *Scope) error {
		panic("nested panic")
	}, SlotMeta{})

	node := &Node{
		Name:     "outer.test",
		Filename: "test.zl",
		Line:     15,
		Col:      1,
	}

	scope := NewScope(nil)

	// Execute should catch the nested panic
	err := eng.Execute(context.Background(), node, scope)

	if err == nil {
		t.Fatal("Expected error from nested panic, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "PANIC") {
		t.Errorf("Error should contain 'PANIC', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "nested panic") {
		t.Errorf("Error should contain panic message, got: %s", errMsg)
	}
}

// TestExecuteNormalExecution tests that normal execution still works
func TestExecuteNormalExecution(t *testing.T) {
	eng := NewEngine()

	executed := false
	eng.Register("normal.test", func(ctx context.Context, node *Node, scope *Scope) error {
		executed = true
		return nil
	}, SlotMeta{})

	node := &Node{
		Name:     "normal.test",
		Filename: "test.zl",
		Line:     1,
		Col:      1,
	}

	scope := NewScope(nil)

	// Execute should work normally
	err := eng.Execute(context.Background(), node, scope)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !executed {
		t.Error("Handler was not executed")
	}
}

// TestExecuteErrorVsPanic tests that normal errors are not confused with panics
func TestExecuteErrorVsPanic(t *testing.T) {
	eng := NewEngine()

	// Register a slot that returns a normal error
	eng.Register("error.test", func(ctx context.Context, node *Node, scope *Scope) error {
		return fmt.Errorf("normal error")
	}, SlotMeta{})

	node := &Node{
		Name:     "error.test",
		Filename: "test.zl",
		Line:     1,
		Col:      1,
	}

	scope := NewScope(nil)

	// Execute should return the error
	err := eng.Execute(context.Background(), node, scope)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	errMsg := err.Error()
	// Should NOT contain "PANIC"
	if strings.Contains(errMsg, "PANIC") {
		t.Errorf("Normal error should not contain 'PANIC', got: %s", errMsg)
	}
	// Should contain the original error
	if !strings.Contains(errMsg, "normal error") {
		t.Errorf("Error should contain original message, got: %s", errMsg)
	}
}

// TestExecutePanicWithStackTrace tests that stack trace is included
func TestExecutePanicWithStackTrace(t *testing.T) {
	eng := NewEngine()

	eng.Register("stack.test", func(ctx context.Context, node *Node, scope *Scope) error {
		panic("test panic")
	}, SlotMeta{})

	node := &Node{
		Name:     "stack.test",
		Filename: "test.zl",
		Line:     1,
		Col:      1,
	}

	scope := NewScope(nil)

	err := eng.Execute(context.Background(), node, scope)

	if err == nil {
		t.Fatal("Expected error from panic, got nil")
	}

	errMsg := err.Error()
	// Verify stack trace is included
	if !strings.Contains(errMsg, "Stack Trace:") {
		t.Errorf("Error should contain stack trace, got: %s", errMsg)
	}
	// Stack trace should contain goroutine info
	if !strings.Contains(errMsg, "goroutine") {
		t.Errorf("Stack trace should contain goroutine info, got: %s", errMsg)
	}
}
