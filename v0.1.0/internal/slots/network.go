package slots

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

func RegisterNetworkSlots(eng *engine.Engine) {
	// HTTP FETCH
	eng.Register("http.fetch", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var url, method, target string
		method = "GET"
		target = "api_response"

		for _, c := range node.Children {
			val := parseNodeValue(c, scope)
			if c.Name == "url" {
				url = coerce.ToString(val)
			}
			if c.Name == "method" {
				method = strings.ToUpper(coerce.ToString(val))
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		if url == "" {
			// Fallback value node (http.fetch: "$url")
			url = coerce.ToString(resolveValue(node.Value, scope))
		}

		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			return err
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		bodyBytes, _ := io.ReadAll(resp.Body)
		var result interface{}

		// Auto-detect JSON
		if strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
			var jsonResult interface{}
			if err := json.Unmarshal(bodyBytes, &jsonResult); err == nil {
				result = jsonResult
			} else {
				result = string(bodyBytes)
			}
		} else {
			result = string(bodyBytes)
		}

		scope.Set(target, result)
		return nil
	}, engine.SlotMeta{Example: "http.fetch: $api_url\n  as: $response"})
}
