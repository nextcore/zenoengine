package slots

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"zeno/pkg/engine"
	"zeno/pkg/fastjson"
	"zeno/pkg/utils/coerce"
)

// RegisterSSESlots registers Server-Sent Events slots
func RegisterSSESlots(eng *engine.Engine) {
	
	// 1. SSE.STREAM - Start SSE connection
	eng.Register("sse.stream", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		w, ok := ctx.Value("httpWriter").(http.ResponseWriter)
		if !ok {
			return fmt.Errorf("sse.stream: not in http context")
		}

		// Set SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Get flusher
		flusher, ok := w.(http.Flusher)
		if !ok {
			return fmt.Errorf("sse.stream: streaming not supported")
		}

		// Store flusher in scope for sse.send to use
		scope.Set("__sse_writer", w)
		scope.Set("__sse_flusher", flusher)
		scope.Set("__sse_active", true)

		// Execute children (sse.send, sse.loop, etc.)
		for _, child := range node.Children {
			if err := eng.Execute(ctx, child, scope); err != nil {
				return err
			}
		}

		return nil
	}, engine.SlotMeta{
		Description: "Start Server-Sent Events stream",
		Example: `sse.stream {
    sse.send {
        event: "message"
        data: "Hello from SSE"
    }
}`,
	})

	// 2. SSE.SEND - Send SSE message
	eng.Register("sse.send", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		w, ok := scope.Get("__sse_writer")
		if !ok {
			return fmt.Errorf("sse.send: must be inside sse.stream")
		}
		writer := w.(http.ResponseWriter)

		flusher, _ := scope.Get("__sse_flusher")
		flush := flusher.(http.Flusher)

		var eventName string
		var data interface{}
		var id string
		var retry int

		// Parse attributes
		for _, child := range node.Children {
			val := parseNodeValue(child, scope)
			switch child.Name {
			case "event":
				eventName = coerce.ToString(val)
			case "data":
				data = val
			case "id":
				id = coerce.ToString(val)
			case "retry":
				retry, _ = coerce.ToInt(val)
			}
		}

		// Send event name
		if eventName != "" {
			fmt.Fprintf(writer, "event: %s\n", eventName)
		}

		// Send ID
		if id != "" {
			fmt.Fprintf(writer, "id: %s\n", id)
		}

		// Send retry
		if retry > 0 {
			fmt.Fprintf(writer, "retry: %d\n", retry)
		}

		// Send data (JSON if object/array, string otherwise)
		if data != nil {
			switch v := data.(type) {
			case string:
				fmt.Fprintf(writer, "data: %s\n", v)
			case map[string]interface{}, []interface{}:
				jsonData, _ := fastjson.Marshal(v)
				fmt.Fprintf(writer, "data: %s\n", string(jsonData))
			default:
				fmt.Fprintf(writer, "data: %v\n", v)
			}
		}

		// End message
		fmt.Fprintf(writer, "\n")
		flush.Flush()

		return nil
	}, engine.SlotMeta{
		Description: "Send SSE message to client",
		Inputs: map[string]engine.InputMeta{
			"event": {Description: "Event name", Type: "string"},
			"data":  {Description: "Data to send", Type: "any"},
			"id":    {Description: "Event ID", Type: "string"},
			"retry": {Description: "Retry interval (ms)", Type: "int"},
		},
		Example: `sse.send {
    event: "notification"
    data: { message: "New update!" }
}`,
	})

	// 3. SSE.LOOP - Keep connection alive with periodic updates
	eng.Register("sse.loop", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		active, ok := scope.Get("__sse_active")
		if !ok || !active.(bool) {
			return fmt.Errorf("sse.loop: must be inside sse.stream")
		}

		// Get interval (default 1 second)
		interval := 1000
		var maxIterations int
		
		for _, child := range node.Children {
			if child.Name == "interval" {
				val := parseNodeValue(child, scope)
				interval, _ = coerce.ToInt(val)
			}
			if child.Name == "max" {
				val := parseNodeValue(child, scope)
				maxIterations, _ = coerce.ToInt(val)
			}
		}

		ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
		defer ticker.Stop()

		iterations := 0
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				// Execute loop body
				for _, child := range node.Children {
					if child.Name == "do" {
						for _, doChild := range child.Children {
							if err := eng.Execute(ctx, doChild, scope); err != nil {
								return err
							}
						}
					}
				}

				iterations++
				if maxIterations > 0 && iterations >= maxIterations {
					return nil
				}
			}
		}
	}, engine.SlotMeta{
		Description: "Loop with periodic SSE updates",
		Inputs: map[string]engine.InputMeta{
			"interval": {Description: "Interval in milliseconds", Type: "int"},
			"max":      {Description: "Max iterations (0 = infinite)", Type: "int"},
		},
		Example: `sse.loop {
    interval: 1000
    do {
        sse.send {
            data: $currentTime
        }
    }
}`,
	})

	// 4. SSE.KEEPALIVE - Send periodic keepalive pings
	eng.Register("sse.keepalive", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		w, ok := scope.Get("__sse_writer")
		if !ok {
			return nil
		}
		writer := w.(http.ResponseWriter)

		flusher, _ := scope.Get("__sse_flusher")
		flush := flusher.(http.Flusher)

		// Default 30 seconds
		interval := 30000
		if node.Value != nil {
			interval, _ = coerce.ToInt(resolveValue(node.Value, scope))
		}

		go func() {
			ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					fmt.Fprintf(writer, ": keepalive\n\n")
					flush.Flush()
				}
			}
		}()

		return nil
	}, engine.SlotMeta{
		Description: "Send periodic keepalive pings",
		Example:     "sse.keepalive: 30000",
	})
}
