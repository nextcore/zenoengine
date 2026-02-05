package slots

import (
	"context"
	"testing"
	"zeno/pkg/engine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheSlots(t *testing.T) {
	eng := engine.NewEngine()
	// Pass nil for rdb since it's disabled/removed
	RegisterCacheSlots(eng, nil)

	t.Run("cache.put is no-op", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "cache.put",
			Children: []*engine.Node{
				{Name: "key", Value: "test_key"},
				{Name: "val", Value: "test_val"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)
	})

	t.Run("cache.get returns default", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "cache.get",
			Children: []*engine.Node{
				{Name: "key", Value: "test_key"},
				{Name: "default", Value: "default_value"},
				{Name: "as", Value: "$res"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		require.NoError(t, err)

		val, ok := scope.Get("res")
		assert.True(t, ok)
		assert.Equal(t, "default_value", val)
	})

	t.Run("cache.forget is no-op", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "cache.forget",
			Value: "test_key",
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)
	})
}
