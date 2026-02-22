package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"zeno/internal/slots"
	"zeno/pkg/engine"
)

func HandleTest(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: zeno test <file.zl>")
		os.Exit(1)
	}

	testFile := args[0]
	fmt.Printf("🧪 Running Tests: %s\n", testFile)

	// Init Engine (Minimal/Mock Environment or Full?)
	// For now, let's load basic utils + mock DB? or Real DB if configured?
	// Tests might need DB access. Best to reuse full env if possible,
	// but for "Unit Testing" usually we want isolation.
	// Let's assume using current env.

	eng := engine.NewEngine()

	// Register Standard Slots
	// We need a way to register ALL slots without app machinery?
	// Or we copy the registration logic from Main?
	// Ideally `RegisterAllSlots` should be reusable.
	// But `app` package depends on `app.AppContext`. circular dependency risk if we move RegisterAllSlots.
	// For this task, I'll register `UtilSlots` and `TestSlots` + minimal placeholders if needed.
	// BUT the user might test logic using DB.
	// Let's try to minimal register for now.

	slots.RegisterUtilSlots(eng)
	slots.RegisterMathSlots(eng)

	// Register Test Slots
	stats := &slots.TestStats{}
	slots.RegisterTestSlots(eng) // Wait, RegisterTestSlots needs 'eng' but internally uses context?
	// Oh in my implementation of RegisterTestSlots I updated context usage for stats but RegisterTestSlots takes 'eng'. Correct.

	// Helper for DB/Math etc would be needed.
	// If the user code calls `db.query`, it will fail if not registered.
	// We should probably just register a "Mock" DB slot or require full env?
	// Let's assume pure logic testing first as requested (math.tax).

	// Load Script
	root, err := engine.LoadScript(testFile)
	if err != nil {
		slog.Error("❌ Failed to load test script", "error", err)
		os.Exit(1)
	}

	// Create Context with Stats
	ctx := context.Background()
	ctx = slots.WithTestStats(ctx, stats)
	scope := engine.NewScope(nil)

	// Execute
	if err := eng.Execute(ctx, root, scope); err != nil {
		// Only log if it's a script error, assertion failures are handled in stats usually?
		// But if test failed it might return error.
		// Our 'test' slot returns nil on failure to continue other tests,
		// but if the script syntax is bad, execute returns error.
		slog.Error("❌ Execution Error", "error", err)
	}

	// Report
	fmt.Println("---------------------------------------------------")
	fmt.Printf("Tests: %d | Passed: %d | Failed: %d\n", stats.Total, stats.Passed, stats.Failed)
	if stats.Failed > 0 {
		fmt.Println("Errors:")
		for _, e := range stats.Errors {
			fmt.Printf(" - %s\n", e)
		}
		os.Exit(1)
	} else {
		fmt.Println("✅ All Tests Passed")
	}
}

// Stub for RegisterAllSlots if we wanted fuller support
func registerMocks(eng *engine.Engine) {
	// Mock DB, etc.
}
