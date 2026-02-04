package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIPBlocker(t *testing.T) {
	// Setup
	handler := IPBlocker(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Allowed"))
	}))

	t.Run("Blocks_Env_IP", func(t *testing.T) {
		os.Setenv("BLOCKED_IPS", "10.0.0.1, 10.0.0.2")
		GlobalBlockList.Load() // Reload to pick up env
		defer func() {
			os.Unsetenv("BLOCKED_IPS")
			GlobalBlockList.Remove("10.0.0.1")
			GlobalBlockList.Remove("10.0.0.2")
		}()

		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("Allows_Clean_IP", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Dynamic_Blocking", func(t *testing.T) {
		ip := "1.2.3.4"
		GlobalBlockList.Add(ip)
		defer GlobalBlockList.Remove(ip)

		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = ip + ":5555"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}
