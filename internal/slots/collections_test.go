package slots

import (
	"context"
	"testing"
	"zeno/pkg/engine"
)

func TestCollectionSlots(t *testing.T) {
	eng := engine.NewEngine()
	RegisterCollectionSlots(eng)

	t.Run("array.push", func(t *testing.T) {
		scope := engine.NewScope(nil)
		scope.Set("my_list", []interface{}{"a"})
		node := &engine.Node{
			Name:  "array.push",
			Value: "$my_list",
			Children: []*engine.Node{
				{Name: "val", Value: "b"},
			},
		}
		err := eng.Execute(context.Background(), node, scope)
		if err != nil {
			t.Fatalf("array.push failed: %v", err)
		}
		val, _ := scope.Get("my_list")
		list := val.([]interface{})
		if len(list) != 2 || list[1] != "b" {
			t.Errorf("Expected [a, b], got %v", val)
		}
	})

	t.Run("collections.get", func(t *testing.T) {
		scope := engine.NewScope(nil)
		scope.Set("list", []interface{}{"x", "y", "z"})
		node := &engine.Node{
			Name:  "collections.get",
			Value: "$list",
			Children: []*engine.Node{
				{Name: "index", Value: 1},
				{Name: "as", Value: "$item"},
			},
		}
		err := eng.Execute(context.Background(), node, scope)
		if err != nil {
			t.Fatalf("collections.get failed: %v", err)
		}
		val, _ := scope.Get("item")
		if val != "y" {
			t.Errorf("Expected 'y', got %v", val)
		}
	})

	t.Run("map.set", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name:  "map.set",
			Value: "$user",
			Children: []*engine.Node{
				{Name: "name", Value: "Budi"},
				{Name: "age", Value: 25},
			},
		}
		err := eng.Execute(context.Background(), node, scope)
		if err != nil {
			t.Fatalf("map.set failed: %v", err)
		}
		val, _ := scope.Get("user")
		m := val.(map[string]interface{})
		if m["name"] != "Budi" || m["age"] != 25 {
			t.Errorf("Expected name=Budi, age=25, got %v", m)
		}
	})

	t.Run("len", func(t *testing.T) {
		scope := engine.NewScope(nil)
		scope.Set("my_str", "hello")
		node := &engine.Node{
			Name:  "len",
			Value: "$my_str",
			Children: []*engine.Node{
				{Name: "as", Value: "$count"},
			},
		}
		err := eng.Execute(context.Background(), node, scope)
		if err != nil {
			t.Fatalf("len failed: %v", err)
		}
		val, _ := scope.Get("count")
		if val != int64(5) {
			t.Errorf("Expected 5, got %v", val)
		}
	})
}
