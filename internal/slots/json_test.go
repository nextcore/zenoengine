package slots

import (
	"context"
	"testing"
	"zeno/pkg/engine"

	"github.com/stretchr/testify/assert"
)

func TestJSONSlots(t *testing.T) {
	eng := engine.NewEngine()
	RegisterJSONSlots(eng)

	t.Run("json.parse success", func(t *testing.T) {
		scope := engine.NewScope(nil)
		jsonStr := `{"name": "Zeno", "version": 1.0}`
		scope.Set("input", jsonStr)

		node := &engine.Node{
			Name: "json.parse",
			Value: "$input",
			Children: []*engine.Node{
				{Name: "as", Value: "$output"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		output, ok := scope.Get("output")
		assert.True(t, ok)

		outputMap, ok := output.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "Zeno", outputMap["name"])
		assert.Equal(t, 1.0, outputMap["version"])
	})

	t.Run("json.parse failure", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "json.parse",
			Value: "{invalid-json}",
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid json")
	})

	t.Run("json.stringify success", func(t *testing.T) {
		scope := engine.NewScope(nil)
		data := map[string]interface{}{
			"active": true,
			"roles":  []interface{}{"admin", "user"},
		}
		scope.Set("data", data)

		node := &engine.Node{
			Name: "json.stringify",
			Value: "$data",
			Children: []*engine.Node{
				{Name: "as", Value: "$str"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		str, ok := scope.Get("str")
		assert.True(t, ok)
		assert.IsType(t, "", str)
		assert.Contains(t, str.(string), `"active":true`)
		assert.Contains(t, str.(string), `"admin"`)
	})
}
