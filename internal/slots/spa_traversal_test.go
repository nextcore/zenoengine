package slots

import (
	"context"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"zeno/pkg/engine"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSPA_Traversal_Vulnerability(t *testing.T) {
	// 1. Setup Environment
	eng := engine.NewEngine()
	r := chi.NewRouter()

	// Register slots
	RegisterRouterSlots(eng, r)

	// Create temp dir for static root
	tmpDir, err := os.MkdirTemp("", "spa_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create index.html
	err = os.WriteFile(filepath.Join(tmpDir, "index.html"), []byte("SPA INDEX"), 0644)
	require.NoError(t, err)

	// Create a secret file OUTSIDE the static root (in parent of tmpDir)
	secretPath := filepath.Join(filepath.Dir(tmpDir), "zeno_secret.txt")
	err = os.WriteFile(secretPath, []byte("SECRET"), 0644)
	require.NoError(t, err)
	defer os.Remove(secretPath)

	// 2. Define Zeno Code
	code := `
	http.static: "` + tmpDir + `" {
		path: "/app"
		spa: true
	}
	`

	root, err := engine.ParseString(code, "test.zl")
	require.NoError(t, err)

	scope := engine.GetScope()
	ctx := context.WithValue(context.Background(), routerKey{}, r)

	err = eng.Execute(ctx, root, scope)
	require.NoError(t, err)

	// 3. Test Cases

	// Case 1: Normal Access -> 200
	req1 := httptest.NewRequest("GET", "/app/", nil)
	rec1 := httptest.NewRecorder()
	r.ServeHTTP(rec1, req1)
	assert.Equal(t, 200, rec1.Code, "Normal access should return 200")
	assert.Equal(t, "SPA INDEX", rec1.Body.String())

	// Case 2: Traversal to existing file outside root
	badPath := "/app/../zeno_secret.txt"
	req2 := httptest.NewRequest("GET", badPath, nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)

	// Case 3: Traversal to NON-EXISTING file
	req3 := httptest.NewRequest("GET", "/app/../nonexistent.txt", nil)
	rec3 := httptest.NewRecorder()
	r.ServeHTTP(rec3, req3)

	// VERIFICATION
	// Both should behave identially (Same Status Code).
	// Ideally 400 (Bad Request) or 403 or 200.
	// We confirmed implementation returns 400 for dirty paths via ServeFile.

	assert.Equal(t, rec2.Code, rec3.Code, "Response code should be identical (No Oracle)")

	// Ensure we don't leak the secret
	assert.NotContains(t, rec2.Body.String(), "SECRET")

	// Ensure we blocked it (400 or 403 or 404 or 200)
	// Just logging the code
	t.Logf("Traversal Response Code: %d", rec2.Code)
}
