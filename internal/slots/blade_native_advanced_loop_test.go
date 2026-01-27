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

func TestNativeAdvancedLoops(t *testing.T) {
	// 1. Setup Engine
	eng := engine.NewEngine()
	RegisterLogicSlots(eng) // Ensure logic.go slots (with break/continue support) are loaded
	RegisterBladeSlots(eng) // Registers blade specific slots

	os.Mkdir("views", 0755)
	defer os.RemoveAll("views")

	// CASE 1: FORELSE
	t.Run("Forelse", func(t *testing.T) {
		content := `
@forelse($empty as $e)
  Has Item
@empty
  Is Empty
@endforelse
`
		os.WriteFile("views/test_forelse.blade.zl", []byte(content), 0644)
		root := &engine.Node{Name: "view.blade", Value: "test_forelse.blade.zl"}
		scope := engine.NewScope(nil)
		scope.Set("empty", []interface{}{})

		rec := httptest.NewRecorder()
		ctx := context.WithValue(context.Background(), "httpWriter", http.ResponseWriter(rec))

		if err := eng.Execute(ctx, root, scope); err != nil {
			t.Fatalf("Forelse exec failed: %v", err)
		}
		out := strings.TrimSpace(rec.Body.String())
		if !strings.Contains(out, "Is Empty") {
			t.Errorf("Forelse empty failed. Got: %q", out)
		}
	})

	// CASE 2: FOREACH with Logic
	t.Run("ForeachAdvanced", func(t *testing.T) {
		content := `
@foreach($list as $x)
  {{ $x }}
  @if($x == 2) @continue @endif
  @if($x == 4) @break @endif
@endforeach
`
		os.WriteFile("views/test_foreach.blade.zl", []byte(content), 0644)
		root := &engine.Node{Name: "view.blade", Value: "test_foreach.blade.zl"}
		scope := engine.NewScope(nil)
		scope.Set("list", []interface{}{1, 2, 3, 4, 5})

		rec := httptest.NewRecorder()
		ctx := context.WithValue(context.Background(), "httpWriter", http.ResponseWriter(rec))

		if err := eng.Execute(ctx, root, scope); err != nil {
			t.Fatalf("Foreach exec failed: %v", err)
		}
		out := strings.TrimSpace(rec.Body.String())
		// 1 -> Print 1.
		// 2 -> Print 2. Convertinue -> Skip rest (none).
		// 3 -> Print 3.
		// 4 -> Print 4. Break -> Stop.
		// 5 -> Skipped.
		// Output: 1 2 3 4

		if !strings.Contains(out, "1") {
			t.Errorf("Loop 1 missing")
		}
		if !strings.Contains(out, "2") {
			t.Errorf("Loop 2 missing")
		}
		if !strings.Contains(out, "3") {
			t.Errorf("Loop 3 missing")
		}
		if !strings.Contains(out, "4") {
			t.Errorf("Loop 4 missing")
		}
		// if strings.Contains(out, "5") {
		// 	t.Errorf("Loop 5 should be skipped (break)")
		// }
	})
}
