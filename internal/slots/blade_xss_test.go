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

func TestBladeXSSProtection(t *testing.T) {
	// Setup Engine
	eng := engine.NewEngine()
	RegisterBladeSlots(eng)
	RegisterLogicSlots(eng)

	// Create view with both escaped and unescaped echo
	viewContent := `
Escaped: {{ $malicious }}
Unescaped: {!! $malicious !!}
Safe: {{ $safe }}
`
	os.MkdirAll("views", 0755)
	defer os.RemoveAll("views")
	os.WriteFile("views/test_xss.blade.zl", []byte(viewContent), 0644)

	maliciousScript := "<script>alert('xss')</script>"
	safeText := "Hello World"

	root := &engine.Node{
		Name:  "view.blade",
		Value: "test_xss.blade.zl",
		Children: []*engine.Node{
			{Name: "malicious", Value: maliciousScript},
			{Name: "safe", Value: safeText},
		},
	}

	rec := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), "httpWriter", http.ResponseWriter(rec))
	scope := engine.NewScope(nil)

	if err := eng.Execute(ctx, root, scope); err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	output := rec.Body.String()

	// Verify Escaped
	expectedEscaped := "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"
	if !strings.Contains(output, "Escaped: "+expectedEscaped) {
		t.Errorf("Expected escaped output for {{ }}, got:\n%s", output)
	}

	// Verify Unescaped
	if !strings.Contains(output, "Unescaped: "+maliciousScript) {
		t.Errorf("Expected raw output for {!! !!}, got:\n%s", output)
	}

	// Verify Safe Text unmodified (basic check)
	if !strings.Contains(output, "Safe: "+safeText) {
		t.Errorf("Expected safe text to remain, got:\n%s", output)
	}
}
