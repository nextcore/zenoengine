package adapter

import (
	"context"
	"fmt"
	"strings"
	"syscall/js"

	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

func RegisterJSSlots(eng *engine.Engine) {
	// JS.CALL: Call a global JS function
	eng.Register("js.call", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		funcName := coerce.ToString(ResolveValue(node.Value, scope))
		var args []interface{}

		for _, c := range node.Children {
			if c.Name == "func" || c.Name == "function" {
				funcName = coerce.ToString(ResolveValue(c.Value, scope))
			} else if c.Name == "args" {
				val := ResolveValue(c.Value, scope)
				if list, err := coerce.ToSlice(val); err == nil {
					args = list
				}
			} else if c.Name == "arg" {
				// Single arg or list of args as children
				args = append(args, ResolveValue(c.Value, scope))
			}
		}

		if funcName == "" {
			return fmt.Errorf("js.call: function name required")
		}

		// Resolve function path (e.g. "console.log" -> global.Get("console").Get("log"))
		fn := resolveJSPath(funcName)
		if fn.Type() != js.TypeFunction {
			return fmt.Errorf("js.call: '%s' is not a function", funcName)
		}

		// Convert args to []any
		jsArgs := make([]any, len(args))
		for i, v := range args {
			jsArgs[i] = js.ValueOf(v)
		}

		fn.Invoke(jsArgs...)
		return nil
	}, engine.SlotMeta{Description: "Call a JavaScript function"})

	// JS.LOG: Console log wrapper
	eng.Register("js.log", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		val := ResolveValue(node.Value, scope)
		js.Global().Get("console").Call("log", fmt.Sprintf("[Zeno] %v", val))
		return nil
	}, engine.SlotMeta{Description: "Log to browser console"})

	// JS.SET: Set global property
	eng.Register("js.set", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		key := coerce.ToString(ResolveValue(node.Value, scope))
		var val interface{}

		for _, c := range node.Children {
			if c.Name == "key" {
				key = coerce.ToString(ResolveValue(c.Value, scope))
			} else if c.Name == "value" || c.Name == "val" {
				val = ResolveValue(c.Value, scope)
			}
		}

		if key == "" {
			return fmt.Errorf("js.set: key required")
		}

		// Handle nested keys: "document.title"
		setJSPath(key, val)
		return nil
	}, engine.SlotMeta{Description: "Set a JavaScript property"})

	// JS.GET: Get global property
	eng.Register("js.get", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		key := coerce.ToString(ResolveValue(node.Value, scope))
		target := "js_result"

		for _, c := range node.Children {
			if c.Name == "key" {
				key = coerce.ToString(ResolveValue(c.Value, scope))
			} else if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		val := resolveJSPath(key)

		// Convert JS Value to Go
		var goVal interface{}
		switch val.Type() {
		case js.TypeString:
			goVal = val.String()
		case js.TypeNumber:
			goVal = val.Float()
		case js.TypeBoolean:
			goVal = val.Bool()
		case js.TypeNull, js.TypeUndefined:
			goVal = nil
		default:
			// Fallback string
			goVal = val.String()
		}

		scope.Set(target, goVal)
		return nil
	}, engine.SlotMeta{Description: "Get a JavaScript property"})
}

// Helpers

func resolveJSPath(path string) js.Value {
	parts := strings.Split(path, ".")
	curr := js.Global()
	for _, part := range parts {
		curr = curr.Get(part)
		if curr.IsUndefined() || curr.IsNull() {
			return curr
		}
	}
	return curr
}

func setJSPath(path string, val interface{}) {
	parts := strings.Split(path, ".")
	if len(parts) == 1 {
		js.Global().Set(parts[0], js.ValueOf(val))
		return
	}

	// Navigate to parent
	parent := js.Global()
	for i := 0; i < len(parts)-1; i++ {
		parent = parent.Get(parts[i])
		if parent.IsUndefined() || parent.IsNull() {
			return // Cannot set property of undefined
		}
	}

	parent.Set(parts[len(parts)-1], js.ValueOf(val))
}
