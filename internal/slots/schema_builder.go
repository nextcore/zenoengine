package slots

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"zeno/pkg/dbmanager"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

type ColumnDef struct {
	Name      string
	Type      string
	Limit     int
	Unique    bool
	Nullable  bool
	Precision int
	Scale     int
}

type ForeignKeyDef struct {
	Column     string
	References string
	On         string
	OnDelete   string
}

type SchemaState struct {
	Table       string
	Columns     []ColumnDef
	ForeignKeys []ForeignKeyDef
	DBName      string
	Dialect     dbmanager.Dialect
}

func RegisterSchemaSlots(eng *engine.Engine, dbMgr *dbmanager.DBManager) {
	// UP Slot
	eng.Register("up", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		dir, _ := scope.Get("_migration_direction")
		if dir == "down" {
			return nil // Skip if we are rolling back
		}
		for _, c := range node.Children {
			if err := eng.Execute(ctx, c, scope); err != nil {
				return err
			}
		}
		return nil
	}, engine.SlotMeta{Description: "Execute block only during up migration"})

	// DOWN Slot
	eng.Register("down", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		dir, _ := scope.Get("_migration_direction")
		if dir != "down" {
			return nil // Skip if we are migrating UP
		}
		for _, c := range node.Children {
			if err := eng.Execute(ctx, c, scope); err != nil {
				return err
			}
		}
		return nil
	}, engine.SlotMeta{Description: "Execute block only during down migration"})

	// DB.CREATE_TABLE
	eng.Register("db.create_table", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		dir, _ := scope.Get("_migration_direction")
		if dir == "down" {
			return nil // Skip creation during rollback
		}

		tableName := coerce.ToString(resolveValue(node.Value, scope))
		dbName := "default"

		for _, c := range node.Children {
			if c.Name == "db" {
				dbName = coerce.ToString(parseNodeValue(c, scope))
			}
		}

		dialect := dbMgr.GetDialect(dbName)
		if dialect == nil {
			dialect = dbmanager.SQLiteDialect{} // Fallback
		}

		state := &SchemaState{
			Table:   tableName,
			DBName:  dbName,
			Dialect: dialect,
		}

		innerScope := engine.NewScope(scope)
		innerScope.Set("_schema_state", state)

		for _, c := range node.Children {
			if c.Name == "db" {
				continue
			}

			callName := c.Name
			if !strings.HasPrefix(callName, "db.") {
				callName = "db." + callName
			}

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

		// Generate SQL and execute
		sql, err := state.BuildSQL()
		if err != nil {
			return err
		}

		slog.Info("Generated SQL", "sql", sql)

		db := dbMgr.GetConnection(dbName)
		if db == nil {
			return fmt.Errorf("database connection '%s' not found", dbName)
		}

		_, err = db.ExecContext(ctx, sql)
		return err
	}, engine.SlotMeta{Description: "Create a database table with schema builder"})

	// DB.DROP_TABLE
	eng.Register("db.drop_table", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		tableName := coerce.ToString(resolveValue(node.Value, scope))
		dbName := "default"

		for _, c := range node.Children {
			if c.Name == "db" {
				dbName = coerce.ToString(parseNodeValue(c, scope))
			}
		}

		db := dbMgr.GetConnection(dbName)
		if db == nil {
			return fmt.Errorf("database connection '%s' not found", dbName)
		}

		dialect := dbMgr.GetDialect(dbName)
		if dialect == nil {
			dialect = dbmanager.SQLiteDialect{}
		}

		sql := fmt.Sprintf("DROP TABLE %s;", dialect.QuoteIdentifier(tableName))
		slog.Info("Generated SQL", "sql", sql)

		_, err := db.ExecContext(ctx, sql)
		return err
	}, engine.SlotMeta{Description: "Drop a database table"})

	// DB.FOREIGN
	eng.Register("db.foreign", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		stateVal, ok := scope.Get("_schema_state")
		if !ok {
			return fmt.Errorf("db.foreign called outside of db.create_table")
		}
		state := stateVal.(*SchemaState)

		colName := coerce.ToString(resolveValue(node.Value, scope))
		references := ""
		on := "id"
		onDelete := "CASCADE"

		for _, c := range node.Children {
			if c.Name == "references" {
				references = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "on" {
				on = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "on_delete" {
				onDelete = coerce.ToString(parseNodeValue(c, scope))
			}
		}

		state.ForeignKeys = append(state.ForeignKeys, ForeignKeyDef{
			Column:     colName,
			References: references,
			On:         on,
			OnDelete:   onDelete,
		})

		return nil
	}, engine.SlotMeta{Description: "Define foreign key constraint"})

	// Column slots
	registerColumnSlot(eng, "db.id", "id")
	registerColumnSlot(eng, "db.string", "string")
	registerColumnSlot(eng, "db.integer", "integer")
	registerColumnSlot(eng, "db.timestamp", "timestamp")
	registerColumnSlot(eng, "db.boolean", "boolean")
	registerColumnSlot(eng, "db.text", "text")
	registerColumnSlot(eng, "db.decimal", "decimal")
	registerColumnSlot(eng, "db.date", "date")
	registerColumnSlot(eng, "db.json", "json")
}

