package transpiler

import (
	"fmt"
	"strings"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

type Transpiler struct{}

func New() *Transpiler {
	return &Transpiler{}
}

// Transpile converts a Zeno AST into a Go source file string.
// Hybrid AOT: Generates native Chi router registration for top-level routes,
// but executes handler bodies via the Engine AST interpreter.
func (t *Transpiler) Transpile(root *engine.Node) (string, error) {
	var sb strings.Builder

	// Header
	sb.WriteString("package main\n\n")
	sb.WriteString("import (\n")
	sb.WriteString("\t\"context\"\n")
	sb.WriteString("\t\"log\"\n")
	sb.WriteString("\t\"net/http\"\n")
	sb.WriteString("\t\"zeno/pkg/engine\"\n")
	sb.WriteString("\t\"zeno/internal/app\"\n")
	sb.WriteString("\t\"zeno/pkg/dbmanager\"\n")
	sb.WriteString("\t\"zeno/pkg/worker\"\n")
	sb.WriteString("\t\"github.com/go-chi/chi/v5\"\n")
	sb.WriteString(")\n\n")

	// Main
	sb.WriteString("func main() {\n")
	sb.WriteString("\t// 1. Setup Engine\n")
	sb.WriteString("\teng := engine.NewEngine()\n")
	sb.WriteString("\tdbMgr := dbmanager.NewDBManager()\n")
	sb.WriteString("\tqueue := worker.NewDBQueue(dbMgr, \"internal\")\n")

	// Create Router
	sb.WriteString("\tr := chi.NewRouter()\n")

	sb.WriteString("\tapp.RegisterAllSlots(eng, r, dbMgr, queue, nil)\n\n")

	sb.WriteString("\t// 2. Hybrid AOT Execution\n")
	sb.WriteString("\tctx := context.Background()\n")
	sb.WriteString("\tscope := engine.NewScope(nil)\n\n")

	// Iterate root children
	// Separate Routes from Init Logic
	for _, child := range root.Children {
		// Detect Route: http.get, http.post, etc.
		method := ""
		if strings.HasPrefix(child.Name, "http.") {
			suffix := strings.TrimPrefix(child.Name, "http.")
			switch suffix {
			case "get", "post", "put", "delete", "patch", "head", "options":
				method = strings.ToUpper(suffix)
			}
		}

		if method != "" {
			// Generate Static Router Call
			// Logic: r.Get("/path", func(...) { ... })

			// Extract Path
			path := ""
			// Path is usually in Value (http.get: "/path")
			if child.Value != nil {
				path = coerce.ToString(child.Value)
			}
			// Fallback to children parsing (simplified for AOT)
			// For prototype, assume Value holds the path

			// Find 'do' block or use all children as body
			var bodyNode *engine.Node
			for _, c := range child.Children {
				if c.Name == "do" {
					bodyNode = c
					break
				}
			}

			// If implicit do, wrap other children
			if bodyNode == nil {
				bodyNode = &engine.Node{Name: "do"}
				for _, c := range child.Children {
					// Filter metadata
					if c.Name != "summary" && c.Name != "desc" && c.Name != "middleware" {
						bodyNode.Children = append(bodyNode.Children, c)
					}
				}
			}

			// Generate Handler Code
			sb.WriteString(fmt.Sprintf("\tr.MethodFunc(\"%s\", \"%s\", func(w http.ResponseWriter, r *http.Request) {\n", method, path))
			sb.WriteString("\t\t// Runtime AST Construction\n")

			// Construct AST for BODY only
			bodyCode := t.nodeToGoCode(bodyNode)
			sb.WriteString(fmt.Sprintf("\t\tbody := %s\n", bodyCode))

			// Execution
			// We need to inject httpWriter/Request into context, similar to router.go handler
			sb.WriteString("\t\thandlerCtx := context.WithValue(r.Context(), \"httpWriter\", w)\n")
			sb.WriteString("\t\thandlerCtx = context.WithValue(handlerCtx, \"httpRequest\", r)\n")
			sb.WriteString("\t\thandlerScope := engine.NewScope(scope)\n") // Child scope
			sb.WriteString("\t\tif err := eng.Execute(handlerCtx, body, handlerScope); err != nil {\n")
			sb.WriteString("\t\t\tlog.Printf(\"Handler Error: %v\", err)\n")
			sb.WriteString("\t\t\thttp.Error(w, \"Internal Server Error\", 500)\n")
			sb.WriteString("\t\t}\n")
			sb.WriteString("\t})\n\n")

		} else {
			// Non-Route (Init Logic)
			// Execute immediately in main
			sb.WriteString(fmt.Sprintf("\t// Init: %s\n", child.Name))
			nodeCode := t.nodeToGoCode(child)
			sb.WriteString(fmt.Sprintf("\tif err := eng.Execute(ctx, %s, scope); err != nil {\n", nodeCode))
			sb.WriteString("\t\tlog.Fatal(err)\n")
			sb.WriteString("\t}\n")
		}
	}

	// Start Server (if not explicitly started by init logic, we usually start it here)
	// But Zeno usually has `http.server` slot.
	// If `http.server` is in the script, it will be executed as "Non-Route" above.
	// But `http.server` in Zeno currently starts the server using the router passed to RegisterAllSlots.
	// In AOT, we passed `r` (our new router).
	// So `http.server` slot should work fine!

	// Wait, `http.server` slot in `RegisterHTTPServerSlots` usually takes `r` from context or global?
	// `internal/slots/http.go` (Server) logic:
	// It uses `package server` (likely `pkg/host` or `server.go`).
	// We need to ensure `http.server` uses OUR `r`.

	// If the script does NOT contain `http.server`, the binary will exit immediately.
	// We should probably block main if server started?
	// Or assume the script contains `http.server`.

	// [AOT FIX] Start Server automatically at the end
	sb.WriteString("\n\t// 4. Start Server (Default Port :3000)\n")
	sb.WriteString("\tlog.Println(\"🚀 AOT Server starting on :3000\")\n")
	sb.WriteString("\tif err := http.ListenAndServe(\":3000\", r); err != nil {\n")
	sb.WriteString("\t\tlog.Fatal(err)\n")
	sb.WriteString("\t}\n")

	sb.WriteString("}\n")

	return sb.String(), nil
}

func (t *Transpiler) nodeToGoCode(n *engine.Node) string {
	if n == nil {
		return "nil"
	}

	var sb strings.Builder
	sb.WriteString("&engine.Node{\n")
	sb.WriteString(fmt.Sprintf("\t\tName: \"%s\",\n", n.Name))

	// Value handling
	if n.Value != nil {
		// Escape quotes
		valStr := fmt.Sprintf("%v", n.Value)
		valStr = strings.ReplaceAll(valStr, "\"", "\\\"")
		sb.WriteString(fmt.Sprintf("\t\tValue: \"%s\",\n", valStr))
	}

	// Children
	if len(n.Children) > 0 {
		sb.WriteString("\t\tChildren: []*engine.Node{\n")
		for _, child := range n.Children {
			sb.WriteString(t.nodeToGoCode(child))
			sb.WriteString(",\n")
		}
		sb.WriteString("\t\t},\n")
	}

	sb.WriteString("\t}")
	return sb.String()
}
