package engine

import (
	"context"
	"fmt"
)

// EngineCallHandler implements vm.ExternalCallHandler using ZenoEngine's slot registry.
//
// OWNERSHIP: Does NOT own engine or scope, only borrows them.
// THREAD-SAFETY: Safe if underlying engine.Registry is not modified during execution.
type EngineCallHandler struct {
	engine *Engine
	scope  *Scope
	ctx    context.Context
}

// NewEngineCallHandler creates a new bridge between VM and Engine.
//
// PRECONDITION: engine and scope must be non-nil and remain valid during VM execution.
func NewEngineCallHandler(ctx context.Context, engine *Engine, scope *Scope) *EngineCallHandler {
	return &EngineCallHandler{
		engine: engine,
		scope:  scope,
		ctx:    ctx,
	}
}

// Call implements vm.ExternalCallHandler.Call
func (h *EngineCallHandler) Call(slotName string, args map[string]interface{}) (interface{}, error) {
	// 1. Lookup handler in registry
	handler, exists := h.engine.Registry[slotName]
	if !exists {
		return nil, fmt.Errorf("slot not found: %s", slotName)
	}

	// 2. Convert args map to Node structure (VM -> Engine bridge)
	mockNode := &Node{Name: slotName}

	// [NEW] Check for implicit value passed from Compiler
	if val, ok := args["__value__"]; ok {
		mockNode.Value = val
		delete(args, "__value__") // Remove so it doesn't appear as a child
	}

	if len(args) > 0 {
		mockNode.Children = make([]*Node, 0, len(args))
		for k, v := range args {
			child := h.expandToNode(k, v, mockNode)
			mockNode.Children = append(mockNode.Children, child)
		}
	}

	// 3. Execute handler
	err := handler(h.ctx, mockNode, h.scope)
	if err != nil {
		return nil, fmt.Errorf("slot '%s' execution failed: %w", slotName, err)
	}

	// 4. Return result (slots typically modify scope, not return values)
	return nil, nil
}

// expandToNode converts a key-value pair to a Node tree.
// This is the inverse of vm.expandToNode - it converts VM data back to Engine format.
func (h *EngineCallHandler) expandToNode(name string, val interface{}, parent *Node) *Node {
	node := &Node{Name: name, Parent: parent}

	if m, ok := val.(map[string]interface{}); ok {
		// Map -> Children
		node.Children = make([]*Node, 0, len(m))
		for k, v := range m {
			child := h.expandToNode(k, v, node)
			node.Children = append(node.Children, child)
		}
	} else {
		// Leaf value
		node.Value = val
	}

	return node
}

// ScopeAdapter implements vm.ScopeInterface using engine.Scope.
//
// OWNERSHIP: Does NOT own scope, only borrows it.
// THREAD-SAFETY: Delegates to engine.Scope's thread-safety.
type ScopeAdapter struct {
	scope *Scope
}

// NewScopeAdapter creates a new adapter for engine.Scope.
func NewScopeAdapter(scope *Scope) *ScopeAdapter {
	return &ScopeAdapter{scope: scope}
}

// Get implements vm.ScopeInterface.Get
func (s *ScopeAdapter) Get(key string) (interface{}, bool) {
	return s.scope.Get(key)
}

// Set implements vm.ScopeInterface.Set
func (s *ScopeAdapter) Set(key string, val interface{}) {
	s.scope.Set(key, val)
}
