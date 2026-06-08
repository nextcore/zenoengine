package cli

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

// HandleKeyGenerate generates a new key and updates/saves it to the .env file.
func HandleKeyGenerate() {
	key := generateRandomKey(32)
	fmt.Printf("🔑 Generated New Security Key: %s\n", key)

	envFile := ".env"
	content, err := os.ReadFile(envFile)
	if err != nil {
		fmt.Printf("⚠️  .env file not found. Creating a new one from .env.example...\n")
		// Try to read .env.example
		content, err = os.ReadFile(".env.example")
		if err != nil {
			fmt.Printf("❌ .env.example not found either. Please create .env manually.\n")
			return
		}
	}

	lines := strings.Split(string(content), "\n")
	found := false
	for i, line := range lines {
		if strings.HasPrefix(line, "JWT_SECRET=") {
			lines[i] = "JWT_SECRET=" + key
			found = true
			break
		}
	}

	if !found {
		lines = append(lines, "JWT_SECRET="+key)
	}

	err = os.WriteFile(envFile, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		fmt.Printf("❌ Failed to update .env: %v\n", err)
		return
	}

	os.Setenv("JWT_SECRET", key)
	fmt.Printf("✅ Success! JWT_SECRET has been updated in your .env file.\n")
}

// EnsureJWTSecret checks if JWT_SECRET is set. If not, it automatically generates one,
// updates the .env file, and sets the env var for the current process.
func EnsureJWTSecret() {
	if os.Getenv("JWT_SECRET") != "" {
		return
	}

	key := generateRandomKey(32)
	fmt.Printf("⚠️  JWT_SECRET is not set. Automatically generating a secure key...\n")

	envFile := ".env"
	content, _ := os.ReadFile(envFile)
	if len(content) == 0 {
		content, _ = os.ReadFile(".env.example")
	}

	lines := strings.Split(string(content), "\n")
	found := false
	for i, line := range lines {
		if strings.HasPrefix(line, "JWT_SECRET=") {
			lines[i] = "JWT_SECRET=" + key
			found = true
			break
		}
	}

	if !found {
		lines = append(lines, "JWT_SECRET="+key)
	}

	_ = os.WriteFile(envFile, []byte(strings.Join(lines, "\n")), 0644)
	os.Setenv("JWT_SECRET", key)
	fmt.Printf("✅ Success! JWT_SECRET has been updated in your .env file and loaded for this session.\n")
}

func generateRandomKey(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "rahasia_" + hex.EncodeToString([]byte("placeholder"))
	}
	return hex.EncodeToString(b)
}
