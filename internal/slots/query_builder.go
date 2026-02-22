package slots

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"zeno/pkg/dbmanager"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

// SQLExecutor interface
type SQLExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

func getExecutor(scope *engine.Scope, dbMgr *dbmanager.DBManager, dbName string) (SQLExecutor, dbmanager.Dialect, error) {
	if val, ok := scope.Get("_active_tx"); ok && val != nil {
		if tx, ok := val.(*sql.Tx); ok {
			// [IMPORTANT] Transaction also needs dialect.
			// For now, we assume it's the default database dialect if not specified.
			return tx, dbMgr.GetDialect(dbName), nil
		}
	}
	db := dbMgr.GetConnection(dbName)
	dialect := dbMgr.GetDialect(dbName)
	if db == nil {
		return nil, nil, fmt.Errorf("database connection '%s' not found", dbName)
	}
	return db, dialect, nil
}

type WhereCond struct {
	Column string
	Op     string
	Value  interface{}
}

type JoinDef struct {
	Type  string // "INNER", "LEFT", "RIGHT"
	Table string
	On    []string // ["t1.col", "=", "t2.col"]
}

type QueryState struct {
	Table   string
	Columns []string
	Joins   []JoinDef
	Where   []WhereCond
	GroupBy []string
	Having  []WhereCond
	Args    []interface{}
	Limit   int
	Offset  int
	OrderBy string
	DBName  string
	Dialect dbmanager.Dialect
}

func (qs *QueryState) Quote(name string) string {
	if strings.Contains(name, " ") || strings.Contains(name, "(") {
		return name
	}
	if strings.Contains(name, ".") {
		parts := strings.Split(name, ".")
		for i, p := range parts {
			parts[i] = qs.Dialect.QuoteIdentifier(p)
		}
		return strings.Join(parts, ".")
	}
	return qs.Dialect.QuoteIdentifier(name)
}

func (qs *QueryState) BuildSQL(queryType string) (string, []interface{}) {
	var sb strings.Builder
	var args []interface{}

	// 1. SELECT
	if queryType == "SELECT" {
		sb.WriteString("SELECT ")
		if len(qs.Columns) > 0 {
			quotedCols := make([]string, len(qs.Columns))
			for i, c := range qs.Columns {
				quotedCols[i] = qs.Quote(c)
			}
			sb.WriteString(strings.Join(quotedCols, ", "))
		} else {
			sb.WriteString("*")
		}
	} else if queryType == "COUNT" {
		sb.WriteString("SELECT COUNT(*)")
	} else if queryType == "DELETE" {
		sb.WriteString("DELETE")
	}

	// 2. FROM
	sb.WriteString(" FROM ")
	sb.WriteString(qs.Dialect.QuoteIdentifier(qs.Table))

	// 3. JOINS
	for _, join := range qs.Joins {
		sb.WriteString(fmt.Sprintf(" %s JOIN %s ON %s %s %s",
			join.Type,
			qs.Quote(join.Table),
			qs.Quote(join.On[0]),
			join.On[1],
			qs.Quote(join.On[2]),
		))
	}

	// 4. WHERE
	if len(qs.Where) > 0 {
		sb.WriteString(" WHERE ")
		for i, cond := range qs.Where {
			if i > 0 {
				sb.WriteString(" AND ")
			}
			// Handle IN / NOT IN
			if strings.ToUpper(cond.Op) == "IN" || strings.ToUpper(cond.Op) == "NOT IN" {
				// Expect Value to be slice
				v := reflect.ValueOf(cond.Value)
				var slice []interface{}
				if v.Kind() == reflect.Slice {
					for k := 0; k < v.Len(); k++ {
						slice = append(slice, v.Index(k).Interface())
					}
				} else if str, ok := cond.Value.(string); ok && strings.HasPrefix(strings.TrimSpace(str), "[") {
					content := strings.TrimSpace(str)
					content = strings.TrimPrefix(content, "[")
					content = strings.TrimSuffix(content, "]")
					parts := strings.Split(content, ",")
					for _, p := range parts {
						slice = append(slice, strings.TrimSpace(p))
					}
				} else {
					// Fallback if single value
					slice = []interface{}{cond.Value}
				}

				placeholders := make([]string, len(slice))
				for j := range slice {
					placeholders[j] = qs.Dialect.Placeholder(len(args) + 1)
					args = append(args, slice[j])
				}
				sb.WriteString(fmt.Sprintf("%s %s (%s)",
					qs.Quote(cond.Column),
					cond.Op,
					strings.Join(placeholders, ", "),
				))
			} else if strings.ToUpper(cond.Op) == "NULL" {
				sb.WriteString(fmt.Sprintf("%s IS NULL", qs.Quote(cond.Column)))
			} else if strings.ToUpper(cond.Op) == "NOT NULL" {
				sb.WriteString(fmt.Sprintf("%s IS NOT NULL", qs.Quote(cond.Column)))
			} else {
				sb.WriteString(fmt.Sprintf("%s %s %s",
					qs.Quote(cond.Column),
					cond.Op,
					qs.Dialect.Placeholder(len(args)+1)))
				args = append(args, cond.Value)
			}
		}
	}

	// 5. GROUP BY
	if len(qs.GroupBy) > 0 {
		sb.WriteString(" GROUP BY ")
		quotedGB := make([]string, len(qs.GroupBy))
		for i, c := range qs.GroupBy {
			quotedGB[i] = qs.Quote(c)
		}
		sb.WriteString(strings.Join(quotedGB, ", "))
	}

	// 6. HAVING
	if len(qs.Having) > 0 {
		sb.WriteString(" HAVING ")
		for i, cond := range qs.Having {
			if i > 0 {
				sb.WriteString(" AND ")
			}
			sb.WriteString(fmt.Sprintf("%s %s %s",
				qs.Quote(cond.Column),
				cond.Op,
				qs.Dialect.Placeholder(len(args)+1)))
			args = append(args, cond.Value)
		}
	}

	// 7. ORDER BY
	if qs.OrderBy != "" {
		sb.WriteString(" ORDER BY " + qs.OrderBy)
	}

	// 8. LIMIT / OFFSET (Handled by Dialect, but appended here for non-delete)
	if queryType != "DELETE" && queryType != "COUNT" { // COUNT typically ignores limit
		sb.WriteString(qs.Dialect.Limit(qs.Limit, qs.Offset))
	}

	return sb.String(), args
}

