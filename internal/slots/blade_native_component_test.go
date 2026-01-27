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

func TestNativeBladeComponents(t *testing.T) {
	// Setup
	eng := engine.NewEngine()
	RegisterBladeSlots(eng)
	
	os.MkdirAll("views/components", 0755)
	defer os.RemoveAll("views")
	
	// Component Definition (views/components/card.blade.zl)
	cardBlade := `
<div class="card">
  <div class="header">{{ $header }}</div>
  <div class="body">{{ $slot }}</div>
  <div class="footer">{{ $type }}</div>
</div>
`
	os.WriteFile("views/components/card.blade.zl", []byte(cardBlade), 0644)
	
	// Usage
	viewContent := `
<x-card type="info" :header="$title">
  <x-slot name="header">Overridden Header</x-slot>
  Main Content Body
</x-card>
`
	// Wait, I passed :header attribute AND x-slot header. 
	// Which one wins? logic.component sets attributes first, then slots override.
	// So x-slot should win.
	
	os.WriteFile("views/test_comp.blade.zl", []byte(viewContent), 0644)
	
	root := &engine.Node{ Name: "view.blade", Value: "test_comp.blade.zl" }
	scope := engine.NewScope(nil)
	scope.Set("title", "Initial Title")
	
	rec := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), "httpWriter", http.ResponseWriter(rec))
	
	if err := eng.Execute(ctx, root, scope); err != nil {
		t.Fatalf("Exec failed: %v", err)
	}
	
	out := strings.TrimSpace(rec.Body.String())
	// Clean newlines for check
	out = strings.ReplaceAll(out, "\n", "")
	
	if !strings.Contains(out, `<div class="card">`) { t.Error("Card wrapper missing") }
	if !strings.Contains(out, `Overridden Header`) { t.Error("Header slot missing") }
	if !strings.Contains(out, `Main Content Body`) { t.Error("Default slot missing") }
	if !strings.Contains(out, `info`) { t.Error("Attribute type missing") }
}
