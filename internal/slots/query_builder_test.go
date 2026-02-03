package slots

import (
	"testing"
	"zeno/pkg/dbmanager"

	"github.com/stretchr/testify/assert"
)

// MockDialect for testing
type MockDialect struct{}

func (m MockDialect) QuoteIdentifier(name string) string {
	return "`" + name + "`"
}
func (m MockDialect) Placeholder(n int) string {
	return "?"
}
func (m MockDialect) Limit(limit, offset int) string {
	if limit > 0 {
		if offset > 0 {
			return " LIMIT ? OFFSET ?" // Simplified for string matching, actual values bound? No, Limit usually string in dialect
		}
		return " LIMIT ?"
	}
	return ""
}
func (m MockDialect) Name() string { return "mock" }

func TestQueryState_BuildSQL(t *testing.T) {
	mockDialect := dbmanager.GetDialect("mysql") // Use real mysql dialect for consistency if available, or just rely on logic

	tests := []struct {
		name      string
		qs        QueryState
		queryType string
		wantSQL   string
		wantArgs  []interface{}
	}{
		{
			name: "Basic Select",
			qs: QueryState{
				Table:   "users",
				Columns: []string{"id", "name"},
				Dialect: mockDialect,
			},
			queryType: "SELECT",
			wantSQL:   "SELECT `id`, `name` FROM `users`",
			wantArgs:  nil,
		},
		{
			name: "Select With Where",
			qs: QueryState{
				Table:   "users",
				Dialect: mockDialect,
				Where: []WhereCond{
					{Column: "age", Op: ">", Value: 18},
					{Column: "status", Op: "=", Value: "active"},
				},
			},
			queryType: "SELECT",
			wantSQL:   "SELECT * FROM `users` WHERE `age` > ? AND `status` = ?",
			wantArgs:  []interface{}{18, "active"},
		},
		{
			name: "Select With Join",
			qs: QueryState{
				Table:   "users",
				Dialect: mockDialect,
				Joins: []JoinDef{
					{Type: "INNER", Table: "orders", On: []string{"users.id", "=", "orders.user_id"}},
				},
			},
			queryType: "SELECT",
			wantSQL:   "SELECT * FROM `users` INNER JOIN `orders` ON `users`.`id` = `orders`.`user_id`",
			wantArgs:  nil,
		},
		{
			name: "Select With IN",
			qs: QueryState{
				Table:   "products",
				Dialect: mockDialect,
				Where: []WhereCond{
					{Column: "id", Op: "IN", Value: []interface{}{1, 2, 3}},
				},
			},
			queryType: "SELECT",
			wantSQL:   "SELECT * FROM `products` WHERE `id` IN (?, ?, ?)",
			wantArgs:  []interface{}{1, 2, 3},
		},
		{
			name: "Delete",
			qs: QueryState{
				Table:   "logs",
				Dialect: mockDialect,
				Where: []WhereCond{
					{Column: "created_at", Op: "<", Value: "2020-01-01"},
				},
			},
			queryType: "DELETE",
			wantSQL:   "DELETE FROM `logs` WHERE `created_at` < ?",
			wantArgs:  []interface{}{"2020-01-01"},
		},
		{
			name: "Count",
			qs: QueryState{
				Table:   "users",
				Dialect: mockDialect,
			},
			queryType: "COUNT",
			wantSQL:   "SELECT COUNT(*) FROM `users`",
			wantArgs:  nil,
		},
		{
			name: "Group By and Having",
			qs: QueryState{
				Table:   "orders",
				Columns: []string{"user_id", "COUNT(*) as total"},
				Dialect: mockDialect,
				GroupBy: []string{"user_id"},
				Having: []WhereCond{
					{Column: "total", Op: ">", Value: 5},
				},
			},
			queryType: "SELECT",
			wantSQL:   "SELECT `user_id`, COUNT(*) as total FROM `orders` GROUP BY `user_id` HAVING `total` > ?",
			wantArgs:  []interface{}{5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSQL, gotArgs := tt.qs.BuildSQL(tt.queryType)
			assert.Equal(t, tt.wantSQL, gotSQL)
			assert.Equal(t, tt.wantArgs, gotArgs)
		})
	}
}
