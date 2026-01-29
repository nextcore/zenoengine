package slots

import (
	"testing"
	"zeno/pkg/engine"

	"github.com/stretchr/testify/assert"
)

func TestBuildQuery(t *testing.T) {
	scope := engine.NewScope(nil)
	scope.Set("userId", 123)
	scope.Set("status", "active")
	scope.Set("userData", map[string]interface{}{
		"role": "admin",
	})

	tests := []struct {
		name     string
		node     *engine.Node
		expected string // Query
		args     []interface{}
	}{
		{
			name: "Standard Bind",
			node: &engine.Node{
				Value: "SELECT * FROM users",
				Children: []*engine.Node{
					{Name: "bind", Children: []*engine.Node{
						{Name: "val", Value: 1},
						{Name: "val", Value: 2},
					}},
				},
			},
			expected: "SELECT * FROM users",
			args:     []interface{}{1, 2},
		},
		{
			name: "Named Keys (Order Preserved)",
			node: &engine.Node{
				Value: "INSERT INTO users (id, name)",
				Children: []*engine.Node{
					{Name: "bind", Children: []*engine.Node{
						{Name: "id", Value: 101},
						{Name: "name", Value: "Alice"},
					}},
				},
			},
			expected: "INSERT INTO users (id, name)",
			args:     []interface{}{101, "Alice"},
		},
		{
			name: "Variable Resolution",
			node: &engine.Node{
				Value: "SELECT * FROM orders WHERE user_id = ?",
				Children: []*engine.Node{
					{Name: "bind", Children: []*engine.Node{
						{Name: "val", Value: "$userId"},
					}},
				},
			},
			expected: "SELECT * FROM orders WHERE user_id = ?",
			args:     []interface{}{123},
		},
		{
			name: "Nested Variable (Map)",
			node: &engine.Node{
				Value: "SELECT * FROM users WHERE role = ?",
				Children: []*engine.Node{
					{Name: "bind", Children: []*engine.Node{
						{Name: "val", Value: "$userData.role"},
					}},
				},
			},
			expected: "SELECT * FROM users WHERE role = ?",
			args:     []interface{}{"admin"},
		},
		{
			name: "Mixed Types",
			node: &engine.Node{
				Value: "UPDATE t SET a=?, b=?, c=?",
				Children: []*engine.Node{
					{Name: "bind", Children: []*engine.Node{
						{Name: "a", Value: 1},
						{Name: "b", Value: "string"},
						{Name: "c", Value: true},
					}},
				},
			},
			expected: "UPDATE t SET a=?, b=?, c=?",
			args:     []interface{}{1, "string", true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, _, args, err := buildQuery(tt.node, scope)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, query)
			assert.Equal(t, tt.args, args)
		})
	}
}
