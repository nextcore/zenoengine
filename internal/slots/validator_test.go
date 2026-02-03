package slots

import (
	"context"
	"testing"
	"zeno/pkg/engine"
)

func TestValidatorSlots(t *testing.T) {
	eng := engine.NewEngine()
	RegisterValidatorSlots(eng)

	t.Run("validate_success", func(t *testing.T) {
		scope := engine.NewScope(nil)
		form := map[string]interface{}{
			"email": "test@example.com",
			"age":   "25",
		}
		scope.Set("form", form)

		node := &engine.Node{
			Name: "validate",
			Children: []*engine.Node{
				{Name: "data", Value: "$form"},
				{
					Name: "rules",
					Children: []*engine.Node{
						{Name: "email", Value: "required|email"},
						{Name: "age", Value: "numeric|min:18"},
					},
				},
				{Name: "as", Value: "$v_errors"},
			},
		}
		if err := eng.Execute(context.Background(), node, scope); err != nil {
			t.Fatalf("validate failed: %v", err)
		}

		any, _ := scope.Get("v_errors_any")
		if any != false {
			errs, _ := scope.Get("v_errors")
			t.Errorf("Expected no errors, got %v", errs)
		}
	})

	t.Run("validate_failure", func(t *testing.T) {
		scope := engine.NewScope(nil)
		form := map[string]interface{}{
			"email": "invalid-email",
			"age":   "15",
		}
		scope.Set("form", form)

		node := &engine.Node{
			Name: "validate",
			Children: []*engine.Node{
				{Name: "data", Value: "$form"},
				{
					Name: "rules",
					Children: []*engine.Node{
						{Name: "email", Value: "required|email"},
						{Name: "age", Value: "numeric|min:18"},
					},
				},
				{Name: "as", Value: "$v_errors"},
			},
		}
		if err := eng.Execute(context.Background(), node, scope); err != nil {
			t.Fatalf("validate failed: %v", err)
		}

		any, _ := scope.Get("v_errors_any")
		if any != true {
			t.Errorf("Expected validation errors, but none found")
		}

		errs, _ := scope.Get("v_errors")
		errMap := errs.(map[string]string)
		if _, ok := errMap["email"]; !ok {
			t.Errorf("Expected email error")
		}
		if _, ok := errMap["age"]; !ok {
			t.Errorf("Expected age error")
		}
	})
}
