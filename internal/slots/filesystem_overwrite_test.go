package slots

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"zeno/pkg/engine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileSystem_Overwrite_Vulnerability(t *testing.T) {
	// 1. Setup Engine
	eng := engine.NewEngine()
	RegisterFileSystemSlots(eng)

	// 2. Create a dummy source file (.zl)
	tmpDir, err := os.MkdirTemp("", "zeno_rce_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	criticalFile := filepath.Join(tmpDir, "critical_logic.zl")
	originalContent := `log: "Safe Logic"`
	err = os.WriteFile(criticalFile, []byte(originalContent), 0644)
	require.NoError(t, err)

	// 3. Attempt to overwrite it using io.file.write
	maliciousContent := `log: "HACKED"`

	code := `
	io.file.write: {
		path: "` + criticalFile + `"
		content: "` + maliciousContent + `"
	}
	`

	// -----------------------------------------------------
	// CASE A: Production Mode (Default/Unset) - SHOULD BLOCK
	// -----------------------------------------------------
	t.Run("Production_Blocks_Critical_Files", func(t *testing.T) {
		os.Setenv("APP_ENV", "production")
		defer os.Unsetenv("APP_ENV")

		// Reset file
		os.WriteFile(criticalFile, []byte(originalContent), 0644)

		root, _ := engine.ParseString(code, "exploit.zl")
		scope := engine.GetScope()
		err = eng.Execute(context.Background(), root, scope)

		// Expect Error
		assert.Error(t, err, "Should block critical file modification in production")
		if err != nil {
			assert.Contains(t, err.Error(), "security violation", "Error message should mention security violation")
		}

		// Expect Content Unchanged
		currentContent, _ := os.ReadFile(criticalFile)
		assert.Equal(t, originalContent, string(currentContent), "File content should remain unchanged")
	})

	// -----------------------------------------------------
	// CASE B: Development Mode - SHOULD ALLOW
	// -----------------------------------------------------
	t.Run("Development_Allows_Critical_Files", func(t *testing.T) {
		os.Setenv("APP_ENV", "development")
		defer os.Unsetenv("APP_ENV")

		// Reset file
		os.WriteFile(criticalFile, []byte(originalContent), 0644)

		root, _ := engine.ParseString(code, "exploit.zl")
		scope := engine.GetScope()
		err = eng.Execute(context.Background(), root, scope)

		// Expect NO Error
		assert.NoError(t, err, "Should allow critical file modification in development")

		// Expect Content Changed
		currentContent, _ := os.ReadFile(criticalFile)
		// We use Contains because Zeno parser might change quoting/spacing slightly
		assert.Contains(t, string(currentContent), "HACKED", "File content SHOULD be changed in dev mode")
	})
}
