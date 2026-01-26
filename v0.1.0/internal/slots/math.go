package slots

import (
	"context"
	"fmt"
	"math"
	"strings"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"

	"github.com/expr-lang/expr"
	"github.com/shopspring/decimal"
)

func RegisterMathSlots(eng *engine.Engine) {

	// ==========================================
	// 1. SLOT: MATH.CALC (General Math - Float64)
	// ==========================================
	eng.Register("math.calc", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		expressionStr := coerce.ToString(node.Value)
		target := "calc_result"

		// Support shorthand & attributes
		for _, c := range node.Children {
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
			if c.Name == "val" || c.Name == "expr" {
				expressionStr = coerce.ToString(c.Value)
			}
		}

		if expressionStr == "" {
			return fmt.Errorf("math.calc: expression is required")
		}

		// 1. Siapkan Environment
		env := make(map[string]interface{})

		// Copy variabel dari scope & AUTO-CONVERT angka
		for k, v := range scope.ToMap() {
			// Cek apakah value ini string yang isinya angka? (misal: "5", "10.5")
			if str, ok := v.(string); ok {
				if f, err := coerce.ToFloat64(str); err == nil {
					env[k] = f // Simpan sebagai Float agar bisa dihitung
				} else {
					env[k] = str // Biarkan string jika bukan angka (misal: "budi")
				}
			} else {
				env[k] = v // Value lain (int, bool, object) biarkan apa adanya
			}
		}

		// [UPGRADE] Inject Fungsi Matematika Standar
		env["ceil"] = math.Ceil
		env["floor"] = math.Floor
		env["round"] = math.Round
		env["abs"] = math.Abs
		env["max"] = math.Max
		env["min"] = math.Min
		env["sqrt"] = math.Sqrt
		env["pow"] = math.Pow

		// 2. Pre-processing ($ -> kosong)
		cleanExpr := strings.ReplaceAll(expressionStr, "$", "")

		// 3. Compile & Run
		program, err := expr.Compile(cleanExpr, expr.Env(env))
		if err != nil {
			return fmt.Errorf("math.calc: syntax error '%s': %v", expressionStr, err)
		}

		output, err := expr.Run(program, env)
		if err != nil {
			return fmt.Errorf("math.calc: runtime error: %v", err)
		}

		scope.Set(target, output)
		return nil
	}, engine.SlotMeta{
		Description: "Melakukan perhitungan matematika menggunakan ekspresi string.",
		Example:     "math.calc: ceil($total / 10)\n  as: $pages",
		Inputs: map[string]engine.InputMeta{
			"as":   {Description: "Variabel penyimpan hasil", Required: false},
			"val":  {Description: "Ekspresi matematika (jika tidak via value utama)", Required: false},
			"expr": {Description: "Alias untuk val", Required: false},
		},
	})

	// ==========================================
	// 2. SLOT: MONEY.CALC (Financial Math - Decimal)
	// ==========================================
	eng.Register("money.calc", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		expressionStr := coerce.ToString(node.Value)
		target := "money_result"

		for _, c := range node.Children {
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
			if c.Name == "val" {
				expressionStr = coerce.ToString(c.Value)
			}
		}

		// 1. Siapkan Environment Decimal
		env := make(map[string]interface{})

		for k, v := range scope.ToMap() {
			s := coerce.ToString(v)
			if d, err := decimal.NewFromString(s); err == nil {
				env[k] = d
			} else {
				env[k] = decimal.Zero
			}
		}

		// 2. Pre-processing
		cleanExpr := strings.ReplaceAll(expressionStr, "$", "")

		// 3. Konfigurasi Expr (Operator Overloading)
		options := []expr.Option{
			expr.Env(env),
			expr.Operator("+", "Add"),
			expr.Operator("-", "Sub"),
			expr.Operator("*", "Mul"),
			expr.Operator("/", "Div"),
		}

		program, err := expr.Compile(cleanExpr, options...)
		if err != nil {
			return fmt.Errorf("money.calc: syntax error '%s': %v", expressionStr, err)
		}

		output, err := expr.Run(program, env)
		if err != nil {
			return fmt.Errorf("money.calc: runtime error: %v", err)
		}

		// Return sebagai String
		if d, ok := output.(decimal.Decimal); ok {
			scope.Set(target, d.String())
		} else {
			scope.Set(target, coerce.ToString(output))
		}

		return nil
	}, engine.SlotMeta{
		Description: "Melakukan perhitungan keuangan menggunakan Decimal untuk presisi tinggi.",
		Example:     "money.calc: ($harga * $qty) - $diskon\n  as: $total",
		Inputs: map[string]engine.InputMeta{
			"as":  {Description: "Variabel penyimpan hasil", Required: false},
			"val": {Description: "Ekspresi keuangan", Required: false},
		},
	})
}
