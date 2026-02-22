package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// HandleInstall downloads a package from a git repository.
// Usage: zeno install github.com/user/repo
func HandleInstall(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: zeno install <package-url>")
		fmt.Println("Example: zeno install github.com/zeno-contrib/auth")
		os.Exit(1)
	}

	pkgURL := args[0]
	// Basic normalization: remove https:// prefix if present
	pkgURL = strings.TrimPrefix(pkgURL, "https://")
	pkgURL = strings.TrimPrefix(pkgURL, "http://")
	pkgURL = strings.TrimSuffix(pkgURL, ".git")

	// Validate Git Availability
	if _, err := exec.LookPath("git"); err != nil {
		fmt.Println("❌ Error: 'git' is not installed or not in PATH.")
		os.Exit(1)
	}

	// Construct target directory
	// packages/github.com/user/repo
	targetDir := filepath.Join("packages", pkgURL)

	// Check if already installed
	if _, err := os.Stat(targetDir); err == nil {
		fmt.Printf("⚠️  Package '%s' already exists. Use 'git pull' inside the directory to update.\n", pkgURL)
		return
	}

	fmt.Printf("📦 Installing %s...\n", pkgURL)

	// Construct Git Clone URL
	gitURL := "https://" + pkgURL + ".git"

	cmd := exec.Command("git", "clone", gitURL, targetDir, "--depth", "1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Installation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Installed to %s\n", targetDir)
	fmt.Println("💡 You can now include it using:")
	fmt.Printf("   include: \"%s/main.zl\"\n", targetDir)
}
