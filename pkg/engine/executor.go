package engine

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sort"
	"strconv"
	"strings" // [BARU] Import strings untuk manipulasi nama variabel

	"github.com/shopspring/decimal"
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
	ValueType      string               `json:"value_type,omitempty"`      // [BARU] Tipe data untuk value utama slot
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

// Execute executes a node with comprehensive panic recovery to ensure runtime immortality.
// Any panic from user scripts will be caught, logged, and converted to an error.
func (e *Engine) Execute(ctx context.Context, node *Node, scope *Scope) (err error) {
	// ==========================================
	// PANIC RECOVERY - IMMORTAL RUNTIME
	// ==========================================
	// This defer ensures that ANY panic from user scripts is caught and converted to an error.
	// The runtime will NEVER crash, even with nil pointer dereferences, division by zero, etc.
	defer func() {
		if r := recover(); r != nil {
			// Capture full stack trace for debugging
			stack := string(debug.Stack())

			// Log the panic with comprehensive context
			slog.Error("🔥 PANIC RECOVERED IN EXECUTOR",
				"panic", r,
				"slot", node.Name,
				"file", node.Filename,
				"line", node.Line,
				"col", node.Col,
				"stack", stack,
			)

			// Convert panic to error for graceful degradation
			// This allows the HTTP middleware to return a proper 500 error
			err = Diagnostic{
				Type:     "panic",
				Message:  fmt.Sprintf("PANIC: %v\n\nStack Trace:\n%s", r, stack),
				Filename: node.Filename,
				Line:     node.Line,
				Col:      node.Col,
				Slot:     node.Name,
			}
		}
	}()

	// Inject engine into context for slots use
	ctx = context.WithValue(ctx, "engine", e)

	// FASTEST PATH: Try optimized fast paths for common operations (2-3x faster)
	if used, err := TryFastPath(ctx, node, scope); used {
		if err != nil {
			return Diagnostic{
				Type:     "error",
				Message:  err.Error(),
				Filename: node.Filename,
				Line:     node.Line,
				Col:      node.Col,
				Slot:     node.Name,
			}
		}
		return nil
	}

	// FAST PATH: Use cached handler if available (7.2x faster than map lookup)
	if node.cachedHandler != nil {
		err := node.cachedHandler(ctx, node, scope)
		if err != nil {
			return Diagnostic{
				Type:     "error",
				Message:  err.Error(),
				Filename: node.Filename,
				Line:     node.Line,
				Col:      node.Col,
				Slot:     node.Name,
			}
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
					if child.Name == "do" || child.Name == "then" || child.Name == "else" || child.Name == "catch" || child.Name == "" || child.Name == "__native_write" || child.Name == "__native_write_safe" {
						continue // Blok spesial diabaikan dari validasi atribut
					}
					if _, allowed := node.cachedMeta.Inputs[child.Name]; !allowed {
						allowedKeys := make([]string, 0, len(node.cachedMeta.Inputs))
						for k := range node.cachedMeta.Inputs {
							allowedKeys = append(allowedKeys, k)
						}
						sort.Strings(allowedKeys)
						return Diagnostic{
							Type:     "error",
							Message:  fmt.Sprintf("validation error: unknown attribute '%s'. Allowed attributes: %s", child.Name, strings.Join(allowedKeys, ", ")),
							Filename: child.Filename,
							Line:     child.Line,
							Col:      child.Col,
							Slot:     node.Name,
						}
					}
				}
			}

			// 2. Cek Atribut Wajib
			for name, input := range node.cachedMeta.Inputs {
				if input.Required {
					found := false
					var attrNode *Node
					for _, child := range node.Children {
						if child.Name == name {
							found = true
							attrNode = child
							break
						}
					}
					if !found {
						return Diagnostic{
							Type:     "error",
							Message:  fmt.Sprintf("validation error: missing required attribute '%s'", name),
							Filename: node.Filename,
							Line:     node.Line,
							Col:      node.Col,
							Slot:     node.Name,
						}
					}

					// --- [BARU] VALIDASI TIPE DATA (Strict Mode) ---
					if input.Type != "" && input.Type != "any" && attrNode != nil {
						val := e.ResolveShorthandValue(attrNode, scope)
						if err := e.ValidateValueType(val, input.Type, attrNode, node.Name); err != nil {
							return err
						}
					}
				} else {
					// Cek tipe data untuk atribut opsional jika diset
					for _, child := range node.Children {
						if child.Name == name && input.Type != "" && input.Type != "any" {
							val := e.ResolveShorthandValue(child, scope)
							if err := e.ValidateValueType(val, input.Type, child, node.Name); err != nil {
								return err
							}
						}
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
					return Diagnostic{
						Type:     "error",
						Message:  fmt.Sprintf("validation error: missing required block '%s:'", blockName),
						Filename: node.Filename,
						Line:     node.Line,
						Col:      node.Col,
						Slot:     node.Name,
					}
				}
			}
		}

		err := handler(ctx, node, scope)
		if err != nil {
			return Diagnostic{
				Type:     "error",
				Message:  err.Error(),
				Filename: node.Filename,
				Line:     node.Line,
				Col:      node.Col,
				Slot:     node.Name,
			}
		}
		return nil
	}

	// 2. [BARU] Cek Variable Shorthand ($var: value)
	// Fitur ini memungkinkan penulisan: $nama: "Budi" atau $user: { name: "Budi" }
	if len(node.Name) > 1 && strings.HasPrefix(node.Name, "$") {
		varName := strings.TrimPrefix(node.Name, "$")

		// Gunakan helper internal untuk resolve value
		val := e.ResolveShorthandValue(node, scope)

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

// ResolveShorthandValue Helper internal untuk memproses nilai pada Variable Shorthand (Exported for analysis)
func (e *Engine) ResolveShorthandValue(n *Node, scope *Scope) interface{} {
	// A. Jika punya children, anggap sebagai Map/Object
	if len(n.Children) > 0 {
		m := make(map[string]interface{})
		for _, c := range n.Children {
			m[c.Name] = e.ResolveShorthandValue(c, scope)
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
	if strings.HasPrefix(valStr, "$") && scope != nil {
		key := strings.TrimPrefix(valStr, "$")
		// Resolusi variabel sederhana dari scope
		if v, ok := scope.Get(key); ok {
			return v
		}
	}

	// E. Fallback (Return raw value: int, bool, dll)
	return n.Value
}

// ValidateValueType Helper internal untuk memvalidasi tipe data (Exported for slots)
func (e *Engine) ValidateValueType(val interface{}, expectedType string, node *Node, slotName string) error {
	if val == nil {
		return nil // Nil is generally allowed unless 'required' check fails (which is separate)
	}

	actualType := fmt.Sprintf("%T", val)
	isValid := false

	switch expectedType {
	case "string":
		_, ok := val.(string)
		isValid = ok
	case "int", "integer":
		switch v := val.(type) {
		case int, int32, int64:
			isValid = true
		case float64:
			isValid = v == float64(int(v))
		case string:
			// [ZENO FRIENDLY] Allow numeric strings for literals
			if _, err := strconv.Atoi(v); err == nil {
				isValid = true
			}
		}
	case "bool", "boolean":
		switch v := val.(type) {
		case bool:
			isValid = true
		case string:
			lower := strings.ToLower(v)
			isValid = (lower == "true" || lower == "false" || lower == "1" || lower == "0")
		}
	case "float", "number":
		switch v := val.(type) {
		case float32, float64, int, int32, int64:
			isValid = true
		case string:
			if _, err := strconv.ParseFloat(v, 64); err == nil {
				isValid = true
			}
		}
	case "decimal":
		switch v := val.(type) {
		case float32, float64, int, int32, int64:
			isValid = true
		case string:
			if _, err := decimal.NewFromString(v); err == nil {
				isValid = true
			}
		}
	case "list", "array":
		actualTypeLower := strings.ToLower(actualType)
		isValid = strings.Contains(actualTypeLower, "slice") || strings.Contains(actualTypeLower, "array")
		if !isValid {
			if s, ok := val.(string); ok && (s == "[]" || s == "[[]]" || strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]")) {
				isValid = true
			}
		}
	case "map", "object":
		actualTypeLower := strings.ToLower(actualType)
		isValid = strings.Contains(actualTypeLower, "map")
		if !isValid {
			if s, ok := val.(string); ok && (s == "{}" || strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")) {
				isValid = true
			}
		}
	default:
		return nil
	}

	if !isValid {
		attrName := node.Name
		if attrName == slotName {
			attrName = "(main value)"
		}

		return Diagnostic{
			Type:     "error",
			Message:  fmt.Sprintf("validation error: type mismatch for '%s'. Expected %s, got %T (%v)", attrName, expectedType, val, val),
			Filename: node.Filename,
			Line:     node.Line,
			Col:      node.Col,
			Slot:     slotName,
		}
	}

	return nil
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
