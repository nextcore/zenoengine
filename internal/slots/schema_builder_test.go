package slots

import (
	"context"
	"testing"
	"zeno/pkg/dbmanager"
	"zeno/pkg/engine"

	"github.com/stretchr/testify/assert"
)

func TestSchemaSlots(t *testing.T) {
	dbMgr := dbmanager.NewDBManager()
	err := dbMgr.AddConnection("default", "sqlite", ":memory:", 1, 1)
	if err != nil {
		t.Fatalf("Failed to create in-memory db: %v", err)
	}
	defer dbMgr.Close()

	eng := engine.NewEngine()
	RegisterSchemaSlots(eng, dbMgr)
	RegisterRawDBSlots(eng, dbMgr) // For verification

	t.Run("schema.create table", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "schema.create",
			Value: "products",
			Children: []*engine.Node{
				{Name: "column.id", Value: "id"},
				{Name: "column.string", Value: "name"},
				{Name: "column.integer", Value: "price"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		// Verify table exists
		// In SQLite: SELECT name FROM sqlite_master WHERE type='table' AND name='products'
		scope.Set("check_query", "SELECT name FROM sqlite_master WHERE type='table' AND name='products'")
		eng.Execute(context.Background(), &engine.Node{
			Name: "db.select",
			Value: "$check_query",
			Children: []*engine.Node{{Name: "as", Value: "$res"}, {Name: "first", Value: true}},
		}, scope)

		resRaw, _ := scope.Get("res")
		assert.NotNil(t, resRaw)
		res := resRaw.(map[string]interface{})
		assert.Equal(t, "products", res["name"])
	})
}
