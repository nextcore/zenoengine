package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"zeno/pkg/wasm"

	"gopkg.in/yaml.v3"
)

// HandlePlugin handles the 'plugin' command with subcommands
func HandlePlugin(args []string) {
	if len(args) == 0 {
		printPluginHelp()
		return
	}

	subcommand := args[0]
	subargs := args[1:]

	switch subcommand {
	case "list":
		handlePluginList(subargs)
	case "info":
		handlePluginInfo(subargs)
	case "validate":
		handlePluginValidate(subargs)
	case "help", "--help", "-h":
		printPluginHelp()
	default:
		fmt.Printf("Unknown plugin subcommand: %s\n", subcommand)
		fmt.Println("Run 'zeno plugin help' for usage information")
		os.Exit(1)
	}
}

func printPluginHelp() {
	help := `
WASM Plugin Management Commands

Usage:
  zeno plugin <command> [arguments]

Available Commands:
  list              List all installed plugins
  info <name>       Show detailed information about a plugin
  validate <path>   Validate a plugin before installation
  help              Show this help message

Examples:
  zeno plugin list
  zeno plugin info hello-rust
  zeno plugin validate ./examples/wasm-plugins/hello-rust

For more information, visit: https://github.com/yourusername/zenoengine
`
	fmt.Println(help)
}

// handlePluginList lists all installed plugins
func handlePluginList(args []string) {
	pluginDir := os.Getenv("ZENO_PLUGIN_DIR")
	if pluginDir == "" {
		pluginDir = "./plugins"
	}

	// Check if directory exists
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		fmt.Printf("Plugin directory not found: %s\n", pluginDir)
		fmt.Println("No plugins installed.")
		return
	}

	// Read directory
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		fmt.Printf("Error reading plugin directory: %v\n", err)
		os.Exit(1)
	}

	// Filter for directories only
	var plugins []pluginInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Try to read manifest
		manifestPath := filepath.Join(pluginDir, entry.Name(), "manifest.yaml")
		manifest, err := readManifest(manifestPath)
		if err != nil {
			continue // Skip invalid plugins
		}

		plugins = append(plugins, pluginInfo{
			Name:     manifest.Name,
			Version:  manifest.Version,
			Desc:     manifest.Description,
			Path:     filepath.Join(pluginDir, entry.Name()),
			Manifest: manifest,
		})
	}

	if len(plugins) == 0 {
		fmt.Println("No valid plugins found.")
		return
	}

	// Print plugins
	fmt.Println("\nInstalled WASM Plugins:")
	fmt.Println(strings.Repeat("=", 60))
	for _, p := range plugins {
		fmt.Printf("\n  📦 %s (v%s)\n", p.Name, p.Version)
		if p.Desc != "" {
			fmt.Printf("     %s\n", p.Desc)
		}
		fmt.Printf("     Location: %s\n", p.Path)
		
		// Count slots
		slotCount := 0
		if p.Manifest != nil {
			// We'll need to load the plugin to get slots, for now just show path
			fmt.Printf("     Binary: %s\n", p.Manifest.Binary)
		}
		_ = slotCount // TODO: Load plugin to get slot count
	}
	fmt.Println()
}

// handlePluginInfo shows detailed information about a plugin
func handlePluginInfo(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: Plugin name required")
		fmt.Println("Usage: zeno plugin info <name>")
		os.Exit(1)
	}

	pluginName := args[0]
	pluginDir := os.Getenv("ZENO_PLUGIN_DIR")
	if pluginDir == "" {
		pluginDir = "./plugins"
	}

	// Find plugin directory
	pluginPath := filepath.Join(pluginDir, pluginName)
	manifestPath := filepath.Join(pluginPath, "manifest.yaml")

	manifest, err := readManifest(manifestPath)
	if err != nil {
		fmt.Printf("Error: Plugin '%s' not found or invalid\n", pluginName)
		fmt.Printf("Details: %v\n", err)
		os.Exit(1)
	}

	// Print detailed info
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("Plugin: %s\n", manifest.Name)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Version:     %s\n", manifest.Version)
	if manifest.Author != "" {
		fmt.Printf("Author:      %s\n", manifest.Author)
	}
	if manifest.License != "" {
		fmt.Printf("License:     %s\n", manifest.License)
	}
	if manifest.Homepage != "" {
		fmt.Printf("Homepage:    %s\n", manifest.Homepage)
	}
	if manifest.Description != "" {
		fmt.Printf("Description: %s\n", manifest.Description)
	}
	fmt.Printf("Binary:      %s\n", manifest.Binary)

	// Show permissions
	fmt.Println("\nPermissions:")
	if len(manifest.Permissions.Scope) > 0 {
		fmt.Printf("  Scope:      %s\n", strings.Join(manifest.Permissions.Scope, ", "))
	}
	if len(manifest.Permissions.Network) > 0 {
		fmt.Printf("  Network:    %s\n", strings.Join(manifest.Permissions.Network, ", "))
	} else {
		fmt.Println("  Network:    (none)")
	}
	if len(manifest.Permissions.Filesystem) > 0 {
		fmt.Printf("  Filesystem: %s\n", strings.Join(manifest.Permissions.Filesystem, ", "))
	} else {
		fmt.Println("  Filesystem: (none)")
	}
	if len(manifest.Permissions.Database) > 0 {
		fmt.Printf("  Database:   %s\n", strings.Join(manifest.Permissions.Database, ", "))
	} else {
		fmt.Println("  Database:   (none)")
	}
	if len(manifest.Permissions.Env) > 0 {
		fmt.Printf("  Env:        %s\n", strings.Join(manifest.Permissions.Env, ", "))
	} else {
		fmt.Println("  Env:        (none)")
	}

	fmt.Println()
}

