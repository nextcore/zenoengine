package blade

import (
	"fmt"
	"strings"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"

	"github.com/expr-lang/expr"
)

func parseNodeValue(n *engine.Node, scope *engine.Scope) interface{} {
	if len(n.Children) > 0 {
		m := make(map[string]interface{})
		for _, c := range n.Children {
			m[c.Name] = parseNodeValue(c, scope)
		}
		return m
	}

	valStr := strings.TrimSpace(fmt.Sprintf("%v", n.Value))

	// Try evaluating as a general expression if it looks like one (e.g. contains math operators, comparison, or starts with '$' and contains other expressions/spaces)
	isExpr := false
	if strings.ContainsAny(valStr, "+-*/%<>=!&|()") {
		isExpr = true
	} else if strings.HasPrefix(valStr, "$") && (strings.ContainsAny(valStr, " \t\n") || strings.Contains(valStr, "??")) {
		isExpr = true
	}

	if isExpr {
		cleanExpr := strings.ReplaceAll(valStr, "$", "")
		env := scope.GetAll()
		// Auto-convert numeric strings to float64 for calculations
		for k, v := range env {
			if str, ok := v.(string); ok {
				if f, err := coerce.ToFloat64(str); err == nil {
					env[k] = f
				}
			}
		}
		program, err := expr.Compile(cleanExpr, expr.Env(env))
		if err == nil {
			output, err := expr.Run(program, env)
			if err == nil {
				return output
			}
		}
	}

	if len(valStr) >= 2 {
		if (strings.HasPrefix(valStr, "\"") && strings.HasSuffix(valStr, "\"")) ||
			(strings.HasPrefix(valStr, "'") && strings.HasSuffix(valStr, "'")) {
			return valStr[1 : len(valStr)-1]
		}
	}

	if strings.HasPrefix(valStr, "$") {
		if strings.Contains(valStr, "??") {
			parts := strings.SplitN(valStr, "??", 2)
			if len(parts) == 2 {
				v1 := strings.TrimSpace(parts[0])
				v2 := strings.TrimSpace(parts[1])

				res1 := parseNodeValue(&engine.Node{Value: v1}, scope)
				if res1 != nil && fmt.Sprintf("%v", res1) != "" && fmt.Sprintf("%v", res1) != "<nil>" {
					return res1
				}
				return parseNodeValue(&engine.Node{Value: v2}, scope)
			}
		}

		if strings.Contains(valStr, " ? ") && strings.Contains(valStr, " : ") {
			parts := strings.SplitN(valStr, " ? ", 2)
			if len(parts) == 2 {
				condStr := strings.TrimSpace(parts[0])
				rest := strings.SplitN(parts[1], " : ", 2)
				if len(rest) == 2 {
					trueStr := strings.TrimSpace(rest[0])
					falseStr := strings.TrimSpace(rest[1])

					condVal := parseNodeValue(&engine.Node{Value: condStr}, scope)
					var b bool
					if condVal != nil {
						if bVal, err := coerce.ToBool(condVal); err == nil {
							b = bVal
						} else {
							strVal := coerce.ToString(condVal)
							b = (strVal != "" && strVal != "false" && strVal != "0" && strVal != "<nil>")
						}
					}

					if b {
						return parseNodeValue(&engine.Node{Value: trueStr}, scope)
					}
					return parseNodeValue(&engine.Node{Value: falseStr}, scope)
				}
			}
		}

		key := strings.TrimPrefix(valStr, "$")

		if strings.Contains(key, ".") || strings.Contains(key, "[") {
			normalizedKey := key
			if strings.Contains(normalizedKey, "[") {
				normalizedKey = strings.ReplaceAll(normalizedKey, "[", ".")
				normalizedKey = strings.ReplaceAll(normalizedKey, "]", "")
			}

			parts := strings.Split(normalizedKey, ".")
			rootKey := strings.TrimSpace(parts[0])

			if rootVal, ok := scope.Get(rootKey); ok {
				curr := rootVal
				isValidPath := true
				for i := 1; i < len(parts); i++ {
					targetKey := strings.TrimSpace(parts[i])
					if targetKey == "" {
						continue
					}

					if m, ok := curr.(map[string]interface{}); ok {
						if val, exists := m[targetKey]; exists {
							curr = val
							continue
						}
					}

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

		if v, ok := scope.Get(strings.TrimSpace(key)); ok {
			return v
		}
		return nil
	}

	return n.Value
}

func resolveValue(val interface{}, scope *engine.Scope) interface{} {
	return parseNodeValue(&engine.Node{Value: val}, scope)
}
