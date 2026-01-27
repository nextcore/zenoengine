package cli

import (
	"fmt"
	"log/slog"
	"os"
	"zeno/internal/app"
	"zeno/pkg/dbmanager"
	"zeno/pkg/engine"
	"zeno/pkg/logger"
	"zeno/pkg/migrator"
	"zeno/pkg/worker"
)

func HandleMigrate() {
	logger.Setup("development")
	dbMgr := dbmanager.NewDBManager()
	dbDriver := os.Getenv("DB_DRIVER")
	if dbDriver == "" {
		dbDriver = "mysql"
	}

	var dsn string
	if dbDriver == "sqlite" {
		dsn = os.Getenv("DB_NAME")
	} else {
		dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&multiStatements=true",
			os.Getenv("DB_USER"), os.Getenv("DB_PASS"),
			os.Getenv("DB_HOST"), os.Getenv("DB_NAME"))
	}

	if err := dbMgr.AddConnection("default", dbDriver, dsn, 10, 5); err != nil {
		slog.Error("❌ Fatal: DB Connection Failed", "error", err)
		os.Exit(1)
	}

	eng := engine.NewEngine()
	// Init DBQueue for migration logic
	queue := worker.NewDBQueue(dbMgr, "default")
	app.RegisterAllSlots(eng, nil, dbMgr, queue, nil)

	migrationDir := "migrations"
	if _, err := os.Stat(migrationDir); os.IsNotExist(err) {
		slog.Info("✨ No migration directory found. Skipping root migrations.")
		os.Exit(0)
	}

	mig := migrator.New(eng, dbMgr, migrationDir)
	if err := mig.Run(); err != nil {
		slog.Error("❌ Migration Failed", "error", err)
		os.Exit(1)
	}
	os.Exit(0)
}
