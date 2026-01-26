package slots

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

// RegisterInertiaSlots registers Inertia.js related slots
func RegisterInertiaSlots(eng *engine.Engine) {
	// inertia.render - Main slot for rendering Inertia responses
	eng.Register("inertia.render", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var component string
		var props map[string]interface{}

		// Parse parameters
		for _, c := range node.Children {
			if c.Name == "component" {
				component = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "props" {
				propsVal := parseNodeValue(c, scope)
				if propsMap, ok := propsVal.(map[string]interface{}); ok {
					props = propsMap
				}
			}
		}

		if component == "" {
			return fmt.Errorf("inertia.render: component is required")
		}

		// Get HTTP context
		w, ok := ctx.Value("httpWriter").(http.ResponseWriter)
		if !ok {
			return fmt.Errorf("inertia.render: http writer not found in context")
		}

		r, ok := ctx.Value("httpRequest").(*http.Request)
		if !ok {
			return fmt.Errorf("inertia.render: http request not found in context")
		}

		// Get shared data from scope
		sharedData := make(map[string]interface{})

		// Add auth if available
		if auth, ok := scope.Get("auth"); ok && auth != nil {
			sharedData["auth"] = auth
		}

		// Add flash messages if available
		if flash, ok := scope.Get("flash"); ok && flash != nil {
			sharedData["flash"] = flash
		}

		// Add errors if available
		if errors, ok := scope.Get("errors"); ok && errors != nil {
			sharedData["errors"] = errors
		}

		// Merge shared data with props
		if props == nil {
			props = make(map[string]interface{})
		}
		for k, v := range sharedData {
			if _, exists := props[k]; !exists {
				props[k] = v
			}
		}

		// Build Inertia page object
		page := map[string]interface{}{
			"component": component,
			"props":     props,
			"url":       r.URL.Path,
			"version":   "1", // TODO: Make this configurable
		}

		// Check if this is an Inertia request
		isInertia := r.Header.Get("X-Inertia") == "true"

		if isInertia {
			// Return JSON for Inertia requests
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Inertia", "true")
			w.Header().Set("Vary", "Accept")

			return json.NewEncoder(w).Encode(page)
		}

		// Return HTML for initial page load
		pageJSON, err := json.Marshal(page)
		if err != nil {
			return fmt.Errorf("inertia.render: failed to marshal page data: %v", err)
		}

		html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Inertia App</title>
    <link rel="stylesheet" href="/inertia/dist/assets/app.css">
</head>
<body>
    <div id="app" data-page='%s'></div>
    <script type="module" src="/inertia/dist/assets/app.js"></script>
</body>
</html>`, string(pageJSON))

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))

		return nil
	}, engine.SlotMeta{
		Description: "Render Inertia.js response",
		Example:     "inertia.render:\n  component: \"Dashboard\"\n  props: { user: $user }",
	})

	// inertia.share - Share data across all Inertia requests
	eng.Register("inertia.share", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		// Parse key-value pairs
		for _, c := range node.Children {
			if c.Name != "do" {
				value := parseNodeValue(c, scope)
				scope.Set(c.Name, value)
			}
		}

		return nil
	}, engine.SlotMeta{
		Description: "Share data across all Inertia requests",
		Example:     "inertia.share:\n  auth: $auth\n  flash: $flash",
	})

	// inertia.location - Force a full page reload to a URL
	eng.Register("inertia.location", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var url string

		for _, c := range node.Children {
			if c.Name == "url" {
				url = coerce.ToString(parseNodeValue(c, scope))
			}
		}

		if url == "" {
			return fmt.Errorf("inertia.location: url is required")
		}

		w, ok := ctx.Value("httpWriter").(http.ResponseWriter)
		if !ok {
			return fmt.Errorf("inertia.location: http writer not found in context")
		}

		w.Header().Set("X-Inertia-Location", url)
		w.WriteHeader(http.StatusConflict)

		return nil
	}, engine.SlotMeta{
		Description: "Force a full page reload to a URL",
		Example:     "inertia.location:\n  url: \"/login\"",
	})
}
