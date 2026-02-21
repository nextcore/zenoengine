package cli

import (
	"context"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"zeno/internal/app"
	"zeno/pkg/apidoc"
	"zeno/pkg/dbmanager"
	"zeno/pkg/engine"
	"zeno/pkg/worker"
)

// HandleRouteList displays all registered routes in a table format.
// Usage: zeno route:list
func HandleRouteList(args []string) {
	fmt.Println("🚀 Booting Zeno to discover routes...")

	// 1. Initialize Engine
	dbMgr := dbmanager.NewDBManager()
	eng := engine.NewEngine()
	queue := worker.NewDBQueue(dbMgr, "internal")

	// Register Slots
	app.RegisterAllSlots(eng, nil, dbMgr, queue, nil)

	// 2. Load and Execute src/main.zl
	if _, err := os.Stat("src/main.zl"); os.IsNotExist(err) {
		fmt.Println("❌ src/main.zl not found. Run this command from your project root.")
		os.Exit(1)
	}

	mainScript, err := engine.LoadScript("src/main.zl")
	if err != nil {
		fmt.Printf("❌ Failed to load script: %v\n", err)
		os.Exit(1)
	}

	scope := engine.NewScope(nil)

	if err := eng.Execute(context.Background(), mainScript, scope); err != nil {
		fmt.Printf("⚠️  Script execution warning: %v\n", err)
	}

	// 3. Render Table
	routes := apidoc.Registry.GetRoutes()
	if len(routes) == 0 {
		fmt.Println("\n⚠️  No routes registered (or src/main.zl didn't define any HTTP slots).")
		return
	}

	// Sort routes by Path then Method
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Path == routes[j].Path {
			return routes[i].Method < routes[j].Method
		}
		return routes[i].Path < routes[j].Path
	})

	fmt.Println("\n✅ Registered Routes:")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "METHOD\tURI\tSUMMARY")
	fmt.Fprintln(w, "------\t---\t-------")

	for _, doc := range routes {
		summary := doc.Summary
		if summary == "" {
			summary = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", doc.Method, doc.Path, summary)
	}
	w.Flush()
	fmt.Println()
}
