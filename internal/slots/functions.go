package slots

import (
	"context"
	"fmt"
	"sync"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

// RegisterFunctionSlots mendaftarkan slot 'fn' dan 'call'
// Mekanisme: 'fn' menyimpan *Node ke GLOBAL REGISTRY, 'call' mengeksekusinya.
func RegisterFunctionSlots(eng *engine.Engine) {

	// Global Function Registry (thread-safe)
	var (
		functionRegistry   = make(map[string]*engine.Node)
		functionRegistryMu sync.RWMutex
	)

	// ==========================================
	// 1. SLOT: FN (Define Function)
	// ==========================================
	// Menyimpan node children sebagai "Function Body" di dalam variabel scope.
	// Tidak dieksekusi saat definisi.
	// Contoh:
	// fn: my_func {
	//    log: "Hello"
	// }
	eng.Register("fn", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		// 1. Ambil Nama Fungsi
		funcName := coerce.ToString(resolveValue(node.Value, scope))
		if funcName == "" {
			return fmt.Errorf("fn: function name is required")
		}

		// 2. Simpan Node ini (atau children-nya) ke Scope
		// Kita simpan pointer ke Node itu sendiri agar bisa di-execute nanti.
		// Prefix "_fn_" bisa digunakan untuk namespace internal jika perlu,
		// tapi untuk fleksibilitas meta-programming, simpan sebagai variable biasa pun oke.
		// Namun agar aman dari overwrite variabel data, kita pakai prefix/suffix atau tipe data khusus.
		// Disini kita simpan raw *engine.Node

		functionRegistryMu.Lock()
		functionRegistry[funcName] = node
		functionRegistryMu.Unlock()
		fmt.Printf("[DEBUG FN] Registered function '%s' in scope\n", funcName)

		return nil
	}, engine.SlotMeta{
		Description: "Mendefinisikan fungsi (menyimpan blok kode untuk dipanggil nanti).",
		Example:     "fn: hitung_gaji {\n  ...\n}",
	})

	// ==========================================
	// 2. SLOT: CALL (Invoke Function)
	// ==========================================
	// Memanggil node yang disimpan oleh 'fn' dari GLOBAL REGISTRY.
	// Contoh: call: my_func
	eng.Register("call", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		funcName := coerce.ToString(resolveValue(node.Value, scope))
		if funcName == "" {
			return fmt.Errorf("call: function name is required")
		}

		// 1. Retrieve Function Node from GLOBAL REGISTRY
		functionRegistryMu.RLock()
		funcNode, found := functionRegistry[funcName]
		functionRegistryMu.RUnlock()

		fmt.Printf("[DEBUG CALL] Looking for function '%s', found=%v (from GLOBAL registry)\n", funcName, found)
		if !found {
			return fmt.Errorf("call: function '%s' not found", funcName)
		}

		// 2. Execute Children of the Function Node
		// Kita eksekusi children-nya seolah-olah mereka ada di sini.
		// Scope: Apakah New Scope atau Current Scope?
		// ZenoLang (HyperLambda style) biasanya mewarisi scope (Dynamic Scope).
		// Jadi variabel $gaji yang ada di scope pemanggil bisa diakses di dalam fungsi.

		for _, child := range funcNode.Children {
			if err := eng.Execute(ctx, child, scope); err != nil {
				return err
			}
		}

		return nil
	}, engine.SlotMeta{
		Description: "Memanggil fungsi yang didefinisikan dengan fn.",
		Example:     "call: hitung_gaji",
	})

	// ==========================================
	// 3. SLOT: RETURN (Early Exit)
	// ==========================================
	// Note: Implementasi return membutuhkan support di engine.Execute loop
	// untuk menangkap error khusus (seperti break/continue).
	// Jika engine belum support handle error "Return", ini hanya akan stop execution blok saat ini.
	// Untuk saat ini kita skip dulu atau implementasi basic error.
}
