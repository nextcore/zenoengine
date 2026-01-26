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

func TestNativeCanDirective(t *testing.T) {
	// Setup
	eng := engine.NewEngine()
	RegisterLogicSlots(eng)
	RegisterBladeSlots(eng)
	
	os.MkdirAll("views", 0755)
	defer os.RemoveAll("views")
	
	// Test template with @can/@cannot
	viewContent := `
@can('edit', $post)
  <a href="/edit">Edit Post</a>
@endcan

@cannot('delete', $post)
  <span>You cannot delete this</span>
@endcannot

@can('create')
  <a href="/create">Create New</a>
@endcan

@cannot('admin')
  <p>Not an admin</p>
@endcannot
`
	os.WriteFile("views/test_can.blade.zl", []byte(viewContent), 0644)
	
	root := &engine.Node{ Name: "view.blade", Value: "test_can.blade.zl" }
	scope := engine.NewScope(nil)
	
	// Mock post object
	post := map[string]interface{}{
		"id": 123,
		"author_id": 1,
	}
	scope.Set("post", post)
	
	// Set authorization callback
	scope.Set("can", func(ability string, resource interface{}) bool {
		switch ability {
		case "edit":
			// Can edit own posts
			if p, ok := resource.(map[string]interface{}); ok {
				return p["author_id"] == 1
			}
			return false
		case "delete":
			// Cannot delete (not admin)
			return false
		case "create":
			// Can create
			return true
		case "admin":
			// Not admin
			return false
		}
		return false
	})
	
	rec := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), "httpWriter", http.ResponseWriter(rec))
	
	if err := eng.Execute(ctx, root, scope); err != nil {
		t.Fatalf("Exec failed: %v", err)
	}
	
	out := strings.TrimSpace(rec.Body.String())
	t.Logf("Output:\n%s", out)
	
	// Check @can('edit') shows (authorized)
	if !strings.Contains(out, "Edit Post") {
		t.Error("Edit link should be visible (user can edit)")
	}
	
	// Check @cannot('delete') shows (not authorized)
	if !strings.Contains(out, "You cannot delete this") {
		t.Error("Delete message should be visible (user cannot delete)")
	}
	
	// Check @can('create') shows
	if !strings.Contains(out, "Create New") {
		t.Error("Create link should be visible")
	}
	
	// Check @cannot('admin') shows
	if !strings.Contains(out, "Not an admin") {
		t.Error("Admin message should be visible (user is not admin)")
	}
}
