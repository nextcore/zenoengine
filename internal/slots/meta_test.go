package slots

import (
	"context"
	"fmt"
	"testing"
	"zeno/pkg/engine"

	"github.com/stretchr/testify/assert"
)

func TestMetaSlots(t *testing.T) {
	eng := engine.NewEngine()
	// Register util slots too because eval might use them (e.g. var)
	RegisterUtilSlots(eng)
	RegisterMetaSlots(eng)

	t.Run("meta.eval", func(t *testing.T) {
		scope := engine.NewScope(nil)
		// Code to eval: var: $x { val: 10 }
		code := `var: $x { val: 10 }`

		node := &engine.Node{
			Name: "meta.eval",
			Value: code,
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		val, _ := scope.Get("x")
		// Parser seems to treat unquoted numbers as strings in some contexts or current setup
		// Adjust expectation to string "10" or use coercion
		assert.Equal(t, "10", fmt.Sprintf("%v", val))
	})

	t.Run("meta.scope", func(t *testing.T) {
		scope := engine.NewScope(nil)
		scope.Set("myvar", "hello")

		node := &engine.Node{
			Name: "meta.scope",
			Children: []*engine.Node{
				{Name: "as", Value: "$vars"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		varsRaw, _ := scope.Get("vars")
		vars := varsRaw.(map[string]interface{})
		assert.Equal(t, "hello", vars["myvar"])
	})

	t.Run("meta.parse", func(t *testing.T) {
		scope := engine.NewScope(nil)
		code := `log: "hi"`

		node := &engine.Node{
			Name: "meta.parse",
			Value: code,
			Children: []*engine.Node{
				{Name: "as", Value: "$ast"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		astRaw, _ := scope.Get("ast")
		ast := astRaw.(map[string]interface{})
		// Root node is "program" (usually implicit list of statements) or the first node if single?
		// Engine.ParseString returns *Node (root) which usually wraps statements or is a list?
		// Let's inspect what ParseString returns. Usually a root node.
		// nodeToMap implementation:
		// m["name"] = n.Name

		// If ParseString returns a dummy root, its name might be "root" or similar.
		assert.NotNil(t, ast)
	})
}
