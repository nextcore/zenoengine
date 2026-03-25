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

// RegisterContainerBridgeSlots mendaftarkan slot khusus untuk memanggil Container lain
func RegisterContainerBridgeSlots(eng *engine.Engine) {

	// ==========================================
	// SLOT: DOCKER.HEALTH
	// ==========================================
	eng.Register("docker.health", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		host := "localhost"
		port := "8000"
		targetAs := "health"

		if node.Value != nil {
			host = coerce.ToString(resolveValue(node.Value, scope))
		}

		for _, c := range node.Children {
			if c.Name == "port" {
				port = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "host" {
				host = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "as" {
				targetAs = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		url := fmt.Sprintf("http://%s:%s/up", host, port)
		
		client := &http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get(url)
		
		result := map[string]interface{}{
			"host": host,
			"port": port,
		}

		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			result["status"] = "healthy"
			resp.Body.Close()
		} else {
			result["status"] = "unhealthy"
			if err != nil {
				result["error"] = err.Error()
			} else {
				result["error"] = fmt.Sprintf("HTTP %d", resp.StatusCode)
				resp.Body.Close()
			}
		}

		scope.Set(targetAs, result)
		return nil
	}, engine.SlotMeta{
		Description: "Check health of a docker sidecar HTTP service.",
		Example:     "docker.health: 'php_worker' {\n  port: 8000\n  as: $h\n}",
		Inputs: map[string]engine.InputMeta{
			"host": {Description: "Hostname of the container", Required: false},
			"port": {Description: "Port of the container", Required: false},
			"as":   {Description: "Variable to store result", Required: false},
		},
	})

	// ==========================================
	// SLOT: DOCKER.CALL
	// ==========================================
	eng.Register("docker.call", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		host := coerce.ToString(resolveValue(node.Value, scope))
		if host == "" {
			return fmt.Errorf("docker.call requires a valid host/container name")
		}

		port := "80"
		endpoint := "/"
		method := "POST"
		var payload interface{}
		targetAs := "response"
		timeoutMs := 15000 // default 15s

		for _, c := range node.Children {
			if c.Name == "port" {
				port = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "endpoint" {
				endpoint = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "method" {
				method = strings.ToUpper(coerce.ToString(parseNodeValue(c, scope)))
			}
			if c.Name == "payload" {
				payload = parseNodeValue(c, scope)
			}
			if c.Name == "timeout" {
				timeoutMs, _ = coerce.ToInt(parseNodeValue(c, scope))
			}
			if c.Name == "as" {
				targetAs = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		// Normalize endpoint
		if !strings.HasPrefix(endpoint, "/") {
			endpoint = "/" + endpoint
		}

		url := fmt.Sprintf("http://%s:%s%s", host, port, endpoint)

		var reqBody io.Reader
		if payload != nil {
			jsonData, err := json.Marshal(payload)
			if err != nil {
				return fmt.Errorf("docker.call payload encode error: %v", err)
			}
			reqBody = bytes.NewBuffer(jsonData)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
		if err != nil {
			return fmt.Errorf("docker.call request error: %v", err)
		}

		if payload != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Header.Set("User-Agent", "ZenoEngine-ContainerBridge/1.0")

		client := &http.Client{Timeout: time.Duration(timeoutMs) * time.Millisecond}
		resp, err := client.Do(req)
		if err != nil {
			scope.Set(targetAs, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
				"host":    host,
			})
			return nil // Don't crash Zeno, just return soft error inside the variable
		}
		defer resp.Body.Close()

		bodyBytes, _ := io.ReadAll(resp.Body)
		
		isOk := resp.StatusCode >= 200 && resp.StatusCode < 300

		// Try parsing JSON response
		var jsonResponse interface{}
		errJson := json.Unmarshal(bodyBytes, &jsonResponse)

		result := map[string]interface{}{
			"success": isOk,
			"code":    resp.StatusCode,
			"raw":     string(bodyBytes),
		}

		if errJson == nil {
			result["data"] = jsonResponse
		}

		scope.Set(targetAs, result)
		return nil
	}, engine.SlotMeta{
		Description: "Call an external docker container microservice.",
		Example:     "docker.call: 'php_worker' {\n  endpoint: '/calculate'\n  payload: { data: 1 }\n  as: $res\n}",
		Inputs: map[string]engine.InputMeta{
			"port":     {Description: "Port (default 80)", Required: false},
			"endpoint": {Description: "HTTP Path (default /)", Required: false},
			"method":   {Description: "HTTP Method (default POST)", Required: false},
			"payload":  {Description: "JSON array/object to send", Required: false},
			"timeout":  {Description: "Timeout in ms (default 15000)", Required: false},
			"as":       {Description: "Variable to store result", Required: false},
		},
	})
}
