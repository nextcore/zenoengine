package cli

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed all:templates
var templatesFS embed.FS

// HandleNew creates a new ZenoEngine project from embedded templates
func HandleNew(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: zeno new <project-name> [--template=mvc|modular]")
		os.Exit(1)
	}

	projectName := args[0]
	templateName := ""

	// Parse flags
	for _, arg := range args[1:] {
		if strings.HasPrefix(arg, "--template=") {
			templateName = strings.TrimPrefix(arg, "--template=")
		}
	}

	// Interactive Prompt if template is not provided
	if templateName == "" {
		fmt.Println("\n🚀 Welcome to ZenoEngine!")
		fmt.Println("Choose a starting boilerplate for your project:")
		fmt.Println("  1) MVC (Classic Laravel-style architecture)")
		fmt.Println("  2) Modular (Domain-Driven, feature-based architecture)")
		fmt.Print("\nEnter choice (1 or 2) [default: 1]: ")

		var choice string
		fmt.Scanln(&choice)

		if choice == "2" || strings.ToLower(choice) == "modular" {
			templateName = "modular"
		} else {
			templateName = "mvc"
		}
	}

	// Validate template name
	if templateName != "mvc" && templateName != "modular" {
		fmt.Printf("❌ Invalid template '%s'. Choose either 'mvc' or 'modular'.\n", templateName)
		os.Exit(1)
	}

	// Create target directory
	targetDir := filepath.Join(".", projectName)
	if _, err := os.Stat(targetDir); !os.IsNotExist(err) {
		fmt.Printf("❌ Directory '%s' already exists. Please choose a different name.\n", projectName)
		os.Exit(1)
	}

	fmt.Printf("\n📦 Creating new %s ZenoEngine project in ./%s...\n", strings.ToUpper(templateName), projectName)

	// Copy embedded files
	srcPrefix := fmt.Sprintf("templates/%s", templateName)
	err := fs.WalkDir(templatesFS, srcPrefix, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path inside the template folder
		relPath, _ := filepath.Rel(srcPrefix, path)
		if relPath == "." {
			return nil
		}

		targetPath := filepath.Join(targetDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		// Ensure parent directory exists before writing file
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		// Read from embedded FS
		content, err := templatesFS.ReadFile(path)
		if err != nil {
			return err
		}

		// Write to disk
		return os.WriteFile(targetPath, content, 0644)
	})

	if err != nil {
		fmt.Printf("❌ Failed to scaffold project: %v\n", err)
		os.Exit(1)
	}

	// Post-processing: .env setup
	setupEnvFile(targetDir)

	fmt.Println("\n✅ Project created successfully!")
	fmt.Println("\nNext steps:")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Println("  ./zeno")
	fmt.Println("\nHappy coding! 🚀")
}

// setupEnvFile renames .env.example to .env and generates secure random keys
func setupEnvFile(targetDir string) {
	envExamplePath := filepath.Join(targetDir, ".env.example")
	envPath := filepath.Join(targetDir, ".env")

	content, err := os.ReadFile(envExamplePath)
	if err != nil {
		// Just skip silently if for some reason it's missing
		return
	}

	envStr := string(content)

	// Replace keys with secure random strings
	envStr = strings.Replace(envStr, "super_secret_jwt_key_here", generateSecureString(32), 1)
	envStr = strings.Replace(envStr, "replace_me_with_random_string", generateSecureString(32), 1)
	envStr = strings.Replace(envStr, "replace_me_with_32_byte_string", generateSecureString(32), 1)

	os.WriteFile(envPath, []byte(envStr), 0644)
}

func generateSecureString(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "fallback_insecure_key_12345"
	}
	return hex.EncodeToString(bytes)[:length]
}
