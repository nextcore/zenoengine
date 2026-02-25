package slots

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

func RegisterMetaSlots(eng *engine.Engine) {
	// 1. meta.eval
	eng.Register("meta.eval", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		// Resolve the value (handle variables like $code)
		val := resolveValue(node.Value, scope)
		code, ok := val.(string)
		if !ok {
			return fmt.Errorf("meta.eval requires a string value, got %T: %v", val, val)
		}

		// Parse the code dynamically
		root, err := engine.ParseString(code, "eval")
		if err != nil {
			return fmt.Errorf("meta.eval parse error: %v", err)
		}

		// Execute the parsed AST
		if err := eng.Execute(ctx, root, scope); err != nil {
			return err
		}

		return nil
	}, engine.SlotMeta{
		Description: "Evaluates a string as ZenoLang code dynamically.",
		Example:     `meta.eval: "http.get: '/api'"`,
		Inputs: map[string]engine.InputMeta{
			"(value)": {Description: "The ZenoLang code string to evaluate", Type: "string"},
		},
	})

	// 2. meta.scope
	eng.Register("meta.scope", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		// Introspection: Return all variables in current scope as a map
		vars := scope.ToMap()

		// Check for "as" attribute to store result
		target := ""
		for _, child := range node.Children {
			if child.Name == "as" {
				if val, ok := child.Value.(string); ok {
					target = strings.TrimPrefix(val, "$")
				}
			}
		}

		if target != "" {
			scope.Set(target, vars)
		}

		return nil
	}, engine.SlotMeta{
		Description: "Returns all variables in the current scope as a map (Introspection).",
		Example:     `$vars: meta.scope`,
		Inputs: map[string]engine.InputMeta{
			"as": {Description: "Variable to store the scope map", Type: "string"},
		},
	})

	// 3. meta.parse
	eng.Register("meta.parse", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		// Resolve the value
		val := resolveValue(node.Value, scope)
		code, ok := val.(string)
		if !ok {
			return fmt.Errorf("meta.parse requires a string value, got %T: %v", val, val)
		}

		// Parse the code
		root, err := engine.ParseString(code, "eval")
		if err != nil {
			return fmt.Errorf("meta.parse error: %v", err)
		}

		// Convert AST to Map
		astMap := nodeToMap(root)

		// Return result via "as" attribute
		target := ""
		for _, child := range node.Children {
			if child.Name == "as" {
				if val, ok := child.Value.(string); ok {
					target = strings.TrimPrefix(val, "$")
				}
			}
		}

		if target != "" {
			scope.Set(target, astMap)
		}

		return nil
	}, engine.SlotMeta{
		Description: "Parses ZenoLang code into an AST Map (Code as Data).",
		Example:     `$ast: meta.parse: "print: 'hello'"`,
		Inputs: map[string]engine.InputMeta{
			"(value)": {Description: "The ZenoLang code string to parse", Type: "string"},
			"as":      {Description: "Variable to store the AST map", Type: "string"},
		},
	})

	// 4. meta.run
	eng.Register("meta.run", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var astMap map[string]interface{}

		// Input can be value or "ast" attribute
		if val, ok := node.Value.(map[string]interface{}); ok {
			astMap = val
		} else {
			// Check children
			for _, child := range node.Children {
				if child.Name == "ast" {
					if m, ok := child.Value.(map[string]interface{}); ok {
						astMap = m
					} else {
						// Try to resolve variable
						val := eng.ResolveShorthandValue(child, scope)
						if m, ok := val.(map[string]interface{}); ok {
							astMap = m
						}
					}
				}
			}
		}

		if astMap == nil {
			// Try to resolve main value if it was a variable reference
			val := eng.ResolveShorthandValue(node, scope)
			if m, ok := val.(map[string]interface{}); ok {
				astMap = m
			}
		}

		if astMap == nil {
			return fmt.Errorf("meta.run requires an AST Map")
		}

		// Convert Map to AST
		root, err := mapToNode(astMap)
		if err != nil {
			return fmt.Errorf("meta.run error: %v", err)
		}

		// Execute
		if err := eng.Execute(ctx, root, scope); err != nil {
			return err
		}

		return nil
	}, engine.SlotMeta{
		Description: "Executes an AST Map as ZenoLang code.",
		Example:     `meta.run: $ast`,
		Inputs: map[string]engine.InputMeta{
			"(value)": {Description: "The AST Map to execute", Type: "map"},
		},
	})

	// 5. meta.template (Render Blade to String)
	eng.Register("meta.template", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		// 1. Prepare Recorder
		rec := httptest.NewRecorder()
		subCtx := context.WithValue(ctx, "httpWriter", http.ResponseWriter(rec))

		// 2. Resolve View File
		var viewFile string
		if node.Value != nil {
			viewFile = coerce.ToString(resolveValue(node.Value, scope))
		}

		// 3. Process Attributes (Children)
		if viewFile == "" {
			for _, c := range node.Children {
				if c.Name == "file" {
					viewFile = coerce.ToString(resolveValue(c.Value, scope))
					continue
				}
				// Bind attribute to scope (variables for the template)
				val := parseNodeValue(c, scope)
				scope.Set(c.Name, val)
			}
		} else {
			for _, c := range node.Children {
				// Check for "as" attribute to store result
				if c.Name == "as" {
					continue // Handle later
				}
				// Bind attribute to scope
				val := parseNodeValue(c, scope)
				scope.Set(c.Name, val)
			}
		}

		if viewFile == "" {
			return fmt.Errorf("meta.template: file required")
		}

		// 4. Load Template (Using shared helpers from blade.go)
		fullPath := filepath.Join(viewRoot(scope), ensureBladeExt(viewFile))
		programNode, err := getCachedOrParse(fullPath)
		if err != nil {
			return err
		}

		// 5. Execute with Recorder Context
		if err := eng.Execute(subCtx, programNode, scope); err != nil {
			return err
		}

		// 6. Capture Output
		output := rec.Body.String()

		// 7. Store in Variable ("as" attribute)
		target := ""
		for _, child := range node.Children {
			if child.Name == "as" {
				val := coerce.ToString(child.Value)
				target = strings.TrimPrefix(val, "$")
			}
		}

		if target != "" {
			scope.Set(target, output)
		}

		return nil

	}, engine.SlotMeta{
		Description: "Renders a Blade template into a string variable (useful for code generation).",
		Example:     `meta.template: 'codegen/route' { resource: 'users'; as: $code }`,
	})

	// 6. engine.slots (Introspection: Get all registered slots and their metadata)
	eng.Register("engine.slots", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		docs := eng.GetDocumentation()
		sortedNames := eng.GetSortedSlotNames()

		slotsList := make([]map[string]interface{}, 0, len(sortedNames))

		for _, name := range sortedNames {
			meta := docs[name]
			slotInfo := make(map[string]interface{})
			slotInfo["name"] = name
			slotInfo["description"] = meta.Description
			slotInfo["example"] = meta.Example

			// Convert inputs map to something manageable for JSON
			inputsInfo := make(map[string]interface{})
			if meta.Inputs != nil {
				for inputName, prop := range meta.Inputs {
					propInfo := make(map[string]interface{})
					propInfo["description"] = prop.Description
					propInfo["required"] = prop.Required
					propInfo["type"] = prop.Type
					inputsInfo[inputName] = propInfo
				}
			}
			slotInfo["inputs"] = inputsInfo

			if meta.RequiredBlocks != nil {
				slotInfo["required_blocks"] = meta.RequiredBlocks
			} else {
				slotInfo["required_blocks"] = []string{}
			}

			if meta.ValueType != "" {
				slotInfo["value_type"] = meta.ValueType
			}

			slotsList = append(slotsList, slotInfo)
		}

		target := ""
		for _, child := range node.Children {
			if child.Name == "as" {
				val := coerce.ToString(child.Value)
				target = strings.TrimPrefix(val, "$")
			}
		}

		if target != "" {
			scope.Set(target, slotsList)
		}

		return nil

	}, engine.SlotMeta{
		Description: "Returns documentation metadata for all registered ZenoLang slots.",
		Example:     `engine.slots: { as: $docs }`,
	})
}

