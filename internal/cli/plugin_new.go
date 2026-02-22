package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// HandlePluginNew scaffolds a new WASM plugin project
// Usage: zeno plugin:new <name> [--lang=rust|go]
func HandlePluginNew(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: zeno plugin:new <name> [--lang=rust|go]")
		os.Exit(1)
	}

	name := args[0]
	lang := "rust" // Default

	// Parse flags
	for _, arg := range args {
		if strings.HasPrefix(arg, "--lang=") {
			lang = strings.TrimPrefix(arg, "--lang=")
		}
	}

	if lang != "rust" && lang != "go" {
		fmt.Printf("❌ Unsupported language: %s. Supported: rust, go\n", lang)
		os.Exit(1)
	}

	pluginDir := filepath.Join("plugins", name)
	if _, err := os.Stat(pluginDir); err == nil {
		fmt.Printf("❌ Plugin directory '%s' already exists\n", pluginDir)
		os.Exit(1)
	}

	fmt.Printf("✨ Creating new %s plugin: %s...\n", lang, name)

	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		fmt.Printf("❌ Failed to create directory: %v\n", err)
		os.Exit(1)
	}

	// 1. Create manifest.yaml
	manifestContent := fmt.Sprintf(`name: %s
version: 0.1.0
description: A Zeno WASM plugin
type: wasm
binary: plugin.wasm
permissions:
  network: []
  env: []
  filesystem: []
config:
  api_key:
    type: string
    required: false
`, name)

	if err := os.WriteFile(filepath.Join(pluginDir, "manifest.yaml"), []byte(manifestContent), 0644); err != nil {
		fmt.Printf("❌ Failed to create manifest: %v\n", err)
		os.Exit(1)
	}

	// 2. Create Source Code
	if lang == "rust" {
		scaffoldRust(pluginDir, name)
	} else if lang == "go" {
		scaffoldGo(pluginDir, name)
	}

	fmt.Printf("✅ Plugin created at %s\n", pluginDir)
	fmt.Println("👉 To build:")
	fmt.Println("   zeno plugin:build " + name)
}

func scaffoldRust(dir, name string) {
	// src directory
	os.MkdirAll(filepath.Join(dir, "src"), 0755)

	// Cargo.toml
	cargoContent := fmt.Sprintf(`[package]
name = "%s"
version = "0.1.0"
edition = "2021"

[lib]
crate-type = ["cdylib"]

[dependencies]
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
extism-pdk = "1.0" # Using Extism PDK for simplicity or custom bindings?
# For Zeno, we likely have a specific PDK or raw WASI.
# Assuming raw JSON-RPC style for now as seen in manager logic?
# Actually manager uses 'wazero'.
# Let's assume a simple guest structure.
`, name)

	os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte(cargoContent), 0644)

	// src/lib.rs
	rsContent := `use extism_pdk::*;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize)]
struct Params {
    name: String,
}

#[plugin_fn]
pub fn greet(input: String) -> FnResult<String> {
    let params: Params = serde_json::from_str(&input)?;
    let output = format!("Hello, {}! From Rust WASM.", params.name);
    Ok(output)
}
`
	os.WriteFile(filepath.Join(dir, "src", "lib.rs"), []byte(rsContent), 0644)
}

func scaffoldGo(dir, name string) {
	// go.mod
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte(fmt.Sprintf("module %s\n\ngo 1.21\n", name)), 0644)

	// main.go
	goContent := `package main

// TinyGo WASM entrypoint
func main() {}

//export greet
func greet(ptr, size uint32) uint64 {
    // Boilerplate for reading/writing memory would go here
    // For now simple stub
    return 0
}
`
	os.WriteFile(filepath.Join(dir, "main.go"), []byte(goContent), 0644)
}
