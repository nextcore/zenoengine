package slots

import (
	"context"
	"testing"
	"zeno/pkg/engine"
)

func TestFunctionSlots(t *testing.T) {
	eng := engine.NewEngine()
	RegisterFunctionSlots(eng)
	RegisterLogicSlots(eng)

	t.Run("fn_and_call", func(t *testing.T) {
		scope := engine.NewScope(nil)

		// Define function
		fnNode := &engine.Node{
			Name:  "fn",
			Value: "my_func",
			Children: []*engine.Node{
				{
					Name:  "scope.set",
					Value: "$result",
					Children: []*engine.Node{
						{Name: "val", Value: "hello from func"},
					},
				},
			},
		}
		if err := eng.Execute(context.Background(), fnNode, scope); err != nil {
			t.Fatalf("fn define failed: %v", err)
		}

		// Call function
		callNode := &engine.Node{
			Name:  "call",
			Value: "my_func",
		}
		if err := eng.Execute(context.Background(), callNode, scope); err != nil {
			t.Fatalf("call failed: %v", err)
		}

		val, _ := scope.Get("result")
		if val != "hello from func" {
			t.Errorf("Expected 'hello from func', got %v", val)
		}
	})
}
