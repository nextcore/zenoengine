package slots

import (
	"context"
	"testing"
	"zeno/pkg/dbmanager"
	"zeno/pkg/engine"

	"github.com/stretchr/testify/assert"
)

func TestORMSlots(t *testing.T) {
	// Setup DB
	dbMgr := dbmanager.NewDBManager()
	err := dbMgr.AddConnection("default", "sqlite", ":memory:", 1, 1)
	if err != nil {
		t.Fatalf("Failed to create in-memory db: %v", err)
	}
	defer dbMgr.Close()

	// Seed DB
	db := dbMgr.GetConnection("default")
	_, err = db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}
	_, err = db.Exec(`INSERT INTO users (name, email) VALUES (?, ?)`, "Alice", "alice@example.com")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	eng := engine.NewEngine()
	// Register dependencies
	RegisterDBSlots(eng, dbMgr)
	RegisterRawDBSlots(eng, dbMgr)
	RegisterORMSlots(eng, dbMgr)

	t.Run("orm.model sets query state", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "orm.model",
			Value: "users",
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		qsRaw, ok := scope.Get("_query_state")
		assert.True(t, ok)
		qs := qsRaw.(*QueryState)
		assert.Equal(t, "users", qs.Table)

		modelName, _ := scope.Get("_active_model")
		assert.Equal(t, "users", modelName)
	})

	t.Run("orm.find", func(t *testing.T) {
		scope := engine.NewScope(nil)
		// Set model first
		eng.Execute(context.Background(), &engine.Node{Name: "orm.model", Value: "users"}, scope)

		node := &engine.Node{
			Name: "orm.find",
			Value: 1,
			Children: []*engine.Node{
				{Name: "as", Value: "$u"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		uRaw, ok := scope.Get("u")
		assert.True(t, ok)
		u := uRaw.(map[string]interface{})
		assert.Equal(t, "Alice", u["name"])
	})

	t.Run("orm.save insert", func(t *testing.T) {
		scope := engine.NewScope(nil)
		eng.Execute(context.Background(), &engine.Node{Name: "orm.model", Value: "users"}, scope)

		data := map[string]interface{}{
			"name": "Bob",
			"email": "bob@example.com",
		}
		scope.Set("new_user", data)

		node := &engine.Node{
			Name: "orm.save",
			Value: "$new_user",
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		// Check if inserted
		// Manually check via DB or slot
		// Let's use db.count
		scope.Set("total", 0)
		eng.Execute(context.Background(), &engine.Node{Name: "db.count", Children: []*engine.Node{{Name: "as", Value: "$total"}}}, scope)
		total, _ := scope.Get("total")
		assert.Equal(t, 2, total)
	})

	t.Run("orm.save update", func(t *testing.T) {
		scope := engine.NewScope(nil)
		eng.Execute(context.Background(), &engine.Node{Name: "orm.model", Value: "users"}, scope)

		data := map[string]interface{}{
			"id": 1,
			"name": "Alice Updated",
		}
		scope.Set("update_user", data)

		node := &engine.Node{
			Name: "orm.save",
			Value: "$update_user",
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		// Verify update
		eng.Execute(context.Background(), &engine.Node{
			Name: "orm.find",
			Value: 1,
			Children: []*engine.Node{{Name: "as", Value: "$u2"}}},
			scope)

		u2Raw, _ := scope.Get("u2")
		u2 := u2Raw.(map[string]interface{})
		assert.Equal(t, "Alice Updated", u2["name"])
	})
}
