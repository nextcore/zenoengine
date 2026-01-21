package slots

import (
	"context"
	"fmt"
	"zeno/pkg/engine"
	"zeno/pkg/engine/vm"
)

// RegisterDebugSlots mendaftarkan slot untuk inspeksi dan debugging VM
func RegisterDebugSlots(eng *engine.Engine) {

	// ==========================================
	// SLOT: debug.dump (Inspect Bytecode)
	// ==========================================
	// Melakukan disassembly terhadap bytecode dari sebuah fungsi atau node root saat ini.
	// Contoh:
	// debug.dump: "myFunc"
	eng.Register("debug.dump", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		val := resolveValue(node.Value, scope)

		// 1. Jika value adalah string, cari fungsi di scope
		if name, ok := val.(string); ok && name != "" {
			fn, found := scope.Get(name)
			if !found {
				return fmt.Errorf("debug.dump: function '%s' not found", name)
			}

			if chunk, ok := fn.(*vm.Chunk); ok {
				chunk.Disassemble(name)
				return nil
			}
			return fmt.Errorf("debug.dump: '%s' is not a compiled function", name)
		}

		// 2. Default: Dump root node bytecode jika ada
		// Kita butuh akses ke rootNode yang sedang dieksekusi.
		// Untuk kemudahan, jika tidak ada argumen, beri instruksi cara pakai.
		fmt.Println("   🔍 [DEBUG] Usage: debug.dump: \"functionName\"")
		return nil
	}, engine.SlotMeta{
		Description: "Melakukan disassembly bytecode dari fungsi yang ditentukan.",
		Example:     "debug.dump: \"hitung_gaji\"",
	})
}