// Helper: Convert Node to Map
func nodeToMap(n *engine.Node) map[string]interface{} {
	if n == nil {
		return nil
	}
	m := make(map[string]interface{})
	m["name"] = n.Name
	m["value"] = n.Value
	m["line"] = n.Line
	m["col"] = n.Col
	m["filename"] = n.Filename

	var children []map[string]interface{}
	for _, c := range n.Children {
		children = append(children, nodeToMap(c))
	}
	m["children"] = children
	return m
}

// Helper: Convert Map to Node
func mapToNode(m map[string]interface{}) (*engine.Node, error) {
	n := &engine.Node{}

	if name, ok := m["name"].(string); ok {
		n.Name = name
	}

	if val, ok := m["value"]; ok {
		n.Value = val
	}

	if line, ok := m["line"]; ok {
		n.Line = coerceToInt(line)
	}
	if col, ok := m["col"]; ok {
		n.Col = coerceToInt(col)
	}
	if filename, ok := m["filename"].(string); ok {
		n.Filename = filename
	}

	if children, ok := m["children"].([]interface{}); ok {
		for _, c := range children {
			if childMap, ok := c.(map[string]interface{}); ok {
				childNode, err := mapToNode(childMap)
				if err != nil {
					return nil, err
				}
				childNode.Parent = n
				n.Children = append(n.Children, childNode)
			}
		}
	} else if childrenList, ok := m["children"].([]map[string]interface{}); ok {
		// Handle specific type slice if applicable
		for _, childMap := range childrenList {
			childNode, err := mapToNode(childMap)
			if err != nil {
				return nil, err
			}
			childNode.Parent = n
			n.Children = append(n.Children, childNode)
		}
	}

	return n, nil
}

func coerceToInt(val interface{}) int {
	if i, ok := val.(int); ok {
		return i
	}
	if f, ok := val.(float64); ok {
		return int(f)
	}
	return 0
}
