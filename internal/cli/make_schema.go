package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gosimple/slug"
)

// HandleMakeSchema generates a new database migration/schema file.
// Usage: zeno make:model <ModelName>
// Alias: zeno make:schema <TableName>
func HandleMakeSchema(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: zeno make:model <ModelName>")
		os.Exit(1)
	}

	name := args[0]
	// Convert ModelName (User) to table_name (users)
	tableName := slug.Make(name)
	tableName = strings.ReplaceAll(tableName, "-", "_")
	// Pluralize? Simple logic: add 's' if not present
	if !strings.HasSuffix(tableName, "s") {
		tableName += "s"
	}

	timestamp := time.Now().Format("2006_01_02_150405")
	fileName := fmt.Sprintf("%s_create_%s_table.zl", timestamp, tableName)

	dir := "database/migrations"
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("❌ Failed to create directory: %v\n", err)
		os.Exit(1)
	}

	fullPath := filepath.Join(dir, fileName)

	// Template for migration
	content := fmt.Sprintf(`
// Migration: Create %s table
schema.create: "%s" {
  column.id
  column.string: "name"
  column.timestamps
}

log: "Migrated table: %s"
`, tableName, tableName, tableName)

	if err := os.WriteFile(fullPath, []byte(strings.TrimSpace(content)), 0644); err != nil {
		fmt.Printf("❌ Failed to write file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Created Migration: %s\n", fullPath)
}
