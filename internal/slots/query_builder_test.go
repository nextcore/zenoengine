package slots

import (
	"context"
	"strings"
	"testing"
	"zeno/pkg/dbmanager"
	"zeno/pkg/engine"

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
					{Logical: "AND", Column: "age", Op: ">", Value: 18},
					{Logical: "AND", Column: "status", Op: "=", Value: "active"},
					{Logical: "OR", Column: "role", Op: "=", Value: "admin"},
				},
			},
			queryType: "SELECT",
			wantSQL:   "SELECT * FROM `users` WHERE `age` > ? AND `status` = ? OR `role` = ?",
			wantArgs:  []interface{}{18, "active", "admin"},
		},
		{
			name: "Select With Between",
			qs: QueryState{
				Table:   "users",
				Dialect: mockDialect,
				Where: []WhereCond{
					{Logical: "AND", Column: "age", Op: "BETWEEN", Value: []interface{}{18, 30}},
				},
			},
			queryType: "SELECT",
			wantSQL:   "SELECT * FROM `users` WHERE `age` BETWEEN ? AND ?",
			wantArgs:  []interface{}{18, 30},
		},
		{
			name: "Select With Not Between String List",
			qs: QueryState{
				Table:   "users",
				Dialect: mockDialect,
				Where: []WhereCond{
					{Logical: "AND", Column: "age", Op: "NOT BETWEEN", Value: "[18, 30]"},
				},
			},
			queryType: "SELECT",
			wantSQL:   "SELECT * FROM `users` WHERE `age` NOT BETWEEN ? AND ?",
			wantArgs:  []interface{}{"18", "30"},
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

func TestDBQuery(t *testing.T) {
	eng := engine.NewEngine()
	// Using sqlite to bypass empty dialect panics
	dbMgr := dbmanager.NewDBManager()
	err := dbMgr.AddConnection("default", "sqlite", ":memory:", 1, 1)
	assert.NoError(t, err)

	RegisterDBSlots(eng, dbMgr)

	tests := []struct {
		name   string
		script string
	}{
		{
			name: "Basic Block Query",
			script: `db.query: 'users' {
				where { col: 'status' val: 'active' }
				limit: 10
			}`,
		},
		{
			name: "Exists Query",
			script: `db.query: 'orders' {
				where { col: 'id' val: 5 }
				exists: { as: $is_found }
			}`,
		},
		{
			name: "Doesnt Exist Query",
			script: `db.query: 'orders' {
				where { col: 'id' val: 5 }
				doesnt_exist: { as: $is_empty }
			}`,
		},
		{
			name: "Aggregates Query",
			script: `db.query: 'sales' {
				sum: 'amount' { as: $total_sales }
				avg: 'amount' { as: $avg_sales }
				min: 'amount' { as: $min_sales }
				max: 'amount' { as: $max_sales }
			}`,
		},
		{
			name: "Pluck Query",
			script: `db.query: 'users' {
				where { col: 'status' val: 'active' }
				pluck: 'id' { as: $user_ids }
			}`,
		},
		{
			name: "Paginate Query",
			script: `db.query: 'logs' {
				where { col: 'level' val: 'error' }
				paginate { page: 2 per_page: 50 as: $results }
			}`,
		},
		{
			name: "Insert Query",
			script: `db.query: 'users' {
				insert { name: 'Alice' role: 'admin' }
			}`,
		},
		{
			name: "Update Query",
			script: `db.query: 'users' {
				where { col: 'id' val: 1 }
				update { status: 'inactive' }
			}`,
		},
		{
			name: "Delete Query",
			script: `db.query: 'users' {
				where { col: 'status' val: 'deleted' }
				delete { as: $deleted_count }
			}`,
		},
		{
			name: "Get Query",
			script: `db.query: 'users' {
				where { col: 'role' val: 'user' }
				get { as: $users }
			}`,
		},
		{
			name: "First Query",
			script: `db.query: 'users' {
				where { col: 'id' val: 1 }
				first { as: $user }
			}`,
		},
		{
			name: "Count Query",
			script: `db.query: 'users' {
				where { col: 'status' val: 'active' }
				count { as: $active_count }
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := engine.ParseString(tt.script, "test.zl")
			assert.NoError(t, err)

			scope := engine.NewScope(nil)
			err = eng.Execute(context.Background(), node, scope)
			// At minimum it should not return errors related to slot binding missing
			if err != nil && !strings.Contains(err.Error(), "no such table") && !strings.Contains(err.Error(), "no connection") {
				t.Fatalf("Unexpected execution error: %v", err)
			}
		})
	}
}
