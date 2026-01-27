package slots

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"zeno/pkg/apidoc"
	"zeno/pkg/engine"
	"zeno/pkg/middleware"
	"zeno/pkg/utils/coerce"

	"github.com/go-chi/chi/v5"
)

// Key context untuk menyimpan router instance
type routerKey struct{}

func RegisterRouterSlots(eng *engine.Engine, rootRouter *chi.Mux) {

	// Helper: Ambil router aktif (Root atau Group)
	getCurrentRouter := func(ctx context.Context) chi.Router {
		if r, ok := ctx.Value(routerKey{}).(chi.Router); ok {
			return r
		}
		return rootRouter
	}

	// Helper: Membuat Handler (Runtime Execution) - OPTIMIZED (Zero Runtime Overhead)
	// Auth is handled by native Chi middleware, injected via context
	createHandler := func(children []*engine.Node, baseScope *engine.Scope) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 1. Get Arena from pool for this request
			arena := engine.GetArena()
			defer engine.PutArena(arena)

			// 2. Create Request Scope in the Arena
			reqScope := arena.AllocScope(baseScope)

			// 2. Inject URL Params (e.g., /news/{id} -> $id)
			rctx := chi.RouteContext(r.Context())
			if rctx != nil && len(rctx.URLParams.Keys) > 0 {
				params := engine.GetMap()
				defer engine.PutMap(params)

				for i, key := range rctx.URLParams.Keys {
					val := rctx.URLParams.Values[i]
					// Set as global scope variable: $id
					reqScope.Set(key, val)
					// Set also in params map: $params.id
					params[key] = val
				}
				reqScope.Set("params", params)
			}

			// 3. Inject Form Data (POST/PUT)
			r.ParseMultipartForm(32 << 20) // 32 MB limit

			formData := engine.GetMap()
			defer engine.PutMap(formData)

			for k, v := range r.Form {
				if len(v) == 1 {
					formData[k] = v[0]
				} else {
					formData[k] = v
				}
			}

			reqScope.Set("form", formData)

			// 4. Parse JSON Body (for API requests)
			var bodyData map[string]interface{}
			if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
				contentType := r.Header.Get("Content-Type")
				if strings.Contains(contentType, "application/json") {
					bodyData = engine.GetMap()
					defer engine.PutMap(bodyData)

					decoder := json.NewDecoder(r.Body)
					if err := decoder.Decode(&bodyData); err == nil {
						// Successfully parsed JSON
					} else {
						// If JSON parse fails, use empty map
						bodyData = make(map[string]interface{})
					}
				} else {
					bodyData = make(map[string]interface{})
				}
			} else {
				bodyData = make(map[string]interface{})
			}

			// 5. Build $request object
			requestObj := engine.GetMap()
			defer engine.PutMap(requestObj)

			requestObj["method"] = r.Method
			requestObj["url"] = r.URL.String()
			requestObj["path"] = r.URL.Path
			requestObj["body"] = bodyData

			// Shortcut variables
			reqScope.Set("path", r.URL.Path)
			reqScope.Set("method", r.Method)

			// Add headers as map
			headersMap := engine.GetMap()
			defer engine.PutMap(headersMap)
			for k, v := range r.Header {
				if len(v) == 1 {
					headersMap[k] = v[0]
				} else {
					headersMap[k] = v
				}
			}
			requestObj["headers"] = headersMap

			// Add query params
			queryMap := engine.GetMap()
			defer engine.PutMap(queryMap)

			for k, v := range r.URL.Query() {
				if len(v) == 1 {
					queryMap[k] = v[0]
				} else {
					queryMap[k] = v
				}
			}
			requestObj["query"] = queryMap

			reqScope.Set("request", requestObj)

			// 6. Inject HTTP context (for middleware/slots that need it)
			ctx := context.WithValue(r.Context(), "httpRequest", r)
			ctx = context.WithValue(ctx, "httpWriter", w)

			// [NEW] 6.1. Add timeout to prevent infinite loops
			// Default: 30 seconds, configurable via ZENO_REQUEST_TIMEOUT
			timeoutStr := os.Getenv("ZENO_REQUEST_TIMEOUT")
			if timeoutStr == "" {
				timeoutStr = "30s" // Default timeout
			}
			timeout, err := time.ParseDuration(timeoutStr)
			if err != nil {
				timeout = 30 * time.Second // Fallback to 30s if parsing fails
			}

			// Create timeout context
			timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			// [NEW] 7. Inject Auth from Chi middleware context to ZenoLang scope
			// This bridges native Chi middleware (MultiTenantAuth) to ZenoLang scope
			middleware.InjectAuthToScope(r, reqScope)

			// 8. Execute Children (Route Logic) - Auth already injected from Chi middleware
			for _, child := range children {
				if err := eng.Execute(timeoutCtx, child, reqScope); err != nil {
					// [NEW] Handle ErrReturn (Normal Halt)
					if errors.Is(err, ErrReturn) || strings.Contains(err.Error(), "return") {
						return
					}

					// Check if error is due to timeout
					if timeoutCtx.Err() == context.DeadlineExceeded {
						http.Error(w, fmt.Sprintf("Request timeout exceeded (%s)", timeout), http.StatusRequestTimeout)
						return
					}
					panic(err) // Will be caught by recovery middleware
				}
			}
		}
	}

	// Helper: Parse Path dari Node (Standardized)
	getPath := func(node *engine.Node, scope *engine.Scope) string {
		path := coerce.ToString(resolveValue(node.Value, scope))
		if path == "" {
			for _, c := range node.Children {
				if c.Name == "path" || c.Name == "url" {
					path = coerce.ToString(parseNodeValue(c, scope))
				}
			}
		}
		return path
	}

	// Helper context for path tracking
	type pathPrefixKey struct{}

	getCurrentPath := func(ctx context.Context) string {
		if p, ok := ctx.Value(pathPrefixKey{}).(string); ok {
			return p
		}
		return ""
	}

	joinPath := func(base, sub string) string {
		if base == "" {
			return sub
		}
		if base == "/" && sub == "/" {
			return "/"
		}
		// Remove trailing slash from base
		base = strings.TrimSuffix(base, "/")
		if !strings.HasPrefix(sub, "/") {
			sub = "/" + sub
		}
		return base + sub
	}

	// ==========================================
	// 1. ROUTE GROUP (Mendukung Implicit Do)
	// ==========================================
	eng.Register("http.group", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		path := getPath(node, scope)

		// Check if group has middleware
		middlewareName := ""
		for _, c := range node.Children {
			if c.Name == "middleware" {
				middlewareName = coerce.ToString(resolveValue(c.Value, scope))
			}
		}

		// Logic: Cari 'do'. Jika tidak ada, pakai 'node' itu sendiri (Implicit)
		var childrenToExec []*engine.Node
		var doNode *engine.Node

		for _, c := range node.Children {
			if c.Name == "do" {
				doNode = c
				break
			}
		}

		if doNode != nil {
			childrenToExec = doNode.Children
		} else {
			// Implicit Mode: filter out config nodes
			for _, c := range node.Children {
				if c.Name != "middleware" && c.Name != "summary" && c.Name != "desc" {
					childrenToExec = append(childrenToExec, c)
				}
			}
		}

		// Create sub-router
		subRouter := chi.NewRouter()

		// [NEW] Apply native Chi middleware if auth is specified
		if middlewareName == "auth" {
			// Use JWT_SECRET from environment (same as auth controller)
			jwtSecret := os.Getenv("JWT_SECRET")
			if jwtSecret == "" {
				// Fallback to .env default
				jwtSecret = "rahasia_dapur_pekalongan_kota_2025_!@#_jgn_disebar"
				fmt.Printf("   ⚠️  Using default JWT_SECRET\n")
			}
			subRouter.Use(middleware.MultiTenantAuth(jwtSecret))
			fmt.Printf("   🔒 [GROUP MIDDLEWARE] Applied native Chi auth to group %s\n", path)
		}

		// Mount sub-router
		getCurrentRouter(ctx).Mount(path, subRouter)

		// Create new context with sub-router
		groupCtx := context.WithValue(ctx, routerKey{}, subRouter)

		// Execute children in group context
		for _, child := range childrenToExec {
			eng.Execute(groupCtx, child, scope)
		}

		return nil
	}, engine.SlotMeta{})

	// ==========================================
	// 2. STANDARD HTTP METHODS (Mendukung Implicit Do)
	// ==========================================
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
	for _, method := range methods {
		m := method // capture loop var
		eng.Register("http."+strings.ToLower(m), func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
			path := getPath(node, scope)

			// 2. Metadata Extraction & Clean Children

			// Resolve Full Path for Documentation
			fullDocPath := joinPath(getCurrentPath(ctx), path)

			routeDoc := &apidoc.RouteDoc{
				Method:    m,
				Path:      fullDocPath,
				Responses: make(map[string]apidoc.ResponseDoc),
			}

			var doNode *engine.Node
			var middlewareName string

			// Scan for Metadata and Logic Container
			for _, c := range node.Children {
				if c.Name == "do" {
					doNode = c
				}

				// Metadata Extraction
				if c.Name == "summary" {
					routeDoc.Summary = coerce.ToString(resolveValue(c.Value, scope))
				}
				if c.Name == "desc" || c.Name == "description" {
					routeDoc.Description = coerce.ToString(resolveValue(c.Value, scope))
				}
				if c.Name == "tags" {
					val := resolveValue(c.Value, scope)
					if list, err := coerce.ToSlice(val); err == nil {
						tags := make([]string, len(list))
						for i, v := range list {
							tags[i] = coerce.ToString(v)
						}
						routeDoc.Tags = tags
					}
				}

				// Capture Middleware (Metadata Level)
				// Support both: middleware: "auth" AND middleware with parameters as route attributes
				if c.Name == "middleware" {
					if c.Value != nil {
						middlewareName = coerce.ToString(resolveValue(c.Value, scope))
					}
				}

				// Extract Query Params
				if c.Name == "query" {
					if m, ok := parseNodeValue(c, scope).(map[string]interface{}); ok {
						for k, v := range m {
							desc := coerce.ToString(v)
							pType := "string"
							required := false

							// Simple syntax parsing: "Description|type|required"
							parts := strings.Split(desc, "|")
							if len(parts) > 0 {
								desc = parts[0]
							}
							if len(parts) > 1 {
								pType = parts[1]
							}
							if strings.Contains(desc, "required") || (len(parts) > 2 && parts[2] == "required") {
								required = true
							}

							routeDoc.Params = append(routeDoc.Params, apidoc.ParamDoc{
								Name:        k,
								In:          "query",
								Description: desc,
								Type:        pType,
								Required:    required,
							})
						}
					}
				}

				// Extract Path Params
				if c.Name == "params" {
					if m, ok := parseNodeValue(c, scope).(map[string]interface{}); ok {
						for k, v := range m {
							desc := coerce.ToString(v)
							pType := "string"
							// Path params are always required

							parts := strings.Split(desc, "|")
							if len(parts) > 0 {
								desc = parts[0]
							}
							if len(parts) > 1 {
								pType = parts[1]
							}

							routeDoc.Params = append(routeDoc.Params, apidoc.ParamDoc{
								Name:        k,
								In:          "path",
								Description: desc,
								Type:        pType,
								Required:    true,
							})
						}
					}
				}
			}

			// Prepare execution children (filtering config nodes)
			var execChildren []*engine.Node
			if doNode != nil {
				for _, child := range doNode.Children {
					execChildren = append(execChildren, child)
				}
			} else {
				for _, child := range node.Children {
					name := child.Name
					if name == "do" || name == "summary" || name == "desc" || name == "tags" || name == "body" || name == "query" || name == "middleware" {
						continue
					}
					execChildren = append(execChildren, child)
				}
			}

			// [NEW] Apply Native Chi Middleware using r.With() pattern
			// This is the idiomatic Go/Chi way for route-specific middleware
			targetRouter := getCurrentRouter(ctx)

			if middlewareName == "auth" {
				// Create a new router chain with middleware applied
				// Use JWT_SECRET from environment (same as auth controller)
				jwtSecret := os.Getenv("JWT_SECRET")
				if jwtSecret == "" {
					// Fallback to .env default
					jwtSecret = "rahasia_dapur_pekalongan_kota_2025_!@#_jgn_disebar"
					fmt.Printf("   ⚠️  Using default JWT_SECRET\n")
				}
				targetRouter = targetRouter.With(middleware.MultiTenantAuth(jwtSecret))
				fmt.Printf("   🔒 [MIDDLEWARE] Applied native Chi auth via r.With() to %s\n", fullDocPath)
			}

			// Register Documentation
			apidoc.Registry.Register(m, fullDocPath, routeDoc)

			fmt.Printf("   ➕ [ROUTE] %-6s %s\n", m, fullDocPath)

			// Register route handler on the middleware-enabled router chain
			targetRouter.MethodFunc(m, path, createHandler(execChildren, scope))
			return nil
		}, engine.SlotMeta{})
	}
}
