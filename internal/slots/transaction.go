package slots

import (
	"context"
	"fmt"
	"zeno/pkg/dbmanager"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

func RegisterTransactionSlots(eng *engine.Engine, dbMgr *dbmanager.DBManager) {

	// DB.TRANSACTION
	eng.Register("db.transaction", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		// 1. Tentukan Nama Database
		dbName := "default"

		// Support: db.transaction: "analytics" (atau $dbname)
		if node.Value != nil {
			dbName = coerce.ToString(resolveValue(node.Value, scope))
		}

		var doNode *engine.Node

		for _, c := range node.Children {
			if c.Name == "db" {
				dbName = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "do" {
				doNode = c
			}
		}

		// 2. Ambil Koneksi
		db := dbMgr.GetConnection(dbName)
		if db == nil {
			return fmt.Errorf("db.transaction: database '%s' not found", dbName)
		}

		// 3. Mulai Transaksi
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		// Simpan state transaksi lama (untuk support nested transaction di masa depan)
		// Saat ini kita replace saja.
		scope.Set("_active_tx", tx)

		// Defer cleanup: Pastikan _active_tx dihapus setelah blok selesai
		defer scope.Set("_active_tx", nil)

		// 4. Eksekusi Blok
		var execErr error

		// Jika ada node 'do', eksekusi isinya.
		// Jika tidak, eksekusi children langsung (shorthand).
		nodesToExec := node.Children
		if doNode != nil {
			nodesToExec = doNode.Children
		}

		for _, child := range nodesToExec {
			// Skip node konfigurasi 'db' atau 'do' wrapper saat looping direct children
			if doNode == nil && (child.Name == "db" || child.Name == "do") {
				continue
			}

			if err := eng.Execute(ctx, child, scope); err != nil {
				execErr = err
				break
			}
		}

		// 5. Commit atau Rollback
		if execErr != nil {
			// Jika ada error di script user, batalkan semua perubahan DB
			tx.Rollback()
			return execErr
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		return nil
	}, engine.SlotMeta{
		Description: "Menjalankan blok kode dalam database transaction (ACID).",
		Example:     "db.transaction\n  do:\n    db.insert: ...\n    db.update: ...",
	})
}
