package middleware

import (
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/andybalholm/brotli"
)

var brWriterPool = sync.Pool{
	New: func() interface{} {
		return brotli.NewWriter(nil)
	},
}

type brotliResponseWriter struct {
	http.ResponseWriter
	w *brotli.Writer
}

func (w *brotliResponseWriter) Write(b []byte) (int, error) {
	return w.w.Write(b)
}

func (w *brotliResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.Header().Del("Content-Length") // Content length changes with compression
	w.ResponseWriter.WriteHeader(code)
}

// BrotliMiddleware compresses responses using Brotli if the client supports it.
func BrotliMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip if Brotli is disabled via env
		if os.Getenv("COMPRESSION_BROTLI_ENABLED") == "false" {
			next.ServeHTTP(w, r)
			return
		}

		// Check if client supports br
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "br") {
			next.ServeHTTP(w, r)
			return
		}

		// Check if already compressed (e.g. by another middleware or pre-compressed file)
		if w.Header().Get("Content-Encoding") != "" {
			next.ServeHTTP(w, r)
			return
		}

		// Prepare for Brotli compression
		w.Header().Set("Content-Encoding", "br")
		w.Header().Add("Vary", "Accept-Encoding")

		// Get writer from pool
		bw := brWriterPool.Get().(*brotli.Writer)
		defer brWriterPool.Put(bw)

		bw.Reset(w)
		defer bw.Close()

		// Wrap response writer
		brw := &brotliResponseWriter{ResponseWriter: w, w: bw}

		next.ServeHTTP(brw, r)
	})
}
