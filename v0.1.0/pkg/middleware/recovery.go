package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
)

// Recoverer adalah middleware custom untuk ZenoEngine
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				// 1. Ambil Stack Trace
				stack := string(debug.Stack())

				// 2. Log Error (Gunakan Slog)
				slog.Error("🔥 PANIC RECOVERED",
					"error", rvr,
					"path", r.URL.Path,
					"method", r.Method,
				)

				// 3. Tentukan Tipe Respon (JSON atau HTML)
				accept := r.Header.Get("Accept")
				isJSON := strings.Contains(accept, "application/json")

				// 4. Siapkan Pesan Error
				errTitle := "Internal Server Error"
				errMsg := fmt.Sprintf("%v", rvr)

				// 5. Cek Environment
				env := os.Getenv("APP_ENV")
				if env == "" {
					env = "production"
				}

				// ==========================================
				// RESPON JSON (API)
				// ==========================================
				if isJSON {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)

					jsonBody := fmt.Sprintf(`{"status":500,"error":"%s"`, errTitle)
					if env == "development" {
						// Di Dev, kirim detail error + stack trace (disederhanakan)
						safeStack := strings.ReplaceAll(stack, "\"", "'")
						safeStack = strings.ReplaceAll(safeStack, "\n", "\\n")
						safeStack = strings.ReplaceAll(safeStack, "\t", "\\t")
						jsonBody += fmt.Sprintf(`,"detail":"%s","stack":"%s"`, errMsg, safeStack)
					}
					jsonBody += "}"
					w.Write([]byte(jsonBody))
					return
				}

				// ==========================================
				// RESPON HTML (BROWSER)
				// ==========================================
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusInternalServerError)

				if env == "development" {
					// MODE DEV: Tampilkan Error Whoops-style
					html := fmt.Sprintf(`
					<html>
					<head>
						<title>Zeno Error</title>
						<style>
							body { font-family: sans-serif; background: #fce4e4; color: #333; padding: 20px; }
							.container { background: white; padding: 30px; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); max-width: 900px; margin: 0 auto; }
							h1 { color: #d32f2f; margin-top: 0; }
							.error-box { background: #ffebee; border-left: 5px solid #d32f2f; padding: 15px; font-family: monospace; font-size: 1.1em; margin-bottom: 20px; }
							.stack { background: #263238; color: #eceff1; padding: 15px; border-radius: 4px; overflow-x: auto; white-space: pre; font-size: 0.85em; }
							.meta { margin-bottom: 10px; color: #666; }
						</style>
					</head>
					<body>
						<div class="container">
							<h1>🔥 Runtime Error</h1>
							<div class="meta">Location: %s %s</div>
							<div class="error-box">%s</div>

							<h3>Stack Trace</h3>
							<div class="stack">%s</div>
						</div>
					</body>
					</html>
					`, r.Method, r.URL.Path, errMsg, stack)
					w.Write([]byte(html))
				} else {
					// MODE PROD: Tampilkan Halaman 500 Cantik
					html := `
					<html>
					<head>
						<title>500 Internal Server Error</title>
						<style>
							body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif; background: #f8f9fa; display: flex; align-items: center; justify-content: center; height: 100vh; margin: 0; }
							.container { text-align: center; }
							h1 { font-size: 6rem; margin: 0; color: #dee2e6; }
							h2 { color: #343a40; margin-bottom: 1rem; }
							p { color: #6c757d; }
							.btn { display: inline-block; margin-top: 20px; padding: 10px 20px; background: #007bff; color: white; text-decoration: none; border-radius: 5px; }
						</style>
					</head>
					<body>
						<div class="container">
							<h1>500</h1>
							<h2>Terjadi Kesalahan Server</h2>
							<p>Maaf, kami sedang mengalami gangguan teknis. Silakan coba beberapa saat lagi.</p>
							<a href="/" class="btn">Kembali ke Beranda</a>
						</div>
					</body>
					</html>
					`
					w.Write([]byte(html))
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}
