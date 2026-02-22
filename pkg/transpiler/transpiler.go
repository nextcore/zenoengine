package transpiler

import (
	"fmt"
	"strings"
	"zeno/pkg/engine"
)

type Transpiler struct{}

func New() *Transpiler {
	return &Transpiler{}
}

// Transpile converts a Zeno AST into a Go source file string.
// This prototype generates code that reconstructs the AST at runtime
// and executes it using the existing engine, bypassing the parsing step.
func (t *Transpiler) Transpile(root *engine.Node) (string, error) {
	var sb strings.Builder

	// Header
	sb.WriteString("package main\n\n")
	sb.WriteString("import (\n")
	sb.WriteString("\t\"context\"\n")
	sb.WriteString("\t\"log\"\n")
	sb.WriteString("\t\"zeno/pkg/engine\"\n")
	sb.WriteString("\t\"zeno/internal/app\"\n")
	sb.WriteString("\t\"zeno/pkg/dbmanager\"\n")
	sb.WriteString("\t\"zeno/pkg/worker\"\n")
	sb.WriteString(")\n\n")

	// Main
	sb.WriteString("func main() {\n")
	sb.WriteString("\t// 1. Setup Engine\n")
	sb.WriteString("\teng := engine.NewEngine()\n")
	sb.WriteString("\tdbMgr := dbmanager.NewDBManager()\n")
	sb.WriteString("\tqueue := worker.NewDBQueue(dbMgr, \"internal\")\n")
	sb.WriteString("\tapp.RegisterAllSlots(eng, nil, dbMgr, queue, nil)\n\n")

	sb.WriteString("\t// 2. Construct AST Programmatically\n")
	sb.WriteString("\troot := ")

	nodeCode := t.nodeToGoCode(root)
	sb.WriteString(nodeCode)
	sb.WriteString("\n\n")

	sb.WriteString("\t// 3. Execute\n")
	sb.WriteString("\tscope := engine.NewScope(nil)\n")
	sb.WriteString("\tif err := eng.Execute(context.Background(), root, scope); err != nil {\n")
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
		// Assume string for simplicity in prototype
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
