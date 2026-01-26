package cli

import (
	"fmt"
	"os"
	"zeno/pkg/engine"
)

func HandleCheck(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: zeno check <path/to/script.zl>")
		os.Exit(1)
	}
	path := args[0]
	if _, err := engine.LoadScript(path); err != nil {
		fmt.Printf("❌ Syntax Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Syntax Valid")
	os.Exit(0)
}
