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

func TestNativeLoopVariable(t *testing.T) {
	// 1. Setup Engine
	eng := engine.NewEngine()
	RegisterBladeSlots(eng) // Native
	RegisterLogicSlots(eng) // For loop
	
	// 2. Mock View
	os.Mkdir("views", 0755)
	defer os.RemoveAll("views")
	
	content := `
@foreach($items as $i)
  {{ $loop.index }}: {{ $i }} | F:{{ $loop.first }} L:{{ $loop.last }} E:{{ $loop.even }}
@endforeach
`
	os.WriteFile("views/test_loop.blade.zl", []byte(content), 0644)
	
	// 3. Execution
	root := &engine.Node{
		Name: "view.blade",
		Value: "test_loop.blade.zl",
	}
	scope := engine.NewScope(nil)
	scope.Set("items", []interface{}{"A", "B", "C"})
	
	rec := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), "httpWriter", http.ResponseWriter(rec))
	
	err := eng.Execute(ctx, root, scope)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}
	
	output := strings.TrimSpace(rec.Body.String())
	
	// Expected:
	// "0: A | F:true L:false E:false"
	// "1: B | F:false L:false E:true"
	// "2: C | F:false L:true E:false"
	
	// Check A
	if !strings.Contains(output, "0: A | F:true") {
		t.Errorf("First iteration Failed. Output: %s", output)
	}
	// Check B
	if !strings.Contains(output, "1: B | F:false") {
		t.Errorf("Second iteration Failed. Output: %s", output)
	}
	// Check C
	if !strings.Contains(output, "2: C | F:false L:true") {
		t.Errorf("Last iteration Failed. Output: %s", output)
	}
}
