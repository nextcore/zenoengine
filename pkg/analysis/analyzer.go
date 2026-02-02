package analysis

import (
	"fmt"
	"strings"
	"zeno/pkg/engine"
)

type AnalysisResult struct {
	Errors   []engine.Diagnostic
	Warnings []engine.Diagnostic
}

type Analyzer struct {
	engine  *engine.Engine
	Visited map[string]bool
}

func NewAnalyzer(eng *engine.Engine) *Analyzer {
	return &Analyzer{
		engine:  eng,
		Visited: make(map[string]bool),
	}
}

func (a *Analyzer) Analyze(root *engine.Node) AnalysisResult {
	res := AnalysisResult{}
	a.Visited = make(map[string]bool) // Reset for fresh analysis
	a.walk(root, &res)
	return res
}

func (a *Analyzer) walk(node *engine.Node, res *AnalysisResult) {
	if node == nil {
		return
	}

	// [NEW] Recursive Analysis for Includes
	if node.Name == "include" {
		path := ""
		if node.Value != nil {
			path = fmt.Sprintf("%v", node.Value)
			// Clean quotes if present
			path = strings.Trim(path, "\"'`")
		}

		if path != "" {
			// Cycle Detection
			if a.Visited[path] {
				// Already visited, skip to prevent infinite recursion
				return
			}
			a.Visited[path] = true

			// Load and Analyze
			includedRoot, err := engine.LoadScript(path)
			if err != nil {
				if diag, ok := err.(engine.Diagnostic); ok {
					res.Errors = append(res.Errors, diag)
				} else {
					res.Errors = append(res.Errors, engine.Diagnostic{
						Type:     "error",
						Message:  fmt.Sprintf("include error: failed to load %s: %v", path, err),
						Filename: node.Filename,
						Line:     node.Line,
						Col:      node.Col,
					})
				}
			} else {
				a.walk(includedRoot, res)
			}
		}
		return
	}

	// Skip root node logic, just recurse
	if node.Name == "root" {
		for _, child := range node.Children {
			a.walk(child, res)
		}
		return
	}

	// 1. Check if it's a slot or a variable assignment
	if len(node.Name) > 0 && !strings.HasPrefix(node.Name, "$") {
		// It's a slot, check validity
		meta, exists := a.engine.Docs[node.Name]
		if !exists {
			// Basic language keywords that might not be in Docs but are handled in executor
			if node.Name != "if" && node.Name != "for" && node.Name != "foreach" &&
				node.Name != "while" && node.Name != "switch" && node.Name != "try" &&
				node.Name != "do" && node.Name != "then" && node.Name != "else" && node.Name != "catch" &&
				node.Name != "as" && node.Name != "rules" && node.Name != "rules_map" && node.Name != "data" {

				// [ADAPTIVE] If parent is 'call' or other dynamic slots, children are likely dynamic arguments
				isCallArg := false
				if node.Parent != nil {
					pMeta, pExists := a.engine.Docs[node.Parent.Name]
					// If parent exists but has NO defined inputs, assume children might be dynamic attributes
					if pExists && pMeta.Inputs == nil {
						isCallArg = true
					}
					// Explicit dynamic slots
					if node.Parent.Name == "call" || node.Parent.Name == "data" || node.Parent.Name == "array.pop" || node.Parent.Name == "date.now" || node.Parent.Name == "system.env" || node.Parent.Name == "coalesce" {
						isCallArg = true
					}
				}

				if !isCallArg {
					res.Errors = append(res.Errors, engine.Diagnostic{
						Type:     "error",
						Message:  fmt.Sprintf("static error: unknown slot '%s'", node.Name),
						Filename: node.Filename,
						Line:     node.Line,
						Col:      node.Col,
						Slot:     node.Name,
					})
				}
			}
		} else {
			// 1.5 Check Main Value Type
			if meta.ValueType != "" && meta.ValueType != "any" {
				// Don't use ResolveShorthandValue here because it treats children as a Map,
				// but for main value validation we ONLY care about the literal value of the node itself.
				if node.Value != nil {
					valStr := fmt.Sprintf("%v", node.Value)
					if valStr != "" && !strings.HasPrefix(valStr, "$") && !strings.Contains(valStr, "??") {
						// Create a temporary dummy node without children for ValidateValueType to get the correct literal
						dummy := *node
						dummy.Children = nil
						if err := a.engine.ValidateValueType(a.engine.ResolveShorthandValue(&dummy, nil), meta.ValueType, node, node.Name); err != nil {
							res.Errors = append(res.Errors, engine.Diagnostic{
								Type:     "error",
								Message:  fmt.Sprintf("static check failed: %v", err),
								Filename: node.Filename,
								Line:     node.Line,
								Col:      node.Col,
								Slot:     node.Name,
							})
						}
					}
				}
			}

			// 2. Check Required Attributes
			firstRequired := ""
			for name, input := range meta.Inputs {
				if input.Required {
					firstRequired = name
					break
				}
			}

			for name, input := range meta.Inputs {
				if input.Required {
					found := false
					for _, child := range node.Children {
						if child.Name == name {
							found = true
							break
						}
					}

					// [NEW] Positional Value Satisfaction
					// If it's a common positional attribute and the node has a main value, it's satisfied.
					isCommonPositional := (name == "id" || name == "spreadsheet_id" || name == "path" || name == "name" || name == "url" || name == "file")
					if !found && (isCommonPositional || name == firstRequired) && node.Value != nil && fmt.Sprintf("%v", node.Value) != "" {
						found = true
					}

					if !found {
						res.Errors = append(res.Errors, engine.Diagnostic{
							Type:     "error",
							Message:  fmt.Sprintf("static error: missing required attribute '%s' for slot '%s'", name, node.Name),
							Filename: node.Filename,
							Line:     node.Line,
							Col:      node.Col,
							Slot:     node.Name,
						})
					}
				}
			}

			// 3. Check Required Blocks
			for _, blockName := range meta.RequiredBlocks {
				found := false
				for _, child := range node.Children {
					if child.Name == blockName {
						found = true
						break
					}
				}
				if !found {
					res.Errors = append(res.Errors, engine.Diagnostic{
						Type:     "error",
						Message:  fmt.Sprintf("static error: missing required block '%s:' for slot '%s'", blockName, node.Name),
						Filename: node.Filename,
						Line:     node.Line,
						Col:      node.Col,
						Slot:     node.Name,
					})
				}
			}

			// 4. Check Type for Constant Values
			for _, child := range node.Children {
				if input, hasInput := meta.Inputs[child.Name]; hasInput {
					if input.Type != "" && input.Type != "any" {
						valStr := fmt.Sprintf("%v", child.Value)
						if valStr != "" && !strings.HasPrefix(valStr, "$") && !strings.Contains(valStr, "??") {
							if err := a.engine.ValidateValueType(a.engine.ResolveShorthandValue(child, nil), input.Type, child, node.Name); err != nil {
								res.Errors = append(res.Errors, engine.Diagnostic{
									Type:     "error",
									Message:  fmt.Sprintf("static check failed: %v", err),
									Filename: child.Filename,
									Line:     child.Line,
									Col:      child.Col,
									Slot:     node.Name,
								})
							}
						}
					}
				}
			}

			// 5. Recurse nested slots (Only children that are NOT attributes)
			for _, child := range node.Children {
				if _, isInput := meta.Inputs[child.Name]; !isInput {
					a.walk(child, res)
				}
			}
			return // Done with this node and its children
		}
	}

	// Default recursion for root or unknown structures (though unknown slots return early above)
	for _, child := range node.Children {
		a.walk(child, res)
	}
}
