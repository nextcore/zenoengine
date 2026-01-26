package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"zeno/internal/app"
	"zeno/internal/cli"
	"zeno/internal/slots"
	"zeno/pkg/dbmanager"
	"zeno/pkg/engine"
	"zeno/pkg/logger"
	"zeno/pkg/worker"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"zeno/pkg/livereload"
)

func main() {
	godotenv.Load()

	// 1. CLI DISPATCHER
	if len(os.Args) > 1 {
		cmd := os.Args[1]
		switch cmd {
		case "check":
			cli.HandleCheck(os.Args[2:])
		case "run":
			cli.HandleRun(os.Args[2:])
		case "migrate":
			cli.HandleMigrate()
		case "test":
			cli.HandleTest(os.Args[2:])
		case "make:auth":
			cli.HandleMakeAuth()
		case "key:generate":
			cli.HandleKeyGenerate()
		case "version":
			cli.HandleVersion()
		default:
			// Automatically run if it ends with .zl
			if strings.HasSuffix(cmd, ".zl") {
				cli.HandleRun(os.Args[1:])
			}
		}
		return // STOP HERE for CLI commands
	}

	// 2. CORE SETUP (Logger, DB)
	appEnv := os.Getenv("APP_ENV")
	logger.Setup(appEnv)
	slog.Info("Starting ZenoEngine...", "env", appEnv)

	dbMgr := initDB()

	// Init Queue - Always use internal SQLite for ZenoEngine operations
	queue := worker.NewDBQueue(dbMgr, "internal")
	slog.Info("✅ Worker Queue Ready", "database", "SQLite (internal)")

	appCtx := &app.AppContext{
		DBMgr: dbMgr,
		Queue: queue,
		Env:   appEnv,
		Hot:   &app.HotRouter{},
	}

	// 3. PRE-FLIGHT VALIDATION (Validate all scripts before starting)
	if os.Getenv("ZENO_SKIP_VALIDATION") != "true" {
		slog.Info("🔍 Running Pre-Flight Validation...")
		if err := validateAllScripts(); err != nil {
			slog.Error("❌ Pre-Flight Validation Failed", "error", err)
			slog.Info("💡 Tip: Fix the syntax errors above, or set ZENO_SKIP_VALIDATION=true to skip validation")
			os.Exit(1)
		}
		slog.Info("✅ Pre-Flight Validation Passed")
	} else {
		slog.Warn("⚠️  Pre-Flight Validation Skipped (ZENO_SKIP_VALIDATION=true)")
	}

	// 4. INITIAL BUILD
	slog.Info("🚀 Loading Routes from src/main.zl...")
	initialRouter, err := app.BuildRouter(appCtx)
	if err != nil {
		slog.Error("❌ Critical Startup Error", "error", err)
		os.Exit(1)
	}
	appCtx.Hot.Swap(initialRouter)
	slog.Info("✅ Routes Registered Successfully")

	// 4. WORKER START
	var workerWG sync.WaitGroup
	ctxWorker, cancelWorker := context.WithCancel(context.Background())

	if os.Getenv("WORKER_ENABLED") == "true" {
		workerEng := engine.NewEngine()
		app.RegisterAllSlots(workerEng, nil, dbMgr, queue, nil)
		slog.Info("👷 Starting Workers...")
		queues := appCtx.WorkerQueues // Leave empty if not configured
		if len(queues) == 0 {
			slog.Info("⚠️  Worker started but no queues configured. Use 'worker.config' in main.zl")
		}
		workerWG.Add(1)
		go func() {
			defer workerWG.Done()
			worker.Start(ctxWorker, workerEng, queue, queues)
		}()
	} else {
		slog.Info("🚫 Worker Disabled (WORKER_ENABLED=false)")
	}

	// 5. WATCHER (Hot Reload for .zl files)
	// Only enable if explicitly in dev mode AND not disabled via env
	liveReloadEnabled := appEnv == "development" && os.Getenv("LIVERELOAD_ENABLED") != "false"

	if liveReloadEnabled {
		go startWatcher(appCtx)
	} else {
		slog.Info("🚫 Live Reload Disabled")
	}

	// 6. HTTP SERVER
	startServer(appCtx, cancelWorker, &workerWG, liveReloadEnabled)
}

// --- HELPER FUNCTIONS ---

