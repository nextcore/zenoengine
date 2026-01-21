package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"zeno/pkg/engine"
	"zeno/pkg/engine/vm"
)

func HandleCompile(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: zeno compile <path/to/script.zl> [output.zbc]")
		os.Exit(1)
	}

	inputFile := args[0]
	outputFile := ""

	if len(args) >= 2 {
		outputFile = args[1]
	} else {
		// Default output: input.zbc
		ext := filepath.Ext(inputFile)
		outputFile = strings.TrimSuffix(inputFile, ext) + ".zbc"
	}

	start := time.Now()
	fmt.Printf("Compiling %s...\n", inputFile)

	// 1. Parsing (The slow part)
	root, err := engine.LoadScript(inputFile)
	if err != nil {
		fmt.Printf("❌ Syntax Error: %v\n", err)
		os.Exit(1)
	}

	// 2. Compilation (AST -> Bytecode)
	compiler := vm.NewCompiler()
	chunk, err := compiler.Compile(root)
	if err != nil {
		fmt.Printf("❌ Compilation Error: %v\n", err)
		os.Exit(1)
	}

	// 3. Save to File
	if err := chunk.SaveToFile(outputFile); err != nil {
		fmt.Printf("❌ Save Error: %v\n", err)
		os.Exit(1)
	}

	elapsed := time.Since(start)
	fmt.Printf("✅ Success! Output: %s\n", outputFile)
	fmt.Printf("⏱️  Time taken: %s\n", elapsed)
	fmt.Printf("ℹ️  Run with: zeno run %s\n", outputFile)
}
