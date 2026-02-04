package slots

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"zeno/pkg/engine"

	"github.com/stretchr/testify/assert"
)

func TestNetworkSlots(t *testing.T) {
	eng := engine.NewEngine()
	RegisterNetworkSlots(eng)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok": true}`))
	}))
	defer ts.Close()

	t.Run("http.fetch", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "http.fetch",
			Value: ts.URL,
			Children: []*engine.Node{
				{Name: "as", Value: "$res"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		resRaw, ok := scope.Get("res")
		assert.True(t, ok)
		res := resRaw.(map[string]interface{})
		assert.Equal(t, true, res["ok"])
	})
}
