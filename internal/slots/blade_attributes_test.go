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

func TestBladeAttributesBag(t *testing.T) {
	// Setup
	eng := engine.NewEngine()
	RegisterBladeSlots(eng)

	os.MkdirAll("views/components", 0755)
	defer os.RemoveAll("views")

	// Component Definition using $attributes
	// Expecting <div {{ $attributes }}>
	// Or at least {{ $attributes }} printing class="foo"
	btnBlade := `
<button {{ $attributes }}>{{ $slot }}</button>
`
	os.WriteFile("views/components/btn.blade.zl", []byte(btnBlade), 0644)

	// Usage
	viewContent := `<x-btn class="primary" type="submit">Click Me</x-btn>`

	os.WriteFile("views/test_attr.blade.zl", []byte(viewContent), 0644)

	root := &engine.Node{ Name: "view.blade", Value: "test_attr.blade.zl" }
	scope := engine.NewScope(nil)

	rec := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), "httpWriter", http.ResponseWriter(rec))

	if err := eng.Execute(ctx, root, scope); err != nil {
		t.Fatalf("Exec failed: %v", err)
	}

	out := rec.Body.String()
	t.Logf("Output: %s", out)

	if !strings.Contains(out, `class="primary"`) { t.Error("Attribute 'class' missing from output") }
	if !strings.Contains(out, `type="submit"`) { t.Error("Attribute 'type' missing from output") }

	// [FIX] Verify that the original <x-tag> is NOT present (transformation success)
	if strings.Contains(out, `<x-btn`) { t.Error("Component tag was not transformed (false positive)") }
}
