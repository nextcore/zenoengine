package engine

import (
	"context"
	"fmt"
	"sort"
	"strings" // [BARU] Import strings untuk manipulasi nama variabel
)

type HandlerFunc func(ctx context.Context, node *Node, scope *Scope) error

type InputMeta struct {
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Type        string `json:"type,omitempty"` // e.g. "string", "int", "bool"
}

// Struct untuk menyimpan Dokumentasi Slot
type SlotMeta struct {
	Description    string               `json:"description"`
	Example        string               `json:"example"` // Snippet kode .zl
	Inputs         map[string]InputMeta `json:"inputs,omitempty"`
	RequiredBlocks []string             `json:"required_blocks,omitempty"` // e.g. ["do"], ["then", "else"]
}

type Engine struct {
	Registry map[string]HandlerFunc
	Docs     map[string]SlotMeta // <--- Database Dokumentasi
}

func NewEngine() *Engine {
	return &Engine{
		Registry: make(map[string]HandlerFunc),
		Docs:     make(map[string]SlotMeta),
	}
}

// Update Register agar menerima Metadata
func (e *Engine) Register(name string, fn HandlerFunc, meta SlotMeta) {
	e.Registry[name] = fn
	e.Docs[name] = meta
}

func (e *Engine) Execute(ctx context.Context, node *Node, scope *Scope) error {
	// FASTEST PATH: Try optimized fast paths for common operations (2-3x faster)
	if used, err := TryFastPath(ctx, node, scope); used {
		if err != nil {
			return fmt.Errorf("[%s:%d:%d] execution error in '%s': %v",
				node.Filename, node.Line, node.Col, node.Name, err)
		}
		return nil
	}

	// FAST PATH: Use cached handler if available (7.2x faster than map lookup)
	if node.cachedHandler != nil {
		err := node.cachedHandler(ctx, node, scope)
		if err != nil {
			return fmt.Errorf("[%s:%d:%d] execution error in '%s': %v",
				node.Filename, node.Line, node.Col, node.Name, err)
		}
		return nil
	}

	// SLOW PATH: Lookup handler and cache for next time
	if handler, exists := e.Registry[node.Name]; exists {
		// Cache handler for future executions
		node.cachedHandler = handler
		if meta, hasMeta := e.Docs[node.Name]; hasMeta {
			metaCopy := meta
			node.cachedMeta = &metaCopy
		}

		// --- VALIDASI ATRIBUT (Strict Mode) ---
		if node.cachedMeta != nil {
			// 1. Cek Atribut Tak Dikenal (Hanya jika Inputs didefinisikan)
			if node.cachedMeta.Inputs != nil {
				for _, child := range node.Children {
					if child.Name == "do" || child.Name == "then" || child.Name == "else" || child.Name == "catch" || child.Name == "" {
						continue // Blok spesial diabaikan dari validasi atribut
					}
					if _, allowed := node.cachedMeta.Inputs[child.Name]; !allowed {
						allowedKeys := make([]string, 0, len(node.cachedMeta.Inputs))
						for k := range node.cachedMeta.Inputs {
							allowedKeys = append(allowedKeys, k)
						}
						sort.Strings(allowedKeys)
						return fmt.Errorf("[%s:%d:%d] validation error: unknown attribute '%s' for slot '%s'. Allowed attributes: %s",
							child.Filename, child.Line, child.Col, child.Name, node.Name, strings.Join(allowedKeys, ", "))
					}
				}
			}

			// 2. Cek Atribut Wajib
			for name, input := range node.cachedMeta.Inputs {
				if input.Required {
					found := false
					for _, child := range node.Children {
						if child.Name == name {
							found = true
							break
						}
					}
					if !found {
						return fmt.Errorf("[%s:%d:%d] validation error: missing required attribute '%s' for slot '%s'",
							node.Filename, node.Line, node.Col, name, node.Name)
					}
				}
			}

			// 3. Cek Blok Wajib (RequiredBlocks)
			for _, blockName := range node.cachedMeta.RequiredBlocks {
				found := false
				for _, child := range node.Children {
					if child.Name == blockName {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("[%s:%d:%d] validation error: missing required block '%s:' for slot '%s'",
						node.Filename, node.Line, node.Col, blockName, node.Name)
				}
			}
		}

		err := handler(ctx, node, scope)
		if err != nil {
			// Jika error sudah punya info baris, biarkan. Jika belum, tambahkan.
			return fmt.Errorf("[%s:%d:%d] execution error in '%s': %v", node.Filename, node.Line, node.Col, node.Name, err)
		}
		return nil
	}

	// 2. [BARU] Cek Variable Shorthand ($var: value)
	// Fitur ini memungkinkan penulisan: $nama: "Budi" atau $user: { name: "Budi" }
	if len(node.Name) > 1 && strings.HasPrefix(node.Name, "$") {
		varName := strings.TrimPrefix(node.Name, "$")

		// Gunakan helper internal untuk resolve value
		val := e.resolveShorthandValue(node, scope)

		scope.Set(varName, val)
		return nil
	}

	// 3. Jika slot tidak ditemukan dan bukan variabel, coba jalankan anak-anaknya (Logic flow)
	// Ini berguna untuk block tanpa nama atau struktur tree murni
	for _, child := range node.Children {
		if err := e.Execute(ctx, child, scope); err != nil {
			return err
		}
	}
	return nil
}

// [BARU] Helper internal untuk memproses nilai pada Variable Shorthand
func (e *Engine) resolveShorthandValue(n *Node, scope *Scope) interface{} {
	// A. Jika punya children, anggap sebagai Map/Object
	if len(n.Children) > 0 {
		m := make(map[string]interface{})
		for _, c := range n.Children {
			m[c.Name] = e.resolveShorthandValue(c, scope)
		}
		return m
	}

	// B. Ambil nilai mentah
	valStr := fmt.Sprintf("%v", n.Value)

	// C. Cek String Literal (Kutip ganda/tunggal) -> "Isi" menjadi Isi
	if len(valStr) >= 2 {
		if (strings.HasPrefix(valStr, "\"") && strings.HasSuffix(valStr, "\"")) ||
			(strings.HasPrefix(valStr, "'") && strings.HasSuffix(valStr, "'")) {
			return valStr[1 : len(valStr)-1]
		}
	}

	// D. Cek Referensi Variabel Lain ($other)
	if strings.HasPrefix(valStr, "$") {
		key := strings.TrimPrefix(valStr, "$")
		// Resolusi variabel sederhana dari scope
		if v, ok := scope.Get(key); ok {
			return v
		}
	}

	// E. Fallback (Return raw value: int, bool, dll)
	return n.Value
}

// Helper untuk mengambil semua docs (Sorted by Name)
func (e *Engine) GetDocumentation() map[string]SlotMeta {
	return e.Docs
}

// Helper untuk mendapatkan list nama slot yang terurut
func (e *Engine) GetSortedSlotNames() []string {
	keys := make([]string, 0, len(e.Docs))
	for k := range e.Docs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
