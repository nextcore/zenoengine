package slots

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"

	"github.com/gosimple/slug"
	"github.com/microcosm-cc/bluemonday"
)

// (STRICT MODE: String Literal diprioritaskan daripada Variable)
func parseNodeValue(n *engine.Node, scope *engine.Scope) interface{} {
	// 1. Nested Object (Map)
	if len(n.Children) > 0 {
		m := make(map[string]interface{})
		for _, c := range n.Children {
			m[c.Name] = parseNodeValue(c, scope)
		}
		return m
	}

	valStr := strings.TrimSpace(fmt.Sprintf("%v", n.Value))

	// 2. [PRIORITAS UTAMA] CEK STRING LITERAL (Kutip)
	if len(valStr) >= 2 {
		if (strings.HasPrefix(valStr, "\"") && strings.HasSuffix(valStr, "\"")) ||
			(strings.HasPrefix(valStr, "'") && strings.HasSuffix(valStr, "'")) {
			return valStr[1 : len(valStr)-1]
		}
	}

	// 3. BARU CEK VARIABLE ($)
	if strings.HasPrefix(valStr, "$") {
		// [NEW] Support Null Coalesce: $var ?? 'default'
		if strings.Contains(valStr, "??") {
			parts := strings.SplitN(valStr, "??", 2)
			v1 := strings.TrimSpace(parts[0])
			v2 := strings.TrimSpace(parts[1])

			res1 := parseNodeValue(&engine.Node{Value: v1}, scope)
			if res1 != nil && fmt.Sprintf("%v", res1) != "" && fmt.Sprintf("%v", res1) != "<nil>" {
				return res1
			}
			return parseNodeValue(&engine.Node{Value: v2}, scope)
		}

		key := strings.TrimPrefix(valStr, "$")
		// fmt.Println("DEBUG RESOLVE:", key)

		// A. Check Dot Notation OR Bracket Notation
		if strings.Contains(key, ".") || strings.Contains(key, "[") {
			// Normalize Bracket Notation: users[0] -> users.0
			normalizedKey := key
			if strings.Contains(normalizedKey, "[") {
				normalizedKey = strings.ReplaceAll(normalizedKey, "[", ".")
				normalizedKey = strings.ReplaceAll(normalizedKey, "]", "")
			}

			parts := strings.Split(normalizedKey, ".")
			rootKey := parts[0]

			if rootVal, ok := scope.Get(rootKey); ok {
				curr := rootVal
				isValidPath := true
				for i := 1; i < len(parts); i++ {
					targetKey := parts[i]

					// 1. Handle Maps
					if m, ok := curr.(map[string]interface{}); ok {
						if val, exists := m[targetKey]; exists {
							curr = val
							continue
						}
					}

					// 2. Handle Arrays/Slices
					if list, err := coerce.ToSlice(curr); err == nil {
						if idx, err := coerce.ToInt(targetKey); err == nil {
							if idx >= 0 && idx < len(list) {
								curr = list[idx]
								continue
							}
						}
					}

					isValidPath = false
					break
				}
				if isValidPath {
					return curr
				}
			}
			return nil
		}

		// B. Cek Variabel Biasa
		if v, ok := scope.Get(key); ok {
			return v
		}
		return nil
	}

	// 4. FALLBACK (Nilai mentah)
	return n.Value
}

// Helper: Resolusi Nilai Tunggal
func resolveValue(val interface{}, scope *engine.Scope) interface{} {
	return parseNodeValue(&engine.Node{Value: val}, scope)
}

// MustResolveValue returns an error if the value is nil or not found.
// Useful for strict validation in slots.
func MustResolveValue(val interface{}, scope *engine.Scope, name string) (interface{}, error) {
	res := resolveValue(val, scope)
	if res == nil || coerce.ToString(res) == "nil" || coerce.ToString(res) == "<nil>" {
		return nil, fmt.Errorf("variable or value '%s' is nil or not found in scope", name)
	}
	return res, nil
}

