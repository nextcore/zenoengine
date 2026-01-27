package cli

import (
	"fmt"
	"os"
	"zeno/internal/app"
	"zeno/pkg/analysis"
	"zeno/pkg/engine"
)

func HandleCheck(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: zeno check <path/to/script.zl>")
		os.Exit(1)
	}
	path := args[0]
	root, err := engine.LoadScript(path)
	if err != nil {
		fmt.Printf("❌ Syntax Error: %v\n", err)
		os.Exit(1)
	}

	// Setup Engine and Register Slots (to get metadata)
	eng := engine.NewEngine()
	// Skip DB/Queue setup for check, pass nil
	app.RegisterAllSlots(eng, nil, nil, nil, nil)

	// Run Static Analysis
	analyzer := analysis.NewAnalyzer(eng)
	result := analyzer.Analyze(root)

	if len(result.Errors) > 0 {
		fmt.Printf("❌ Static Analysis Failed (%d errors):\n", len(result.Errors))
		for _, errStr := range result.Errors {
			fmt.Printf("  - %s\n", errStr)
		}
		os.Exit(1)
	}

	for _, warn := range result.Warnings {
		fmt.Printf("⚠️  Warning: %s\n", warn)
	}

	fmt.Println("✅ Code Valid (Static Analysis Passed)")
	os.Exit(0)
}
