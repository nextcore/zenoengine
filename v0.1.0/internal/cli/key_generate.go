package cli

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

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

	fmt.Printf("✅ Success! JWT_SECRET has been updated in your .env file.\n")
}

func generateRandomKey(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "rahasia_" + hex.EncodeToString([]byte("placeholder"))
	}
	return hex.EncodeToString(b)
}
