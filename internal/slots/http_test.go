package slots

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"zeno/pkg/engine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPServerSlots(t *testing.T) {
	eng := engine.NewEngine()
	RegisterHTTPServerSlots(eng)

	t.Run("http.response", func(t *testing.T) {
		rec := httptest.NewRecorder()
		ctx := context.WithValue(context.Background(), "httpWriter", http.ResponseWriter(rec))

		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "http.response",
			Children: []*engine.Node{
				{Name: "status", Value: 201},
				{Name: "body", Value: map[string]interface{}{"msg": "hello"}},
			},
		}

		err := eng.Execute(ctx, node, scope)
		require.NoError(t, err)

		assert.Equal(t, 201, rec.Code)
		assert.JSONEq(t, `{"msg": "hello"}`, rec.Body.String())
	})

	t.Run("http.redirect", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		ctx := context.WithValue(context.Background(), "httpWriter", http.ResponseWriter(rec))
		ctx = context.WithValue(ctx, "httpRequest", req)

		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "http.redirect",
			Value: "/new-location",
		}

		err := eng.Execute(ctx, node, scope)
		// Redirect returns ErrReturn to stop execution, checking if that's the case
		// Actually looking at code: return ErrReturn
		assert.Error(t, err)
		assert.Equal(t, "return", err.Error()) // ErrReturn usually has "return" or similar

		assert.Equal(t, http.StatusFound, rec.Code)
		assert.Equal(t, "/new-location", rec.Header().Get("Location"))
	})

	t.Run("cookie.set", func(t *testing.T) {
		rec := httptest.NewRecorder()
		ctx := context.WithValue(context.Background(), "httpWriter", http.ResponseWriter(rec))

		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "cookie.set",
			Children: []*engine.Node{
				{Name: "name", Value: "session"},
				{Name: "val", Value: "123"},
			},
		}

		err := eng.Execute(ctx, node, scope)
		require.NoError(t, err)

		cookies := rec.Result().Cookies()
		require.Len(t, cookies, 1)
		assert.Equal(t, "session", cookies[0].Name)
		assert.Equal(t, "123", cookies[0].Value)
	})

	t.Run("http.query", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/?page=2", nil)
		ctx := context.WithValue(context.Background(), "httpRequest", req)

		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "http.query",
			Value: "page",
			Children: []*engine.Node{
				{Name: "as", Value: "$p"},
			},
		}

		err := eng.Execute(ctx, node, scope)
		require.NoError(t, err)

		val, ok := scope.Get("p")
		assert.True(t, ok)
		assert.Equal(t, "2", val)
	})

	t.Run("http.header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-API-Key", "secret")
		ctx := context.WithValue(context.Background(), "httpRequest", req)

		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "http.header",
			Value: "X-API-Key",
			Children: []*engine.Node{
				{Name: "as", Value: "$key"},
			},
		}

		err := eng.Execute(ctx, node, scope)
		require.NoError(t, err)

		val, ok := scope.Get("key")
		assert.True(t, ok)
		assert.Equal(t, "secret", val)
	})

	t.Run("http.ok helper", func(t *testing.T) {
		rec := httptest.NewRecorder()
		ctx := context.WithValue(context.Background(), "httpWriter", http.ResponseWriter(rec))

		scope := engine.NewScope(nil)
		scope.Set("data", "some data")

		node := &engine.Node{
			Name: "http.ok",
			Children: []*engine.Node{
				{Name: "result", Value: "$data"},
			},
		}

		err := eng.Execute(ctx, node, scope)
		require.NoError(t, err)

		assert.Equal(t, 200, rec.Code)

		var body map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &body)
		assert.Equal(t, true, body["success"])
		assert.Equal(t, "some data", body["result"])
	})
}
