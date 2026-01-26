package slots

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"zeno/pkg/engine"
)

func TestNativeErrorDirective(t *testing.T) {
	// Setup
	eng := engine.NewEngine()
	RegisterLogicSlots(eng)
	RegisterBladeSlots(eng)
	
	os.MkdirAll("views", 0755)
	defer os.RemoveAll("views")
	
	// Test template with @error
	viewContent := `
<form>
  <input name="email">
  @error('email')
    <span class="error">{{ $message }}</span>
  @enderror
  
  <input name="password">
  @error('password')
    <div class="alert">{{ $message }}</div>
  @enderror
  
  @error('username')
    <p>This should not appear</p>
  @enderror
</form>
`
	os.WriteFile("views/test_error.blade.zl", []byte(viewContent), 0644)
	
	root := &engine.Node{ Name: "view.blade", Value: "test_error.blade.zl" }
	scope := engine.NewScope(nil)
	
	// Set validation errors
	scope.Set("errors", map[string][]string{
		"email": {"Email is required", "Email must be valid"},
		"password": {"Password must be at least 8 characters"},
		// username has no errors
	})
	
	rec := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), "httpWriter", http.ResponseWriter(rec))
	
	if err := eng.Execute(ctx, root, scope); err != nil {
		t.Fatalf("Exec failed: %v", err)
	}
	
	out := strings.TrimSpace(rec.Body.String())
	
	// Check email error appears
	if !strings.Contains(out, "Email is required") {
		t.Error("Email error message missing")
	}
	if !strings.Contains(out, `<span class="error">`) {
		t.Error("Email error wrapper missing")
	}
	
	// Check password error appears
	if !strings.Contains(out, "Password must be at least 8 characters") {
		t.Error("Password error message missing")
	}
	if !strings.Contains(out, `<div class="alert">`) {
		t.Error("Password error wrapper missing")
	}
	
	// Check username error does NOT appear
	if strings.Contains(out, "This should not appear") {
		t.Error("Username error should not be displayed (no errors for that field)")
	}
}
