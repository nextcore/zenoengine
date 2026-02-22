package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"zeno/pkg/engine"
	"zeno/pkg/transpiler"
)

// HandleBuild compiles Zeno script to a Go binary.
// Usage: zeno build [src/main.zl]
func HandleBuild(args []string) {
	entryPoint := "src/main.zl"
	if len(args) > 0 {
		entryPoint = args[0]
	}

	fmt.Printf("🏗️  Building %s...\n", entryPoint)

	// 1. Load Script
	root, err := engine.LoadScript(entryPoint)
	if err != nil {
		fmt.Printf("❌ Failed to load script: %v\n", err)
		os.Exit(1)
	}

	// 2. Transpile
	tr := transpiler.New()
	goCode, err := tr.Transpile(root)
	if err != nil {
		fmt.Printf("❌ Transpilation failed: %v\n", err)
		os.Exit(1)
	}

	// 3. Write generated code
	buildDir := "build"
	os.MkdirAll(buildDir, 0755)

	genFile := filepath.Join(buildDir, "main_gen.go")
	if err := os.WriteFile(genFile, []byte(goCode), 0644); err != nil {
		fmt.Printf("❌ Failed to write generated code: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Generated Go code: %s\n", genFile)

	// 4. Compile with Go
	// We need to run 'go build' inside the project root, targeting the generated file.
	// But the generated file declares 'package main'.
	// And it imports "zeno/pkg/..." which assumes we are in the module.
	// So `go build -o dist/app build/main_gen.go` should work if go.mod is in root.

	outFile := "dist/app"
	if os.PathSeparator == '\\' {
		outFile += ".exe"
	}
	os.MkdirAll("dist", 0755)

	fmt.Println("⚙️  Compiling binary...")
	cmd := exec.Command("go", "build", "-o", outFile, genFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Build failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🚀 Build success! Binary: %s\n", outFile)
}
