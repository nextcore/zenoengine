package livereload

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// InjectMiddleware wraps the handler to inject live reload script in development mode
func InjectMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip injection in production
		if os.Getenv("APP_ENV") != "development" {
			next.ServeHTTP(w, r)
			return
		}

		// Create a response recorder to capture the output
		recorder := &responseRecorder{
			ResponseWriter: w,
			body:           &bytes.Buffer{},
			statusCode:     http.StatusOK,
			headers:        make(http.Header),
		}

		// Call the next handler
		next.ServeHTTP(recorder, r)

		// Copy headers from recorder to actual response
		for key, values := range recorder.headers {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		// Check if response is HTML
		contentType := recorder.headers.Get("Content-Type")
		bodySize := recorder.body.Len()

		// Debug logging
		fmt.Printf("[LIVERELOAD] Path: %s, Content-Type: %s, Body Size: %d\n", r.URL.Path, contentType, bodySize)

		// If Content-Type is empty, assume HTML (Blade doesn't set it)
		// If Content-Type is set but not HTML, skip injection
		if contentType != "" && !strings.Contains(contentType, "text/html") {
			// Not HTML, write as-is
			fmt.Printf("[LIVERELOAD] Skipping injection (not HTML)\n")
			w.WriteHeader(recorder.statusCode)
			w.Write(recorder.body.Bytes())
			return
		}

		// Inject live reload script before </body>
		body := recorder.body.String()
		injectedBody := injectScript(body)

		fmt.Printf("[LIVERELOAD] ✅ Injected script! Original: %d bytes, New: %d bytes\n", len(body), len(injectedBody))

		// Update Content-Length and write the modified response
		// Set Content-Type if not already set
		if contentType == "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		}
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(injectedBody)))
		w.WriteHeader(recorder.statusCode)
		w.Write([]byte(injectedBody))
	})
}

// responseRecorder captures the response
type responseRecorder struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
	headers    http.Header
}

func (r *responseRecorder) Header() http.Header {
	return r.headers
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	return r.body.Write(b)
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
}

// injectScript adds the live reload SSE script before </body>
func injectScript(html string) string {
	script := `
<script>
(function() {
	const source = new EventSource('/livereload');
	source.onmessage = function(e) {
		if (e.data === 'reload') {
			console.log('🔄 Live reload triggered');
			location.reload();
		}
	};
	source.onerror = function() {
		console.warn('⚠️ Live reload disconnected');
	};
	console.log('✅ Live reload connected');
})();
</script>
`
	// Find </body> and inject before it
	if idx := strings.LastIndex(html, "</body>"); idx != -1 {
		return html[:idx] + script + html[idx:]
	}

	// If no </body>, append at the end
	return html + script
}
