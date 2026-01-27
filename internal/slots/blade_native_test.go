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

func TestNativeRenderer(t *testing.T) {
	// 1. Setup Engine
	eng := engine.NewEngine()
	RegisterBladeSlots(eng) // Updated function name
	RegisterLogicSlots(eng)
	
	// 2. Create Temp View File
	viewContent := `<h1>Hello {{ $name }}</h1>
@if($admin)
  <span>Admin</span>
@endif
<ul>
@foreach($items as $item)
  <li>{{ $item }}</li>
@endforeach
</ul>`
	
	os.MkdirAll("views", 0755)
	defer os.RemoveAll("views")
	os.WriteFile("views/test_native.blade.zl", []byte(viewContent), 0644)
	
	// 3. Construct Zeno Node to Call view.blade.native
	// view.blade.native: "test_native.blade.zl" { name: "Zeno"; admin: "true"; items: ["A", "B"] }
	
	root := &engine.Node{
		Name: "view.blade", // Updated slot name
		Value: "test_native.blade.zl",
		Children: []*engine.Node{
			{Name: "name", Value: "Zeno"},
			{Name: "admin", Value: "true"}, // String "true" for simple coercion in text check
			{Name: "items", Value: []interface{}{"A", "B"}},
		},
	}
	
	// 4. Mock HTTP Writer
	rec := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), "httpWriter", http.ResponseWriter(rec))
	scope := engine.NewScope(nil)

	// 5. Execute
	err := eng.Execute(ctx, root, scope)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}
	
	// 6. Verify Output
	output := rec.Body.String()
	
	expectedSubstrings := []string{
		"<h1>Hello Zeno</h1>",
		"<span>Admin</span>",
		"<li>A</li>",
		"<li>B</li>",
	}
	
	for _, exp := range expectedSubstrings {
		if !strings.Contains(output, exp) {
			t.Errorf("Output missing: %s. Got:\n%s", exp, output)
		}
	}
}
