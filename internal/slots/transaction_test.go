package slots

import (
	"context"
	"testing"
	"zeno/pkg/dbmanager"
	"zeno/pkg/engine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionSlots(t *testing.T) {
	// Setup DB Manager with In-Memory SQLite
	dbMgr := dbmanager.NewDBManager()
	err := dbMgr.AddConnection("default", "sqlite", ":memory:", 1, 1)
	require.NoError(t, err)
	defer dbMgr.Close()

	// Initialize DB Schema
	db := dbMgr.GetConnection("default")
	_, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
	require.NoError(t, err)

	// Setup Engine
	eng := engine.NewEngine()
	RegisterTransactionSlots(eng, dbMgr)
	RegisterRawDBSlots(eng, dbMgr)

	t.Run("db.transaction commit success", func(t *testing.T) {
		// Clear table
		_, _ = db.Exec("DELETE FROM users")

		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "db.transaction",
			Children: []*engine.Node{
				{
					Name: "do",
					Children: []*engine.Node{
						{
							Name: "db.execute",
							Value: "INSERT INTO users (name) VALUES ('alice')",
						},
						{
							Name: "db.execute",
							Value: "INSERT INTO users (name) VALUES ('bob')",
						},
					},
				},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		require.NoError(t, err)

		// Verify data
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})

	t.Run("db.transaction rollback on error", func(t *testing.T) {
		// Clear table
		_, _ = db.Exec("DELETE FROM users")
		_, _ = db.Exec("INSERT INTO users (name) VALUES ('existing')")

		scope := engine.NewScope(nil)

		// Register a slot that always fails
		eng.Register("fail_now", func(ctx context.Context, n *engine.Node, s *engine.Scope) error {
			return assert.AnError
		}, engine.SlotMeta{})

		node := &engine.Node{
			Name: "db.transaction",
			Children: []*engine.Node{
				{
					Name: "do",
					Children: []*engine.Node{
						{
							Name: "db.execute",
							Value: "INSERT INTO users (name) VALUES ('should_rollback')",
						},
						{
							Name: "fail_now", // Trigger error
						},
					},
				},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.Error(t, err)

		// Verify rollback: 'should_rollback' should NOT exist
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count) // Only 'existing' remains
	})
}
