package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"zeno/internal/app"
	"zeno/pkg/engine"
)

// LSPCompletionItem mimics VS Code CompletionItem structure
type LSPCompletionItem struct {
	Label         string `json:"label"`
	Kind          int    `json:"kind"` // 3 = Function, 14 = Keyword
	Detail        string `json:"detail,omitempty"`
	Documentation string `json:"documentation,omitempty"`
	InsertText    string `json:"insertText,omitempty"`
}

// HandleLSPCompletion provides autocomplete suggestions for editors.
// Usage: zeno lsp:completion --prefix="http."
func HandleLSPCompletion(args []string) {
	prefix := ""
	for _, arg := range args {
		if strings.HasPrefix(arg, "--prefix=") {
			prefix = strings.TrimPrefix(arg, "--prefix=")
		}
	}

	// 1. Initialize Engine (Minimal)
	eng := engine.NewEngine()
	// Mock dependencies
	mockSetConfig := func(cfg []string) {}
	app.RegisterAllSlots(eng, nil, nil, nil, mockSetConfig)

	// 2. Get Documentation
	docs := eng.GetDocumentation()

	// 3. Filter and Build Completions
	var items []LSPCompletionItem

	for name, meta := range docs {
		// Simple Prefix Match
		if strings.HasPrefix(name, prefix) {
			// Determine Kind
			kind := 3 // Function
			if meta.Group == "Logic" || meta.Group == "Control Flow" {
				kind = 14 // Keyword
			}

			// Build Detail string
			detail := meta.Description
			if meta.Returns != "" {
				detail = fmt.Sprintf("(%s) %s", meta.Returns, meta.Description)
			}

			// Build Insert Text (Snippet)
			// e.g. http.get: "/path"
			insertText := name + ": "
			if meta.Example != "" {
				// Naive snippet generation from example?
				// Just append colon for now
			}

			items = append(items, LSPCompletionItem{
				Label:         name,
				Kind:          kind,
				Detail:        detail,
				Documentation: meta.Example,
				InsertText:    insertText,
			})
		}
	}

	// 4. Output JSON
	jsonBytes, err := json.Marshal(items)
	if err != nil {
		fmt.Printf("[]") // Return empty list on error
		return
	}
	fmt.Println(string(jsonBytes))
}
