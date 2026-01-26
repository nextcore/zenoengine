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

func TestNativeLDX(t *testing.T) {
	eng := engine.NewEngine()
	RegisterBladeSlots(eng)

	os.Mkdir("views", 0755)
	defer os.RemoveAll("views")

	content := `
<form>
  @method('PUT')
  @for($i=1; $i<=5; $i++)
    Item {{ $i }}
    @break($i == 2)
  @endfor
</form>
`
	os.WriteFile("views/test_ldx.blade.zl", []byte(content), 0644)

	root := &engine.Node{Name: "view.blade", Value: "test_ldx.blade.zl"}
	scope := engine.NewScope(nil)

	rec := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), "httpWriter", http.ResponseWriter(rec))

	if err := eng.Execute(ctx, root, scope); err != nil {
		t.Fatalf("LDX Exec failed: %v", err)
	}

	out := rec.Body.String()

	// 1. Check @method
	if !strings.Contains(out, `<input type="hidden" name="_method" value="PUT">`) {
		t.Errorf("Expected hidden method input, got: %s", out)
	}

	// 2. Check conditional @break
	if !strings.Contains(out, "Item 2") {
		t.Errorf("Expected Item 2")
	}
	if strings.Contains(out, "Item 3") {
		t.Errorf("Should have broken at 2, but found Item 3")
	}
}
