package slots

import (
	"context"
	"testing"
	"zeno/pkg/engine"

	"github.com/stretchr/testify/assert"
)

func TestCacheSlots(t *testing.T) {
	eng := engine.NewEngine()
	RegisterCacheSlots(eng, nil)

	t.Run("cache.put no-op", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "cache.put",
			Children: []*engine.Node{
				{Name: "key", Value: "foo"},
				{Name: "val", Value: "bar"},
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
				{Name: "key", Value: "foo"},
				{Name: "default", Value: "default_val"},
				{Name: "as", Value: "$res"},
			},
		}
		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		res, _ := scope.Get("res")
		assert.Equal(t, "default_val", res)
	})
}
