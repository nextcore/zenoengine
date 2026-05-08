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
		batch INTEGER,
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

	// Hitung batch berikutnya
	var currentBatch int
	err = m.DB.QueryRowContext(ctx, "SELECT COALESCE(MAX(batch), 0) FROM schema_migrations").Scan(&currentBatch)
	if err != nil {
		return fmt.Errorf("failed to calculate next batch: %w", err)
	}
	nextBatch := currentBatch + 1

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
		versionKey := filepath.Join(m.Dir, filename)
		if applied[versionKey] {
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
		insertQuery := fmt.Sprintf("INSERT INTO schema_migrations (version, batch) VALUES (%s, %s)", m.Dialect.Placeholder(1), m.Dialect.Placeholder(2))
		_, err = m.DB.ExecContext(ctx, insertQuery, filepath.Join(m.Dir, filename), nextBatch)
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

func (m *Migrator) Rollback() error {
	ctx := context.Background()

	// 1. Ambil batch terakhir
	var lastBatch int
	err := m.DB.QueryRowContext(ctx, "SELECT COALESCE(MAX(batch), 0) FROM schema_migrations").Scan(&lastBatch)
	if err != nil {
		return fmt.Errorf("failed to get last batch: %w", err)
	}

	if lastBatch == 0 {
		slog.Info("✨ No migrations to rollback.")
		return nil
	}

	// 2. Ambil semua migrasi dari batch terakhir (urut terbalik)
	query := fmt.Sprintf("SELECT version FROM schema_migrations WHERE batch = %s ORDER BY version DESC", m.Dialect.Placeholder(1))
	rows, err := m.DB.QueryContext(ctx, query, lastBatch)
	if err != nil {
		return fmt.Errorf("failed to fetch migrations for rollback: %w", err)
	}
	defer rows.Close()

	var toRollback []string
	for rows.Next() {
		var ver string
		rows.Scan(&ver)
		toRollback = append(toRollback, ver)
	}

	slog.Info("🔄 Rolling back...", "batch", lastBatch, "count", len(toRollback))

	for _, versionKey := range toRollback {
		// Parse file
		filename := filepath.Base(versionKey)
		root, err := engine.LoadScript(versionKey)
		if err != nil {
			return fmt.Errorf("failed to parse migration '%s': %w", filename, err)
		}

		scope := engine.NewScope(nil)
		scope.Set("migration_ver", filename)
		scope.Set("_migration_direction", "down") // Set direction to down!

		// Jalankan Script Zenolang
		if err := m.Engine.Execute(ctx, root, scope); err != nil {
			return fmt.Errorf("failed to execute rollback for '%s': %w", filename, err)
		}

		// Hapus dari DB
		deleteQuery := fmt.Sprintf("DELETE FROM schema_migrations WHERE version = %s", m.Dialect.Placeholder(1))
		_, err = m.DB.ExecContext(ctx, deleteQuery, versionKey)
		if err != nil {
			return fmt.Errorf("failed to delete migration record '%s': %w", filename, err)
		}

		slog.Info("Rolled back", "file", filename)
	}

	slog.Info("🎉 Rollback Complete", "batch", lastBatch)
	return nil
}
