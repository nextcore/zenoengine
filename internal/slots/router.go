package slots

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"zeno/pkg/apidoc"
	"zeno/pkg/engine"
	hostPkg "zeno/pkg/host"
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
	// 0. HOST / DOMAIN GROUP
	// ==========================================
	eng.Register("http.host", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		host := coerce.ToString(resolveValue(node.Value, scope))
		if host == "" {
			return fmt.Errorf("http.host: domain/host is required")
		}

		// Create host-specific router
		hostRouter := chi.NewRouter()

		// [AUTOMATIC] Register to Native Host Map (O(1) lookup)
		// This is much faster than linear middleware checks
		hostPkg.GlobalManager.RegisterRouter(host, hostRouter)

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
			for _, c := range node.Children {
				if c.Name != "summary" && c.Name != "desc" {
					childrenToExec = append(childrenToExec, c)
				}
			}
		}

		// Create new context with host-router
		hostCtx := context.WithValue(ctx, routerKey{}, hostRouter)

		// Execute children in host context
		for _, child := range childrenToExec {
			eng.Execute(hostCtx, child, scope)
		}

		fmt.Printf("   🌐 [VHOST] Registered domain: %s\n", host)
		return nil
	}, engine.SlotMeta{
		Description: "Mengelompokkan route berdasarkan Domain atau Subdomain tertentu.",
		Group:       "HTTP",
		Example:     "http.host: \"api.zeno.dev\"\n  do:\n    http.get: \"/v1/users\" { ... }",
	})

	// ==========================================
	// 1. ROUTE GROUP (Mendukung Implicit Do)
	// ==========================================
	eng.Register("http.group", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		path := getPath(node, scope)

		// Check if group has middleware
		var middlewares []string
		for _, c := range node.Children {
			if c.Name == "middleware" {
				val := resolveValue(c.Value, scope)
				if slice, err := coerce.ToSlice(val); err == nil {
					for _, item := range slice {
						middlewares = append(middlewares, coerce.ToString(item))
					}
				} else {
					// Fallback: Parse string representation "[a, b]"
					s := coerce.ToString(val)
					if strings.HasPrefix(strings.TrimSpace(s), "[") {
						content := strings.Trim(strings.TrimSpace(s), "[]")
						parts := strings.Split(content, ",")
						for _, p := range parts {
							middlewares = append(middlewares, strings.Trim(strings.TrimSpace(p), "\"'"))
						}
					} else {
						middlewares = append(middlewares, s)
					}
				}
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

		// [NEW] Apply Middleware Stack
		for _, m := range middlewares {
			if m == "auth" {
				// Use JWT_SECRET from environment
				jwtSecret := os.Getenv("JWT_SECRET")
				if jwtSecret == "" {
					jwtSecret = "458127c2cffdd41a448b5d37b825188bf12db10e5c98cb03b681da667ac3b294_pekalongan_kota_2025_!@#_jgn_disebar"
					fmt.Printf("   ⚠️  Using default JWT_SECRET\n")
				}
				subRouter.Use(middleware.MultiTenantAuth(jwtSecret))
				fmt.Printf("   🔒 [GROUP MIDDLEWARE] Applied 'auth' to group %s\n", path)
			}
			// Future: Add other middlewares here (e.g. "throttle", "cors")
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
	}, engine.SlotMeta{
		Description: "Groups routes under a common path prefix and optional middleware.",
		Group:       "HTTP",
		Example:     "http.group: '/admin' {\n  middleware: 'auth'\n  do: { ... }\n}",
	})

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
			var middlewares []string

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
				if c.Name == "middleware" {
					if c.Value != nil {
						val := resolveValue(c.Value, scope)
						if slice, err := coerce.ToSlice(val); err == nil {
							for _, item := range slice {
								middlewares = append(middlewares, coerce.ToString(item))
							}
						} else {
							// Fallback: Parse string representation "[a, b]"
							s := coerce.ToString(val)
							if strings.HasPrefix(strings.TrimSpace(s), "[") {
								content := strings.Trim(strings.TrimSpace(s), "[]")
								parts := strings.Split(content, ",")
								for _, p := range parts {
									middlewares = append(middlewares, strings.Trim(strings.TrimSpace(p), "\"'"))
								}
							} else {
								middlewares = append(middlewares, s)
							}
						}
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

			for _, m := range middlewares {
				if m == "auth" {
					// Use JWT_SECRET from environment
					jwtSecret := os.Getenv("JWT_SECRET")
					if jwtSecret == "" {
						jwtSecret = "458127c2cffdd41a448b5d37b825188bf12db10e5c98cb03b681da667ac3b294_pekalongan_kota_2025_!@#_jgn_disebar"
						fmt.Printf("   ⚠️  Using default JWT_SECRET\n")
					}
					targetRouter = targetRouter.With(middleware.MultiTenantAuth(jwtSecret))
					fmt.Printf("   🔒 [MIDDLEWARE] Applied 'auth' to %s\n", fullDocPath)
				}
			}

			// Register Documentation
			apidoc.Registry.Register(m, fullDocPath, routeDoc)

			fmt.Printf("   ➕ [ROUTE] %-6s %s\n", m, fullDocPath)

			// Register route handler on the middleware-enabled router chain
			targetRouter.MethodFunc(m, path, createHandler(execChildren, scope))
			return nil
		}, engine.SlotMeta{
			Description: fmt.Sprintf("Register a %s route handler.", m),
			Group:       "HTTP",
			Example:     fmt.Sprintf("http.%s: '/users' { ... }", strings.ToLower(m)),
		})
	}

	// ==========================================
	// 3. REVERSE PROXY SLOT (Caddy-Style)
	// ==========================================
	eng.Register("http.proxy", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		targetStr := coerce.ToString(resolveValue(node.Value, scope))
		if targetStr == "" {
			for _, c := range node.Children {
				if c.Name == "to" || c.Name == "target" {
					targetStr = coerce.ToString(parseNodeValue(c, scope))
				}
			}
		}

		if targetStr == "" {
			return fmt.Errorf("http.proxy: target URL is required")
		}

		targetURL, err := url.Parse(targetStr)
		if err != nil {
			return fmt.Errorf("http.proxy: invalid target URL: %v", err)
		}

		path := "/"
		for _, c := range node.Children {
			if c.Name == "path" {
				path = coerce.ToString(parseNodeValue(c, scope))
			}
		}

		// Create Reverse Proxy
		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		// [OPTIONAL] Customizing the Director to handle headers correctly
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			req.Host = targetURL.Host // Critical for some backends
			req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
		}

		// Register to router
		getCurrentRouter(ctx).Handle(path+"*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Strip prefix if not root
			if path != "/" {
				http.StripPrefix(strings.TrimSuffix(path, "/"), proxy).ServeHTTP(w, r)
			} else {
				proxy.ServeHTTP(w, r)
			}
		}))

		fmt.Printf("   🔄 [PROXY] Registered proxy: %s -> %s\n", path, targetStr)
		return nil
	}, engine.SlotMeta{
		Description: "Meneruskan request ke backend service lain (Reverse Proxy).",
		Group:       "HTTP",
		Example:     "http.proxy: \"http://localhost:8080\"\n  path: \"/api\"",
	})

	// ==========================================
	// 4. STATIC / SPA HOSTING SLOT
	// ==========================================
	eng.Register("http.static", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		root := coerce.ToString(resolveValue(node.Value, scope))
		path := "/"
		isSPA := false

		for _, c := range node.Children {
			if c.Name == "root" || c.Name == "dir" {
				root = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "path" {
				path = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "spa" {
				isSPA, _ = coerce.ToBool(parseNodeValue(c, scope))
			}
		}

		if root == "" {
			return fmt.Errorf("http.static: root directory is required")
		}

		// Ensure path ends with * for Chi wildcard matching
		routePath := path
		if !strings.HasSuffix(routePath, "/") {
			routePath += "/"
		}

		fileServer := http.FileServer(http.Dir(root))

		getCurrentRouter(ctx).Handle(routePath+"*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Clean path and check if file exists
			cleanPath := filepath.Join(root, strings.TrimPrefix(r.URL.Path, path))

			// [SECURITY] Prevent Path Traversal
			// Ensure cleanPath is effectively inside root
			// We use filepath.Rel to check if the path attempts to go above root
			rel, err := filepath.Rel(root, cleanPath)
			if err != nil || strings.HasPrefix(rel, "..") {
				// Traversal attempt detected
				if isSPA {
					// For SPA, treat traversal as "page not found" -> serve index
					// This prevents Oracle attacks (distinguishing files via 404 vs 200)
					http.ServeFile(w, r, filepath.Join(root, "index.html"))
					return
				}
				// For non-SPA, return 404
				http.NotFound(w, r)
				return
			}

			_, err = os.Stat(cleanPath)

			// 2. If SPA and file not found, serve index.html
			if isSPA && os.IsNotExist(err) {
				http.ServeFile(w, r, filepath.Join(root, "index.html"))
				return
			}

			// 3. Regular file serving
			if path != "/" {
				http.StripPrefix(strings.TrimSuffix(path, "/"), fileServer).ServeHTTP(w, r)
			} else {
				fileServer.ServeHTTP(w, r)
			}
		}))

		mode := "Static Site"
		if isSPA {
			mode = "SPA (Single Page App)"
		}
		fmt.Printf("   📁 [STATIC] Registered %s: %s -> %s\n", mode, path, root)
		return nil
	}, engine.SlotMeta{
		Description: "Hosting aplikasi SPA (React/Vue) atau Static Site.",
		Group:       "HTTP",
		Example:     "http.static: \"./dist\"\n  path: \"/\"\n  spa: true",
	})
}
