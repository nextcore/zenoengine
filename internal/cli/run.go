package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"zeno/internal/app"
	"zeno/internal/slots"
	"zeno/pkg/dbmanager"
	"zeno/pkg/engine"
	"zeno/pkg/logger"
	"zeno/pkg/worker"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func HandleRun(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: zeno run <path/to/script.zl>")
		os.Exit(1)
	}
	godotenv.Load()
	logger.Setup("development")
	path := args[0]
	root, err := engine.LoadScript(path)
	if err != nil {
		fmt.Printf("❌ Syntax Error: %v\n", err)
		os.Exit(1)
	}

	dbMgr := dbmanager.NewDBManager()
	eng := engine.NewEngine()

	// Setup DB Connection
	dbDriver := os.Getenv("DB_DRIVER")
	if dbDriver == "" {
		dbDriver = "mysql"
	}

	var dsn string
	if dbDriver == "sqlite" {
		dsn = os.Getenv("DB_NAME")
	} else {
		dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true",
			os.Getenv("DB_USER"), os.Getenv("DB_PASS"),
			os.Getenv("DB_HOST"), os.Getenv("DB_NAME"))
	}

	if err := dbMgr.AddConnection("default", dbDriver, dsn, 10, 5); err != nil {
		fmt.Printf("❌ Fatal: DB Connection Failed: %v\n", err)
		os.Exit(1)
	}

	// Auto-detect additional DBs
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
				dbNamePart := strings.ToLower(strings.TrimSuffix(strings.TrimPrefix(key, "DB_"), s))
				if dbNamePart != "" {
					detectedDBs[dbNamePart] = true
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

		if err := dbMgr.AddConnection(dbName, driver, dsn, 10, 5); err != nil {
			fmt.Printf("⚠️  Failed to connect to database %s: %v\n", dbName, err)
		} else {
			fmt.Printf("✅ Additional Database Connected! db=%s\n", dbName)
		}
	}

	// Use the newly created helper registry
	queue := worker.NewDBQueue(dbMgr, "default")
	r := chi.NewRouter()
	app.RegisterAllSlots(eng, r, dbMgr, queue, nil)

	// Pass remaining CLI arguments through context and scope
	cliArgs := args[1:]
	ifaceArgs := make([]interface{}, len(cliArgs))
	for i, v := range cliArgs {
		ifaceArgs[i] = v
	}

	ctx := context.WithValue(context.Background(), "zenoArgs", cliArgs)
	scope := engine.NewScope(nil)
	scope.Set("args", ifaceArgs)

	if err := eng.Execute(ctx, root, scope); err != nil {
		// Ignore control flow signals at the root level (return, break, continue)
		// These are used by slots like 'return' to halt execution gracefully.
		if errors.Is(err, slots.ErrReturn) || errors.Is(err, slots.ErrBreak) || errors.Is(err, slots.ErrContinue) {
			os.Exit(0)
		}
		fmt.Printf("❌ Execution Error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
