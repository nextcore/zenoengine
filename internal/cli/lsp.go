package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"zeno/internal/app"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

// LSPCompletionItem mimics VS Code CompletionItem structure
type LSPCompletionItem struct {
	Label         string `json:"label"`
	Kind          int    `json:"kind"` // 3=Function, 6=Variable, 14=Keyword
	Detail        string `json:"detail,omitempty"`
	Documentation string `json:"documentation,omitempty"`
	InsertText    string `json:"insertText,omitempty"`
}

// HandleLSPCompletion provides autocomplete suggestions for editors.
// Usage: zeno lsp:completion --file=x.zl --line=10 --col=5 --content="..."
	var content string
func HandleLSPCompletion(args []string) {
	var line, col int

	// Parse Args
	for _, arg := range args {
		if strings.HasPrefix(arg, "--content=") {
			content = strings.TrimPrefix(arg, "--content=")
		} else if strings.HasPrefix(arg, "--line=") {
			line, _ = strconv.Atoi(strings.TrimPrefix(arg, "--line="))
		} else if strings.HasPrefix(arg, "--col=") {
			col, _ = strconv.Atoi(strings.TrimPrefix(arg, "--col="))
		} else if strings.HasPrefix(arg, "--prefix=") {
			// Legacy support
			// We won't use it directly if line/col is present, but fallback
		}
	}

	// Initialize Engine
	eng := engine.NewEngine()
	mockSetConfig := func(cfg []string) {}
	app.RegisterAllSlots(eng, nil, nil, nil, mockSetConfig)
	docs := eng.GetDocumentation()

	var items []LSPCompletionItem

	// Strategy:
	// 1. If content is provided, parse it to find context (Variables, etc.)
	// 2. Determine Trigger Character (dot, dollar, space)
	// 3. Generate suggestions

	// Determine Context from Content
	// Simple Lexing around cursor
	trigger := ""
	prefix := ""

	if content != "" {
		// Split lines
		lines := strings.Split(content, "\n")
		if line > 0 && line <= len(lines) {
			targetLine := lines[line-1]
			// Safe slice
			if col > 0 && col <= len(targetLine) {
				textBefore := targetLine[:col]
				textAfter := targetLine[col:]
				_ = textAfter // unused for now

				// Analyze textBefore
				// Check for $ (Variable trigger)
				if strings.HasSuffix(textBefore, "$") {
					trigger = "$"
				} else if idx := strings.LastIndex(textBefore, "$"); idx != -1 && idx == len(textBefore)-1 {
					trigger = "$" // Same as HasSuffix
				} else if idx := strings.LastIndexAny(textBefore, " \t:({"); idx != -1 {
					// Check if we are typing a variable e.g. "$us"
					potentialVar := textBefore[idx+1:]
					if strings.HasPrefix(potentialVar, "$") {
						trigger = "$"
						prefix = potentialVar // includes $
					} else {
						// Slot completion
						prefix = potentialVar
					}
				} else {
					prefix = textBefore
				}
			}
		}
	}

	// 1. Variable Completion
	if trigger == "$" || strings.HasPrefix(prefix, "$") {
		// Add Globals
		globals := []string{"$request", "$auth", "$env", "$input", "$query", "$headers"}
		for _, v := range globals {
			if strings.HasPrefix(v, prefix) {
				items = append(items, LSPCompletionItem{
					Label:      v,
					Kind:       6, // Variable
					Detail:     "Global Variable",
					InsertText: v,
				})
			}
		}

		// Add Local Variables (from AST Analysis)
		// We try to parse the content and walk to find defined variables
		if content != "" {
			locals := getVariablesAtCursor(content, line, col)
			for _, v := range locals {
				if strings.HasPrefix(v, prefix) {
					// Dedup against globals
					isGlobal := false
					for _, g := range globals { if g == v { isGlobal = true; break } }
					if !isGlobal {
						items = append(items, LSPCompletionItem{
							Label:      v,
							Kind:       6,
							Detail:     "Local Variable",
							InsertText: v,
						})
					}
				}
			}
		}

	} else {
		// 2. Slot Completion (Default)
		for name, meta := range docs {
			if strings.HasPrefix(name, prefix) {
				kind := 3 // Function
				if meta.Group == "Logic" || meta.Group == "Control Flow" {
					kind = 14 // Keyword
				}

				detail := meta.Description
				if meta.Returns != "" {
					detail = fmt.Sprintf("(%s) %s", meta.Returns, meta.Description)
				}

				// Insert Text: "slot: "
				insertText := name + ": "

				items = append(items, LSPCompletionItem{
					Label:         name,
					Kind:          kind,
					Detail:        detail,
					Documentation: meta.Example,
					InsertText:    insertText,
				})
			}
		}
	}

	jsonBytes, _ := json.Marshal(items)
	fmt.Println(string(jsonBytes))
}

// getVariablesAtCursor parses script and finds variable assignments visible at the cursor line.
func getVariablesAtCursor(content string, targetLine, targetCol int) []string {
	// Tolerant Parse: We might fail if script is incomplete.
	// But `engine.ParseString` handles errors by returning what it could?
	// Actually `ParseString` loops token by token.
	// We can reuse `engine.ParseString` but we need to modify it to be tolerant?
	// Or just use the existing one and ignore error if Root is returned?
	// `parser.go`: returns nil on error.

	// Quick Hack: Remove the line being edited to make it parsable?
	// Or just try.

	// Truncate content to previous line to ensure valid syntax
	lines := strings.Split(content, "\n")
	if targetLine > 1 && targetLine <= len(lines) + 1 {
		content = strings.Join(lines[:targetLine-1], "\n")
	}
	root, _ := engine.ParseString(content, "lsp_temp.zl")
	if root == nil {
		return []string{}
	}

	// Walk
	vars := make(map[string]bool)
	walkScope(root, targetLine, vars)

	var list []string
	for k := range vars {
		list = append(list, k)
	}
	return list
}

func walkScope(node *engine.Node, targetLine int, vars map[string]bool) {
	if node == nil {
		return
	}

	// Stop if we passed the line? No, variables defined AFTER might not be visible,
	// but we are statically analyzing "what variables exist in this file/block".
	// Ideally we only show variables defined BEFORE targetLine.

	if node.Line > targetLine {
		return
	}

	// Detect assignments
	// 1. "var: $x"
	if node.Name == "var" {
		valStr := coerce.ToString(node.Value) // "$x"
		if strings.HasPrefix(valStr, "$") {
			vars[valStr] = true
		}
	}

	// 2. "as: $x" (used in db.get, etc)
	if node.Name == "as" {
		valStr := coerce.ToString(node.Value)
		if strings.HasPrefix(valStr, "$") {
			vars[valStr] = true
		}
	}

	// 3. "$x: value" (Shorthand assignment)
	if strings.HasPrefix(node.Name, "$") {
		vars[node.Name] = true
	}

	for _, child := range node.Children {
		walkScope(child, targetLine, vars)
	}
}