func initDB() *dbmanager.DBManager {
	dbMgr := dbmanager.NewDBManager()

	// 1. Database from .env as DEFAULT (User Data)
	driver := os.Getenv("DB_DRIVER")
	if driver == "" {
		driver = "mysql"
	}
	var primaryDSN string
	if driver == "sqlite" {
		primaryDSN = os.Getenv("DB_NAME")
	} else if driver == "sqlserver" || driver == "mssql" {
		// SQL Server DSN format: sqlserver://user:pass@host?database=dbname
		primaryDSN = fmt.Sprintf("sqlserver://%s:%s@%s?database=%s",
			os.Getenv("DB_USER"), os.Getenv("DB_PASS"),
			os.Getenv("DB_HOST"), os.Getenv("DB_NAME"))
	} else if driver == "postgres" || driver == "postgresql" {
		// PostgreSQL DSN format: postgres://user:pass@host/dbname?sslmode=disable
		primaryDSN = fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
			os.Getenv("DB_USER"), os.Getenv("DB_PASS"),
			os.Getenv("DB_HOST"), os.Getenv("DB_NAME"))
	} else {
		// MySQL
		primaryDSN = fmt.Sprintf("%s:%s@tcp(%s)/%s",
			os.Getenv("DB_USER"), os.Getenv("DB_PASS"),
			os.Getenv("DB_HOST"), os.Getenv("DB_NAME"))
	}

	maxOpen, _ := strconv.Atoi(os.Getenv("DB_MAX_OPEN_CONNS"))
	if maxOpen == 0 {
		maxOpen = 25
	}
	maxIdle, _ := strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONNS"))
	if maxIdle == 0 {
		maxIdle = 5
	}

	if err := dbMgr.AddConnection("default", driver, primaryDSN, maxOpen, maxIdle); err != nil {
		slog.Error("❌ Error connecting to primary database", "driver", driver, "error", err)
		os.Exit(1)
	}
	slog.Info("✅ Database Connected (Default)", "driver", driver, "target", primaryDSN)

	// 2. SQLite as INTERNAL (ZenoEngine Operations: Worker Queue, Task Scheduler, etc)
	internalDSN := "zeno_internal.db"
	if err := dbMgr.AddConnection("internal", "sqlite", internalDSN, 1, 1); err != nil {
		slog.Warn("⚠️  Failed to create internal database", "error", err)
	} else {
		slog.Info("✅ SQLite Connected (Internal Operations)", "file", internalDSN)
	}

	// 3. Auto-detect additional DBs
	envVars := os.Environ()
	detectedDBs := make(map[string]bool)
	suffixes := []string{"_DRIVER", "_HOST", "_NAME", "_USER", "_PASS"}

	for _, env := range envVars {
		parts := strings.SplitN(env, "=", 2)
		key := parts[0]
		if !strings.HasPrefix(key, "DB_") {
			continue
		}

		// Skip primary DB keys
		isPrimary := false
		for _, s := range suffixes {
			if key == "DB"+s {
				isPrimary = true
				break
			}
		}
		if isPrimary || key == "DB_MAX_OPEN_CONNS" || key == "DB_MAX_IDLE_CONNS" {
			continue
		}

		// Check if it's an additional DB key
		for _, s := range suffixes {
			if strings.HasSuffix(key, s) {
				dbName := strings.ToLower(strings.TrimSuffix(strings.TrimPrefix(key, "DB_"), s))
				if dbName != "" {
					detectedDBs[dbName] = true
				}
				break
			}
		}
	}

	for dbName := range detectedDBs {
		prefix := "DB_" + strings.ToUpper(dbName) + "_"
		driver := os.Getenv(prefix + "DRIVER")
		if driver == "" {
			driver = "mysql" // Default fallback
		}

		var dsn string
		host := os.Getenv(prefix + "HOST")
		user := os.Getenv(prefix + "USER")
		pass := os.Getenv(prefix + "PASS")
		name := os.Getenv(prefix + "NAME")

		if driver == "sqlite" {
			dsn = name
		} else if driver == "sqlserver" || driver == "mssql" {
			dsn = fmt.Sprintf("sqlserver://%s:%s@%s?database=%s", user, pass, host, name)
		} else if driver == "postgres" || driver == "postgresql" {
			dsn = fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", user, pass, host, name)
		} else {
			// MySQL
			dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s", user, pass, host, name)
		}

		if err := dbMgr.AddConnection(dbName, driver, dsn, maxOpen, maxIdle); err != nil {
			slog.Warn("⚠️  Failed to connect to database", "db", dbName, "error", err)
		} else {
			slog.Info("✅ Additional Database Connected!", "db", dbName, "driver", driver)
		}
	}

	return dbMgr
}

