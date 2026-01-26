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

func TestNativeForLoop(t *testing.T) {
	// 1. Setup Engine
	eng := engine.NewEngine()
	RegisterBladeSlots(eng) // Native
	
	// 2. Mock View
	os.Mkdir("views", 0755)
	defer os.RemoveAll("views")
	
	content := `
@for ($i = 0; $i < 3; $i++)
  Current: {{ $i }}
@endfor
`
	os.WriteFile("views/test_for.blade.zl", []byte(content), 0644)
	
	// 3. Execution
	root := &engine.Node{ Name: "view.blade", Value: "test_for.blade.zl" }
	scope := engine.NewScope(nil)
	
	rec := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), "httpWriter", http.ResponseWriter(rec))
	
	if err := eng.Execute(ctx, root, scope); err != nil {
		t.Fatalf("For Exec failed: %v", err)
	}
	
	out := strings.TrimSpace(rec.Body.String())
	
	// Expect: Current: 0\nCurrent: 1\nCurrent: 2
	if !strings.Contains(out, "Current: 0") { t.Errorf("Expected Current: 0") }
	if !strings.Contains(out, "Current: 1") { t.Errorf("Expected Current: 1") }
	if !strings.Contains(out, "Current: 2") { t.Errorf("Expected Current: 2") }
	if strings.Contains(out, "Current: 3") { t.Errorf("Should not see Current: 3") }
}