func registerColumnSlot(eng *engine.Engine, slotName, colType string) {
	eng.Register(slotName, func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		stateVal, ok := scope.Get("_schema_state")
		if !ok {
			return fmt.Errorf("%s called outside of db.create_table", slotName)
		}
		state := stateVal.(*SchemaState)

		colName := coerce.ToString(resolveValue(node.Value, scope))
		limit := 0
		unique := false
		nullable := true
		precision := 0
		scale := 0

		for _, c := range node.Children {
			if c.Name == "limit" || c.Name == "length" {
				limit, _ = coerce.ToInt(parseNodeValue(c, scope))
			}
			if c.Name == "unique" {
				unique, _ = coerce.ToBool(parseNodeValue(c, scope))
			}
			if c.Name == "nullable" {
				nullable, _ = coerce.ToBool(parseNodeValue(c, scope))
			}
			if c.Name == "precision" {
				precision, _ = coerce.ToInt(parseNodeValue(c, scope))
			}
			if c.Name == "scale" {
				scale, _ = coerce.ToInt(parseNodeValue(c, scope))
			}
		}

		state.Columns = append(state.Columns, ColumnDef{
			Name:      colName,
			Type:      colType,
			Limit:     limit,
			Unique:    unique,
			Nullable:  nullable,
			Precision: precision,
			Scale:     scale,
		})

		return nil
	}, engine.SlotMeta{})
}

func (s *SchemaState) BuildSQL() (string, error) {
	var sb strings.Builder
	sb.WriteString("CREATE TABLE ")
	sb.WriteString(s.Dialect.QuoteIdentifier(s.Table))
	sb.WriteString(" (\n")

	var colStrings []string
	for _, col := range s.Columns {
		colStr := "  " + s.Dialect.QuoteIdentifier(col.Name) + " " + s.translateType(col)
		if !col.Nullable {
			colStr += " NOT NULL"
		}
		if col.Unique {
			colStr += " UNIQUE"
		}
		colStrings = append(colStrings, colStr)
	}

	for _, fk := range s.ForeignKeys {
		fkStr := fmt.Sprintf("  FOREIGN KEY (%s) REFERENCES %s(%s)",
			s.Dialect.QuoteIdentifier(fk.Column),
			s.Dialect.QuoteIdentifier(fk.References),
			s.Dialect.QuoteIdentifier(fk.On))
		if fk.OnDelete != "" {
			fkStr += " ON DELETE " + fk.OnDelete
		}
		colStrings = append(colStrings, fkStr)
	}

	sb.WriteString(strings.Join(colStrings, ",\n"))
	sb.WriteString("\n);")

	return sb.String(), nil
}

func (s *SchemaState) translateType(col ColumnDef) string {
	dialectName := s.Dialect.Name()

	switch col.Type {
	case "id":
		switch dialectName {
		case "sqlite":
			return "INTEGER PRIMARY KEY AUTOINCREMENT"
		case "mysql":
			return "INT AUTO_INCREMENT PRIMARY KEY"
		case "postgres":
			return "SERIAL PRIMARY KEY"
		default:
			return "INTEGER PRIMARY KEY"
		}
	case "string":
		limit := col.Limit
		if limit == 0 {
			limit = 255
		}
		switch dialectName {
		case "sqlite":
			return "TEXT"
		case "mysql", "postgres":
			return fmt.Sprintf("VARCHAR(%d)", limit)
		default:
			return "TEXT"
		}
	case "integer":
		switch dialectName {
		case "sqlite":
			return "INTEGER"
		case "mysql", "postgres":
			return "INT"
		default:
			return "INTEGER"
		}
	case "timestamp":
		switch dialectName {
		case "sqlite":
			return "DATETIME"
		case "mysql", "postgres":
			return "TIMESTAMP"
		default:
			return "DATETIME"
		}
	case "boolean":
		switch dialectName {
		case "sqlite":
			return "BOOLEAN"
		case "mysql":
			return "TINYINT(1)"
		case "postgres":
			return "BOOLEAN"
		default:
			return "BOOLEAN"
		}
	case "text":
		return "TEXT"
	case "decimal":
		precision := col.Precision
		scale := col.Scale
		if precision == 0 {
			precision = 10
			scale = 2
		}
		return fmt.Sprintf("DECIMAL(%d,%d)", precision, scale)
	case "date":
		return "DATE"
	case "json":
		switch dialectName {
		case "sqlite":
			return "TEXT"
		case "mysql":
			return "JSON"
		case "postgres":
			return "JSONB"
		default:
			return "TEXT"
		}
	default:
		return "TEXT"
	}
}

