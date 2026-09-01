package slots

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"github.com/nextcore/zenoengine/pkg/apidoc"
	"github.com/nextcore/zeno-go/pkg/engine"
	"github.com/nextcore/zenoengine/pkg/middleware"
	"github.com/nextcore/zeno-go/pkg/utils/coerce"
	pkgslots "github.com/nextcore/zeno-go/pkg/slots"

	"github.com/go-chi/chi/v5"
)

// Key context untuk menyimpan router instance
type routerKey struct{}

// [NEW] Registry for ZenoLang-defined custom middlewares
var customMiddlewares = make(map[string]*engine.Node)

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
			// [NEW] Execute optional route-specific custom middleware first?
			// Actually, Chi middleware chain handles this better.
			// But if we want to support multiple custom middlewares, we'd need a more complex system.
			// For now, custom middlewares are usually pre-processed or handled via `r.With`.

			// 1. Get Scope from pool for this request (avoids Arena GC pointer hiding issues)
			reqScope := engine.GetScope()
			reqScope.Reset() // Ensure clean state
			reqScope.SetParent(baseScope)
			defer engine.PutScope(reqScope)

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
					if errors.Is(err, pkgslots.ErrReturn) || strings.Contains(err.Error(), "return") {
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
				jwtSecret = "458127c2cffdd41a448b5d37b825188bf12db10e5c98cb03b681da667ac3b294_pekalongan_kota_2025_!@#_jgn_disebar"
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
					jwtSecret = "458127c2cffdd41a448b5d37b825188bf12db10e5c98cb03b681da667ac3b294_pekalongan_kota_2025_!@#_jgn_disebar"
					fmt.Printf("   🔒 [MIDDLEWARE] Applied default JWT secret\n")
				}
				targetRouter = targetRouter.With(middleware.MultiTenantAuth(jwtSecret))
				fmt.Printf("   🔒 [MIDDLEWARE] Applied native Chi auth via r.With() to %s\n", fullDocPath)
			} else if midNode, exists := customMiddlewares[middlewareName]; exists {
				// [NEW] Bridge ZenoLang middleware to Chi middleware
				targetRouter = targetRouter.With(func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						// Create request scope from pool
						reqScope := engine.GetScope()
						reqScope.Reset()
						reqScope.SetParent(scope)
						defer engine.PutScope(reqScope)

						// Inject HTTP context
						ctx := context.WithValue(r.Context(), "httpRequest", r)
						ctx = context.WithValue(ctx, "httpWriter", w)

						// Execute middleware node
						// We look for 'do' block inside the middleware definition
						var doBlock *engine.Node
						for _, child := range midNode.Children {
							if child.Name == "do" {
								doBlock = child
								break
							}
						}

						if doBlock != nil {
							if err := eng.Execute(ctx, doBlock, reqScope); err != nil {
								// If middleware calls return or has error, don't call next
								if errors.Is(err, pkgslots.ErrReturn) || strings.Contains(err.Error(), "return") {
									return
								}
								// Log error and stop
								fmt.Printf("   ❌ [MIDDLEWARE ERROR] %s: %v\n", middlewareName, err)
								http.Error(w, "Middleware Error", http.StatusInternalServerError)
								return
							}
						}

						// If we reach here and 'http.next' was called (or just default to continue)
						// We check if $http_next was set in scope?
						// Actually, standard ZenoLang middleware pattern uses 'return' to STOP.
						// So if it finished normally, we continue.
						next.ServeHTTP(w, r)
					})
				})
				fmt.Printf("   🛡️ [MIDDLEWARE] Applied custom ZenoLang middleware '%s' to %s\n", middlewareName, fullDocPath)
			}

			// Register Documentation
			apidoc.Registry.Register(m, fullDocPath, routeDoc)

			fmt.Printf("   ➕ [ROUTE] %-6s %s\n", m, fullDocPath)

			// Register route handler on the middleware-enabled router chain
			targetRouter.MethodFunc(m, path, createHandler(execChildren, scope))
			return nil
		}, engine.SlotMeta{})
	}



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
		Example:     "http.static: \"./dist\"\n  path: \"/\"\n  spa: true",
	})

	// ==========================================
	// 5. HTTP ROUTES INTROSPECTION
	// ==========================================
	eng.Register("http.routes", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		target := "routes"
		for _, c := range node.Children {
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		routes := make([]map[string]interface{}, 0)

		fmt.Println("   [DEBUG] Walking rootRouter...")
		err := chi.Walk(rootRouter, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
			// clean up method because chi can return multiple methods
			methods := strings.Split(method, ",")
			for _, m := range methods {
				if m != "" && m != "*" {
					r := make(map[string]interface{})
					r["method"] = strings.TrimSpace(m)
					r["path"] = route
					routes = append(routes, r)
				}
			}
			return nil
		})

		if err != nil {
			fmt.Printf("   [DEBUG] chi.Walk Error: %v\n", err)
			return err
		}

		fmt.Printf("   [DEBUG] chi.Walk found %d routes\n", len(routes))

		scope.Set(target, routes)
		return nil
	}, engine.SlotMeta{
		Description: "Mengambil daftar semua rute HTTP yang terdaftar di engine.",
		Example:     "http.routes: { as: $routes }",
	})

	// ==========================================
	// 6. CUSTOM MIDDLEWARE DEFINITION
	// ==========================================
	eng.Register("http.middleware", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		name := coerce.ToString(resolveValue(node.Value, scope))
		if name == "" {
			return fmt.Errorf("http.middleware: name is required")
		}

		// Store the entire node for later execution
		customMiddlewares[name] = node
		fmt.Printf("   🛡️ [MIDDLEWARE] Defined ZenoLang middleware: %s\n", name)
		return nil
	}, engine.SlotMeta{
		Description: "Mendefinisikan middleware kustom menggunakan ZenoLang.",
		Example:     "http.middleware: 'auth' {\n  do: {\n    session.get: 'user_id' { as: $uid }\n    if: $uid == null { then: { http.redirect: '/login' } }\n  }\n}",
	})

	// 7. HTTP.NEXT (Middleware Continuity)
	eng.Register("http.next", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		// Currently a no-op as the middleware bridge proceeds by default.
		return nil
	}, engine.SlotMeta{Description: "Melanjutkan ke handler berikutnya dalam rantai middleware."})

	// 8. MIDDLEWARE SPOOF (Laravel/PHP Security Obfuscation)
	eng.Register("middleware.spoof", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		r := getCurrentRouter(ctx)
		if r == nil {
			return fmt.Errorf("middleware.spoof: router context not found")
		}

		enabled := true
		xPoweredBy := "PHP/8.3.0"
		laravelSession := "eyJpdiI6IlZGVk..."
		phpSessID := "sess_89a7f3..."

		for _, c := range node.Children {
			val := parseNodeValue(c, scope)
			switch c.Name {
			case "enabled":
				if b, err := coerce.ToBool(val); err == nil {
					enabled = b
				}
			case "x_powered_by":
				xPoweredBy = coerce.ToString(val)
			case "laravel_session":
				laravelSession = coerce.ToString(val)
			case "php_sessid":
				phpSessID = coerce.ToString(val)
			}
		}

		if !enabled {
			return nil
		}

		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("X-Powered-By", xPoweredBy)
				isSecure := req.TLS != nil || strings.ToLower(req.Header.Get("X-Forwarded-Proto")) == "https"

				if _, err := req.Cookie("laravel_session"); err != nil {
					http.SetCookie(w, &http.Cookie{
						Name:     "laravel_session",
						Value:    laravelSession,
						Path:     "/",
						HttpOnly: true,
						Secure:   isSecure,
						SameSite: http.SameSiteLaxMode,
					})
				}
				if _, err := req.Cookie("PHPSESSID"); err != nil {
					http.SetCookie(w, &http.Cookie{
						Name:     "PHPSESSID",
						Value:    phpSessID,
						Path:     "/",
						HttpOnly: true,
						Secure:   isSecure,
						SameSite: http.SameSiteLaxMode,
					})
				}
				next.ServeHTTP(w, req)
			})
		})
		fmt.Printf("   🛡️ [MIDDLEWARE] Applied SpoofLaravelMiddleware\n")
		return nil
	}, engine.SlotMeta{
		Description: "Memasang header & cookie penyamaran PHP/Laravel untuk obfuscation keamanan server.",
		Example:     "middleware.spoof:\n  enabled: true\n  x_powered_by: 'PHP/8.3.0'",
	})

	// 9. MIDDLEWARE API KEY
	eng.Register("middleware.api_key", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		r := getCurrentRouter(ctx)
		if r == nil {
			return fmt.Errorf("middleware.api_key: router context not found")
		}

		envKey := "API_KEY"
		defaultKey := ""

		for _, c := range node.Children {
			val := parseNodeValue(c, scope)
			switch c.Name {
			case "env":
				envKey = coerce.ToString(val)
			case "key", "default_key":
				defaultKey = coerce.ToString(val)
			}
		}

		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				expectedKey := os.Getenv(envKey)
				if expectedKey == "" {
					expectedKey = defaultKey
				}

				clientKey := req.Header.Get("X-API-KEY")
				if clientKey == "" {
					clientKey = req.URL.Query().Get("api_key")
				}
				if clientKey == "" {
					authHeader := req.Header.Get("Authorization")
					if strings.HasPrefix(authHeader, "Bearer ") {
						clientKey = strings.TrimPrefix(authHeader, "Bearer ")
					}
				}

				if expectedKey != "" && clientKey != expectedKey {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`{"status":"error","message":"Unauthorized: Invalid or missing API Key"}`))
					return
				}
				next.ServeHTTP(w, req)
			})
		})
		fmt.Printf("   🔑 [MIDDLEWARE] Applied ApiKeyMiddleware (env: %s)\n", envKey)
		return nil
	}, engine.SlotMeta{
		Description: "Memverifikasi X-API-KEY header, query param, atau Bearer token pada request HTTP.",
		Example:     "middleware.api_key:\n  env: 'MY_API_KEY'\n  key: 'secret123'",
	})
}
