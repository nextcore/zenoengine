package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"zeno/internal/app"
	"zeno/internal/slots"
	"zeno/pkg/dbmanager"
	"zeno/pkg/engine"
	"zeno/pkg/worker"
)

// HandleTest executes .test.zl files in the tests/ directory.
// Usage: zeno test [path/to/file]
func HandleTest(args []string) {
	fmt.Println("🧪 Starting Zeno Test Runner...")
	start := time.Now()

	// 1. Find Test Files
	var testFiles []string

	target := "tests"
	if len(args) > 0 {
		target = args[0]
	}

	info, err := os.Stat(target)
	if os.IsNotExist(err) {
		fmt.Printf("❌ Target '%s' not found.\n", target)
		os.Exit(1)
	}

	if !info.IsDir() {
		testFiles = append(testFiles, target)
	} else {
		filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() && strings.HasSuffix(path, ".test.zl") {
				testFiles = append(testFiles, path)
			}
			return nil
		})
	}

	if len(testFiles) == 0 {
		fmt.Println("⚠️  No test files found (looking for *.test.zl).")
		return
	}

	fmt.Printf("🔍 Found %d test file(s)\n\n", len(testFiles))

	// 2. Execute Tests
	passed := 0
	failed := 0

	for _, file := range testFiles {
		// Run each test file in isolation
		err := runSingleTestFile(file)
		if err == nil {
			fmt.Printf("✅ PASS: %s\n", file)
			passed++
		} else {
			fmt.Printf("❌ FAIL: %s\n", file)
			fmt.Printf("   Error: %v\n", err)
			failed++
		}
	}

	// 3. Summary
	duration := time.Since(start)
	fmt.Println("\n" + strings.Repeat("-", 40))
	if failed == 0 {
		fmt.Printf("🎉 All tests passed! (%s)\n", duration)
	} else {
		fmt.Printf("💥 %d passed, %d failed. (%s)\n", passed, failed, duration)
		os.Exit(1)
	}
}

func runSingleTestFile(path string) error {
	// Initialize Isolated Engine
	eng := engine.NewEngine()

	// Create DB Manager (In-Memory for speed/safety default)
	dbMgr := dbmanager.NewDBManager()
	dbMgr.AddConnection("default", "sqlite", ":memory:", 1, 1)

	queue := worker.NewDBQueue(dbMgr, "internal")

	// Register Standard Slots
	app.RegisterAllSlots(eng, nil, dbMgr, queue, nil)

	// Register Test Slots (Overrides or Additions)
	slots.RegisterTestSlots(eng)

	// Load Script
	node, err := engine.LoadScript(path)
	if err != nil {
		return err
	}

	// Execute
	scope := engine.NewScope(nil)

	// Create Stats for this file
	stats := &slots.TestStats{}
	ctx := slots.WithTestStats(context.Background(), stats)

	err = eng.Execute(ctx, node, scope)

	// Check results
	if err != nil {
		return err // Script error (syntax/runtime)
	}

	if stats.Failed > 0 {
		return fmt.Errorf("%d/%d assertions failed", stats.Failed, stats.Total)
	}

	if stats.Total == 0 {
		// Only warn if no tests found? Or pass?
		// Usually pass.
	}

	return nil
}
