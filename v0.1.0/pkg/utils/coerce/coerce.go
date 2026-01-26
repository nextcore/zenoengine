package coerce

import (
	"fmt"

	"github.com/spf13/cast"
)

// ============================================================================
// SAFE COERCION HELPERS
// Fungsi-fungsi ini mencoba mengonversi input (interface{}) ke tipe tujuan
// dengan aman. Jika gagal, akan mengembalikan error yang jelas, BUKAN panic.
// ============================================================================

// ToString mencoba mengubah input menjadi string.
// Hampir semua tipe bisa diubah jadi string, jadi error jarang terjadi.
// Nil akan menjadi string kosong "".
func ToString(input interface{}) string {
	if input == nil {
		return ""
	}
	// cast.ToStringE menangani hampir semua kasus dengan aman
	s, err := cast.ToStringE(input)
	if err != nil {
		// Fallback terakhir yang sangat aman: fmt.Sprintf
		// Ini memastikan kita selalu dapat string, walau mungkin bentuknya aneh (misal map jadi string)
		return fmt.Sprintf("%v", input)
	}
	return s
}

// ToInt mencoba mengubah input menjadi integer (int).
// Mendukung input string angka ("123"), float bulat (123.0), dll.
// Gagal jika inputnya string non-angka ("budi"), object, dll.
func ToInt(input interface{}) (int, error) {
	if input == nil {
		return 0, nil // Anggap nil sebagai 0
	}
	i, err := cast.ToIntE(input)
	if err != nil {
		return 0, fmt.Errorf("failed to coerce value '%v' (type %T) to int", input, input)
	}
	return i, nil
}

// ToInt64 sama seperti ToInt tapi untuk angka yang lebih besar.
// Berguna untuk ID database atau timestamp.
func ToInt64(input interface{}) (int64, error) {
	if input == nil {
		return 0, nil
	}
	i, err := cast.ToInt64E(input)
	if err != nil {
		return 0, fmt.Errorf("failed to coerce value '%v' (type %T) to int64", input, input)
	}
	return i, nil
}

// ToFloat64 mencoba mengubah input menjadi angka desimal.
// Mendukung string ("123.45"), integer, dll.
func ToFloat64(input interface{}) (float64, error) {
	if input == nil {
		return 0.0, nil
	}
	f, err := cast.ToFloat64E(input)
	if err != nil {
		return 0.0, fmt.Errorf("failed to coerce value '%v' (type %T) to float64", input, input)
	}
	return f, nil
}

// ToBool mencoba mengubah input menjadi boolean.
// Sangat cerdas: mendukung true/false, 1/0, "true"/"false", "on"/"off", "yes"/"no".
func ToBool(input interface{}) (bool, error) {
	if input == nil {
		return false, nil
	}
	b, err := cast.ToBoolE(input)
	if err != nil {
		return false, fmt.Errorf("failed to coerce value '%v' (type %T) to bool", input, input)
	}
	return b, nil
}

// ToMap mencoba mengubah input menjadi map[string]interface{}.
// Berguna jika inputnya adalah struct atau JSON object.
func ToMap(input interface{}) (map[string]interface{}, error) {
	if input == nil {
		return nil, nil
	}
	m, err := cast.ToStringMapE(input)
	if err != nil {
		return nil, fmt.Errorf("failed to coerce value (type %T) to map", input)
	}
	return m, nil
}

// ToSlice mencoba mengubah input menjadi array of interface{}.
func ToSlice(input interface{}) ([]interface{}, error) {
	if input == nil {
		return nil, nil
	}
	s, err := cast.ToSliceE(input)
	if err != nil {
		return nil, fmt.Errorf("failed to coerce value (type %T) to slice", input)
	}
	return s, nil
}

// Helper: Konversi ke Float64 dengan nilai default jika gagal
func ToFloat64Def(input interface{}, defaultVal float64) float64 {
	val, err := ToFloat64(input)
	if err != nil {
		return defaultVal
	}
	return val
}
