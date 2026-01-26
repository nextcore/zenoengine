package dbmanager

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite"
)

// DBManager mengelola multiple database connections
type DBManager struct {
	mu          sync.RWMutex
	connections map[string]*sql.DB
	dialects    map[string]Dialect
	defaultName string
}

// NewDBManager membuat instance DBManager baru
func NewDBManager() *DBManager {
	return &DBManager{
		connections: make(map[string]*sql.DB),
		dialects:    make(map[string]Dialect),
		defaultName: "default",
	}
}

// AddConnection menambahkan koneksi database baru
// driverName: nama driver database (contoh: "mysql", "sqlite")
// name: nama identifier untuk koneksi (contoh: "default", "warehouse", "analytics")
// dsn: Data Source Name (format tergantung driver)
func (m *DBManager) AddConnection(name, driverName, dsn string, maxOpen, maxIdle int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Cek apakah sudah ada
	if _, exists := m.connections[name]; exists {
		return fmt.Errorf("database connection '%s' already exists", name)
	}

	// Buka koneksi
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return fmt.Errorf("failed to open database '%s': %w", name, err)
	}

	// Ping untuk validasi koneksi
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database '%s': %w", name, err)
	}

	// Optimize connection pool settings for performance
	m.optimizePool(db, maxOpen, maxIdle)

	m.connections[name] = db
	m.dialects[name] = GetDialect(driverName)
	return nil
}

// GetConnection mengambil koneksi database berdasarkan nama
// Mengembalikan nil jika tidak ditemukan
func (m *DBManager) GetConnection(name string) *sql.DB {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connections[name]
}

// GetDialect mengambil dialect database berdasarkan nama koneksi
func (m *DBManager) GetDialect(name string) Dialect {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.dialects[name]
}

// GetDefault mengambil koneksi database default (primary)
func (m *DBManager) GetDefault() (*sql.DB, Dialect) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connections[m.defaultName], m.dialects[m.defaultName]
}

// SetDefault mengubah nama koneksi default
func (m *DBManager) SetDefault(name string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, exists := m.connections[name]; !exists {
		return fmt.Errorf("database connection '%s' not found", name)
	}

	m.defaultName = name
	return nil
}

// GetConnectionNames mengembalikan daftar semua nama koneksi
func (m *DBManager) GetConnectionNames() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.connections))
	for name := range m.connections {
		names = append(names, name)
	}
	return names
}

// optimizePool configures optimal connection pool settings
// Based on benchmarking and best practices for high-performance applications
func (m *DBManager) optimizePool(db *sql.DB, maxOpen, maxIdle int) {
	// Use provided values or defaults
	if maxOpen == 0 {
		maxOpen = 100 // Default: 100 concurrent connections
	}
	if maxIdle == 0 {
		maxIdle = 25 // Default: 25 idle connections (25% of max)
	}

	// Set maximum number of open connections
	// Higher values allow more concurrent queries but use more resources
	db.SetMaxOpenConns(maxOpen)

	// Set maximum number of idle connections
	// Keeping connections warm reduces connection overhead
	// Rule of thumb: 20-30% of MaxOpenConns
	db.SetMaxIdleConns(maxIdle)

	// Set maximum lifetime of a connection (5 minutes)
	// Prevents stale connections and helps with load balancing
	db.SetConnMaxLifetime(5 * 60 * 1000000000) // 5 minutes in nanoseconds

	// Set maximum idle time for a connection (1 minute)
	// Closes idle connections faster to free resources
	db.SetConnMaxIdleTime(1 * 60 * 1000000000) // 1 minute in nanoseconds
}

// Close menutup semua koneksi database
func (m *DBManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	for name, db := range m.connections {
		if err := db.Close(); err != nil {
			lastErr = fmt.Errorf("failed to close database '%s': %w", name, err)
		}
	}

	return lastErr
}
