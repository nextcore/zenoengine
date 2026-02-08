package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"syscall/js"

	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

func RegisterFetchSlots(eng *engine.Engine) {
	// HTTP.FETCH: Async Native Browser Fetch
	// Usage:
	// http.fetch: 'https://api.example.com/data' {
	//    method: 'POST'
	//    body: { ... }
	//    then: { as: $resp ... }
	//    catch: { as: $err ... }
	// }
	eng.Register("http.fetch", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		url := coerce.ToString(ResolveValue(node.Value, scope))
		method := "GET"
		var body interface{}
		var headers map[string]string

		var thenBlock, catchBlock *engine.Node
		var asVar = "response"
		var errVar = "error"

		for _, c := range node.Children {
			if c.Name == "method" {
				method = strings.ToUpper(coerce.ToString(ResolveValue(c.Value, scope)))
			} else if c.Name == "body" {
				body = ResolveValue(c.Value, scope)
			} else if c.Name == "headers" {
				// Parse map
				val := ResolveValue(c.Value, scope)
				if m, ok := val.(map[string]interface{}); ok {
					headers = make(map[string]string)
					for k, v := range m {
						headers[k] = coerce.ToString(v)
					}
				}
			} else if c.Name == "then" {
				thenBlock = c
				// Check inner 'as'
				for _, sub := range c.Children {
					if sub.Name == "as" {
						asVar = strings.TrimPrefix(coerce.ToString(sub.Value), "$")
					}
				}
			} else if c.Name == "catch" {
				catchBlock = c
				for _, sub := range c.Children {
					if sub.Name == "as" {
						errVar = strings.TrimPrefix(coerce.ToString(sub.Value), "$")
					}
				}
			}
		}

		if url == "" {
			return fmt.Errorf("http.fetch: url required")
		}

		// Prepare Fetch Options
		opts := js.Global().Get("Object").New()
		opts.Set("method", method)

		if headers != nil {
			jsHeaders := js.Global().Get("Object").New()
			for k, v := range headers {
				jsHeaders.Set(k, v)
			}
			opts.Set("headers", jsHeaders)
		}

		if body != nil && method != "GET" && method != "HEAD" {
			jsonBody, _ := json.Marshal(body)
			opts.Set("body", string(jsonBody))
			// Auto content-type if not set?
			// Ideally user sets it, or we default to application/json
		}

		// Call Fetch Promise
		promise := js.Global().Call("fetch", url, opts)

		// HANDLE PROMISE (Non-Blocking / Async)
		// Since Zeno Engine is synchronous, we cannot "wait" here without freezing UI.
		// Instead, we attach .then() callbacks that will re-trigger engine execution for the blocks.

		// Capture Engine and Scope for async execution
		asyncScope := scope.Clone() // Clone scope to avoid race conditions if needed, or share?
		// Sharing scope is better for SPA state usually, but let's be safe.
		// Actually, for a callback, we usually want to update the UI state.

		successFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			resp := args[0]

			// Parse JSON?
			// We define a promise chain to parse JSON first
			// But for simplicity, let's assume we want JSON.
			// Ideally we chain .json() in JS.

			// Handle response body parsing
			resp.Call("json").Call("then", js.FuncOf(func(this2 js.Value, args2 []js.Value) interface{} {
				jsonResult := args2[0]

				// Convert JS Object to Map/Slice for Go
				// Quick hack: stringify back to JSON and unmarshal in Go (safest for complex types)
				jsonStr := js.Global().Get("JSON").Call("stringify", jsonResult).String()
				var goData interface{}
				json.Unmarshal([]byte(jsonStr), &goData)

				// Set Variable
				asyncScope.Set(asVar, goData)
				asyncScope.Set("status", resp.Get("status").Int())

				// Execute THEN Block
				if thenBlock != nil {
					// We need a fresh context? Or reuse?
					// Creating new context is safer.
					// Also we need to render the output?
					// If this block does "view.render", where does it write?
					// It needs to write to the DOM element "app"?
					// This is tricky. Zeno slots write to "httpWriter".
					// In Async mode, we might need to capture output and apply it?
					// Or just let logic run (e.g. updating variables for Datastar).

					// Strategy: Logic Only. Datastar handles the UI update if variables change?
					// But we are in a disconnected scope.
					// Ideally we call "render" again?

					// For MVP: Just execute logic.
					// If user wants to update UI, they might call 'zenoRender' again via JS interop?
					// OR: We provide a way to trigger reactivity.

					// Let's pass a dummy writer for now.
					// But wait, if they use 'view: ...', they expect UI update.
					// We can reuse the main Render logic if we had access.

					// Let's execute.
					err := eng.Execute(context.Background(), thenBlock, asyncScope)
					if err != nil {
						fmt.Println("Fetch Then Error:", err)
					}

					// Trigger Datastar Patch if available?
					// Not easy without explicit signal updates.
				}
				return nil
			}), js.FuncOf(func(this3 js.Value, args3 []js.Value) interface{} {
				// JSON Parse Fail (maybe text?)
				fmt.Println("Fetch JSON Parse Error")
				return nil
			}))

			return nil
		})

		failureFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			errJs := args[0]
			errStr := errJs.Get("message").String()

			asyncScope.Set(errVar, errStr)

			if catchBlock != nil {
				eng.Execute(context.Background(), catchBlock, asyncScope)
			}
			return nil
		})

		promise.Call("then", successFunc).Call("catch", failureFunc)

		return nil
	}, engine.SlotMeta{
		Description: "Native Browser Fetch (Async)",
		Inputs: map[string]engine.InputMeta{
			"method": {Description: "HTTP Method (GET, POST, etc.)"},
			"body":   {Description: "Request body"},
			"then":   {Description: "Block executed on success"},
			"catch":  {Description: "Block executed on error"},
		},
	})
}
