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

func TestNativeBladeFeatures(t *testing.T) {
	// 1. Setup Engine
	eng := engine.NewEngine()
	RegisterBladeSlots(eng)
	RegisterLogicSlots(eng) // For if/else/for

	// 2. Mock View Content
	// Issues to test:
	// A. Function Calls: {{ formatDate($date, "02 Jan 2006") }}
	// B. @else Logic: @if($cond) YES @else NO @endif
	// C. @csrf: @csrf
	viewContent := `START
@csrf
@if($show)
  SHOW
@else
  HIDE
@endif
{{ formatDate($date, "Y-m-d") }}
END`

	// 3. Mock Data & Scope
	scope := engine.NewScope(nil)
	scope.Set("show", false) // Should trigger @else
	scope.Set("date", "2024-01-01")
	// CSRF fields are usually set by the slot from context, we can mock them in scope for simplicity if checking transpiler 
	// But let's mock the slot's expected behavior validation through context too if possible, 
	// OR just set them in scope manually to verify `{{ $csrf_field }}` works.
	scope.Set("csrf_field", "<input name='_token' value='123'>")

	// 4. Register Mock `formatDate` helper slot? 
	// If the transpiler converts `formatDate(...)` to a slot call, we need that slot.
	// If it converts to a Go function call (old way), native engine can't do that naturally without a "native_eval" slot.
	// Let's assume we want to map it to a slot "formatDate".
	eng.Register("formatDate", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		// Mock implementation
		// Expect 2 args
		if len(node.Children) < 2 {
			return nil
		}
		// Resolve args (using resolveValue helper logic duplicate or manually)
		// Since we don't have access to interal resolveValue here easily without export?
		// Actually `resolveValue` is in blade_native logic, not exported?
		// But here we are in same package `slots`.
		// But wait, `resolveValue` (lowercase) is unexported in `slots` package.
		// So we CAN use it in test if test is `package slots`. Yes it is.
		
		val1 := resolveValue(node.Children[0].Value, scope)
		val2 := resolveValue(node.Children[1].Value, scope)
		
		// fmt.Printf("FormatDate Arg1: %v, Arg2: %v\n", val1, val2)
		
		w, _ := ctx.Value("httpWriter").(http.ResponseWriter)
		// Simple mock logic
		if val1 == "2024-01-01" && val2 == "Y-m-d" {
			w.Write([]byte("2024-01-01 Formatted"))
		} else {
			w.Write([]byte("Invalid Args"))
		}
		return nil
	}, engine.SlotMeta{})


	// 5. Transpile & Execute
	// We use the internal Transpiler directly or via "view.blade" manually? 
	// Let's use `transpileBladeNative` (internal) if possible, but it's not exported.
	// So we rely on `RegisterBladeSlots` and use `view.blade` with `file: ...`? 
	// Or we can mock `os.ReadFile`.
	// Easier: Create a temporary file.

	// ... File Creation skipped for brevity, let's use a "direct string" approach if possible?
	// `view.blade` accepts a file path. We need to write to disk.
	// OR we modify `view.blade` to accept `content`? No, stick to file.
	
	// Create File
	fileName := "views/test_features.blade.zl"
	os.Mkdir("views", 0755)
	os.WriteFile(fileName, []byte(viewContent), 0644)
	defer os.Remove(fileName)
	defer os.Remove("views")

	// 6. Execute
	root := &engine.Node{
		Name: "view.blade",
		Value: "test_features.blade.zl",
	}

	rec := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), "httpWriter", http.ResponseWriter(rec))

	err := eng.Execute(ctx, root, scope)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	output := strings.TrimSpace(rec.Body.String())
	// Expected behavior (if fixed):
	// START
	// <input...> (CSRF)
	// HIDE (Else block)
	// 2024-01-01 Formatted (Function call)
	// END
	
	t.Logf("Output:\n%s", output)
	
	if !strings.Contains(output, "HIDE") {
		t.Error("Expected @else block 'HIDE' to be rendered")
	}
	if strings.Contains(output, "@else") {
		t.Error("Syntactic @else leaked into output")
	}
	if strings.Contains(output, "csrf_field") && !strings.Contains(output, "<input") {
		t.Error("Expected CSRF token, got literal 'csrf_field'")
	}
	if strings.Contains(output, "formatDate") {
		t.Error("Expected function call result, got literal 'formatDate'")
	}
}
