package slots

import (
	"context"
	"net/http/httptest"
	"testing"
	"zeno/pkg/engine"
	"zeno/pkg/middleware"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIPSecuritySlots(t *testing.T) {
	eng := engine.NewEngine()
	RegisterIPSecuritySlots(eng)

	t.Run("sec.block_ip", func(t *testing.T) {
		ip := "5.5.5.5"
		defer middleware.GlobalBlockList.Remove(ip)

		code := `sec.block_ip: "` + ip + `"`
		root, err := engine.ParseString(code, "test.zl")
		require.NoError(t, err)

		err = eng.Execute(context.Background(), root, engine.GetScope())
		require.NoError(t, err)

		assert.True(t, middleware.GlobalBlockList.IsBlocked(ip))
	})

	t.Run("sec.block_ip_current_request", func(t *testing.T) {
		ip := "6.6.6.6"
		defer middleware.GlobalBlockList.Remove(ip)

		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = ip + ":1234"

		ctx := context.WithValue(context.Background(), "httpRequest", req)

		code := `sec.block_ip` // Implicit from request
		root, err := engine.ParseString(code, "test.zl")
		require.NoError(t, err)

		err = eng.Execute(ctx, root, engine.GetScope())
		require.NoError(t, err)

		assert.True(t, middleware.GlobalBlockList.IsBlocked(ip))
	})

	t.Run("sec.is_blocked_and_unblock", func(t *testing.T) {
		ip := "7.7.7.7"
		middleware.GlobalBlockList.Add(ip)
		defer middleware.GlobalBlockList.Remove(ip)

		code := `
		sec.is_blocked: "` + ip + `" { as: $status }
		sec.unblock_ip: "` + ip + `"
		sec.is_blocked: "` + ip + `" { as: $status2 }
		`
		root, err := engine.ParseString(code, "test.zl")
		require.NoError(t, err)

		scope := engine.GetScope()
		err = eng.Execute(context.Background(), root, scope)
		require.NoError(t, err)

		val1, _ := scope.Get("status")
		val2, _ := scope.Get("status2")

		assert.Equal(t, true, val1)
		assert.Equal(t, false, val2)
	})
}
