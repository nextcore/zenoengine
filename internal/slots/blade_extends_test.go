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

func TestBladeExtends(t *testing.T) {
	// 1. Setup Engine
	eng := engine.NewEngine()
	RegisterBladeSlots(eng)

	// 2. Mock File System
	os.MkdirAll("views/test_extends", 0755)
	defer os.RemoveAll("views")

	layoutContent := `<html><body>@yield('content')</body></html>`
	childContent := `@extends('test_extends/layout.blade.zl')
@section('content')
<h1>Hello World</h1>
@endsection`

	os.WriteFile("views/test_extends/layout.blade.zl", []byte(layoutContent), 0644)
	os.WriteFile("views/test_extends/child.blade.zl", []byte(childContent), 0644)

	// 3. Construct Node
	root := &engine.Node{
		Name: "view.blade",
		Value: "test_extends/child.blade.zl",
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
	expected := "<html><body><h1>Hello World</h1></body></html>"
	
	// Normalize spaces/newlines for comparison
	output = strings.ReplaceAll(output, "\n", "")
	output = strings.ReplaceAll(output, "\r", "")
	if output != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, output)
	}
}
