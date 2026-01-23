package cli

import (
	"fmt"
	"os"
)

func HandleCompile(args []string) {
	fmt.Println("❌ Error: Bytecode compilation is no longer supported.")
	fmt.Println("ℹ️  ZenoVM has been separated into a standalone project.")
	fmt.Println("ℹ️  ZenoEngine now uses AST interpretation only.")
	fmt.Println("")
	fmt.Println("To use bytecode compilation:")
	fmt.Println("  1. Install ZenoVM: go install github.com/yourusername/ZenoVM/cmd/zeno-vm@latest")
	fmt.Println("  2. Compile: zeno-vm compile script.zl")
	fmt.Println("  3. Run: zeno-vm run script.zbc")
	os.Exit(1)
}
