package slots

import (
	"context"
	"testing"
	"zeno/pkg/engine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestSlots(t *testing.T) {
	eng := engine.NewEngine()
	RegisterTestSlots(eng)

	t.Run("assert.eq success", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "assert.eq",
			Value: 10,
			Children: []*engine.Node{
				{Name: "expected", Value: 10},
			},
		}
		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)
	})

	t.Run("assert.eq failure", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "assert.eq",
			Value: 10,
			Children: []*engine.Node{
				{Name: "expected", Value: 20},
			},
		}
		err := eng.Execute(context.Background(), node, scope)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected 20, got 10")
	})

	t.Run("test slot aggregates stats", func(t *testing.T) {
		scope := engine.NewScope(nil)
		stats := &TestStats{}
		ctx := WithTestStats(context.Background(), stats)

		// Test Case 1: Passing
		nodePass := &engine.Node{
			Name: "test",
			Value: "Passing Test",
			Children: []*engine.Node{
				{Name: "assert.eq", Value: 1, Children: []*engine.Node{{Name: "expected", Value: 1}}},
			},
		}

		// Test Case 2: Failing
		nodeFail := &engine.Node{
			Name: "test",
			Value: "Failing Test",
			Children: []*engine.Node{
				{Name: "assert.eq", Value: 1, Children: []*engine.Node{{Name: "expected", Value: 2}}},
			},
		}

		err := eng.Execute(ctx, nodePass, scope)
		require.NoError(t, err)

		err = eng.Execute(ctx, nodeFail, scope)
		require.NoError(t, err) // 'test' slot returns nil even on failure, but updates stats

		assert.Equal(t, 2, stats.Total)
		assert.Equal(t, 1, stats.Passed)
		assert.Equal(t, 1, stats.Failed)
		require.Len(t, stats.Errors, 1)
		assert.Contains(t, stats.Errors[0], "FAIL [Failing Test]")
	})
}
