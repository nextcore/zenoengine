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

func TestNativeConditionals(t *testing.T) {
	// 1. Setup Engine
	eng := engine.NewEngine()
	RegisterBladeSlots(eng) // Native
	
	// Register logic slots if we use them, or rely on native transpiler handling
	// We might need to register logic.isset/empty if we use that approach.
	// For now, assume we will implement them.
	
	// 2. Mock View
	os.Mkdir("views", 0755)
	defer os.RemoveAll("views")
	
	content := `
@isset($definedVar)
  DEFINED
@endisset

@isset($undefinedVar)
  UNDEFINED_SHOWN
@endisset

@empty($emptyList)
  EMPTY_LIST
@endempty

@empty($filledList)
  FILLED_LIST_SHOWN
@endempty

@unless($isTrue)
  UNLESS_TRUE_SHOWN
@endunless

@unless($isFalse)
  UNLESS_FALSE
@endunless
`
	os.WriteFile("views/test_conditionals.blade.zl", []byte(content), 0644)
	
	// 3. Execution
	root := &engine.Node{
		Name: "view.blade",
		Value: "test_conditionals.blade.zl",
	}
	scope := engine.NewScope(nil)
	scope.Set("definedVar", "exists")
	scope.Set("emptyList", []string{})
	scope.Set("filledList", []string{"a"})
	scope.Set("isTrue", true)
	scope.Set("isFalse", false)
	
	rec := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), "httpWriter", http.ResponseWriter(rec))
	
	err := eng.Execute(ctx, root, scope)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}
	
	output := strings.TrimSpace(rec.Body.String())
	
	// Expectations
	if !strings.Contains(output, "DEFINED") { t.Errorf("Expected DEFINED") }
	if strings.Contains(output, "UNDEFINED_SHOWN") { t.Errorf("Did not expect UNDEFINED_SHOWN") }
	
	if !strings.Contains(output, "EMPTY_LIST") { t.Errorf("Expected EMPTY_LIST") }
	if strings.Contains(output, "FILLED_LIST_SHOWN") { t.Errorf("Did not expect FILLED_LIST_SHOWN") }
	
	if strings.Contains(output, "UNLESS_TRUE_SHOWN") { t.Errorf("Did not expect UNLESS_TRUE_SHOWN") }
	if !strings.Contains(output, "UNLESS_FALSE") { t.Errorf("Expected UNLESS_FALSE") }
}
