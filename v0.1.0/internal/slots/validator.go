package slots

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

func RegisterValidatorSlots(eng *engine.Engine) {

	// 1. VALIDATOR.VALIDATE & VALIDATE (Alias)
	validateHandler := func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var inputData map[string]interface{}
		var rules map[string]interface{}
		target := "errors"

		// Support Shorthand: validate: $form_data
		if node.Value != nil {
			val := resolveValue(node.Value, scope)
			if m, ok := val.(map[string]interface{}); ok {
				inputData = m
			}
		}

		// Prepare implicit data map
		implicitData := make(map[string]interface{})

		for _, c := range node.Children {
			val := parseNodeValue(c, scope)

			if c.Name == "input" || c.Name == "data" {
				if m, ok := val.(map[string]interface{}); ok {
					inputData = m
				}
				continue
			}
			if c.Name == "rules" {
				if m, ok := val.(map[string]interface{}); ok {
					rules = m
				}
				continue
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
				continue
			}

			// Treat other children as data fields
			implicitData[c.Name] = val
		}

		// Merge implicit data if inputData is still nil
		if inputData == nil {
			inputData = implicitData
		}

		if inputData == nil {
			return fmt.Errorf("validate: input data is missing or not a map")
		}

		errors := make(map[string]string)
		emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

		// Iterate Rules
		for field, ruleRaw := range rules {
			ruleStr := coerce.ToString(ruleRaw) // e.g. "required|email|min:5"
			ruleParts := strings.Split(ruleStr, "|")

			val, exists := inputData[field]
			strVal := coerce.ToString(val)

			for _, r := range ruleParts {
				// 1. REQUIRED
				if r == "required" {
					if !exists || strVal == "" {
						errors[field] = fmt.Sprintf("%s is required", field)
						break
					}
				}

				// Skip check if empty and not required
				if strVal == "" {
					continue
				}

				// 2. EMAIL
				if r == "email" {
					if !emailRegex.MatchString(strVal) {
						errors[field] = fmt.Sprintf("%s must be a valid email", field)
						break
					}
				}

				// 3. NUMERIC
				if r == "numeric" {
					if _, err := strconv.ParseFloat(strVal, 64); err != nil {
						errors[field] = fmt.Sprintf("%s must be a number", field)
						break
					}
				}

				// 4. MIN:X (Length or Value)
				if strings.HasPrefix(r, "min:") {
					param := strings.TrimPrefix(r, "min:")
					minVal, _ := strconv.ParseFloat(param, 64)

					// Jika input angka, cek value. Jika string, cek panjang.
					if num, err := strconv.ParseFloat(strVal, 64); err == nil {
						if num < minVal {
							errors[field] = fmt.Sprintf("%s must be at least %v", field, minVal)
							break
						}
					} else {
						if float64(len(strVal)) < minVal {
							errors[field] = fmt.Sprintf("%s must be at least %v characters", field, minVal)
							break
						}
					}
				}

				// 5. MAX:X
				if strings.HasPrefix(r, "max:") {
					param := strings.TrimPrefix(r, "max:")
					maxVal, _ := strconv.ParseFloat(param, 64)

					if num, err := strconv.ParseFloat(strVal, 64); err == nil {
						if num > maxVal {
							errors[field] = fmt.Sprintf("%s must not exceed %v", field, maxVal)
							break
						}
					} else {
						if float64(len(strVal)) > maxVal {
							errors[field] = fmt.Sprintf("%s must not exceed %v characters", field, maxVal)
							break
						}
					}
				}
			}
		}

		// Set Result
		if len(errors) > 0 {
			scope.Set(target, errors)
			scope.Set(target+"_any", true) // Flag helper untuk IF check
		} else {
			scope.Set(target, nil)
			scope.Set(target+"_any", false)
		}

		return nil
	}

	meta := engine.SlotMeta{
		Example: `validate: $form
  rules:
    email: "required|email"
    age: "numeric|min:18"
  as: $errs`}

	eng.Register("validator.validate", validateHandler, meta)
	eng.Register("validate", validateHandler, meta)
}
