package dbmanager

import (
	"fmt"
	"strings"
)

// Dialect abstraction for different SQL databases
type Dialect interface {
	Name() string
	QuoteIdentifier(name string) string
	Placeholder(n int) string
	Limit(limit, offset int) string
}

// MySQLDialect implementation
type MySQLDialect struct{}

func (d MySQLDialect) Name() string { return "mysql" }

func (d MySQLDialect) QuoteIdentifier(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

func (d MySQLDialect) Placeholder(n int) string {
	return "?"
}

func (d MySQLDialect) Limit(limit, offset int) string {
	if limit > 0 {
		if offset > 0 {
			return fmt.Sprintf(" LIMIT %d, %d", offset, limit)
		}
		return fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		// MySQL hack for OFFSET without LIMIT
		return fmt.Sprintf(" LIMIT 18446744073709551615 OFFSET %d", offset)
	}
	return ""
}

// SQLiteDialect implementation
type SQLiteDialect struct{}

func (d SQLiteDialect) Name() string { return "sqlite" }

func (d SQLiteDialect) QuoteIdentifier(name string) string {
	return "\"" + strings.ReplaceAll(name, "\"", "\"\"") + "\""
}

func (d SQLiteDialect) Placeholder(n int) string {
	return "?"
}

func (d SQLiteDialect) Limit(limit, offset int) string {
	res := ""
	if limit > 0 {
		res += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		if limit <= 0 {
			res += " LIMIT -1" // SQLite requires LIMIT for OFFSET
		}
		res += fmt.Sprintf(" OFFSET %d", offset)
	}
	return res
}

// SQLServerDialect implementation
type SQLServerDialect struct{}

func (d SQLServerDialect) Name() string { return "sqlserver" }

func (d SQLServerDialect) QuoteIdentifier(name string) string {
	// SQL Server uses square brackets
	return "[" + strings.ReplaceAll(name, "]", "]]") + "]"
}

func (d SQLServerDialect) Placeholder(n int) string {
	// SQL Server uses @p1, @p2, @p3
	return fmt.Sprintf("@p%d", n)
}

func (d SQLServerDialect) Limit(limit, offset int) string {
	// SQL Server uses OFFSET-FETCH (SQL Server 2012+)
	res := ""
	if offset > 0 {
		res += fmt.Sprintf(" OFFSET %d ROWS", offset)
		if limit > 0 {
			res += fmt.Sprintf(" FETCH NEXT %d ROWS ONLY", limit)
		}
	} else if limit > 0 {
		res += fmt.Sprintf(" OFFSET 0 ROWS FETCH NEXT %d ROWS ONLY", limit)
	}
	return res
}

// PostgreSQLDialect implementation
type PostgreSQLDialect struct{}

func (d PostgreSQLDialect) Name() string { return "postgres" }

func (d PostgreSQLDialect) QuoteIdentifier(name string) string {
	// PostgreSQL uses double quotes
	return "\"" + strings.ReplaceAll(name, "\"", "\"\"") + "\""
}

func (d PostgreSQLDialect) Placeholder(n int) string {
	// PostgreSQL uses $1, $2, $3
	return fmt.Sprintf("$%d", n)
}

func (d PostgreSQLDialect) Limit(limit, offset int) string {
	// PostgreSQL uses standard LIMIT/OFFSET
	res := ""
	if limit > 0 {
		res += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		res += fmt.Sprintf(" OFFSET %d", offset)
	}
	return res
}

// GetDialect returns the appropriate dialect for the driver name
func GetDialect(driverName string) Dialect {
	switch strings.ToLower(driverName) {
	case "mysql":
		return MySQLDialect{}
	case "sqlite", "sqlite3":
		return SQLiteDialect{}
	case "postgres", "postgresql", "pgx":
		return PostgreSQLDialect{}
	case "sqlserver", "mssql":
		return SQLServerDialect{}
	default:
		// Default to MySQL-like behavior for unknown drivers
		return MySQLDialect{}
	}
}
