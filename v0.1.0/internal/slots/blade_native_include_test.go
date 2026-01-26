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

func TestNativeInclude(t *testing.T) {
	// 1. Setup Engine
	eng := engine.NewEngine()
	RegisterBladeSlots(eng) // Native
	
	// 2. Mock Views
	os.Mkdir("views", 0755)
	defer os.RemoveAll("views")
	
	mainContent := `Start @include('partial.blade.zl', ['name' => 'World']) End`
	partialContent := `Partial: {{ $name }}`
	
	os.WriteFile("views/main.blade.zl", []byte(mainContent), 0644)
	os.WriteFile("views/partial.blade.zl", []byte(partialContent), 0644)
	
	// 3. Construct Node
	root := &engine.Node{
		Name: "view.blade",
		Value: "main.blade.zl",
	}
	
	scope := engine.NewScope(nil)
	
	rec := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), "httpWriter", http.ResponseWriter(rec))
	
	// 4. Execute
	err := eng.Execute(ctx, root, scope)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}
	
	// 5. Verify
	output := strings.TrimSpace(rec.Body.String())
	expected := "Start Partial: World End"
	
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}
