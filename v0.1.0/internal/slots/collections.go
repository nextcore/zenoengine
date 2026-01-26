package slots

import (
	"context"
	"fmt"
	"strings"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

func RegisterCollectionSlots(eng *engine.Engine) {
	// ==========================================
	// ARRAY MANIPULATION
	// ==========================================

	// ARRAY.PUSH
	eng.Register("array.push", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var targetName string
		var items []interface{}

		// Parse arguments
		for _, c := range node.Children {
			if c.Name == "in" || c.Name == "list" {
				targetName = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
			if c.Name == "val" || c.Name == "value" || c.Name == "item" {
				items = append(items, parseNodeValue(c, scope))
			}
		}

		// Also support main value as list name
		if node.Value != nil {
			targetName = strings.TrimPrefix(coerce.ToString(node.Value), "$")
		}

		if targetName == "" {
			return fmt.Errorf("array.push: target list not specified")
		}

		// Get existing list
		currentVal, ok := scope.Get(targetName)
		if !ok || currentVal == nil {
			// Initialize if not exists
			currentVal = []interface{}{}
		}

		// Convert to Slice
		list, err := coerce.ToSlice(currentVal)
		if err != nil {
			// If it's single item, make it a list
			list = []interface{}{currentVal}
		}

		// Append new items
		list = append(list, items...)

		// Update Scope
		scope.Set(targetName, list)
		return nil
	}, engine.SlotMeta{
		Description: "Menambahkan elemen baru ke akhir array.",
		Example:     "array.push: $my_list\n  val: 'New Item'",
	})

	// ARRAY.POP
	eng.Register("array.pop", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		targetName := strings.TrimPrefix(coerce.ToString(node.Value), "$")
		dstName := "popped_item"

		for _, c := range node.Children {
			if c.Name == "in" || c.Name == "list" {
				targetName = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
			if c.Name == "as" {
				dstName = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		currentVal, ok := scope.Get(targetName)
		if !ok {
			return fmt.Errorf("array.pop: variable '%s' not found", targetName)
		}

		list, err := coerce.ToSlice(currentVal)
		if err != nil || len(list) == 0 {
			scope.Set(dstName, nil)
			return nil
		}

		// Pop last item
		lastItem := list[len(list)-1]
		newList := list[:len(list)-1]

		scope.Set(targetName, newList)
		scope.Set(dstName, lastItem)
		return nil
	}, engine.SlotMeta{Example: "array.pop: $stack\n  as: $item"})

	// ARRAY.JOIN
	eng.Register("array.join", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var list []interface{}
		separator := ","
		target := "joined_string"

		if node.Value != nil {
			list, _ = coerce.ToSlice(resolveValue(node.Value, scope))
		}

		for _, c := range node.Children {
			if c.Name == "list" {
				list, _ = coerce.ToSlice(parseNodeValue(c, scope))
			}
			if c.Name == "sep" || c.Name == "separator" {
				separator = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		var strList []string
		for _, item := range list {
			strList = append(strList, coerce.ToString(item))
		}

		scope.Set(target, strings.Join(strList, separator))
		return nil
	}, engine.SlotMeta{Example: "array.join: $tags\n  sep: ', '\n  as: $tag_str"})

	// ==========================================
	// MAP MANIPULATION
	// ==========================================

	// MAP.SET
	eng.Register("map.set", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		targetName := strings.TrimPrefix(coerce.ToString(node.Value), "$")

		for _, c := range node.Children {
			if c.Name == "map" {
				targetName = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		if targetName == "" {
			return fmt.Errorf("map.set: target map not specified")
		}

		// Get or Init Map
		currentVal, ok := scope.Get(targetName)
		mapVal, okMap := currentVal.(map[string]interface{})
		if !ok || !okMap || mapVal == nil {
			mapVal = make(map[string]interface{})
		}

		// Set key-values
		// Usage: map.set: $my_map { key: "foo"; val: "bar" }
		// OR shorthand attributes: map.set: $my_map { name: "Budi"; age: 20 }

		explicitKey := ""
		explicitVal := interface{}(nil)
		hasExplicit := false

		for _, c := range node.Children {
			if c.Name == "key" {
				explicitKey = coerce.ToString(parseNodeValue(c, scope))
				hasExplicit = true
			}
			if c.Name == "val" || c.Name == "value" {
				explicitVal = parseNodeValue(c, scope)
			}
			if c.Name != "map" && c.Name != "key" && c.Name != "val" && c.Name != "value" {
				// Shorthand: name: "Budi"
				mapVal[c.Name] = parseNodeValue(c, scope)
			}
		}

		if hasExplicit && explicitKey != "" {
			mapVal[explicitKey] = explicitVal
		}

		scope.Set(targetName, mapVal)
		return nil
	}, engine.SlotMeta{Example: "map.set: $user\n  age: 30"})

	// MAP.KEYS
	eng.Register("map.keys", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		val := resolveValue(node.Value, scope)
		target := "keys"

		for _, c := range node.Children {
			if c.Name == "map" {
				val = parseNodeValue(c, scope)
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		m, ok := val.(map[string]interface{})
		var keys []interface{}

		if ok {
			for k := range m {
				keys = append(keys, k)
			}
		}

		scope.Set(target, keys)
		return nil
	}, engine.SlotMeta{Example: "map.keys: $user\n  as: $fields"})
	// LEN
	eng.Register("len", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var val interface{}
		target := "len"

		// Value dari main node? e.g. len: $var
		if node.Value != nil {
			val = resolveValue(node.Value, scope)
		}

		for _, c := range node.Children {
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
			if c.Name == "in" || c.Name == "list" || c.Name == "val" || c.Name == "value" {
				val = parseNodeValue(c, scope)
			}
		}

		var length int64 = 0

		if val == nil {
			length = 0
		} else if str, ok := val.(string); ok {
			length = int64(len(str))
		} else if slice, err := coerce.ToSlice(val); err == nil {
			length = int64(len(slice))
		} else if m, ok := val.(map[string]interface{}); ok {
			length = int64(len(m))
		}

		scope.Set(target, length)
		return nil
	}, engine.SlotMeta{Example: "len: $my_list { as: $count }"})
}
