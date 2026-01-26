package slots

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

func RegisterHTTPClientSlots(eng *engine.Engine) {
	eng.Register("http.request", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var url, method string
		var headers map[string]interface{}
		var body interface{}
		var timeoutSeconds int = 30
		target := "response"

		// 1. Parse Arguments (Value & Children)
		if node.Value != nil {
			url = coerce.ToString(resolveValue(node.Value, scope))
		}

		for _, c := range node.Children {
			val := parseNodeValue(c, scope)
			switch c.Name {
			case "url":
				url = coerce.ToString(val)
			case "method":
				method = strings.ToUpper(coerce.ToString(val))
			case "headers":
				if h, ok := val.(map[string]interface{}); ok {
					headers = h
				}
			case "body", "data", "payload":
				body = val
			case "timeout":
				if t, err := coerce.ToInt(val); err == nil {
					timeoutSeconds = t
				}
			case "as":
				// Handle variable assignment target
				if str, ok := c.Value.(string); ok {
					target = strings.TrimPrefix(str, "$")
				}
			}
		}

		// Defaults
		if url == "" {
			return fmt.Errorf("http.request: url is required")
		}
		if method == "" {
			method = "GET"
		}

		// 2. Prepare Request Body
		var reqBody io.Reader
		if body != nil {
			// If explicitly string, send raw
			if strBody, ok := body.(string); ok {
				reqBody = bytes.NewBufferString(strBody)
			} else {
				// Otherwise, auto-marshal to JSON
				jsonBytes, err := json.Marshal(body)
				if err != nil {
					return fmt.Errorf("http.request: failed to marshal body: %w", err)
				}
				reqBody = bytes.NewBuffer(jsonBytes)
			}
		}

		// 3. Create Request
		req, err := http.NewRequest(method, url, reqBody)
		if err != nil {
			return fmt.Errorf("http.request: failed to create request: %w", err)
		}

		// 4. Set Headers
		// Default Content-Type if body exists and not set
		if body != nil && headers["Content-Type"] == nil {
			req.Header.Set("Content-Type", "application/json")
		}

		for k, v := range headers {
			req.Header.Set(k, coerce.ToString(v))
		}

		// 5. Execute Request with Timeout
		client := &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		}

		resp, err := client.Do(req)
		if err != nil {
			// Return error structure instead of hard failing?
			// Ideally we want to let user handle connection errors too
			// For now, let's return a structured error response in the target variable
			// to avoid crashing the script execution flow completely.
			/*
				scope.Set(target, map[string]interface{}{
					"status": 0,
					"error":  err.Error(),
				})
				return nil
			*/
			return fmt.Errorf("http.request: connection failed: %w", err)
		}
		defer resp.Body.Close()

		// 6. Process Response
		respBodyBytes, _ := io.ReadAll(resp.Body)
		var respBody interface{}

		// Auto-parse Response JSON
		contentType := resp.Header.Get("Content-Type")
		if strings.Contains(contentType, "application/json") {
			var jsonResult interface{}
			if err := json.Unmarshal(respBodyBytes, &jsonResult); err == nil {
				respBody = jsonResult
			} else {
				respBody = string(respBodyBytes) // Fallback to string if invalid JSON
			}
		} else {
			respBody = string(respBodyBytes)
		}

		// Convert Headers to Map
		respHeaders := make(map[string]interface{})
		for k, v := range resp.Header {
			if len(v) > 0 {
				respHeaders[k] = v[0]
			}
		}

		// 7. Store Result
		result := map[string]interface{}{
			"status":  resp.StatusCode,
			"body":    respBody,
			"headers": respHeaders,
		}

		scope.Set(target, result)
		return nil

	}, engine.SlotMeta{Example: "http.request: 'https://api.com'\n  method: 'POST'\n  body: $data\n  as: $res"})
}