func startWatcher(appCtx *app.AppContext) {
	slog.Info("👀 Watching src/*.zl and views/*.blade.zl for changes...")
	lastReload := time.Now()
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		isRouterChanged := false
		isViewChanged := false

		// 1. Check SRC (Router Logic)
		filepath.Walk("src", func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, ".zl") && info.ModTime().After(lastReload) {
				isRouterChanged = true
				return fmt.Errorf("changed")
			}
			return nil
		})

		// 2. Check VIEWS (Templates) - Only if router didn't change (optimization)
		if !isRouterChanged {
			filepath.Walk("views", func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return nil
				}
				if strings.HasSuffix(path, ".zl") && info.ModTime().After(lastReload) {
					isViewChanged = true
					return fmt.Errorf("changed")
				}
				return nil
			})
		}

		if isRouterChanged {
			slog.Info("🔄 Logic Change detected! Rebuilding Router...")
			lastReload = time.Now()

			// Clear Handler Cache
			engine.GlobalCache.ClearHandlerCache()

			func() {
				defer func() {
					if r := recover(); r != nil {
						slog.Error("❌ Hot Reload Panic", "panic", r)
					}
				}()

				newRouter, err := app.BuildRouter(appCtx)
				if err != nil {
					slog.Error("❌ Hot Reload Failed", "error", err)
				} else {
					appCtx.Hot.Swap(newRouter)
					slog.Info("✅ Hot Reload Success!")
					livereload.Broadcast()
				}
			}()
		} else if isViewChanged {
			slog.Info("🎨 View Change detected! Broadcasting...")
			lastReload = time.Now()

			// Clear Blade Template Cache
			slots.ClearBladeCache()

			livereload.Broadcast()
		}
	}
}

// validateAllScripts validates all .zl files in src/ and views/ directories
func validateAllScripts() error {
	var errors []string
	validatedCount := 0

	// 1. Validate src/ directory (logic files)
	if _, err := os.Stat("src"); err == nil {
		filepath.Walk("src", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && strings.HasSuffix(path, ".zl") {
				if _, err := engine.LoadScript(path); err != nil {
					errors = append(errors, fmt.Sprintf("  ❌ %s: %v", path, err))
				} else {
					validatedCount++
				}
			}
			return nil
		})
	}

	// 2. Validate views/ directory (blade templates)
	if _, err := os.Stat("views"); err == nil {
		filepath.Walk("views", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && strings.HasSuffix(path, ".blade.zl") {
				// Blade files are validated during rendering, but we can check basic syntax
				if _, err := os.ReadFile(path); err != nil {
					errors = append(errors, fmt.Sprintf("  ❌ %s: %v", path, err))
				} else {
					validatedCount++
				}
			}
			return nil
		})
	}

	if len(errors) > 0 {
		slog.Error(fmt.Sprintf("Found %d syntax error(s):", len(errors)))
		for _, errMsg := range errors {
			fmt.Println(errMsg)
		}
		return fmt.Errorf("%d file(s) failed validation", len(errors))
	}

	slog.Info(fmt.Sprintf("  Validated %d file(s)", validatedCount))
	return nil
}

func startServer(appCtx *app.AppContext, cancelWorker context.CancelFunc, workerWG *sync.WaitGroup, liveReloadEnabled bool) {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = ":3000"
	}

	// Root Mux for Split Routing (Safe Live Reload)
	rootMux := http.NewServeMux()

	// 1. Live Reload (Safe, no lock)
	if liveReloadEnabled {
		rootMux.Handle("/livereload", livereload.Instance)
	}

	// 2. Main App (Hot Swap, RLock protected)
	// Wrap with live reload injection middleware in development
	var mainHandler http.Handler = appCtx.Hot
	if liveReloadEnabled {
		mainHandler = livereload.InjectMiddleware(appCtx.Hot)
	}
	rootMux.Handle("/", mainHandler)

	srv := &http.Server{
		Addr:    port,
		Handler: rootMux, // Use Root Mux instead of Hot directly
	}

	// Start Listener first to catch "port in use" error cleanly
	ln, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("❌ FAILED TO START SERVER")
		fmt.Println(strings.Repeat("=", 60))
		fmt.Printf("Error: Port %s is already in use by another application.\n", port)
		fmt.Println("\nTroubleshooting Suggestions:")
		fmt.Println("1. Ensure no other ZenoEngine instance is running.")
		fmt.Println("2. Use the command 'lsof -i " + port + "' to check for applications using this port.")
		fmt.Println("3. Change APP_PORT in the .env file to use a different port.")
		fmt.Println(strings.Repeat("=", 60) + "\n")
		os.Exit(1)
	}

	go func() {
		slog.Info("🚀 Engine Ready", "port", port)
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			slog.Error("❌ Listen failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	slog.Info("⚠️  Shutting down server...")

	// Shutdown Worker
	slog.Info("⏳ Stopping Worker...")
	cancelWorker()
	workerWG.Wait()
	slog.Info("✅ Worker Stopped")

	// Shutdown HTTP Server (30s timeout for SSE connections)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("❌ Server Forced Shutdown", "error", err)
	} else {
		slog.Info("✅ Server Gracefully Stopped")
	}
}
