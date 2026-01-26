package migrator

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"zeno/pkg/dbmanager"
	"zeno/pkg/engine"
)

type Migrator struct {
	Engine  *engine.Engine
	DB      *sql.DB
	Dialect dbmanager.Dialect
	Dir     string
}

func New(eng *engine.Engine, dbMgr *dbmanager.DBManager, dir string) *Migrator {
	db, dialect := dbMgr.GetDefault()
	return &Migrator{
		Engine:  eng,
		DB:      db,
		Dialect: dialect,
		Dir:     dir,
	}
}

func (m *Migrator) Run() error {
	ctx := context.Background()

	// 1. Pastikan tabel tracking ada
	queryInit := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := m.DB.ExecContext(ctx, queryInit); err != nil {
		return fmt.Errorf("failed to init migration table: %w", err)
	}

	// 2. Ambil migrasi yang sudah diaplikasikan
	rows, err := m.DB.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("failed to fetch applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var ver string
		rows.Scan(&ver)
		applied[ver] = true
	}

	// 3. Baca file migrasi dari folder
	files, err := os.ReadDir(m.Dir)
	if err != nil {
		return fmt.Errorf("failed to read migration dir: %w", err)
	}

	var pending []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".zl") {
			pending = append(pending, f.Name())
		}
	}
	sort.Strings(pending) // Pastikan urut (001, 002, ...)

	// 4. Eksekusi yang belum diaplikasikan
	count := 0
	for _, filename := range pending {
		if applied[filename] {
			continue // Skip jika sudah
		}

		slog.Info("🚀 Migrating...", "file", filename)

		fullPath := filepath.Join(m.Dir, filename)
		root, err := engine.LoadScript(fullPath)
		if err != nil {
			return fmt.Errorf("failed to parse migration '%s': %w", filename, err)
		}

		scope := engine.NewScope(nil)
		scope.Set("migration_ver", filename)

		// Jalankan Script Zenolang
		if err := m.Engine.Execute(ctx, root, scope); err != nil {
			return fmt.Errorf("failed to execute migration '%s': %w", filename, err)
		}

		// Catat ke DB
		insertQuery := fmt.Sprintf("INSERT INTO schema_migrations (version) VALUES (%s)", m.Dialect.Placeholder(1))
		_, err = m.DB.ExecContext(ctx, insertQuery, filename)
		if err != nil {
			return fmt.Errorf("failed to record migration '%s': %w", filename, err)
		}

		slog.Info("✅ Applied", "file", filename)
		count++
	}

	if count == 0 {
		slog.Info("✨ Database is up to date.")
	} else {
		slog.Info("🎉 Migration Complete", "applied", count)
	}

	return nil
}
