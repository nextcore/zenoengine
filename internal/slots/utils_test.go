package slots

import (
	"context"
	"os"
	"testing"
	"zeno/pkg/engine"

	"github.com/stretchr/testify/assert"
)

func TestUtilsSlots(t *testing.T) {
	eng := engine.NewEngine()
	RegisterUtilSlots(eng)

	t.Run("strings.concat", func(t *testing.T) {
		scope := engine.NewScope(nil)
		scope.Set("name", "World")

		node := &engine.Node{
			Name: "strings.concat",
			Value: "Hello ",
			Children: []*engine.Node{
				{Name: "val", Value: "$name"},
				{Name: "as", Value: "$result"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		res, _ := scope.Get("result")
		assert.Equal(t, "Hello World", res)
	})

	t.Run("string.replace", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "string.replace",
			Value: "foobarfoo",
			Children: []*engine.Node{
				{Name: "find", Value: "foo"},
				{Name: "replace", Value: "baz"},
				{Name: "as", Value: "$result"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		res, _ := scope.Get("result")
		assert.Equal(t, "bazbarbaz", res)
	})

	t.Run("text.slugify", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "text.slugify",
			Value: "Hello World 123",
			Children: []*engine.Node{
				{Name: "as", Value: "$slug"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		res, _ := scope.Get("slug")
		assert.Equal(t, "hello-world-123", res)
	})

	t.Run("system.env", func(t *testing.T) {
		os.Setenv("TEST_ENV_VAR", "test_value")
		defer os.Unsetenv("TEST_ENV_VAR")

		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "system.env",
			Value: "TEST_ENV_VAR",
			Children: []*engine.Node{
				{Name: "as", Value: "$val"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		res, _ := scope.Get("val")
		assert.Equal(t, "test_value", res)
	})

	t.Run("if condition true", func(t *testing.T) {
		scope := engine.NewScope(nil)
		scope.Set("age", 20)

		// Create a mock slot to verify execution
		executed := false
		eng.Register("mock.exec", func(ctx context.Context, n *engine.Node, s *engine.Scope) error {
			executed = true
			return nil
		}, engine.SlotMeta{})

		node := &engine.Node{
			Name: "if",
			Value: "$age > 18",
			Children: []*engine.Node{
				{Name: "then", Children: []*engine.Node{
					{Name: "mock.exec"},
				}},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)
		assert.True(t, executed)
	})

	t.Run("arrays.length", func(t *testing.T) {
		scope := engine.NewScope(nil)
		scope.Set("list", []interface{}{1, 2, 3})

		node := &engine.Node{
			Name: "arrays.length",
			Value: "$list",
			Children: []*engine.Node{
				{Name: "as", Value: "$len"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		res, _ := scope.Get("len")
		assert.Equal(t, 3, res)
	})

	t.Run("var", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "var",
			Value: "$myvar",
			Children: []*engine.Node{
				{Name: "val", Value: "hello"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		res, _ := scope.Get("myvar")
		assert.Equal(t, "hello", res)
	})

	t.Run("coalesce", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "coalesce",
			Value: nil, // null input
			Children: []*engine.Node{
				{Name: "default", Value: "default_val"},
				{Name: "as", Value: "$res"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		res, _ := scope.Get("res")
		assert.Equal(t, "default_val", res)
	})

	t.Run("cast.to_int", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "cast.to_int",
			Value: "123",
			Children: []*engine.Node{
				{Name: "as", Value: "$num"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		res, _ := scope.Get("num")
		assert.Equal(t, 123, res)
	})
}
