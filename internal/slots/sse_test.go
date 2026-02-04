package slots

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"zeno/pkg/engine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSESlots(t *testing.T) {
	eng := engine.NewEngine()
	RegisterSSESlots(eng)

	t.Run("sse.stream sets headers", func(t *testing.T) {
		rec := httptest.NewRecorder()
		ctx := context.WithValue(context.Background(), "httpWriter", http.ResponseWriter(rec))

		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "sse.stream",
		}

		err := eng.Execute(ctx, node, scope)
		require.NoError(t, err)

		assert.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))
		assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
		assert.Equal(t, "keep-alive", rec.Header().Get("Connection"))
	})

	t.Run("sse.send writes formatted data", func(t *testing.T) {
		rec := httptest.NewRecorder()
		ctx := context.WithValue(context.Background(), "httpWriter", http.ResponseWriter(rec))

		scope := engine.NewScope(nil)

		// Setup sse.stream wrapping sse.send
		node := &engine.Node{
			Name: "sse.stream",
			Children: []*engine.Node{
				{
					Name: "sse.send",
					Children: []*engine.Node{
						{Name: "event", Value: "ping"},
						{Name: "data", Value: "pong"},
						{Name: "id", Value: "123"},
					},
				},
			},
		}

		err := eng.Execute(ctx, node, scope)
		require.NoError(t, err)

		body := rec.Body.String()
		assert.Contains(t, body, "event: ping\n")
		assert.Contains(t, body, "id: 123\n")
		assert.Contains(t, body, "data: pong\n")
		assert.True(t, strings.HasSuffix(body, "\n\n") || strings.Contains(body, "\n\n"))
	})

	t.Run("sse.loop limited iterations", func(t *testing.T) {
		rec := httptest.NewRecorder()
		ctx := context.WithValue(context.Background(), "httpWriter", http.ResponseWriter(rec))

		scope := engine.NewScope(nil)

		// Loop 2 times
		node := &engine.Node{
			Name: "sse.stream",
			Children: []*engine.Node{
				{
					Name: "sse.loop",
					Children: []*engine.Node{
						{Name: "interval", Value: 10}, // 10ms
						{Name: "max", Value: 2},
						{
							Name: "do",
							Children: []*engine.Node{
								{
									Name: "sse.send",
									Children: []*engine.Node{
										{Name: "data", Value: "tick"},
									},
								},
							},
						},
					},
				},
			},
		}

		start := time.Now()
		err := eng.Execute(ctx, node, scope)
		require.NoError(t, err)
		duration := time.Since(start)

		// Should take at least 20ms (2 * 10ms)
		assert.True(t, duration >= 20*time.Millisecond)

		body := rec.Body.String()
		assert.Equal(t, 2, strings.Count(body, "data: tick"))
	})
}
