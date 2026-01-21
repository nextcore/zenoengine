package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"zeno/internal/app"
	"zeno/pkg/dbmanager"
	"zeno/pkg/engine"
	"zeno/pkg/engine/vm"
	"zeno/pkg/logger"
	"zeno/pkg/worker"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func HandleRun(args []string) {
	// 1. Load .env FIRST to check configuration
	godotenv.Load()

	// 2. Check VM Toggle (Env Priority, then Flag Override)
	useVM := os.Getenv("ZENO_VM_ENABLED") == "true"
	var scriptArgs []string

	// Simple Flag Parsing
	for _, arg := range args {
		if arg == "--vm" {
			useVM = true
		} else {
			scriptArgs = append(scriptArgs, arg)
		}
	}

	if len(scriptArgs) < 1 {
		fmt.Println("Usage: zeno run <path/to/script.zl> [--vm]")
		os.Exit(1)
	}

	// 3. Setup Logger
	logger.Setup("development")

	path := scriptArgs[0]
	var root *engine.Node
	var chunk *vm.Chunk
	var err error

	// A. Check for Bytecode File (.zbc) -> FAST PATH
	if strings.HasSuffix(path, ".zbc") {
		useVM = true // Force VM mode for bytecode
		chunk, err = vm.LoadFromFile(path)
		if err != nil {
			fmt.Printf("❌ Load Bytecode Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		// B. Standard Source File (.zl) -> Parse AST
		root, err = engine.LoadScript(path)
		if err != nil {
			fmt.Printf("❌ Syntax Error: %v\n", err)
			os.Exit(1)
		}
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

	if useVM {
		fmt.Println("🚀 Running in VM Mode (Experimental)")
		// If chunk not loaded from file (meaning we came from .zl), compile now
		if chunk == nil {
			compiler := vm.NewCompiler()
			var err error
			chunk, err = compiler.Compile(root)
			if err != nil {
				fmt.Printf("❌ Compilation Error: %v\n", err)
				os.Exit(1)
			}
		}

		virtualMachine := vm.NewVM()

		// Map Engine Slots to Context
		ctx := context.WithValue(context.Background(), "engine", eng)

		// Run VM
		if err := virtualMachine.Run(ctx, chunk, engine.NewScope(nil)); err != nil {
			fmt.Printf("❌ Runtime Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := eng.Execute(context.Background(), root, engine.NewScope(nil)); err != nil {
			fmt.Printf("❌ Execution Error: %v\n", err)
			os.Exit(1)
		}
	}
	os.Exit(0)
}