func RegisterDBSlots(eng *engine.Engine, dbMgr *dbmanager.DBManager) {

	// DB.TABLE
	eng.Register("db.table", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		// Main value: db.table: "users" (or $tablename)
		tableName := coerce.ToString(resolveValue(node.Value, scope))
		dbName := "default"

		for _, c := range node.Children {
			if c.Name == "name" {
				tableName = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "db" {
				dbName = coerce.ToString(parseNodeValue(c, scope))
			}
		}
		dialect := dbMgr.GetDialect(dbName)
		scope.Set("_query_state", &QueryState{
			Table:   tableName,
			DBName:  dbName,
			Dialect: dialect,
		})

		// [NESTED QUERY SUPPORT]
		// Iterate children to execute nested builders/commands
		// e.g. db.table: "users" { where: {...}, limit: 10, get: $users }
		for _, c := range node.Children {
			if c.Name == "name" || c.Name == "db" {
				continue // Skip config keys
			}

			// 1. Map simple keys to query slots
			slotName := ""
			switch c.Name {
			case "where":
				slotName = "db.where"
			case "where_in":
				slotName = "db.where_in"
			case "where_not_in":
				slotName = "db.where_not_in"
			case "where_null":
				slotName = "db.where_null"
			case "where_not_null":
				slotName = "db.where_not_null"
			case "or_where": // TODO: Add db.or_where slot
				slotName = "db.or_where"
			case "join":
				slotName = "db.join"
			case "left_join":
				slotName = "db.left_join"
			case "right_join":
				slotName = "db.right_join"
			case "order_by", "sort":
				slotName = "db.order_by"
			case "group_by":
				slotName = "db.group_by"
			case "having":
				slotName = "db.having"
			case "limit", "take":
				slotName = "db.limit"
			case "offset", "skip":
				slotName = "db.offset"
			case "select", "columns":
				slotName = "db.columns"
			// Execution Commands
			case "get", "all":
				slotName = "db.get"
			case "first", "find":
				slotName = "db.first"
			case "count":
				slotName = "db.count"
			case "insert", "create":
				slotName = "db.insert"
			case "update":
				slotName = "db.update"
			case "delete", "destroy":
				slotName = "db.delete"
			}

			if slotName != "" {
				// Create a virtual node for the slot
				// We need to preserve the children/value structure
				// Construct new node to execute
				childNode := &engine.Node{
					Name:     slotName,
					Value:    c.Value,
					Children: c.Children,
					Filename: c.Filename,
					Line:     c.Line,
					Col:      c.Col,
				}

				// If the child is a simple key-value like "limit: 10",
				// the value is in c.Value. The slot handler usually expects value or children.
				// Most handlers (db.limit, db.order_by) use node.Value.
				// Handlers like db.where usually inspect children (col, val) but we want to support shorthand too.

				if err := eng.Execute(ctx, childNode, scope); err != nil {
					return err
				}
			}
		}

		return nil
	}, engine.SlotMeta{
		Description: "Start a query builder chain for a table.",
		Group:       "Database",
		Example:     "db.table: 'users'",
		Inputs: map[string]engine.InputMeta{
			"name":           {Description: "Table name (Optional if specified in main value)", Required: false},
			"db":             {Description: "Database connection name (Default: 'default')", Required: false},
			"where":          {Description: "Filter conditions (Nested)", Required: false},
			"where_in":       {Description: "Filter conditions (IN)", Required: false},
			"where_not_in":   {Description: "Filter conditions (NOT IN)", Required: false},
			"where_null":     {Description: "Filter conditions (NULL)", Required: false},
			"where_not_null": {Description: "Filter conditions (NOT NULL)", Required: false},
			"or_where":       {Description: "Filter conditions (OR)", Required: false},
			"join":           {Description: "Join another table", Required: false},
			"left_join":      {Description: "Left Join another table", Required: false},
			"right_join":     {Description: "Right Join another table", Required: false},
			"order_by":       {Description: "Sort results", Required: false},
			"group_by":       {Description: "Group results", Required: false},
			"having":         {Description: "Filter groups", Required: false},
			"limit":          {Description: "Limit results", Required: false},
			"offset":         {Description: "Offset results", Required: false},
			"select":         {Description: "Select specific columns", Required: false},
			"columns":        {Description: "Select specific columns (Alias)", Required: false},
			"get":            {Description: "Execute SELECT query", Required: false},
			"first":          {Description: "Execute SELECT query (Single row)", Required: false},
			"count":          {Description: "Execute COUNT query", Required: false},
			"insert":         {Description: "Execute INSERT query", Required: false},
			"update":         {Description: "Execute UPDATE query", Required: false},
			"delete":         {Description: "Execute DELETE query", Required: false},
		},
	})

	// DB.COLUMNS (Select specific columns)
	eng.Register("db.columns", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		qsVal, ok := scope.Get("_query_state")
		if !ok {
			return nil
		}
		qs := qsVal.(*QueryState)

		var cols []string
		// Check explicit columns list passed as value
		if node.Value != nil {
			val := resolveValue(node.Value, scope)
			v := reflect.ValueOf(val)
			if v.Kind() == reflect.Slice {
				for i := 0; i < v.Len(); i++ {
					cols = append(cols, coerce.ToString(v.Index(i).Interface()))
				}
			} else if str, ok := val.(string); ok && strings.HasPrefix(strings.TrimSpace(str), "[") {
				// Fallback: Parse string representation "[ a, b ]"
				content := strings.TrimSpace(str)
				content = strings.TrimPrefix(content, "[")
				content = strings.TrimSuffix(content, "]")
				parts := strings.Split(content, ",")
				for _, p := range parts {
					cols = append(cols, strings.TrimSpace(p))
				}
			} else {
				cols = append(cols, coerce.ToString(val))
			}
		}
		// Also children
		for _, c := range node.Children {
			cols = append(cols, coerce.ToString(parseNodeValue(c, scope)))
		}
		qs.Columns = cols
		return nil
	}, engine.SlotMeta{
		Example: "db.columns: ['id', 'name']",
		Group:   "Database",
	})

	// DB.JOIN / LEFT_JOIN
	joinHandler := func(joinType string) func(context.Context, *engine.Node, *engine.Scope) error {
		return func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
			qsVal, ok := scope.Get("_query_state")
			if !ok {
				return fmt.Errorf("db.join called without db.table")
			}
			qs := qsVal.(*QueryState)

			table := ""
			var on []string

			for _, c := range node.Children {
				if c.Name == "table" {
					table = coerce.ToString(parseNodeValue(c, scope))
				}
				if c.Name == "on" {
					val := parseNodeValue(c, scope)
					// Universal list handling
					var parts []string
					v := reflect.ValueOf(val)
					if v.Kind() == reflect.Slice {
						for k := 0; k < v.Len(); k++ {
							parts = append(parts, coerce.ToString(v.Index(k).Interface()))
						}
					} else if str, ok := val.(string); ok && strings.HasPrefix(strings.TrimSpace(str), "[") {
						content := strings.TrimSpace(str)
						content = strings.TrimPrefix(content, "[")
						content = strings.TrimSuffix(content, "]")
						rawParts := strings.Split(content, ",")
						for _, p := range rawParts {
							parts = append(parts, strings.TrimSpace(p))
						}
					}

					if len(parts) == 3 {
						on = parts
					}
				}
			}

			if table != "" && len(on) == 3 {
				qs.Joins = append(qs.Joins, JoinDef{Type: joinType, Table: table, On: on})
			}
			return nil
		}
	}
	eng.Register("db.join", joinHandler("INNER"), engine.SlotMeta{
		Example: "db.join {\n  table: posts\n  on: ['users.id', '=', 'posts.user_id']\n}",
		Group:   "Database",
	})
	eng.Register("db.left_join", joinHandler("LEFT"), engine.SlotMeta{
		Example: "db.left_join ...",
		Group:   "Database",
	})
	eng.Register("db.right_join", joinHandler("RIGHT"), engine.SlotMeta{
		Group: "Database",
	})

	// DB.WHERE_IN
	eng.Register("db.where_in", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		qsVal, ok := scope.Get("_query_state")
		if !ok {
			return nil
		}
		qs := qsVal.(*QueryState)

		col := ""
		var val interface{}

		for _, c := range node.Children {
			if c.Name == "col" {
				col = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "val" {
				val = parseNodeValue(c, scope)
			}
		}

		if col != "" && val != nil {
			qs.Where = append(qs.Where, WhereCond{Column: col, Op: "IN", Value: val})
		}
		return nil
	}, engine.SlotMeta{
		Example: "db.where_in {\n  col: id\n  val: [1, 2, 3]\n}",
		Group:   "Database",
	})

	// DB.WHERE_NULL
	eng.Register("db.where_null", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		qsVal, ok := scope.Get("_query_state")
		if !ok {
			return nil
		}
		qs := qsVal.(*QueryState)
		col := coerce.ToString(resolveValue(node.Value, scope))
		if col != "" {
			qs.Where = append(qs.Where, WhereCond{Column: col, Op: "NULL", Value: nil})
		}
		return nil
	}, engine.SlotMeta{
		Example: "db.where_null: deleted_at",
		Group:   "Database",
	})

	// DB.WHERE_NOT_NULL
	eng.Register("db.where_not_null", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		qsVal, ok := scope.Get("_query_state")
		if !ok {
			return nil
		}
		qs := qsVal.(*QueryState)
		col := coerce.ToString(resolveValue(node.Value, scope))
		if col != "" {
			qs.Where = append(qs.Where, WhereCond{Column: col, Op: "NOT NULL", Value: nil})
		}
		return nil
	}, engine.SlotMeta{
		Example: "db.where_not_null: created_at",
		Group:   "Database",
	})

	// DB.WHERE_NOT_IN
	eng.Register("db.where_not_in", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		qsVal, ok := scope.Get("_query_state")
		if !ok {
			return nil
		}
		qs := qsVal.(*QueryState)

		col := ""
		var val interface{}

		for _, c := range node.Children {
			if c.Name == "col" {
				col = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "val" {
				val = parseNodeValue(c, scope)
			}
		}

		if col != "" && val != nil {
			qs.Where = append(qs.Where, WhereCond{Column: col, Op: "NOT IN", Value: val})
		}
		return nil
	}, engine.SlotMeta{
		Example: "db.where_not_in {\n  col: status\n  val: ['archived', 'deleted']\n}",
		Group:   "Database",
	})

	// DB.GROUP_BY
	eng.Register("db.group_by", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		qsVal, ok := scope.Get("_query_state")
		if !ok {
			return nil
		}
		qs := qsVal.(*QueryState)

		if node.Value != nil {
			qs.GroupBy = append(qs.GroupBy, coerce.ToString(resolveValue(node.Value, scope)))
		}
		return nil
	}, engine.SlotMeta{
		Example: "db.group_by: 'status'",
		Group:   "Database",
	})

	// DB.HAVING
	eng.Register("db.having", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		qsVal, ok := scope.Get("_query_state")
		if !ok {
			return nil
		}
		qs := qsVal.(*QueryState)

		col, op := "", ">"
		var val interface{}

		for _, c := range node.Children {
			if c.Name == "col" {
				col = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "op" {
				op = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "val" {
				val = parseNodeValue(c, scope)
			}
		}
		if col != "" {
			qs.Having = append(qs.Having, WhereCond{Column: col, Op: op, Value: val})
		}
		return nil
	}, engine.SlotMeta{
		Example: "db.having {\n  col: count\n  op: '>'\n  val: 5\n}",
		Group:   "Database",
	})

	// DB.WHERE
	eng.Register("db.where", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		qsVal, ok := scope.Get("_query_state")
		if !ok {
			return fmt.Errorf("db.where called without db.table")
		}
		qs := qsVal.(*QueryState)

		col := ""
		op := "="
		var val interface{}

		// Shorthand: where: { id: 1, status: 'active' }
		if node.Value != nil {
			v := resolveValue(node.Value, scope)
			if m, ok := v.(map[string]interface{}); ok {
				for k, v := range m {
					qs.Where = append(qs.Where, WhereCond{Column: k, Op: "=", Value: v})
					qs.Args = append(qs.Args, v)
				}
				// If we processed a map value, we might also have children that are key-values for the map
				// But typically if Value is a map, it means shorthand object syntax
			}
		}

		// Standard Syntax: where { col: 'id', op: '=', val: 1 }
		// OR Multi-Where: where { id: 1, name: 'Budi' } (Children as Key-Value)
		// We need to distinguish between explicit config (col/op/val) and implicit key-value
		isExplicitConfig := false
		for _, c := range node.Children {
			if c.Name == "col" || c.Name == "op" || c.Name == "val" {
				isExplicitConfig = true
				break
			}
		}

		if isExplicitConfig {
			for _, c := range node.Children {
				if c.Name == "col" {
					col = coerce.ToString(parseNodeValue(c, scope))
				}
				if c.Name == "op" {
					op = coerce.ToString(parseNodeValue(c, scope))
				}
				if c.Name == "val" {
					val = parseNodeValue(c, scope)
				}
			}
			if col != "" {
				qs.Where = append(qs.Where, WhereCond{Column: col, Op: op, Value: val})
				qs.Args = append(qs.Args, val)
			}
		} else {
			// Implicit Key-Value from children
			// e.g. where { id: 1, status: 'active' }
			for _, c := range node.Children {
				k := c.Name
				v := parseNodeValue(c, scope)
				qs.Where = append(qs.Where, WhereCond{Column: k, Op: "=", Value: v})
				qs.Args = append(qs.Args, v)
			}
		}

		return nil
	}, engine.SlotMeta{
		Description: "Add a WHERE filter to the query. Supports both explicit config (col, op, val) and implicit key-value pairs.",
		Group:       "Database",
		Example:     "db.where\n  col: id\n  val: $user_id\n\n# Or implicit:\ndb.where { id: 1, status: 'active' }",
	})

	// DB.ORDER_BY
	eng.Register("db.order_by", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		qsVal, ok := scope.Get("_query_state")
		if !ok {
			return nil
		}
		qs := qsVal.(*QueryState)

		// [STANDARISASI] Support variable ($sort) dan auto-clean quotes
		if node.Value != nil {
			qs.OrderBy = coerce.ToString(resolveValue(node.Value, scope))
		}
		return nil
	}, engine.SlotMeta{
		Example: "db.order_by: 'id DESC'",
		Group:   "Database",
	})

	// DB.LIMIT
	eng.Register("db.limit", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		qsVal, ok := scope.Get("_query_state")
		if !ok {
			return nil
		}
		qs := qsVal.(*QueryState)

		val := resolveValue(node.Value, scope)
		limit, _ := coerce.ToInt(val)
		qs.Limit = limit
		return nil
	}, engine.SlotMeta{
		Example: "db.limit: $limit",
		Group:   "Database",
	})

	// DB.OFFSET
	eng.Register("db.offset", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		qsVal, ok := scope.Get("_query_state")
		if !ok {
			return nil
		}
		qsVal.(*QueryState).Offset, _ = coerce.ToInt(resolveValue(node.Value, scope))
		return nil
	}, engine.SlotMeta{
		Example: "db.offset: $offset",
		Group:   "Database",
	})

	// =========================================================================
	// EXECUTION SLOTS
	// =========================================================================

	// DB.GET
	eng.Register("db.get", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		qsVal, ok := scope.Get("_query_state")
		if !ok {
			return fmt.Errorf("db.get called without db.table")
		}
		qs := qsVal.(*QueryState)
		target := "rows"

		for _, c := range node.Children {
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		// Use BuildSQL
		query, args := qs.BuildSQL("SELECT")

		executor, _, err := getExecutor(scope, dbMgr, qs.DBName)
		if err != nil {
			return err
		}

		rows, err := executor.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		cols, _ := rows.Columns()
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

		scope.Set(target, results)
		return nil
	}, engine.SlotMeta{
		Description: "Retrieve multiple rows from the database based on the current query state.",
		Group:       "Database",
		Example:     "db.get\n  as: $users",
		Inputs: map[string]engine.InputMeta{
			"as": {Description: "Variable name to store results", Required: true},
		},
	})

	// DB.FIRST
	eng.Register("db.first", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		qsVal, ok := scope.Get("_query_state")
		if !ok {
			return fmt.Errorf("db.first called without db.table")
		}
		qs := qsVal.(*QueryState)
		target := "row"
		for _, c := range node.Children {
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		// Use BuildSQL
		// LIMIT 1 handling by dialect inside BuildSQL is tricky if we pass raw SELECT
		// We can override limit in state temporarily
		oldLimit := qs.Limit
		qs.Limit = 1
		query, args := qs.BuildSQL("SELECT")
		qs.Limit = oldLimit // Restore

		executor, _, err := getExecutor(scope, dbMgr, qs.DBName)
		if err != nil {
			return err
		}

		rows, err := executor.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		cols, _ := rows.Columns()
		if rows.Next() {
			columns := make([]interface{}, len(cols))
			columnPointers := make([]interface{}, len(cols))
			for i := range columns {
				columnPointers[i] = &columns[i]
			}
			rows.Scan(columnPointers...)
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
			scope.Set(target, m)
			scope.Set(target+"_found", true)
		} else {
			scope.Set(target, nil)
			scope.Set(target+"_found", false)
		}
		return nil
	}, engine.SlotMeta{
		Description: "Retrieve the first row from the database based on the current query state.",
		Group:       "Database",
		Example:     "db.first\n  as: $user",
		Inputs: map[string]engine.InputMeta{
			"as": {Description: "Variable name to store result", Required: true},
		},
	})

	// DB.INSERT
	eng.Register("db.insert", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		qsVal, ok := scope.Get("_query_state")
		if !ok {
			return fmt.Errorf("db.insert called without db.table")
		}
		qs := qsVal.(*QueryState)
		var cols []string
		var vals []interface{}
		var placeholders []string

		for i, c := range node.Children {
			cols = append(cols, qs.Dialect.QuoteIdentifier(c.Name))
			placeholders = append(placeholders, qs.Dialect.Placeholder(i+1))
			// Use parseNodeValue to support $variable
			val := parseNodeValue(c, scope)
			vals = append(vals, val)
		}
		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			qs.Dialect.QuoteIdentifier(qs.Table), strings.Join(cols, ", "), strings.Join(placeholders, ", "))

		executor, dialect, err := getExecutor(scope, dbMgr, qs.DBName)
		if err != nil {
			return err
		}
		res, err := executor.ExecContext(ctx, query, vals...)
		if err != nil {
			return err
		}

		// [IMPORTANT] LastInsertId behavior is dialect-dependent.
		// Some DBs (Postgres) don't support it directly without RETURNING.
		// For now we keep it but it might return 0 or error on some DBs.
		if dialect.Name() != "postgres" {
			id, _ := res.LastInsertId()
			scope.Set("db_last_id", id)
		}
		return nil
	}, engine.SlotMeta{
		Description: "Insert a new row into the table.",
		Example:     "db.insert\n  name: $name\n  email: 'test@example.com'",
		Group:       "Database",
	})

	// DB.UPDATE
	eng.Register("db.update", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		qsVal, ok := scope.Get("_query_state")
		if !ok {
			return fmt.Errorf("db.update called without db.table")
		}
		qs := qsVal.(*QueryState)
		var sets []string
		var vals []interface{}
		for i, c := range node.Children {
			sets = append(sets, fmt.Sprintf("%s = %s", qs.Dialect.QuoteIdentifier(c.Name), qs.Dialect.Placeholder(i+1)))
			vals = append(vals, parseNodeValue(c, scope))
		}

		whereClause := ""
		if len(qs.Where) > 0 {
			var whereParts []string
			baseIdx := len(vals)
			for i, cond := range qs.Where {
				whereParts = append(whereParts, fmt.Sprintf("%s %s %s",
					qs.Dialect.QuoteIdentifier(cond.Column),
					cond.Op,
					qs.Dialect.Placeholder(baseIdx+i+1)))
				vals = append(vals, cond.Value)
			}
			whereClause = " WHERE " + strings.Join(whereParts, " AND ")
		}

		query := fmt.Sprintf("UPDATE %s SET %s%s", qs.Dialect.QuoteIdentifier(qs.Table), strings.Join(sets, ", "), whereClause)
		executor, _, err := getExecutor(scope, dbMgr, qs.DBName)
		if err != nil {
			return err
		}
		_, err = executor.ExecContext(ctx, query, vals...)
		return err
	}, engine.SlotMeta{
		Description: "Update rows in the table.",
		Example:     "db.update\n  status: 'active'",
		Group:       "Database",
	})

	// DB.DELETE
	eng.Register("db.delete", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		qsVal, ok := scope.Get("_query_state")
		if !ok {
			return fmt.Errorf("db.delete called without db.table")
		}
		qs := qsVal.(*QueryState)
		target := ""
		for _, c := range node.Children {
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		// Use BuildSQL
		query, args := qs.BuildSQL("DELETE")
		executor, _, err := getExecutor(scope, dbMgr, qs.DBName)
		if err != nil {
			return err
		}
		res, err := executor.ExecContext(ctx, query, args...)
		if err != nil {
			return err
		}

		if target != "" {
			count, _ := res.RowsAffected()
			scope.Set(target, count)
		}
		return nil
	}, engine.SlotMeta{
		Example: "db.delete\n  as: $count",
		Group:   "Database",
	})

	// DB.COUNT
	eng.Register("db.count", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		qsVal, ok := scope.Get("_query_state")
		if !ok {
			return fmt.Errorf("db.count called without db.table")
		}
		qs := qsVal.(*QueryState)
		target := "count"
		for _, c := range node.Children {
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}
		// Use BuildSQL
		query, args := qs.BuildSQL("COUNT")

		executor, _, err := getExecutor(scope, dbMgr, qs.DBName)
		if err != nil {
			return err
		}
		var total int
		err = executor.QueryRowContext(ctx, query, args...).Scan(&total)
		if err != nil {
			return err
		}
		scope.Set(target, total)
		return nil
	}, engine.SlotMeta{
		Description: "Count the number of rows based on the current query state.",
		Group:       "Database",
		Example:     "db.count\n  as: $total",
		Inputs: map[string]engine.InputMeta{
			"as": {Description: "Variable name to store result", Required: true},
		},
	})
}
