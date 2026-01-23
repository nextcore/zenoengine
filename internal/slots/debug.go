package slots

import (
	"context"
	"fmt"
	"zeno/pkg/engine"
)

// RegisterDebugSlots mendaftarkan slot untuk inspeksi dan debugging
func RegisterDebugSlots(eng *engine.Engine) {

	// ==========================================
	// SLOT: debug.dump (Inspect AST Node)
	// ==========================================
	// Melakukan dump terhadap AST node dari sebuah fungsi atau variabel.
	// Contoh:
	// debug.dump: "myFunc"
	eng.Register("debug.dump", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		val := resolveValue(node.Value, scope)

		// 1. Jika value adalah string, cari di scope
		if name, ok := val.(string); ok && name != "" {
			fn, found := scope.Get(name)
			if !found {
				return fmt.Errorf("debug.dump: '%s' not found in scope", name)
			}

			// Dump value
			fmt.Printf("   🔍 [DEBUG] %s = %+v (type: %T)\n", name, fn, fn)
			return nil
		}

		// 2. Default: Dump current node
		fmt.Printf("   🔍 [DEBUG] Node: %+v\n", node)
		return nil
	}, engine.SlotMeta{
		Description: "Melakukan dump AST node atau variabel untuk debugging.",
		Example:     "debug.dump: \"myVariable\"",
	})
}
