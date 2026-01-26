package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

// SecurityHeaders adds "Helmet"-like security headers
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// --- HEADER STANDAR (Selalu Aktif) ---
		// Header ini aman dan tidak akan merusak tampilan/fungsi di dev mode
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// --- PRODUCTION ONLY HEADERS (HSTS) ---
		if os.Getenv("APP_ENV") == "production" {
			// 1. Strict Transport Security (Force HTTPS)
			w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		}

		// --- OPTIONAL CSP (Disabled by Default) ---
		// Enable only if CSP_ENABLED is "true" or "enable" in .env
		cspEnv := os.Getenv("CSP_ENABLED")
		if cspEnv == "true" || cspEnv == "enable" {
			// 2. Content Security Policy (CSP)
			// Kita buat dinamis agar production tetap fleksibel via .env
			defaultDomains := "https://cdn.jsdelivr.net https://cdnjs.cloudflare.com https://fonts.googleapis.com https://fonts.gstatic.com"

			// Tambahan domain dari .env (opsional)
			extraDomains := os.Getenv("CSP_ALLOWED_DOMAINS")
			if extraDomains != "" {
				defaultDomains = defaultDomains + " " + extraDomains
			}

			// Racik string CSP
			// Kita masukkan allowed domains ke script, style, font, dan connect
			csp := fmt.Sprintf(
				"default-src 'self'; "+
					"script-src 'self' 'unsafe-inline' 'unsafe-eval' %s; "+
					"style-src 'self' 'unsafe-inline' %s; "+
					"font-src 'self' %s; "+
					"img-src 'self' data: https:; "+
					"connect-src 'self' ws: wss: %s;", // ws: wss: penting untuk hot reload jika production pakai websocket
				defaultDomains, // script-src
				defaultDomains, // style-src
				defaultDomains, // font-src
				defaultDomains, // connect-src
			)

			w.Header().Set("Content-Security-Policy", strings.TrimSpace(csp))
		}

		next.ServeHTTP(w, r)
	})
}
