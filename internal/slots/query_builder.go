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
	Logical string // "AND" or "OR"
	Column  string
	Op      string
	Value   interface{}
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
				logical := cond.Logical
				if logical == "" {
					logical = "AND"
				}
				sb.WriteString(fmt.Sprintf(" %s ", logical))
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
			} else if strings.ToUpper(cond.Op) == "BETWEEN" || strings.ToUpper(cond.Op) == "NOT BETWEEN" {
				// Expect Value to be slice of 2 items
				v := reflect.ValueOf(cond.Value)
				var val1, val2 interface{}

				if v.Kind() == reflect.Slice && v.Len() >= 2 {
					val1 = v.Index(0).Interface()
					val2 = v.Index(1).Interface()
				} else if str, ok := cond.Value.(string); ok && strings.HasPrefix(strings.TrimSpace(str), "[") {
					content := strings.TrimSpace(str)
					content = strings.TrimPrefix(content, "[")
					content = strings.TrimSuffix(content, "]")
					parts := strings.Split(content, ",")
					if len(parts) >= 2 {
						val1 = strings.TrimSpace(parts[0])
						val2 = strings.TrimSpace(parts[1])
					}
				}

				// Apply bound values if valid
				if val1 != nil && val2 != nil {
					p1 := qs.Dialect.Placeholder(len(args) + 1)
					args = append(args, val1)
					p2 := qs.Dialect.Placeholder(len(args) + 1)
					args = append(args, val2)

					sb.WriteString(fmt.Sprintf("%s %s %s AND %s",
						qs.Quote(cond.Column),
						cond.Op,
						p1,
						p2,
					))
				}
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

	// DB.QUERY (Fluent Block-Style Builder)
	eng.Register("db.query", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		// Main value: db.query: 'users' (table name)
		tableName := coerce.ToString(resolveValue(node.Value, scope))
		dbName := "default"

		// Cari nama opsional dan db di anak level atas
		for _, c := range node.Children {
			if c.Name == "table" || c.Name == "name" {
				tableName = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "db" {
				dbName = coerce.ToString(parseNodeValue(c, scope))
			}
		}

		dialect := dbMgr.GetDialect(dbName)

		// Create a local inner scope for this entire query building process
		// So "_query_state" doesn't leak into the global scope
		innerScope := engine.NewScope(scope)
		innerScope.Set("_query_state", &QueryState{
			Table:   tableName,
			DBName:  dbName,
			Dialect: dialect,
		})

		// Execute all children instructions inside the query block
		for _, c := range node.Children {
			// Skip basic configuration keys to avoid executing them as slots
			if c.Name == "table" || c.Name == "name" || c.Name == "db" {
				continue
			}

			// We dynamically adjust child names to append "db." prefix if missing
			// Example: where { ... } becomes db.where { ... } in the engine's perspective
			callName := c.Name
			if !strings.HasPrefix(callName, "db.") {
				callName = "db." + callName
			}

			// Mock a node call for execution
			callNode := &engine.Node{
				Name:     callName,
				Value:    c.Value,
				Children: c.Children,
				Line:     c.Line,
				Col:      c.Col,
				Filename: c.Filename,
			}

			if err := eng.Execute(ctx, callNode, innerScope); err != nil {
				return err
			}
		}

		return nil
	}, engine.SlotMeta{
		Description: "A fluent block wrapper for constructing and executing a query on a specific table.",
		Example:     "db.query: 'users' {\n  where: { col: 'status', val: 'active' }\n  get: { as: $users }\n}",
	})

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
		return nil
	}, engine.SlotMeta{
		Description: "Set the table to be used for subsequent database operations.",
		Example:     "db.table: 'users'",
		Inputs: map[string]engine.InputMeta{
			"name": {Description: "Table name (Optional if specified in main value)", Required: false},
			"db":   {Description: "Database connection name (Default: 'default')", Required: false},
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
	}, engine.SlotMeta{Example: "db.columns: ['id', 'name']"})

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
	eng.Register("db.join", joinHandler("INNER"), engine.SlotMeta{Example: "db.join {\n  table: posts\n  on: ['users.id', '=', 'posts.user_id']\n}"})
	eng.Register("db.left_join", joinHandler("LEFT"), engine.SlotMeta{Example: "db.left_join ..."})
	eng.Register("db.right_join", joinHandler("RIGHT"), engine.SlotMeta{})

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
	}, engine.SlotMeta{Example: "db.where_in {\n  col: id\n  val: [1, 2, 3]\n}"})

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
	}, engine.SlotMeta{Example: "db.where_null: deleted_at"})

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
	}, engine.SlotMeta{Example: "db.where_not_null: created_at"})

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
	}, engine.SlotMeta{Example: "db.where_not_in {\n  col: status\n  val: ['archived', 'deleted']\n}"})

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
	}, engine.SlotMeta{Example: "db.group_by: 'status'"})

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
	}, engine.SlotMeta{Example: "db.having {\n  col: count\n  op: '>'\n  val: 5\n}"})

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

		if len(node.Children) > 0 {
			for _, c := range node.Children {
				if c.Name == "col" {
					// [STANDARDIZATION] Use parseNodeValue so quotes are stripped & variables supported
					col = coerce.ToString(parseNodeValue(c, scope))
				}
				if c.Name == "op" {
					// [STANDARDIZATION] Same, so op: "LIKE" and op: LIKE are both valid
					op = coerce.ToString(parseNodeValue(c, scope))
				}
				if c.Name == "val" {
					val = parseNodeValue(c, scope)
				}
			}
		}

		if col != "" {
			qs.Where = append(qs.Where, WhereCond{Logical: "AND", Column: col, Op: op, Value: val})
			qs.Args = append(qs.Args, val)
		}
		return nil
	}, engine.SlotMeta{
		Description: "Add a WHERE filter to the query.",
		Example:     "db.where\n  col: id\n  val: $user_id",
		Inputs: map[string]engine.InputMeta{
			"col": {Description: "Column name", Required: false},
			"op":  {Description: "Operator (Default: '=')", Required: false},
			"val": {Description: "Filter value", Required: false},
		},
	})

	// DB.OR_WHERE
	eng.Register("db.or_where", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		qsVal, ok := scope.Get("_query_state")
		if !ok {
			return fmt.Errorf("db.or_where called without db.table")
		}
		qs := qsVal.(*QueryState)

		col := ""
		op := "="
		var val interface{}

		if len(node.Children) > 0 {
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
		}

		if col != "" {
			qs.Where = append(qs.Where, WhereCond{Logical: "OR", Column: col, Op: op, Value: val})
			qs.Args = append(qs.Args, val)
		}
		return nil
	}, engine.SlotMeta{
		Description: "Add an OR WHERE filter to the query.",
		Example:     "db.or_where\n  col: role\n  val: 'admin'",
	})

	// DB.WHERE_BETWEEN
	eng.Register("db.where_between", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		qsVal, ok := scope.Get("_query_state")
		if !ok {
			return fmt.Errorf("db.where_between called without db.table")
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
			qs.Where = append(qs.Where, WhereCond{Logical: "AND", Column: col, Op: "BETWEEN", Value: val})
		}
		return nil
	}, engine.SlotMeta{
		Description: "Add a WHERE BETWEEN filter to the query.",
		Example:     "db.where_between\n  col: age\n  val: [18, 30]",
	})

	// DB.WHERE_NOT_BETWEEN
	eng.Register("db.where_not_between", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		qsVal, ok := scope.Get("_query_state")
		if !ok {
			return fmt.Errorf("db.where_not_between called without db.table")
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
			qs.Where = append(qs.Where, WhereCond{Logical: "AND", Column: col, Op: "NOT BETWEEN", Value: val})
		}
		return nil
	}, engine.SlotMeta{
		Description: "Add a WHERE NOT BETWEEN filter to the query.",
		Example:     "db.where_not_between\n  col: age\n  val: [18, 30]",
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
	}, engine.SlotMeta{Example: "db.order_by: 'id DESC'"})

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
	}, engine.SlotMeta{Example: "db.limit: $limit"})

	// DB.OFFSET
	eng.Register("db.offset", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		qsVal, ok := scope.Get("_query_state")
		if !ok {
			return nil
		}
		qs := qsVal.(*QueryState)

		val := resolveValue(node.Value, scope)
		offset, _ := coerce.ToInt(val)
		qs.Offset = offset
		return nil
	}, engine.SlotMeta{Example: "db.offset: $offset"})

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
	}, engine.SlotMeta{Example: "db.insert\n  name: $name"})

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
	}, engine.SlotMeta{Example: "db.update\n  status: 'active'"})

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
	}, engine.SlotMeta{Example: "db.delete\n  as: $count"})

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
		Example:     "db.count\n  as: $total",
		Inputs: map[string]engine.InputMeta{
			"as": {Description: "Variable name to store result", Required: true},
		},
	})

	// DB.EXISTS
	eng.Register("db.exists", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		qsVal, ok := scope.Get("_query_state")
		if !ok {
			return fmt.Errorf("db.exists called without db.table")
		}
		qs := qsVal.(*QueryState)
		target := "exists"
		for _, c := range node.Children {
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		oldLimit := qs.Limit
		qs.Limit = 1
		query, args := qs.BuildSQL("SELECT")
		qs.Limit = oldLimit

		executor, _, err := getExecutor(scope, dbMgr, qs.DBName)
		if err != nil {
			return err
		}

		rows, err := executor.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		exists := rows.Next()
		scope.Set(target, exists)
		return nil
	}, engine.SlotMeta{
		Description: "Check if at least one row exists based on the current query state.",
		Example:     "db.exists\n  as: $has_users",
	})

	// DB.DOESNT_EXIST
	eng.Register("db.doesnt_exist", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		qsVal, ok := scope.Get("_query_state")
		if !ok {
			return fmt.Errorf("db.doesnt_exist called without db.table")
		}
		qs := qsVal.(*QueryState)
		target := "doesnt_exist"
		for _, c := range node.Children {
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		oldLimit := qs.Limit
		qs.Limit = 1
		query, args := qs.BuildSQL("SELECT")
		qs.Limit = oldLimit

		executor, _, err := getExecutor(scope, dbMgr, qs.DBName)
		if err != nil {
			return err
		}

		rows, err := executor.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		exists := rows.Next()
		scope.Set(target, !exists)
		return nil
	}, engine.SlotMeta{
		Description: "Check if no rows exist based on the current query state.",
		Example:     "db.doesnt_exist\n  as: $is_empty",
	})

	// AGGREGATE HANDLER GENERATOR
	aggregateHandler := func(funcName string) func(context.Context, *engine.Node, *engine.Scope) error {
		return func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
			qsVal, ok := scope.Get("_query_state")
			if !ok {
				return fmt.Errorf("db.%s called without db.table", strings.ToLower(funcName))
			}
			qs := qsVal.(*QueryState)
			target := strings.ToLower(funcName)
			col := "*"

			if node.Value != nil {
				col = coerce.ToString(resolveValue(node.Value, scope))
			}

			for _, c := range node.Children {
				if c.Name == "col" {
					col = coerce.ToString(parseNodeValue(c, scope))
				}
				if c.Name == "as" {
					target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
				}
			}

			// We temporarily modify Columns to the aggregate, then switch back
			oldCols := qs.Columns
			qs.Columns = []string{fmt.Sprintf("%s(%s)", funcName, qs.Quote(col))}
			query, args := qs.BuildSQL("SELECT")
			qs.Columns = oldCols

			executor, _, err := getExecutor(scope, dbMgr, qs.DBName)
			if err != nil {
				return err
			}
			var result interface{}
			err = executor.QueryRowContext(ctx, query, args...).Scan(&result)
			if err != nil && err != sql.ErrNoRows {
				return err
			}

			// Coerce to proper type based on funcName. sum/min/max usually float or int.
			b, ok := result.([]byte)
			if ok {
				result = string(b)
			}

			scope.Set(target, result)
			return nil
		}
	}

	eng.Register("db.sum", aggregateHandler("SUM"), engine.SlotMeta{Example: "db.sum: 'price'\n  as: $total_price"})
	eng.Register("db.avg", aggregateHandler("AVG"), engine.SlotMeta{Example: "db.avg: 'rating'\n  as: $average"})
	eng.Register("db.min", aggregateHandler("MIN"), engine.SlotMeta{Example: "db.min: 'age'\n  as: $youngest"})
	eng.Register("db.max", aggregateHandler("MAX"), engine.SlotMeta{Example: "db.max: 'age'\n  as: $oldest"})

	// DB.PLUCK
	eng.Register("db.pluck", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		qsVal, ok := scope.Get("_query_state")
		if !ok {
			return fmt.Errorf("db.pluck called without db.table")
		}
		qs := qsVal.(*QueryState)

		col := ""
		target := "plucks"

		if node.Value != nil {
			col = coerce.ToString(resolveValue(node.Value, scope))
		}

		for _, c := range node.Children {
			if c.Name == "col" {
				col = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		if col == "" {
			return fmt.Errorf("db.pluck requires a column name")
		}

		// Save old columns
		oldCols := qs.Columns
		qs.Columns = []string{qs.Quote(col)}
		query, args := qs.BuildSQL("SELECT")
		qs.Columns = oldCols

		executor, _, err := getExecutor(scope, dbMgr, qs.DBName)
		if err != nil {
			return err
		}

		rows, err := executor.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		var results []interface{}

		for rows.Next() {
			var val interface{}
			if err := rows.Scan(&val); err != nil {
				return err
			}
			b, ok := val.([]byte)
			if ok {
				val = string(b)
			}
			results = append(results, val)
		}

		scope.Set(target, results)
		return nil
	}, engine.SlotMeta{
		Description: "Retrieve a single column's values as a flat array.",
		Example:     "db.pluck: 'id'\n  as: $user_ids",
	})

	// DB.PAGINATE
	eng.Register("db.paginate", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		qsVal, ok := scope.Get("_query_state")
		if !ok {
			return fmt.Errorf("db.paginate called without db.table")
		}
		qs := qsVal.(*QueryState)

		target := "paginator"
		perPage := 15
		page := 1

		for _, c := range node.Children {
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
			if c.Name == "per_page" {
				val, _ := coerce.ToFloat64(parseNodeValue(c, scope))
				perPage = int(val)
			}
			if c.Name == "page" {
				val, _ := coerce.ToFloat64(parseNodeValue(c, scope))
				page = int(val)
			}
		}
		if page < 1 {
			page = 1
		}
		if perPage < 1 {
			perPage = 15
		}

		executor, _, err := getExecutor(scope, dbMgr, qs.DBName)
		if err != nil {
			return err
		}

		// 1. Get Total Count
		countQuery, countArgs := qs.BuildSQL("COUNT")
		var total int
		err = executor.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
		if err != nil && err != sql.ErrNoRows {
			return err
		}

		// 2. Compute Pagination Meta
		// We avoid importing math by doing integer arithmetic
		lastPage := (total + perPage - 1) / perPage
		if lastPage < 1 {
			lastPage = 1
		}

		// 3. Set Limit and Offset for actual query
		qs.Limit = perPage
		qs.Offset = (page - 1) * perPage
		query, args := qs.BuildSQL("SELECT")

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

		paginator := map[string]interface{}{
			"data":         results,
			"total":        total,
			"per_page":     perPage,
			"current_page": page,
			"last_page":    lastPage,
			"from":         qs.Offset + 1,
			"to":           qs.Offset + len(results),
		}
		if len(results) == 0 {
			paginator["from"] = 0
			paginator["to"] = 0
		}

		scope.Set(target, paginator)
		return nil
	}, engine.SlotMeta{
		Description: "Retrieve rows paginated with metadata.",
		Example:     "db.paginate\n  page: 1\n  per_page: 20\n  as: $users_paginator",
	})
}
