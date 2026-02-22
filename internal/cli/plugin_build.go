package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// HandlePluginBuild compiles a WASM plugin
// Usage: zeno plugin:build <name>
func HandlePluginBuild(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: zeno plugin:build <name>")
		os.Exit(1)
	}

	name := args[0]
	pluginDir := filepath.Join("plugins", name)

	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		fmt.Printf("❌ Plugin directory '%s' not found\n", pluginDir)
		os.Exit(1)
	}

	// Detect language
	if _, err := os.Stat(filepath.Join(pluginDir, "Cargo.toml")); err == nil {
		buildRust(pluginDir, name)
	} else if _, err := os.Stat(filepath.Join(pluginDir, "go.mod")); err == nil {
		buildGo(pluginDir, name)
	} else {
		fmt.Println("❌ Unknown plugin type (no Cargo.toml or go.mod)")
		os.Exit(1)
	}
}

func buildRust(dir, name string) {
	fmt.Printf("🦀 Building Rust plugin '%s'...\n", name)

	// Check for cargo
	if _, err := exec.LookPath("cargo"); err != nil {
		fmt.Println("❌ 'cargo' not found. Please install Rust.")
		os.Exit(1)
	}

	// cargo build --target wasm32-wasi --release
	cmd := exec.Command("cargo", "build", "--target", "wasm32-wasi", "--release")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Build failed: %v\n", err)
		os.Exit(1)
	}

	// Copy binary to root of plugin dir
	// target/wasm32-wasi/release/name.wasm -> plugin.wasm
	// Note: Rust sanitizes names (dashes to underscores)
	crateName := strings.ReplaceAll(name, "-", "_")
	src := filepath.Join(dir, "target", "wasm32-wasi", "release", crateName+".wasm")
	dst := filepath.Join(dir, "plugin.wasm") // As per default manifest

	// Check if src exists
	if _, err := os.Stat(src); os.IsNotExist(err) {
		// Fallback: maybe wasm32-unknown-unknown?
		src = filepath.Join(dir, "target", "wasm32-unknown-unknown", "release", crateName+".wasm")
	}

	copyFile(src, dst)
	fmt.Printf("✅ Built: %s\n", dst)
}

func buildGo(dir, name string) {
	fmt.Printf("🐹 Building Go plugin '%s'...\n", name)

	// Check for tinygo (preferred) or go
	compiler := "tinygo"
	if _, err := exec.LookPath("tinygo"); err != nil {
		compiler = "go"
		fmt.Println("⚠️  TinyGo not found, using standard Go (larger binaries).")
	}

	dst := filepath.Join(dir, "plugin.wasm")

	var cmd *exec.Cmd
	if compiler == "tinygo" {
		cmd = exec.Command("tinygo", "build", "-o", "plugin.wasm", "-target", "wasi", ".")
	} else {
		// Standard Go (1.21+) supports wasip1
		cmd = exec.Command("go", "build", "-o", "plugin.wasm", ".")
		cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm")
	}

	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Build failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Built: %s\n", dst)
}

func copyFile(src, dst string) {
	input, err := os.ReadFile(src)
	if err != nil {
		fmt.Printf("❌ Failed to read build artifact: %v\n", err)
		return
	}
	err = os.WriteFile(dst, input, 0644)
	if err != nil {
		fmt.Printf("❌ Failed to copy to plugin.wasm: %v\n", err)
	}
}
