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

	root, err := engine.ParseString(code, "exploit.zl")
	require.NoError(t, err)

	scope := engine.GetScope()
	err = eng.Execute(context.Background(), root, scope)

	// 4. Verification

	// Expect Error (Security Violation)
	assert.Error(t, err, "Should block critical file modification")
	if err != nil {
		assert.Contains(t, err.Error(), "security violation", "Error message should mention security violation")
	}

	// Expect Content Unchanged
	currentContent, readErr := os.ReadFile(criticalFile)
	require.NoError(t, readErr)
	assert.Equal(t, originalContent, string(currentContent), "File content should remain unchanged")

	// 5. Test Safe File Write (Regression Check)
	safeFile := filepath.Join(tmpDir, "data.json")
	safeCode := `
	io.file.write: {
		path: "` + safeFile + `"
		content: "{}"
	}
	`
	safeRoot, _ := engine.ParseString(safeCode, "safe.zl")
	err = eng.Execute(context.Background(), safeRoot, scope)
	assert.NoError(t, err, "Should allow writing to safe extensions")
}
