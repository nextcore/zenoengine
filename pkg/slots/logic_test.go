package slots

import (
	"context"
	"testing"
	"zeno/pkg/engine"
)

func TestLogicSlots(t *testing.T) {
	eng := engine.NewEngine()
	RegisterLogicSlots(eng)

	t.Run("scope.set", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name:  "scope.set",
			Value: "$test_var",
			Children: []*engine.Node{
				{Name: "val", Value: 123},
			},
		}
		err := eng.Execute(context.Background(), node, scope)
		if err != nil {
			t.Fatalf("scope.set failed: %v", err)
		}
		val, _ := scope.Get("test_var")
		if val != 123 {
			t.Errorf("Expected 123, got %v", val)
		}
	})

	t.Run("logic.compare", func(t *testing.T) {
		scope := engine.NewScope(nil)
		scope.Set("age", 20)
		node := &engine.Node{
			Name: "logic.compare",
			Children: []*engine.Node{
				{Name: "v1", Value: "$age"},
				{Name: "op", Value: ">"},
				{Name: "v2", Value: 18},
				{Name: "as", Value: "is_adult"},
			},
		}
		err := eng.Execute(context.Background(), node, scope)
		if err != nil {
			t.Fatalf("logic.compare failed: %v", err)
		}
		val, _ := scope.Get("is_adult")
		if val != true {
			t.Errorf("Expected true, got %v", val)
		}
	})

	t.Run("try_catch", func(t *testing.T) {
		scope := engine.NewScope(nil)
		// Register a slot that fails
		eng.Register("fail_slot", func(ctx context.Context, n *engine.Node, s *engine.Scope) error {
			return engine.Diagnostic{Message: "intended failure"}
		}, engine.SlotMeta{})

		node := &engine.Node{
			Name: "try",
			Children: []*engine.Node{
				{
					Name: "do",
					Children: []*engine.Node{
						{Name: "fail_slot"},
					},
				},
				{
					Name: "catch",
					Children: []*engine.Node{
						{
							Name:  "scope.set",
							Value: "$caught",
							Children: []*engine.Node{
								{Name: "val", Value: true},
							},
						},
					},
				},
			},
		}
		err := eng.Execute(context.Background(), node, scope)
		if err != nil {
			t.Fatalf("try block returned error despite catch: %v", err)
		}
		val, _ := scope.Get("caught")
		if val != true {
			t.Errorf("Expected catch block to execute")
		}
	})

	t.Run("for_loop", func(t *testing.T) {
		scope := engine.NewScope(nil)
		scope.Set("items", []interface{}{"a", "b", "c"})
		node := &engine.Node{
			Name:  "for",
			Value: "$items",
			Children: []*engine.Node{
				{Name: "as", Value: "$item"},
				{
					Name: "do",
					Children: []*engine.Node{
						{
							Name:  "scope.set",
							Value: "$last_item",
							Children: []*engine.Node{
								{Name: "val", Value: "$item"},
							},
						},
					},
				},
			},
		}
		err := eng.Execute(context.Background(), node, scope)
		if err != nil {
			t.Fatalf("for loop failed: %v", err)
		}
		val, _ := scope.Get("last_item")
		if val != "c" {
			t.Errorf("Expected 'c', got %v", val)
		}
	})

	t.Run("switch_case", func(t *testing.T) {
		scope := engine.NewScope(nil)
		scope.Set("status", "active")
		node := &engine.Node{
			Name:  "switch",
			Value: "$status",
			Children: []*engine.Node{
				{
					Name:  "case",
					Value: "inactive",
					Children: []*engine.Node{
						{Name: "scope.set", Value: "$res", Children: []*engine.Node{{Name: "val", Value: 1}}},
					},
				},
				{
					Name:  "case",
					Value: "active",
					Children: []*engine.Node{
						{Name: "scope.set", Value: "$res", Children: []*engine.Node{{Name: "val", Value: 2}}},
					},
				},
				{
					Name: "default",
					Children: []*engine.Node{
						{Name: "scope.set", Value: "$res", Children: []*engine.Node{{Name: "val", Value: 3}}},
					},
				},
			},
		}
		err := eng.Execute(context.Background(), node, scope)
		if err != nil {
			t.Fatalf("switch failed: %v", err)
		}
		val, _ := scope.Get("res")
		if val != 2 {
			t.Errorf("Expected 2, got %v", val)
		}
	})
}
