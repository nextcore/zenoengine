package middleware

import (
	"net"
	"net/http"
	"zeno/pkg/host"
)

// HostDispatcher performs an O(1) lookup to route requests based on the Host header.
// This is much more efficient than linear middleware checks for environments with many domains.
func HostDispatcher(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqHost := r.Host

		// Handle host with port (e.g. localhost:3000)
		if h, _, err := net.SplitHostPort(reqHost); err == nil {
			reqHost = h
		}

		// O(1) lookup in the native host map
		if handler, ok := host.GlobalManager.GetHandler(reqHost); ok {
			handler.ServeHTTP(w, r)
			return
		}

		// Fallback to default router
		next.ServeHTTP(w, r)
	})
}
