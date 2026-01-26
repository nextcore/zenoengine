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

func TestNativePushStack(t *testing.T) {
	// 1. Setup Engine
	eng := engine.NewEngine()
	RegisterBladeSlots(eng) // Native
	
	// 2. Mock View
	os.Mkdir("views", 0755)
	defer os.RemoveAll("views")
	
	content := `
Start
@push('scripts')
  <script>console.log('pushed');</script>
@endpush
Middle
@stack('scripts')
End
`
	os.WriteFile("views/test_push.blade.zl", []byte(content), 0644)
	
	// 3. Execution
	root := &engine.Node{
		Name: "view.blade",
		Value: "test_push.blade.zl",
	}
	scope := engine.NewScope(nil)
	
	rec := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), "httpWriter", http.ResponseWriter(rec))
	
	err := eng.Execute(ctx, root, scope)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}
	
	output := strings.TrimSpace(rec.Body.String())
	
	// Expected:
	// "Start"
	// "Middle"
	// "<script>..."
	// "End"
	
	if !strings.Contains(output, "Start") || !strings.Contains(output, "Middle") || !strings.Contains(output, "End") {
		t.Errorf("Missing main content structure")
	}
	
	// Verify order: Middle before Script
	middleIdx := strings.Index(output, "Middle")
	scriptIdx := strings.Index(output, "<script>console.log('pushed');</script>")
	
	if scriptIdx == -1 {
		t.Errorf("Stack content not rendered")
	}
	if middleIdx > scriptIdx {
		t.Errorf("Stack rendered too early? Script at %d, Middle at %d", scriptIdx, middleIdx)
	}
}