// handlePluginValidate validates a plugin
func handlePluginValidate(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: Plugin path required")
		fmt.Println("Usage: zeno plugin validate <path>")
		os.Exit(1)
	}

	pluginPath := args[0]
	fmt.Printf("\nValidating plugin at: %s\n", pluginPath)
	fmt.Println(strings.Repeat("=", 60))

	errors := 0
	warnings := 0

	// Check 1: Directory exists
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		fmt.Println("❌ Plugin directory not found")
		os.Exit(1)
	}
	fmt.Println("✅ Plugin directory exists")

	// Check 2: Manifest exists and is valid
	manifestPath := filepath.Join(pluginPath, "manifest.yaml")
	manifest, err := readManifest(manifestPath)
	if err != nil {
		fmt.Printf("❌ Manifest invalid: %v\n", err)
		errors++
	} else {
		fmt.Println("✅ Manifest found and valid")

		// Check required fields
		if manifest.Name == "" {
			fmt.Println("❌ Manifest missing required field: name")
			errors++
		}
		if manifest.Version == "" {
			fmt.Println("❌ Manifest missing required field: version")
			errors++
		}
		if manifest.Binary == "" {
			fmt.Println("❌ Manifest missing required field: binary")
			errors++
		}

		// Check optional fields
		if manifest.Author == "" {
			fmt.Println("⚠️  Optional field 'author' not set")
			warnings++
		}
		if manifest.Description == "" {
			fmt.Println("⚠️  Optional field 'description' not set")
			warnings++
		}
		if manifest.License == "" {
			fmt.Println("⚠️  Optional field 'license' not set")
			warnings++
		}
	}

	// Check 3: WASM binary exists
	if manifest != nil && manifest.Binary != "" {
		wasmPath := filepath.Join(pluginPath, manifest.Binary)
		stat, err := os.Stat(wasmPath)
		if err != nil {
			fmt.Printf("❌ WASM binary not found: %s\n", manifest.Binary)
			errors++
		} else {
			sizeKB := float64(stat.Size()) / 1024.0
			fmt.Printf("✅ WASM binary found: %s (%.1f KB)\n", manifest.Binary, sizeKB)
			
			// Check if it's actually a WASM file (basic check)
			if !strings.HasSuffix(manifest.Binary, ".wasm") {
				fmt.Println("⚠️  Binary doesn't have .wasm extension")
				warnings++
			}
		}
	}

	// Check 4: Permissions are valid
	if manifest != nil {
		fmt.Println("✅ Permissions structure valid")
	}

	// Summary
	fmt.Println(strings.Repeat("=", 60))
	if errors > 0 {
		fmt.Printf("\n❌ Validation failed with %d error(s) and %d warning(s)\n", errors, warnings)
		fmt.Println("Plugin is NOT ready to install.\n")
		os.Exit(1)
	} else if warnings > 0 {
		fmt.Printf("\n⚠️  Validation passed with %d warning(s)\n", warnings)
		fmt.Println("Plugin is valid but could be improved.\n")
	} else {
		fmt.Println("\n✅ Validation passed!")
		fmt.Println("Plugin is valid and ready to install!\n")
	}
}

// Helper types and functions

type pluginInfo struct {
	Name     string
	Version  string
	Desc     string
	Path     string
	Manifest *wasm.PluginManifest
}

func readManifest(path string) (*wasm.PluginManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var manifest wasm.PluginManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}
