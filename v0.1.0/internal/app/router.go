package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"zeno/internal/console"
	"zeno/pkg/apidoc"
	"zeno/pkg/engine"
	"zeno/pkg/logger"
	"zeno/pkg/metrics"
	"zeno/pkg/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/gorilla/csrf"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// BuildRouter: Membaca main.zl dan membangun Router Chi baru
func BuildRouter(app *AppContext) (*chi.Mux, error) {
	r := chi.NewRouter()
	r.Use(logger.Middleware)
	r.Use(metrics.Middleware) // PROMETHEUS METRICS (Place early)
	r.Use(middleware.Recoverer)
	r.Use(middleware.SecurityHeaders) // New Security Middleware

	// Rate Limiting
	// Rate Limiting
	rlReqStr := os.Getenv("RATE_LIMIT_REQUESTS")
	if rlReqStr == "" {
		slog.Info("⚠️  Rate Limiting Disabled (RATE_LIMIT_REQUESTS not set)")
	} else {
		rlRequests, _ := strconv.Atoi(rlReqStr)
		if rlRequests == 0 {
			rlRequests = 100
		}
		rlWindow, _ := strconv.Atoi(os.Getenv("RATE_LIMIT_WINDOW"))
		if rlWindow == 0 {
			rlWindow = 60
		}
		r.Use(httprate.LimitByIP(rlRequests, time.Duration(rlWindow)*time.Second))
	}

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
	}))

	// CSRF Protection
	port := os.Getenv("APP_PORT")
	CSRF := csrf.Protect(
		[]byte(os.Getenv("CSRF_TOKEN")),
		csrf.Secure(false),
		csrf.Path("/"),
		csrf.TrustedOrigins([]string{
			"localhost",
			"localhost:3000",
			"http://localhost",
			"http://localhost:" + port,
			"127.0.0.1",
			"127.0.0.1:" + port,
			"http://127.0.0.1",
			"http://127.0.0.1:" + port,
		}),
		csrf.SameSite(csrf.SameSiteLaxMode),
	)

	// Apply CSRF but skip for /api and /health
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			path := req.URL.Path
			if strings.HasPrefix(path, "/api") || path == "/health" {
				next.ServeHTTP(w, req)
				return
			}
			// CSRF middleware will automatically parse form data when needed
			CSRF(next).ServeHTTP(w, req)
		})
	})

	// Health Check (No CSRF)
	r.Get("/health", func(w http.ResponseWriter, req *http.Request) {
		if err := app.DBMgr.GetConnection("default").Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("DOWN: Database Error"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Prometheus Metrics Endpoint
	r.Handle("/metrics", promhttp.Handler())

	// 4. Update signatures
	eng := engine.NewEngine()
	RegisterAllSlots(eng, r, app.DBMgr, app.Queue, func(queues []string) {
		app.WorkerQueues = queues
		slog.Info("🔧 Worker Configuration Updated", "queues", queues)
	})

	// Static Files
	workDir, _ := os.Getwd()
	filesDir := filepath.Join(workDir, "public")
	r.Get("/public/*", func(w http.ResponseWriter, req *http.Request) {
		rctx := chi.RouteContext(req.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(http.Dir(filesDir)))
		fs.ServeHTTP(w, req)
	})

	// Console Developer & API Docs
	if app.Env == "development" {
		console.RegisterRoutes(r, eng)

		// OpenAPI JSON
		r.Get("/api/docs/json", func(w http.ResponseWriter, req *http.Request) {
			json, err := apidoc.Registry.ToJSON()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(json)
		})

		// Swagger UI
		r.Get("/api/docs", func(w http.ResponseWriter, req *http.Request) {
			html := `<!DOCTYPE html>
<html lang="en">
<head>
    <title>Zeno API Docs</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css" />
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
<script>
window.onload = function() {
  window.ui = SwaggerUIBundle({
    url: "/api/docs/json",
    dom_id: '#swagger-ui',
  });
};
</script>
</body>
</html>`
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(html))
		})
	}

	// Exec Main Script
	mainScriptPath := "src/main.zl"
	root, err := engine.LoadScript(mainScriptPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load script: %v", err)
	}

	globalScope := engine.GetScope()
	defer engine.PutScope(globalScope)
	globalScope.Set("APP_ENV", app.Env)

	if err := eng.Execute(context.Background(), root, globalScope); err != nil {
		return nil, fmt.Errorf("execution error: %v", err)
	}

	return r, nil
}
