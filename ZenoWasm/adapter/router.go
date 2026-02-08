package adapter

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

// Router Structure
type Route struct {
	Method  string
	Pattern string
	Handler *engine.Node
	Params  []string
	Regex   *regexp.Regexp
}

var (
	Routes []*Route
	RouterMu sync.RWMutex
)

// Normalize path: remove trailing slash, ensure leading slash
func normalizePath(p string) string {
	if p == "" {
		return "/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	if len(p) > 1 && strings.HasSuffix(p, "/") {
		p = strings.TrimSuffix(p, "/")
	}
	return p
}

// Convert path pattern (e.g. /users/:id) to Regex
func patternToRegex(p string) (*regexp.Regexp, []string) {
	segments := strings.Split(p, "/")
	var params []string
	var regexParts []string

	for _, seg := range segments {
		if strings.HasPrefix(seg, ":") {
			params = append(params, seg[1:])
			regexParts = append(regexParts, "([^/]+)")
		} else if seg == "*" {
			params = append(params, "wildcard")
			regexParts = append(regexParts, "(.*)")
		} else {
			regexParts = append(regexParts, regexp.QuoteMeta(seg))
		}
	}

	pattern := "^" + strings.Join(regexParts, "/") + "$"
	return regexp.MustCompile(pattern), params
}

func RegisterRouterSlots(eng *engine.Engine) {
	// ROUTE DEFINITION: route: '/path' { ... }
	eng.Register("route", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		path := coerce.ToString(ResolveValue(node.Value, scope))

		var viewName string
		var handlerNode *engine.Node

		// Check children for view or do block
		for _, c := range node.Children {
			if c.Name == "view" {
				viewName = coerce.ToString(ResolveValue(c.Value, scope))
			} else if c.Name == "do" {
				handlerNode = c
			}
		}

		// If simple view shorthand
		if handlerNode == nil && viewName != "" {
			handlerNode = &engine.Node{
				Name: "view.blade",
				Value: viewName,
			}
		}

		if handlerNode != nil {
			addRoute("GET", path, handlerNode)
		}
		return nil
	}, engine.SlotMeta{Description: "Define a client-side route"})

	// ALIAS: router.get
	eng.Register("router.get", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		path := coerce.ToString(ResolveValue(node.Value, scope))
		var handlerNode *engine.Node
		for _, c := range node.Children {
			if c.Name == "do" {
				handlerNode = c
				break
			}
		}
		if handlerNode != nil {
			addRoute("GET", path, handlerNode)
		}
		return nil
	}, engine.SlotMeta{Description: "Define a GET route"})
}

func addRoute(method, path string, handler *engine.Node) {
	RouterMu.Lock()
	defer RouterMu.Unlock()

	normPath := normalizePath(path)
	regex, params := patternToRegex(normPath)

	Routes = append(Routes, &Route{
		Method:  method,
		Pattern: normPath,
		Handler: handler,
		Params:  params,
		Regex:   regex,
	})
	fmt.Printf("Route Registered: %s -> %s\n", method, normPath)
}

// MatchRoute finds the first matching route for a given path
func MatchRoute(path string) (*Route, map[string]interface{}) {
	RouterMu.RLock()
	defer RouterMu.RUnlock()

	normPath := normalizePath(path)

	for _, r := range Routes {
		matches := r.Regex.FindStringSubmatch(normPath)
		if matches != nil {
			params := make(map[string]interface{})
			// matches[0] is full string, matches[1:] are groups
			for i, paramName := range r.Params {
				if i+1 < len(matches) {
					params[paramName] = matches[i+1]
				}
			}
			return r, params
		}
	}
	return nil, nil
}
