package slots

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"zeno/pkg/engine"
	"zeno/pkg/fastjson"
	"zeno/pkg/utils/coerce"
)

func RegisterHTTPServerSlots(eng *engine.Engine) {

	// 1. HTTP.RESPONSE
	eng.Register("http.response", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		w, ok := ctx.Value("httpWriter").(http.ResponseWriter)
		if !ok {
			return fmt.Errorf("http.response: not in http context")
		}

		var status int = 200
		var contentType string = "application/json"
		var body interface{}

		if node.Value != nil {
			val := resolveValue(node.Value, scope)
			if s, err := coerce.ToInt(val); err == nil {
				status = s
			}
		}

		for _, c := range node.Children {
			val := parseNodeValue(c, scope)

			if c.Name == "status" {
				if s, err := coerce.ToInt(val); err == nil {
					status = s
				}
			}
			if c.Name == "type" {
				contentType = coerce.ToString(val)
			}
			if c.Name == "data" || c.Name == "body" {
				body = val
			}
		}

		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(status)

		if body != nil {
			if contentType == "application/json" {
				// Use fast JSON encoder (2-3x faster than encoding/json)
				fastjson.NewEncoder(w).Encode(body)
			} else {
				fmt.Fprint(w, coerce.ToString(body))
			}
		}
		return nil
	}, engine.SlotMeta{Example: "http.response: 200\n  body: $data"})

	// 2. HTTP.REDIRECT
	eng.Register("http.redirect", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		w, ok := ctx.Value("httpWriter").(http.ResponseWriter)
		r, ok2 := ctx.Value("httpRequest").(*http.Request)
		if !ok || !ok2 {
			return fmt.Errorf("http.redirect: not in http context")
		}

		url := coerce.ToString(resolveValue(node.Value, scope))
		if url == "" {
			for _, c := range node.Children {
				if c.Name == "to" || c.Name == "url" {
					url = coerce.ToString(parseNodeValue(c, scope))
				}
			}
		}

		if url == "" {
			return fmt.Errorf("http.redirect: url is required")
		}

		http.Redirect(w, r, url, http.StatusFound)
		return ErrReturn
	}, engine.SlotMeta{Example: "http.redirect: '/home'"})

	// 3. COOKIE.SET
	eng.Register("cookie.set", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		w, ok := ctx.Value("httpWriter").(http.ResponseWriter)
		if !ok {
			return fmt.Errorf("cookie.set: missing context")
		}

		name := ""
		value := ""
		maxAge := 3600

		for _, c := range node.Children {
			val := parseNodeValue(c, scope)
			if c.Name == "name" {
				name = coerce.ToString(val)
			}
			if c.Name == "val" {
				value = coerce.ToString(val)
			}
			if c.Name == "age" {
				if i, err := coerce.ToInt(val); err == nil {
					maxAge = i
				}
			}
		}

		if name == "" {
			return fmt.Errorf("cookie.set: name required")
		}

		http.SetCookie(w, &http.Cookie{
			Name: name, Value: value, MaxAge: maxAge, Path: "/", HttpOnly: true,
		})
		return nil
	}, engine.SlotMeta{Example: "cookie.set\n  name: 'token'\n  val: $token"})

	// 4. HTTP.FORM
	eng.Register("http.form", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		key := coerce.ToString(resolveValue(node.Value, scope))
		target := key

		for _, c := range node.Children {
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		// Get form data injected by Router
		if formData, ok := scope.Get("form"); ok {
			if m, ok := formData.(map[string]interface{}); ok {
				if val, exists := m[key]; exists {
					scope.Set(target, val)
					return nil
				}
			}
		}

		// If not present, set empty string
		scope.Set(target, "")
		return nil
	}, engine.SlotMeta{Example: "http.form: 'email'\n  as: $email"})

	// 5. HTTP.QUERY (Fixed: Use resolveValue)
	eng.Register("http.query", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		reqVal := ctx.Value("httpRequest")
		if reqVal == nil {
			return nil
		}
		r := reqVal.(*http.Request)

		// [FIX] Use resolveValue so quotes ("page") are cleaned up to (page)
		param := coerce.ToString(resolveValue(node.Value, scope))
		target := param

		for _, c := range node.Children {
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		val := r.URL.Query().Get(param)
		scope.Set(target, val)
		return nil
	}, engine.SlotMeta{Example: "http.query: 'page'\n  as: $page_param"})

	// 6. HTTP.HEADER (Extract HTTP headers)
	eng.Register("http.header", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		reqVal := ctx.Value("httpRequest")
		if reqVal == nil {
			return nil
		}
		r := reqVal.(*http.Request)

		// Get header name from node value
		headerName := coerce.ToString(resolveValue(node.Value, scope))
		target := headerName

		for _, c := range node.Children {
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		val := r.Header.Get(headerName)
		scope.Set(target, val)
		return nil
	}, engine.SlotMeta{Example: "http.header: 'X-Tenant-ID'\n  as: $tenant_id"})

	// 7. HTTP.HOST (Extract HTTP host)
	eng.Register("http.host", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		reqVal := ctx.Value("httpRequest")
		if reqVal == nil {
			return nil
		}
		r := reqVal.(*http.Request)

		target := "host"
		for _, c := range node.Children {
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		scope.Set(target, r.Host)
		return nil
	}, engine.SlotMeta{Example: "http.host: { as: $host }"})

	// ==========================================
	// HTTP RESPONSE HELPERS (Syntax Sugar)
	// ==========================================

	// Helper function to send JSON response with auto success field
	sendJSONResponse := func(ctx context.Context, statusCode int, node *engine.Node, scope *engine.Scope, success bool) error {
		w, ok := ctx.Value("httpWriter").(http.ResponseWriter)
		if !ok {
			return fmt.Errorf("http response helper: not in http context")
		}

		// Build response body
		responseBody := make(map[string]interface{})
		responseBody["success"] = success

		// Parse children for message, data, errors, etc.
		for _, child := range node.Children {
			val := parseNodeValue(child, scope)
			responseBody[child.Name] = val
		}

		// Send JSON response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		fastjson.NewEncoder(w).Encode(responseBody)
		return nil
	}

	// Success Responses (2xx)
	eng.Register("http.ok", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		return sendJSONResponse(ctx, 200, node, scope, true)
	}, engine.SlotMeta{
		Description: "Send 200 OK response with auto JSON wrapping",
		Example:     "http.ok: {\n  data: $posts\n}",
	})

	eng.Register("http.created", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		return sendJSONResponse(ctx, 201, node, scope, true)
	}, engine.SlotMeta{
		Description: "Send 201 Created response",
		Example:     "http.created: {\n  message: \"Resource created\"\n  id: $db_last_id\n}",
	})

	eng.Register("http.accepted", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		return sendJSONResponse(ctx, 202, node, scope, true)
	}, engine.SlotMeta{
		Description: "Send 202 Accepted response",
		Example:     "http.accepted: {\n  message: \"Request accepted\"\n}",
	})

	eng.Register("http.no_content", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		w, ok := ctx.Value("httpWriter").(http.ResponseWriter)
		if !ok {
			return fmt.Errorf("http.no_content: not in http context")
		}
		w.WriteHeader(204)
		return nil
	}, engine.SlotMeta{
		Description: "Send 204 No Content response",
		Example:     "http.no_content",
	})

	// Client Error Responses (4xx)
	eng.Register("http.bad_request", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		return sendJSONResponse(ctx, 400, node, scope, false)
	}, engine.SlotMeta{
		Description: "Send 400 Bad Request response",
		Example:     "http.bad_request: {\n  message: \"Invalid parameters\"\n  errors: $errors\n}",
	})

	eng.Register("http.unauthorized", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		return sendJSONResponse(ctx, 401, node, scope, false)
	}, engine.SlotMeta{
		Description: "Send 401 Unauthorized response",
		Example:     "http.unauthorized: {\n  message: \"Authentication required\"\n}",
	})

	eng.Register("http.forbidden", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		return sendJSONResponse(ctx, 403, node, scope, false)
	}, engine.SlotMeta{
		Description: "Send 403 Forbidden response",
		Example:     "http.forbidden: {\n  message: \"Access denied\"\n}",
	})

	eng.Register("http.not_found", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		return sendJSONResponse(ctx, 404, node, scope, false)
	}, engine.SlotMeta{
		Description: "Send 404 Not Found response",
		Example:     "http.not_found: {\n  message: \"Resource not found\"\n}",
	})

	eng.Register("http.validation_error", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		return sendJSONResponse(ctx, 422, node, scope, false)
	}, engine.SlotMeta{
		Description: "Send 422 Validation Error response",
		Example:     "http.validation_error: {\n  message: \"Validation failed\"\n  errors: $errors\n}",
	})

	// Server Error Responses (5xx)
	eng.Register("http.server_error", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		return sendJSONResponse(ctx, 500, node, scope, false)
	}, engine.SlotMeta{
		Description: "Send 500 Internal Server Error response",
		Example:     "http.server_error: {\n  message: \"Internal error\"\n  error: $error\n}",
	})

}
