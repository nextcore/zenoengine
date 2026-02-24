package slots

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"zeno/pkg/engine"
)

func TestBladeLaravelFeatures(t *testing.T) {
	eng := engine.NewEngine()
	RegisterBladeSlots(eng)
	RegisterLogicSlots(eng)
	RegisterUtilSlots(eng)

	tests := []struct {
		name     string
		template string
		scope    map[string]interface{}
		expected string
	}{
		{
			name:     "Method Put",
			template: `<html>@method('PUT')</html>`,
			expected: `<html><input type="hidden" name="_method" value="PUT"></html>`,
		},
		{
			name:     "Method Delete",
			template: `<html>@method("DELETE")</html>`,
			expected: `<html><input type="hidden" name="_method" value="DELETE"></html>`,
		},
		{
			name:     "Class array mapping",
			template: `<html><div @class(['p-4', 'font-bold' => $isActive, 'bg-red' => $hasError])></div></html>`,
			scope: map[string]interface{}{
				"isActive": true,
				"hasError": false,
			},
			expected: `<html><div class="p-4 font-bold"></div></html>`,
		},
		{
			name:     "Class simple false mapping",
			template: `<html><div @class(['p-4' => false])></div></html>`,
			expected: `<html><div ></div></html>`,
		},
		{
			name:     "PHP block execution",
			template: "<html>@php\nvar: $msg { val: \"hello from php\" }\n@endphp{{ $msg }}</html>",
			expected: `<html>hello from php</html>`,
		},
		{
			name:     "Zeno block execution",
			template: "<html>@zeno\nvar: $msg { val: \"hello from zeno\" }\n@endzeno{{ $msg }}</html>",
			expected: `<html>hello from zeno</html>`,
		},
		{
			name:     "Loop variable injection",
			template: `<ul>@foreach($items as $item)<li>{{ $loop.index }}:{{ $item }}{{ $loop.first ? " first" : "" }}{{ $loop.last ? " last" : "" }}</li>@endforeach</ul>`,
			scope: map[string]interface{}{
				"items": []interface{}{1, 2, 3},
			},
			expected: `<ul><li>0:1 first</li><li>1:2</li><li>2:3 last</li></ul>`,
		},
		{
			name:     "Error directive injection",
			template: `<div>@error('email')<span class="error">{{ $message }}</span>@enderror</div>`,
			scope: map[string]interface{}{
				"errors": map[string][]string{
					"email": {"Invalid email format"},
				},
			},
			expected: `<div><span class="error">Invalid email format</span></div>`,
		},
		{
			name:     "CSRF token injection",
			template: `<form>@csrf</form>`,
			scope: map[string]interface{}{
				"csrf_field": `<input type="hidden" name="_token" value="abc123xyz">`,
			},
			expected: `<form><input type="hidden" name="_token" value="abc123xyz"></form>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := transpileBladeNative(tt.template)
			if err != nil {
				t.Fatalf("Failed to transpile: %v", err)
			}

			scope := engine.NewScope(nil)
			if tt.scope != nil {
				for k, v := range tt.scope {
					scope.Set(k, v)
				}
			}

			rec := httptest.NewRecorder()
			ctx := context.WithValue(context.Background(), "httpWriter", http.ResponseWriter(rec))

			err = eng.Execute(ctx, node, scope)
			if err != nil {
				t.Fatalf("Failed to execute: %v", err)
			}

			result := rec.Body.String()
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected result to contain %q, but got %q", tt.expected, result)
			}
		})
	}
}
