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

func TestNativeSwitch(t *testing.T) {
	// 1. Setup Engine
	eng := engine.NewEngine()
	RegisterBladeSlots(eng) // Native

	// Register switch logic if needed (or rely on native transpiled nodes)
	// If transpiler emits "logic.switch", we need that slot.
	// For now let's assume valid "switch" support via "logic.switch" or similar.
	// Check if RegisterLogicSlots provides switch?
	// If not, we need to register it here or in blade.go temporarilly/permanently.
	// Assuming implementation plan suggests registering "logic.switch" or custom logic.

	// 2. Mock View
	os.Mkdir("views", 0755)
	defer os.RemoveAll("views")

	content := `
@switch($role)
    @case('admin')
        User is Admin
        @break
    @case('editor')
        User is Editor
        @break
    @default
        User is Guest
@endswitch
`
	os.WriteFile("views/test_switch.blade.zl", []byte(content), 0644)

	// 3. Test Cases
	tests := []struct {
		role     string
		expected string
	}{
		{"admin", "User is Admin"},
		{"editor", "User is Editor"},
		{"guest", "User is Guest"},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			root := &engine.Node{
				Name:  "view.blade",
				Value: "test_switch.blade.zl",
			}
			scope := engine.NewScope(nil)
			scope.Set("role", tt.role)

			rec := httptest.NewRecorder()
			ctx := context.WithValue(context.Background(), "httpWriter", http.ResponseWriter(rec))

			err := eng.Execute(ctx, root, scope)
			if err != nil {
				t.Fatalf("Execution failed: %v", err)
			}

			output := strings.TrimSpace(rec.Body.String())
			if !strings.Contains(output, tt.expected) {
				t.Errorf("Role %s: Expected %q, got %q", tt.role, tt.expected, output)
			}
		})
	}
}
