package adapter

import (
	"context"
	"fmt"
	"strings"
	"syscall/js"

	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

func RegisterStorageSlots(eng *engine.Engine) {
	// STORAGE.SET: Save to LocalStorage
	eng.Register("storage.set", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
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
			return fmt.Errorf("storage.set: key required")
		}

		// Convert value to string (JSON if object?)
		// LocalStorage only supports strings.
		strVal := coerce.ToString(val)
		js.Global().Get("localStorage").Call("setItem", key, strVal)
		return nil
	}, engine.SlotMeta{Description: "Save data to LocalStorage"})

	// STORAGE.GET: Retrieve from LocalStorage
	eng.Register("storage.get", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		key := coerce.ToString(ResolveValue(node.Value, scope))
		target := "storage_result"
		defVal := ""

		for _, c := range node.Children {
			if c.Name == "key" {
				key = coerce.ToString(ResolveValue(c.Value, scope))
			} else if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			} else if c.Name == "default" {
				defVal = coerce.ToString(ResolveValue(c.Value, scope))
			}
		}

		item := js.Global().Get("localStorage").Call("getItem", key)
		var res string
		if item.IsNull() {
			res = defVal
		} else {
			res = item.String()
		}

		scope.Set(target, res)
		return nil
	}, engine.SlotMeta{Description: "Get data from LocalStorage"})

	// STORAGE.REMOVE
	eng.Register("storage.remove", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		key := coerce.ToString(ResolveValue(node.Value, scope))
		js.Global().Get("localStorage").Call("removeItem", key)
		return nil
	}, engine.SlotMeta{Description: "Remove item from LocalStorage"})

	// STORAGE.CLEAR
	eng.Register("storage.clear", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		js.Global().Get("localStorage").Call("clear")
		return nil
	}, engine.SlotMeta{Description: "Clear all LocalStorage"})
}
