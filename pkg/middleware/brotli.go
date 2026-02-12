package middleware

import (
	"bufio"
	"fmt"
	"net"
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
	w           *brotli.Writer
	wroteHeader bool
	code        int
}

func (w *brotliResponseWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.code = code

	// Check if we should compress
	// 1. Skip if not 200 OK (simple heuristic for now, avoid 204, 304, 302 redirections)
	// We can expand this list later (e.g. 404 is compressible).
	// But critically, 204/304 MUST NOT have body/encoding.
	// 302 Found has small body, but usually negligible gain.
	// Safe bet for now: Only compress 200.
	if code != http.StatusOK {
		w.ResponseWriter.WriteHeader(code)
		w.wroteHeader = true
		return
	}

	// 2. Check Content-Type
	ct := w.Header().Get("Content-Type")
	// If empty, sniff? Or skip?
	// Skip images, videos, etc if already compressed.
	if strings.HasPrefix(ct, "image/") && !strings.Contains(ct, "svg") {
		w.ResponseWriter.WriteHeader(code)
		w.wroteHeader = true
		return
	}
	if strings.HasPrefix(ct, "video/") || strings.HasPrefix(ct, "audio/") {
		w.ResponseWriter.WriteHeader(code)
		w.wroteHeader = true
		return
	}

	// OK to compress
	w.Header().Del("Content-Length")
	w.Header().Set("Content-Encoding", "br")
	w.Header().Add("Vary", "Accept-Encoding")
	w.ResponseWriter.WriteHeader(code)
	w.wroteHeader = true
}

func (w *brotliResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	// If we decided NOT to compress (header not set), pass through
	if w.Header().Get("Content-Encoding") != "br" {
		return w.ResponseWriter.Write(b)
	}

	return w.w.Write(b)
}

// Hijack implements the http.Hijacker interface
func (w *brotliResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, fmt.Errorf("brotliResponseWriter: underlying ResponseWriter does not support Hijacker")
}

// Flush implements the http.Flusher interface
func (w *brotliResponseWriter) Flush() {
	if w.Header().Get("Content-Encoding") == "br" {
		w.w.Flush()
	}
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
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

		// Check if already compressed
		if w.Header().Get("Content-Encoding") != "" {
			next.ServeHTTP(w, r)
			return
		}

		// Don't set header yet! Wait for WriteHeader decision.

		// Get writer from pool
		bw := brWriterPool.Get().(*brotli.Writer)
		defer brWriterPool.Put(bw)

		bw.Reset(w)
		defer func() {
			// Don't close if we didn't compress, or if we need to clean up
			// But bw.Close() writes footer.
			// Only close if we actually used it (encoding is br)
			// Wait, the wrapper uses 'w.w' which is 'bw'.
			// We can check header on the underlying writer 'w'.
			if w.Header().Get("Content-Encoding") == "br" {
				bw.Close()
			}
		}()

		// Wrap response writer
		brw := &brotliResponseWriter{ResponseWriter: w, w: bw}

		next.ServeHTTP(brw, r)
	})
}
