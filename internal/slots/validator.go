package slots

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"zeno/pkg/dbmanager"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

func RegisterValidatorSlots(eng *engine.Engine, dbMgr *dbmanager.DBManager) {

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
		urlRegex := regexp.MustCompile(`^(http|https)://[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,}(?:/[^ "]+)?$`)
		alphaNumRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)

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

				// 6. CONFIRMED
				if r == "confirmed" {
					confirmField := field + "_confirmation"
					confirmVal, ok := inputData[confirmField]
					if !ok || coerce.ToString(confirmVal) != strVal {
						errors[field] = fmt.Sprintf("%s confirmation does not match", field)
						break
					}
				}

				// 7. ALPHA_NUM
				if r == "alpha_num" {
					if !alphaNumRegex.MatchString(strVal) {
						errors[field] = fmt.Sprintf("%s must contain only letters and numbers", field)
						break
					}
				}

				// 8. URL
				if r == "url" {
					if !urlRegex.MatchString(strVal) {
						errors[field] = fmt.Sprintf("%s must be a valid URL", field)
						break
					}
				}

				// 9. DATE
				if r == "date" {
					// Default format YYYY-MM-DD
					if _, err := time.Parse("2006-01-02", strVal); err != nil {
						errors[field] = fmt.Sprintf("%s must be a valid date (YYYY-MM-DD)", field)
						break
					}
				}

				// 10. IN:foo,bar,baz
				if strings.HasPrefix(r, "in:") {
					param := strings.TrimPrefix(r, "in:")
					options := strings.Split(param, ",")
					found := false
					for _, opt := range options {
						if strVal == strings.TrimSpace(opt) {
							found = true
							break
						}
					}
					if !found {
						errors[field] = fmt.Sprintf("%s is invalid", field)
						break
					}
				}

				// 11. UNIQUE:table,column
				if strings.HasPrefix(r, "unique:") {
					if dbMgr == nil {
						// Skip if DB not available (e.g. testing without DB)
						continue
					}
					param := strings.TrimPrefix(r, "unique:")
					parts := strings.Split(param, ",")
					if len(parts) >= 2 {
						table := parts[0]
						column := parts[1]

						// Basic check to prevent SQL injection in table/column names
						// Allow only alphanumeric and underscores
						if !alphaNumRegex.MatchString(table) || !alphaNumRegex.MatchString(column) {
							continue
						}

						// Use default connection
						db := dbMgr.GetConnection("default")
						if db != nil {
							var count int
							// EXCEPTION: If validating update, ignore own ID
							// unique:users,email,id,1  (ignore id=1)
							query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = ?", table, column)
							args := []interface{}{strVal}

							if len(parts) >= 4 {
								ignoreCol := parts[2]
								ignoreVal := parts[3]
								if alphaNumRegex.MatchString(ignoreCol) {
									query += fmt.Sprintf(" AND %s != ?", ignoreCol)
									args = append(args, ignoreVal)
								}
							}

							err := db.QueryRow(query, args...).Scan(&count)
							if err == nil && count > 0 {
								errors[field] = fmt.Sprintf("%s has already been taken", field)
								break
							}
						}
					}
				}

				// 12. EXISTS:table,column
				if strings.HasPrefix(r, "exists:") {
					if dbMgr == nil {
						continue
					}
					param := strings.TrimPrefix(r, "exists:")
					parts := strings.Split(param, ",")
					if len(parts) >= 2 {
						table := parts[0]
						column := parts[1]

						if !alphaNumRegex.MatchString(table) || !alphaNumRegex.MatchString(column) {
							continue
						}

						db := dbMgr.GetConnection("default")
						if db != nil {
							var count int
							query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = ?", table, column)
							err := db.QueryRow(query, strVal).Scan(&count)
							if err == nil && count == 0 {
								errors[field] = fmt.Sprintf("%s is invalid", field)
								break
							}
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

			// Helper: Safe Data (Whitelist)
			// Return only fields that are in 'rules'
			safeTarget := ""
			for _, c := range node.Children {
				if c.Name == "as_safe" {
					safeTarget = strings.TrimPrefix(coerce.ToString(c.Value), "$")
				}
			}

			if safeTarget != "" {
				safeData := make(map[string]interface{})
				for field := range rules {
					if val, ok := inputData[field]; ok {
						safeData[field] = val
					}
				}
				scope.Set(safeTarget, safeData)
			}
		}

		return nil
	}

	meta := engine.SlotMeta{
		Example: `validate: $form
  rules:
    email: "required|email|unique:users,email"
    password: "required|confirmed|min:8"
    role: "in:admin,user"
  as: $errs
  as_safe: $valid_data`}

	eng.Register("validator.validate", validateHandler, meta)
	eng.Register("validate", validateHandler, meta)
}
