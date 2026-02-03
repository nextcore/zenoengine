package slots

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"zeno/pkg/engine"

	"github.com/stretchr/testify/assert"
)

func TestInertiaSlots(t *testing.T) {
	eng := engine.NewEngine()
	RegisterInertiaSlots(eng)

	t.Run("inertia.render HTML", func(t *testing.T) {
		scope := engine.NewScope(nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/dashboard", nil)

		ctx := context.Background()
		ctx = context.WithValue(ctx, "httpWriter", http.ResponseWriter(rec))
		ctx = context.WithValue(ctx, "httpRequest", req)

		node := &engine.Node{
			Name: "inertia.render",
			Children: []*engine.Node{
				{Name: "component", Value: "Dashboard"},
				{Name: "props", Value: map[string]interface{}{"user": "Alice"}},
			},
		}

		err := eng.Execute(ctx, node, scope)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "<!DOCTYPE html>")
		assert.Contains(t, rec.Body.String(), `data-page='`)
	})

	t.Run("inertia.render JSON (X-Inertia)", func(t *testing.T) {
		scope := engine.NewScope(nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/dashboard", nil)
		req.Header.Set("X-Inertia", "true")

		ctx := context.Background()
		ctx = context.WithValue(ctx, "httpWriter", http.ResponseWriter(rec))
		ctx = context.WithValue(ctx, "httpRequest", req)

		node := &engine.Node{
			Name: "inertia.render",
			Children: []*engine.Node{
				{Name: "component", Value: "Dashboard"},
				{Name: "props", Value: map[string]interface{}{"user": "Bob"}},
			},
		}

		err := eng.Execute(ctx, node, scope)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var page map[string]interface{}
		json.NewDecoder(rec.Body).Decode(&page)
		assert.Equal(t, "Dashboard", page["component"])
		props := page["props"].(map[string]interface{})
		assert.Equal(t, "Bob", props["user"])
	})

	t.Run("inertia.share", func(t *testing.T) {
		scope := engine.NewScope(nil)

		node := &engine.Node{
			Name: "inertia.share",
			Children: []*engine.Node{
				{Name: "app_name", Value: "Zeno"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		val, ok := scope.Get("app_name")
		assert.True(t, ok)
		assert.Equal(t, "Zeno", val)
	})

	t.Run("inertia.location", func(t *testing.T) {
		scope := engine.NewScope(nil)
		rec := httptest.NewRecorder()
		ctx := context.WithValue(context.Background(), "httpWriter", http.ResponseWriter(rec))

		node := &engine.Node{
			Name: "inertia.location",
			Children: []*engine.Node{
				{Name: "url", Value: "/login"},
			},
		}

		err := eng.Execute(ctx, node, scope)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusConflict, rec.Code)
		assert.Equal(t, "/login", rec.Header().Get("X-Inertia-Location"))
	})
}