func RegisterUtilSlots(eng *engine.Engine) {
	// 1. LOG
	eng.Register("log", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		val := resolveValue(node.Value, scope)
		fmt.Println("[LOG]:", val)
		return nil
	}, engine.SlotMeta{Example: "log: $user.name"})

	// 2. STRINGS CONCAT
	eng.Register("strings.concat", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var builder strings.Builder
		target := "concat_result"
		for _, c := range node.Children {
			if c.Name == "as" {
				// [FIX] Bersihkan awalan $ agar variabel tersimpan benar
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
				continue
			}
			if c.Name == "val" {
				val := parseNodeValue(c, scope)
				builder.WriteString(coerce.ToString(val))
			}
		}
		scope.Set(target, builder.String())
		return nil
	}, engine.SlotMeta{
		Description: "Menggabungkan beberapa string menjadi satu.",
		Example:     "strings.concat\n  val: 'Hello '\n  val: $name\n  as: $greeting",
		Inputs: map[string]engine.InputMeta{
			"val": {Description: "Nilai yang akan digabung (Bisa banyak)", Required: true},
			"as":  {Description: "Variabel penyimpan hasil", Required: false},
		},
	})

	// 2.5 STRINGS REPLACE
	eng.Register("string.replace", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		original := coerce.ToString(resolveValue(node.Value, scope))
		var find, replace string
		target := "replace_result"
		limit := -1

		for _, c := range node.Children {
			val := parseNodeValue(c, scope)
			if c.Name == "find" {
				find = coerce.ToString(val)
			}
			if c.Name == "replace" {
				replace = coerce.ToString(val)
			}
			if c.Name == "limit" {
				limit, _ = coerce.ToInt(val)
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		result := strings.Replace(original, find, replace, limit)
		scope.Set(target, result)
		return nil
	}, engine.SlotMeta{
		Description: "Mengganti substring dalam string dengan string lain.",
		Example:     "string.replace: $text\n  find: 'old'\n  replace: 'new'\n  as: $result",
		Inputs: map[string]engine.InputMeta{
			"find":    {Description: "Substring yang dicari", Required: true},
			"replace": {Description: "Substring pengganti", Required: true},
			"limit":   {Description: "Jumlah penggantian maksimum (-1 untuk semua)", Required: false},
			"as":      {Description: "Variabel penyimpan hasil", Required: false},
		},
	})

	// 3. TEXT SLUGIFY
	eng.Register("text.slugify", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		input := coerce.ToString(resolveValue(node.Value, scope))
		target := "slug_result"
		for _, c := range node.Children {
			if c.Name == "as" {
				// [FIX] Bersihkan awalan $
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
			if c.Name == "text" || c.Name == "val" {
				input = coerce.ToString(parseNodeValue(c, scope))
			}
		}
		scope.Set(target, slug.Make(input))
		return nil
	}, engine.SlotMeta{
		Description: "Mengubah teks menjadi format URL-friendly slug.",
		Example:     "text.slugify: 'Halo Dunia'\n  as: $my_slug",
		Inputs: map[string]engine.InputMeta{
			"text": {Description: "Teks sumber", Required: false},
			"val":  {Description: "Alias untuk text", Required: false},
			"as":   {Description: "Variabel penyimpan hasil", Required: false},
		},
	})

	// 4. TEXT SANITIZE
	eng.Register("text.sanitize", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		p := bluemonday.UGCPolicy()
		var input, target string
		target = "clean_text"
		for _, c := range node.Children {
			if c.Name == "input" || c.Name == "val" {
				input = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "as" {
				// [FIX] Bersihkan awalan $
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}
		scope.Set(target, p.Sanitize(input))
		return nil
	}, engine.SlotMeta{
		Description: "Membersihkan teks dari tag HTML berbahaya (XSS prevention).",
		Example:     "text.sanitize: $user_input\n  as: $clean_input",
		Inputs: map[string]engine.InputMeta{
			"input": {Description: "Teks sumber", Required: false},
			"val":   {Description: "Alias untuk input", Required: false},
			"as":    {Description: "Variabel penyimpan hasil", Required: false},
		},
	})

	// 5. SYS INCLUDE
	eng.Register("include", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		path := coerce.ToString(resolveValue(node.Value, scope))
		root, err := engine.LoadScript(path)
		if err != nil {
			return err
		}
		return eng.Execute(ctx, root, scope)
	}, engine.SlotMeta{})

	// 6. SYSTEM ENV
	// 6. SYSTEM ENV
	eng.Register("system.env", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		envKey := coerce.ToString(node.Value)
		val := os.Getenv(envKey)
		targetBucket := envKey // Default: use env key as var name

		for _, c := range node.Children {
			if c.Name == "as" {
				targetBucket = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		scope.Set(targetBucket, val)
		return nil
	}, engine.SlotMeta{})

	// 7. HTTP RESPONSE
	eng.Register("http.response", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		w, ok := ctx.Value("httpWriter").(http.ResponseWriter)
		if !ok {
			return fmt.Errorf("http.response: not in http context")
		}

		status := 200
		var data interface{}

		if node.Value != nil {
			status, _ = coerce.ToInt(node.Value)
		}
		for _, c := range node.Children {
			if c.Name == "status" {
				status, _ = coerce.ToInt(c.Value)
			}
			if c.Name == "data" || c.Name == "body" {
				data = parseNodeValue(c, scope)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		return json.NewEncoder(w).Encode(data)
	}, engine.SlotMeta{
		Description: "Mengirim response HTTP dalam format JSON.",
		Example:     "http.response: 200\n  data: $user_info",
		Inputs: map[string]engine.InputMeta{
			"status": {Description: "HTTP Status Code", Required: false},
			"data":   {Description: "Data JSON (Alias untuk body)", Required: false},
			"body":   {Description: "Data JSON", Required: false},
		},
	})

	// 8. HTTP REDIRECT
	eng.Register("http.redirect", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		w, ok := ctx.Value("httpWriter").(http.ResponseWriter)
		r, ok2 := ctx.Value("httpRequest").(*http.Request)
		if !ok || !ok2 {
			return fmt.Errorf("http.redirect: context missing")
		}

		url := coerce.ToString(resolveValue(node.Value, scope))
		http.Redirect(w, r, url, http.StatusSeeOther)
		return nil
	}, engine.SlotMeta{})

	// 9. IF (UPGRADED: Support ==, !=, >, <, >=, <=)
	eng.Register("if", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		expression := coerce.ToString(node.Value)
		isTrue := false

		// Helper untuk parsing bagian kiri dan kanan operator
		parseParts := func(expr, op string) (interface{}, interface{}) {
			parts := strings.SplitN(expr, op, 2)
			left := resolveValue(strings.TrimSpace(parts[0]), scope)
			right := resolveValue(strings.TrimSpace(parts[1]), scope)
			return left, right
		}

		if strings.Contains(expression, "==") {
			left, right := parseParts(expression, "==")
			// Handle Explicit Nil Check
			if (left == nil && coerce.ToString(right) == "nil") || (right == nil && coerce.ToString(left) == "nil") {
				isTrue = true
			} else {
				isTrue = (coerce.ToString(left) == coerce.ToString(right))
			}

		} else if strings.Contains(expression, "!=") {
			left, right := parseParts(expression, "!=")
			isTrue = (coerce.ToString(left) != coerce.ToString(right))

		} else if strings.Contains(expression, ">=") {
			left, right := parseParts(expression, ">=")
			l, err1 := coerce.ToFloat64(left)
			r, err2 := coerce.ToFloat64(right)
			if err1 == nil && err2 == nil {
				isTrue = (l >= r)
			}
		} else if strings.Contains(expression, "<=") {
			left, right := parseParts(expression, "<=")
			l, err1 := coerce.ToFloat64(left)
			r, err2 := coerce.ToFloat64(right)
			if err1 == nil && err2 == nil {
				isTrue = (l <= r)
			}
		} else if strings.Contains(expression, ">") {
			left, right := parseParts(expression, ">")
			l, err1 := coerce.ToFloat64(left)
			r, err2 := coerce.ToFloat64(right)
			if err1 == nil && err2 == nil {
				isTrue = (l > r)
			}
		} else if strings.Contains(expression, "<") {
			left, right := parseParts(expression, "<")
			l, err1 := coerce.ToFloat64(left)
			r, err2 := coerce.ToFloat64(right)
			if err1 == nil && err2 == nil {
				isTrue = (l < r)
			}
		} else {
			// Logic: Truthy Check (Single Value)
			val := resolveValue(node.Value, scope)
			if b, err := coerce.ToBool(val); err == nil {
				isTrue = b
			} else {
				s := coerce.ToString(val)
				isTrue = (s != "" && s != "false" && s != "0" && s != "<nil>")
			}
		}

		// Eksekusi Blok Then/Else
		var target *engine.Node
		if isTrue {
			for _, c := range node.Children {
				if c.Name == "then" {
					target = c
					break
				}
			}
		} else {
			for _, c := range node.Children {
				if c.Name == "else" {
					target = c
					break
				}
			}
		}

		if target != nil {
			for _, c := range target.Children {
				if err := eng.Execute(ctx, c, scope); err != nil {
					return err
				}
			}
		}
		return nil
	}, engine.SlotMeta{
		Description: "Kondisional if-then-else. Support: ==, !=, >, <, >=, <=",
		Inputs: map[string]engine.InputMeta{
			"then": {Description: "Blok kode jika kondisi benar", Required: false},
			"else": {Description: "Blok kode jika kondisi salah", Required: false},
		},
		RequiredBlocks: []string{"then"},
	})

	// 10. ARRAY LENGTH
	eng.Register("arrays.length", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		val := parseNodeValue(node, scope)
		list, _ := coerce.ToSlice(val)
		target := fmt.Sprintf("%v_length", node.Value)
		for _, c := range node.Children {
			if c.Name == "as" {
				// [FIX] Bersihkan awalan $
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}
		scope.Set(target, len(list))
		return nil
	}, engine.SlotMeta{
		Description: "Mengambil jumlah elemen dalam sebuah array atau list.",
		Example:     "arrays.length: $users\n  as: $count",
		Inputs: map[string]engine.InputMeta{
			"as": {Description: "Variabel penyimpan hasil", Required: true},
		},
	})

	eng.Register("doc.api", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error { return nil }, engine.SlotMeta{})

	// SLOT: VAR
	eng.Register("var", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		rawName := coerce.ToString(node.Value)
		varName := strings.TrimPrefix(rawName, "$")

		if varName == "" {
			return fmt.Errorf("var: variable name required (usage: var: $name)")
		}

		var val interface{}
		for _, c := range node.Children {
			if c.Name == "val" || c.Name == "value" {
				val = parseNodeValue(c, scope)
			}
		}
		scope.Set(varName, val)
		return nil
	}, engine.SlotMeta{
		Description: "Membuat atau mengubah nilai variabel dalam scope saat ini.",
		Example:     "var: $count\n  val: 10",
		Inputs: map[string]engine.InputMeta{
			"val":   {Description: "Nilai variabel", Required: false},
			"value": {Description: "Alias untuk val", Required: false},
		},
	})
	// 11. SLEEP
	eng.Register("sleep", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		ms, _ := coerce.ToInt(node.Value)
		time.Sleep(time.Duration(ms) * time.Millisecond)
		return nil
	}, engine.SlotMeta{Example: "sleep: 1000"})

	// ==========================================
	// SLOT: COALESCE (Null Safety)
	// ==========================================
	eng.Register("coalesce", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var val, def interface{}
		target := "coalesce_result"

		// Support value as main input
		if node.Value != nil {
			val = resolveValue(node.Value, scope)
		}

		for _, c := range node.Children {
			if c.Name == "value" || c.Name == "val" {
				val = parseNodeValue(c, scope)
			}
			// Default value
			if c.Name == "default" || c.Name == "def" {
				def = parseNodeValue(c, scope)
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		result := val

		// Check for true nil or string "nil"
		if result == nil || coerce.ToString(result) == "nil" || coerce.ToString(result) == "<nil>" {
			result = def
		}

		scope.Set(target, result)
		return nil
	}, engine.SlotMeta{
		Description: "Mengembalikan nilai default jika input bernilai null.",
		Example:     "coalesce: $user.name { default: 'Guest'; as: $name }",
	})

	// ==========================================
	// SLOT: IS_NULL
	// ==========================================
	eng.Register("is_null", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		val := resolveValue(node.Value, scope)
		target := "is_null"

		for _, c := range node.Children {
			if c.Name == "val" || c.Name == "value" {
				val = parseNodeValue(c, scope)
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		scope.Set(target, val == nil)
		return nil
	}, engine.SlotMeta{})

	// ==========================================
	// SLOT: CAST.TO_INT
	// ==========================================
	eng.Register("cast.to_int", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		val := resolveValue(node.Value, scope)
		target := "cast_result"

		for _, c := range node.Children {
			if c.Name == "val" || c.Name == "value" {
				val = parseNodeValue(c, scope)
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		res, err := coerce.ToInt(val)
		if err != nil {
			// Jika gagal, set 0 atau null? Default 0 aman untuk ID.
			res = 0
		}
		scope.Set(target, res)
		return nil
	}, engine.SlotMeta{
		Description: "Mengubah variabel menjadi Integer.",
		Example:     "cast.to_int: $id { as: $id_int }",
	})
}
