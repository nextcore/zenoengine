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

func TestHTTPClientSlots(t *testing.T) {
	eng := engine.NewEngine()
	RegisterHTTPClientSlots(eng)

	// Create a mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/json" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"message": "success"}`))
			return
		}
		if r.URL.Path == "/echo" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			json.NewEncoder(w).Encode(body)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	t.Run("http.request GET JSON", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "http.request",
			Value: ts.URL + "/json",
			Children: []*engine.Node{
				{Name: "method", Value: "GET"},
				{Name: "as", Value: "$res"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		resRaw, ok := scope.Get("res")
		assert.True(t, ok)
		res := resRaw.(map[string]interface{})

		assert.Equal(t, 200, res["status"])
		body := res["body"].(map[string]interface{})
		assert.Equal(t, "success", body["message"])
	})

	t.Run("http.request POST Body", func(t *testing.T) {
		scope := engine.NewScope(nil)
		scope.Set("payload", map[string]interface{}{"foo": "bar"})

		node := &engine.Node{
			Name: "http.request",
			Value: ts.URL + "/echo",
			Children: []*engine.Node{
				{Name: "method", Value: "POST"},
				{Name: "body", Value: "$payload"},
				{Name: "as", Value: "$res"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		resRaw, ok := scope.Get("res")
		assert.True(t, ok)
		res := resRaw.(map[string]interface{})

		assert.Equal(t, 200, res["status"])
		body := res["body"].(map[string]interface{})
		assert.Equal(t, "bar", body["foo"])
	})
}
