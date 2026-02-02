package slots

import (
	"context"
	"testing"
	"zeno/pkg/engine"
)

func TestMoneyCalc(t *testing.T) {
	eng := engine.NewEngine()
	RegisterMathSlots(eng)

	tests := []struct {
		name     string
		expr     string
		scope    map[string]interface{}
		expected string
	}{
		{
			name: "Basic Addition with Constructor",
			expr: "decimal('0.1') + decimal('0.2')",
			expected: "0.3",
		},
		{
			name: "Variable Calculation",
			expr: "($price * $qty) - $discount",
			scope: map[string]interface{}{
				"price":    "100.50",
				"qty":      "2",
				"discount": "1.25",
			},
			expected: "199.75",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope := engine.NewScope(nil)
			if tt.scope != nil {
				for k, v := range tt.scope {
					scope.Set(k, v)
				}
			}

			node := &engine.Node{
				Name:  "money.calc",
				Value: tt.expr,
				Children: []*engine.Node{
					{Name: "as", Value: "result"},
				},
			}

			err := eng.Execute(context.Background(), node, scope)
			if err != nil {
				t.Fatalf("Execution failed: %v", err)
			}

			res, _ := scope.Get("result")
			if res != tt.expected {
				t.Errorf("Expected %s, got %v", tt.expected, res)
			}
		})
	}
}
