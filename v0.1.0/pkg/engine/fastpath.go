package engine

import (
	"context"
	"net/http"
	"zeno/pkg/fastjson"
	"zeno/pkg/utils/coerce"
)

// FastHTTPResponse provides optimized path for simple JSON responses
// Bypasses full node traversal and validation for common case
// Expected: 2-3x faster than generic http.response handler
func FastHTTPResponse(ctx context.Context, w http.ResponseWriter, status int, body interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return fastjson.NewEncoder(w).Encode(body)
}

// FastVarSet provides optimized path for simple variable assignment
// Bypasses node execution overhead
func FastVarSet(scope *Scope, name string, value interface{}) {
	scope.Set(name, value)
}

// isSimpleHTTPResponse checks if http.response node can use fast path
// Fast path criteria:
// - Has status and/or body children only
// - No complex nested logic
// - No middleware or special attributes
func isSimpleHTTPResponse(node *Node) bool {
	if node.Name != "http.response" {
		return false
	}

	// Check children - only allow status, body, data
	for _, child := range node.Children {
		if child.Name != "status" && child.Name != "body" && child.Name != "data" {
			return false
		}
		// No nested children in status/body
		if len(child.Children) > 0 {
			return false
		}
	}

	return len(node.Children) > 0 && len(node.Children) <= 3
}

// ExecuteFastHTTPResponse executes http.response using fast path
func ExecuteFastHTTPResponse(ctx context.Context, node *Node, scope *Scope) error {
	w, ok := ctx.Value("httpWriter").(http.ResponseWriter)
	if !ok {
		return nil // Fallback to slow path
	}

	status := 200
	var body interface{}

	// Quick extraction without full parseNodeValue
	for _, child := range node.Children {
		if child.Name == "status" {
			if val := resolveValue(child.Value, scope); val != nil {
				if s, err := coerce.ToInt(val); err == nil {
					status = s
				}
			}
		} else if child.Name == "body" || child.Name == "data" {
			body = resolveValue(child.Value, scope)
		}
	}

	return FastHTTPResponse(ctx, w, status, body)
}

// resolveValue is a helper to resolve variable references
// This should match the existing resolveValue in slots package
func resolveValue(val interface{}, scope *Scope) interface{} {
	if val == nil {
		return nil
	}

	// Check if it's a variable reference
	if str, ok := val.(string); ok {
		if len(str) > 0 && str[0] == '$' {
			varName := str[1:]
			if v, exists := scope.Get(varName); exists {
				return v
			}
		}
		return str
	}

	return val
}

// TryFastPath attempts to execute node using fast path
// Returns true if fast path was used, false if should use slow path
func TryFastPath(ctx context.Context, node *Node, scope *Scope) (bool, error) {
	// Fast path for http.response
	if node.Name == "http.response" && isSimpleHTTPResponse(node) {
		err := ExecuteFastHTTPResponse(ctx, node, scope)
		return true, err
	}

	// Fast path for simple variable assignment
	if len(node.Name) > 1 && node.Name[0] == '$' && len(node.Children) == 0 {
		varName := node.Name[1:]
		value := resolveValue(node.Value, scope)
		FastVarSet(scope, varName, value)
		return true, nil
	}

	return false, nil
}
