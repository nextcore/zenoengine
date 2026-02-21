package slots

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"zeno/pkg/dbmanager"
	"zeno/pkg/engine"

	"github.com/go-chi/chi/v5"
)

func TestRouteModelBinding(t *testing.T) {
	// 1. Setup Engine & DB
	eng := engine.NewEngine()
	dbMgr := dbmanager.NewDBManager()

	// Use In-Memory SQLite
	dbMgr.AddConnection("default", "sqlite", ":memory:", 1, 1)

	// Create Table & Seed
	db, _ := dbMgr.GetDefault()
	db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
	db.Exec("INSERT INTO users (name) VALUES ('Budi')")

	// Register Slots
	RegisterRouterSlots(eng, chi.NewRouter())
	RegisterDBSlots(eng, dbMgr)

	// 2. Define Route Script with Binding
	script := `
http.get: "/users/{user}" {
  bind: { user: "users" }
  do: {
    // If bound, $user should be a map, not a string ID
    log: $user.name
  }
}
`
	root, _ := engine.ParseString(script, "test.zl")

	// 3. Execute Route Registration
	router := chi.NewRouter()
	ctx := context.WithValue(context.Background(), routerKey{}, router)
	scope := engine.NewScope(nil)
	eng.Execute(ctx, root, scope)

	// 4. Simulate Request GET /users/1
	req := httptest.NewRequest("GET", "/users/1", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// 5. Verify Output
	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", rec.Code)
	}

	// 6. Simulate Not Found GET /users/999
	req404 := httptest.NewRequest("GET", "/users/999", nil)
	rec404 := httptest.NewRecorder()
	router.ServeHTTP(rec404, req404)

	if rec404.Code != http.StatusNotFound {
		t.Errorf("Expected 404 Not Found, got %d", rec404.Code)
	}
}
