package slots

import (
	"context"
	"fmt"
	"strings"
	"github.com/nextcore/zenoengine/pkg/dbmanager"
	"github.com/nextcore/zeno-go/pkg/engine"
	"github.com/nextcore/zeno-go/pkg/utils/coerce"
)

// Helper: Build query from Node
func buildQuery(node *engine.Node, scope *engine.Scope) (string, string, []interface{}, error) {
	query := ""
	dbName := "default"
	var args []interface{}

	// Value utama: mysql.select: "SELECT * FROM..." (atau $query_var)
	if node.Value != nil {
		val, err := MustResolveValue(node.Value, scope, coerce.ToString(node.Value))
		if err != nil {
			return "", "", nil, err
		}
		query = coerce.ToString(val)
	}

	for _, c := range node.Children {
		if c.Name == "sql" {
			query = coerce.ToString(parseNodeValue(c, scope))
		}
		if c.Name == "db" || c.Name == "connection" {
			dbName = coerce.ToString(parseNodeValue(c, scope))
		}

		// Add support for 'bind' parameter
		if c.Name == "bind" {
			// Check if bind has children (any keyname is allowed, order is preserved)
			if len(c.Children) > 0 {
				for _, child := range c.Children {
					val := parseNodeValue(child, scope)
					args = append(args, val)
				}
			}
		}
	}
	return query, dbName, args, nil
}

// replacePlaceholders replaces positional '?' with dialect-specific placeholders
func replacePlaceholders(query string, dialect dbmanager.Dialect) string {
	var result strings.Builder
	placeholderCount := 0
	inSingleQuote := false
	inDoubleQuote := false
	inBackTick := false

	for i := 0; i < len(query); i++ {
		char := query[i]
		switch char {
		case '\'':
			if !inDoubleQuote && !inBackTick {
				inSingleQuote = !inSingleQuote
			}
			result.WriteByte(char)
		case '"':
			if !inSingleQuote && !inBackTick {
				inDoubleQuote = !inDoubleQuote
			}
			result.WriteByte(char)
		case '`':
			if !inSingleQuote && !inDoubleQuote {
				inBackTick = !inBackTick
			}
			result.WriteByte(char)
		case '?':
			if !inSingleQuote && !inDoubleQuote && !inBackTick {
				placeholderCount++
				result.WriteString(dialect.Placeholder(placeholderCount))
			} else {
				result.WriteByte(char)
			}
		default:
			result.WriteByte(char)
		}
	}
	return result.String()
}

func RegisterRawDBSlots(eng *engine.Engine, dbMgr *dbmanager.DBManager) {

	// DB.SELECT (Generic)
	handlerSelect := func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		query, dbName, args, err := buildQuery(node, scope)
		if err != nil {
			return err
		}
		target := "rows"
		onlyFirst := false
		for _, c := range node.Children {
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
			if c.Name == "first" {
				onlyFirst, _ = coerce.ToBool(parseNodeValue(c, scope))
			}
		}

		if query == "" {
			return fmt.Errorf("db.select: query cannot be empty")
		}

		executor, dialect, err := getExecutor(scope, dbMgr, dbName)
		if err != nil {
			return err
		}

		// Normalize placeholders for the current dialect
		query = replacePlaceholders(query, dialect)

		rows, err := executor.QueryContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("db.select execution error: %w", err)
		}
		defer rows.Close()

		cols, err := rows.Columns()
		if err != nil {
			return fmt.Errorf("db.select columns error: %w", err)
		}

		var results []map[string]interface{}

		for rows.Next() {
			columns := make([]interface{}, len(cols))
			columnPointers := make([]interface{}, len(cols))
			for i := range columns {
				columnPointers[i] = &columns[i]
			}

			if err := rows.Scan(columnPointers...); err != nil {
				return err
			}

			m := make(map[string]interface{})
			for i, colName := range cols {
				val := columns[i]
				b, ok := val.([]byte)
				if ok {
					m[colName] = string(b)
				} else {
					m[colName] = val
				}
			}
			results = append(results, m)
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("db.select iteration error: %w", err)
		}


		if onlyFirst {
			if len(results) > 0 {
				scope.Set(target, results[0])
			} else {
				scope.Set(target, nil)
			}
		} else {
			scope.Set(target, results)
		}
		return nil
	}

	eng.Register("db.select", handlerSelect, engine.SlotMeta{
		Description: "Perform a SELECT query and retrieve multiple rows.",
		Example:     "db.select: 'SELECT * FROM users'\n  as: $users",
		ValueType:   "string",
		Inputs: map[string]engine.InputMeta{
			"as":     {Description: "Variable to store results", Required: false, Type: "string"},
			"first":  {Description: "Return only the first row as a map (Default: false)", Required: false, Type: "bool"},
			"db":     {Description: "Database connection name", Required: false, Type: "string"},
			"bind":   {Description: "Bind parameters container", Required: false, Type: "any"},
		},
	})
	eng.Register("mysql.select", handlerSelect, engine.SlotMeta{Description: "Alias for db.select"})
	eng.Register("db.query", handlerSelect, engine.SlotMeta{Description: "Alias for db.select"})

	// DB.EXECUTE (Generic)
	handlerExecute := func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		query, dbName, args, err := buildQuery(node, scope)
		if err != nil {
			return err
		}

		if query == "" {
			return fmt.Errorf("db.execute: query cannot be empty")
		}

		executor, dialect, err := getExecutor(scope, dbMgr, dbName)
		if err != nil {
			return err
		}

		// Normalize placeholders
		query = replacePlaceholders(query, dialect)

		res, err := executor.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("db.execute execution error: %w", err)
		}

		affected, _ := res.RowsAffected()
		lastId, _ := res.LastInsertId()


		scope.Set("db_affected", affected)
		scope.Set("db_last_id", lastId)

		return nil
	}

	eng.Register("db.execute", handlerExecute, engine.SlotMeta{
		Description: "Execute a raw SQL query (INSERT, UPDATE, DELETE, etc.).",
		Example:     "db.execute: 'UPDATE users SET x=1'",
		ValueType:   "string",
		Inputs: map[string]engine.InputMeta{
			"db":     {Description: "Database connection name", Required: false, Type: "string"},
			"bind":   {Description: "Bind parameters container", Required: false, Type: "any"},
		},
	})
	eng.Register("mysql.execute", handlerExecute, engine.SlotMeta{Description: "Alias for db.execute"})
}
