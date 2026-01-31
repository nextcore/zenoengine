package slots

import (
	"context"
	"fmt"
	"strings"
	"zeno/pkg/dbmanager"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

type ColumnDef struct {
	Name          string
	Type          string
	IsPrimary     bool
	IsNullable    bool
	AutoIncrement bool
	DefaultValue  string
}

type TableSchema struct {
	Name    string
	Columns []ColumnDef
}

func RegisterSchemaSlots(eng *engine.Engine, dbMgr *dbmanager.DBManager) {

	// SCHEMA.CREATE: 'table_name' { ... columns ... }
	eng.Register("schema.create", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		tableName := coerce.ToString(resolveValue(node.Value, scope))
		dbName := "default"

		// Target: Collect column definitions from children execution
		schema := &TableSchema{
			Name: tableName,
		}

		// Set temporary state for column slots to collect data
		oldSchema, hasOld := scope.Get("_schema_draft")
		scope.Set("_schema_draft", schema)
		defer func() {
			if hasOld {
				scope.Set("_schema_draft", oldSchema)
			} else {
				scope.Delete("_schema_draft")
			}
		}()

		// Execute children to fill schema columns
		for _, c := range node.Children {
			if c.Name == "db" || c.Name == "connection" {
				dbName = coerce.ToString(parseNodeValue(c, scope))
				continue
			}
			// Execute other children (column slots)
			if err := eng.Execute(ctx, c, scope); err != nil {
				return err
			}
		}

		if schema.Name == "" {
			return fmt.Errorf("schema.create: table name is required")
		}

		// Build SQL DDL based on Dialect
		dialect := dbMgr.GetDialect(dbName)
		ddl := buildCreateSQL(schema, dialect)

		// Execute DDL
		executor, _, err := getExecutor(scope, dbMgr, dbName)
		if err != nil {
			return err
		}

		_, err = executor.ExecContext(ctx, ddl)
		if err != nil {
			return fmt.Errorf("schema.create error: %v\nSQL: %s", err, ddl)
		}

		return nil
	}, engine.SlotMeta{
		Description: "Create a new database table using fluent schema builder.",
		Example:     "schema.create: 'users' {\n  column.id: 'id'\n  column.string: 'name'\n}",
	})

	// Helper for column slots
	registerColumn := func(typeName string, defaultType string) {
		eng.Register("column."+typeName, func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
			val, _ := scope.Get("_schema_draft")
			if val == nil {
				return fmt.Errorf("column.%s must be called inside schema.create", typeName)
			}
			schema := val.(*TableSchema)

			colName := coerce.ToString(resolveValue(node.Value, scope))
			col := ColumnDef{
				Name: colName,
				Type: defaultType,
			}

			for _, c := range node.Children {
				if c.Name == "name" {
					col.Name = coerce.ToString(parseNodeValue(c, scope))
				}
				if c.Name == "primary" {
					col.IsPrimary, _ = coerce.ToBool(parseNodeValue(c, scope))
				}
				if c.Name == "nullable" {
					col.IsNullable, _ = coerce.ToBool(parseNodeValue(c, scope))
				}
				if c.Name == "default" {
					col.DefaultValue = coerce.ToString(parseNodeValue(c, scope))
				}
			}

			if typeName == "id" {
				col.IsPrimary = true
				col.AutoIncrement = true
			}

			schema.Columns = append(schema.Columns, col)
			return nil
		}, engine.SlotMeta{})
	}

	registerColumn("id", "INTEGER")
	registerColumn("string", "VARCHAR(255)")
	registerColumn("text", "TEXT")
	registerColumn("integer", "INTEGER")
	registerColumn("boolean", "BOOLEAN")
	registerColumn("datetime", "DATETIME")

	// Special: Timestamps
	eng.Register("column.timestamps", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		val, _ := scope.Get("_schema_draft")
		if val == nil {
			return fmt.Errorf("column.timestamps must be called inside schema.create")
		}
		schema := val.(*TableSchema)

		schema.Columns = append(schema.Columns, ColumnDef{Name: "created_at", Type: "TIMESTAMP", DefaultValue: "CURRENT_TIMESTAMP"})
		schema.Columns = append(schema.Columns, ColumnDef{Name: "updated_at", Type: "TIMESTAMP", DefaultValue: "CURRENT_TIMESTAMP"})
		return nil
	}, engine.SlotMeta{})
}

func buildCreateSQL(schema *TableSchema, dialect dbmanager.Dialect) string {
	var sb strings.Builder
	sb.WriteString("CREATE TABLE IF NOT EXISTS ")
	sb.WriteString(dialect.QuoteIdentifier(schema.Name))
	sb.WriteString(" (\n")

	var colStrings []string
	for _, col := range schema.Columns {
		s := fmt.Sprintf("    %s %s", dialect.QuoteIdentifier(col.Name), translateType(col.Type, dialect))
		if col.IsPrimary {
			// Dialect specific primary key / auto increment
			if dialect.Name() == "sqlite" && col.AutoIncrement {
				s += " PRIMARY KEY AUTOINCREMENT"
			} else if col.IsPrimary {
				s += " PRIMARY KEY"
			}
		}
		if !col.IsNullable && !col.IsPrimary {
			s += " NOT NULL"
		}
		if col.DefaultValue != "" {
			s += " DEFAULT " + col.DefaultValue
		}
		colStrings = append(colStrings, s)
	}

	sb.WriteString(strings.Join(colStrings, ",\n"))
	sb.WriteString("\n);")

	return sb.String()
}

func translateType(typeName string, dialect dbmanager.Dialect) string {
    // Simple mapping for now
    if dialect.Name() == "postgres" {
        if typeName == "TIMESTAMP" { return "TIMESTAMPTZ" }
    }
    return typeName
}
