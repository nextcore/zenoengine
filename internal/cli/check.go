package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"zeno/internal/app"
	"zeno/pkg/analysis"
	"zeno/pkg/engine"
)

func HandleCheck(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: zeno check [--json] <path/to/script.zl>")
		os.Exit(1)
	}

	isJSON := false
	path := ""

	for _, arg := range args {
		if arg == "--json" {
			isJSON = true
		} else {
			path = arg
		}
	}

	if path == "" {
		fmt.Println("Usage: zeno check [--json] <path/to/script.zl>")
		os.Exit(1)
	}

	root, err := engine.LoadScript(path)
	if err != nil {
		if isJSON {
			if diag, ok := err.(engine.Diagnostic); ok {
				out, _ := json.MarshalIndent(map[string]interface{}{
					"success": false,
					"errors":  []engine.Diagnostic{diag},
				}, "", "  ")
				fmt.Println(string(out))
			} else {
				out, _ := json.MarshalIndent(map[string]interface{}{
					"success": false,
					"errors": []engine.Diagnostic{{
						Type:    "error",
						Message: err.Error(),
					}},
				}, "", "  ")
				fmt.Println(string(out))
			}
		} else {
			fmt.Printf("❌ Syntax Error: %v\n", err)
		}
		os.Exit(1)
	}

	// Setup Engine and Register Slots (to get metadata)
	eng := engine.NewEngine()
	// Skip DB/Queue setup for check, pass nil
	app.RegisterAllSlots(eng, nil, nil, nil, nil)

	// Run Static Analysis
	analyzer := analysis.NewAnalyzer(eng)
	result := analyzer.Analyze(root)

	if isJSON {
		success := len(result.Errors) == 0
		out, _ := json.MarshalIndent(map[string]interface{}{
			"success":  success,
			"errors":   result.Errors,
			"warnings": result.Warnings,
		}, "", "  ")
		fmt.Println(string(out))
		if !success {
			os.Exit(1)
		}
		os.Exit(0)
	}

	if len(result.Errors) > 0 {
		fmt.Printf("❌ Static Analysis Failed (%d errors):\n", len(result.Errors))
		for _, diag := range result.Errors {
			fmt.Printf("  - [%s:%d:%d] %s\n", diag.Filename, diag.Line, diag.Col, diag.Message)
		}
		os.Exit(1)
	}

	for _, diag := range result.Warnings {
		fmt.Printf("⚠️  Warning: [%s:%d:%d] %s\n", diag.Filename, diag.Line, diag.Col, diag.Message)
	}

	fmt.Println("✅ Code Valid (Static Analysis Passed)")
	os.Exit(0)
}
