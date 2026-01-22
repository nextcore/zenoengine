package engine

import (
	"context"
	"fmt"
)

// ZenoHost implements vm.HostInterface using ZenoEngine's slot registry and scope.
//
// OWNERSHIP: Does NOT own engine or scope, only borrows them.
// THREAD-SAFETY: Safe if underlying engine.Registry and Scope are not modified concurrently in unsafe ways.
type ZenoHost struct {
	engine *Engine
	scope  *Scope
	ctx    context.Context
}

// NewZenoHost creates a new host adapter for the VM.
//
// PRECONDITION: engine and scope must be non-nil.
func NewZenoHost(ctx context.Context, engine *Engine, scope *Scope) *ZenoHost {
	return &ZenoHost{
		engine: engine,
		scope:  scope,
		ctx:    ctx,
	}
}

// Call implements vm.HostInterface.Call
func (h *ZenoHost) Call(slotName string, args map[string]interface{}) (interface{}, error) {
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

	// 4. Return result (slots typically modify scope, not return values because of legacy architecture)
	return nil, nil
}

// Get implements vm.HostInterface.Get
func (h *ZenoHost) Get(key string) (interface{}, bool) {
	return h.scope.Get(key)
}

// Set implements vm.HostInterface.Set
func (h *ZenoHost) Set(key string, val interface{}) {
	h.scope.Set(key, val)
}

// expandToNode converts a key-value pair to a Node tree.
func (h *ZenoHost) expandToNode(name string, val interface{}, parent *Node) *Node {
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
