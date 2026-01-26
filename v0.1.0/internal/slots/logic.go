package slots

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrBreak    = errors.New("break")
	ErrContinue = errors.New("continue")
	ErrReturn   = errors.New("return")
)

func RegisterLogicSlots(eng *engine.Engine) {
	// ==========================================
	// SLOT: RETURN / STOP
	// ==========================================
	eng.Register("return", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		return ErrReturn
	}, engine.SlotMeta{Description: "Halt execution of the current block/handler."})

	// ==========================================
	// SLOT: SCOPE SET (Legacy alias for 'var')
	// ==========================================
	eng.Register("scope.set", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var key string
		var val interface{}

		// Support Shorthand: scope.set: $nama_var
		if node.Value != nil {
			raw := coerce.ToString(node.Value)
			key = strings.TrimPrefix(raw, "$")
		}

		for _, c := range node.Children {
			if c.Name == "key" || c.Name == "name" {
				raw := coerce.ToString(c.Value)
				key = strings.TrimPrefix(raw, "$")
			}
			if c.Name == "val" || c.Name == "value" {
				// First parse the node value
				parsedVal := parseNodeValue(c, scope)

				// Check if it's still a variable reference that needs resolution
				valStr := fmt.Sprintf("%v", parsedVal)
				if strings.HasPrefix(valStr, "$") {
					// Recursively resolve the variable
					resolvedVal := valStr
					maxDepth := 10
					depth := 0

					for depth < maxDepth && strings.HasPrefix(resolvedVal, "$") {
						varName := strings.TrimPrefix(resolvedVal, "$")

						// Handle dot notation (e.g., tenants.0.db_connection_name)
						if strings.Contains(varName, ".") {
							parts := strings.Split(varName, ".")
							rootKey := parts[0]

							if rootVal, ok := scope.Get(rootKey); ok {
								// Navigate through the path
								curr := rootVal
								for j := 1; j < len(parts); j++ {
									targetKey := parts[j]

									// Handle maps
									if m, ok := curr.(map[string]interface{}); ok {
										if v, exists := m[targetKey]; exists {
											curr = v
											continue
										}
									}

									// Handle arrays/slices with numeric index
									if list, err := coerce.ToSlice(curr); err == nil {
										if idx, err := coerce.ToInt(targetKey); err == nil {
											if idx >= 0 && idx < len(list) {
												curr = list[idx]
												continue
											}
										}
									}

									// Not found
									curr = nil
									break
								}

								if curr != nil {
									// Check if result is still a variable reference
									currStr := fmt.Sprintf("%v", curr)
									if strings.HasPrefix(currStr, "$") {
										resolvedVal = currStr
										depth++
										continue
									} else {
										// Final value
										val = curr
										break
									}
								} else {
									val = nil
									break
								}
							} else {
								val = nil
								break
							}
						} else {
							// Simple variable
							if v, ok := scope.Get(varName); ok {
								vStr := fmt.Sprintf("%v", v)
								if strings.HasPrefix(vStr, "$") {
									resolvedVal = vStr
									depth++
									continue
								} else {
									val = v
									break
								}
							} else {
								val = nil
								break
							}
						}
					}

					if depth >= maxDepth {
						val = nil
					}
				} else {
					// Not a variable reference, use parsed value as-is
					val = parsedVal
				}
			}
		}

		if key != "" {
			scope.Set(key, val)
		}
		return nil
	}, engine.SlotMeta{
		Description: "Create a variable (Legacy alias for 'var').",
		Example:     "scope.set: $my_var\n  val: 123",
		Inputs: map[string]engine.InputMeta{
			"key":   {Description: "Variable name", Required: false},
			"name":  {Description: "Variable name (alias for key)", Required: false},
			"val":   {Description: "Variable value", Required: false},
			"value": {Description: "Variable value (alias for val)", Required: false},
		},
	})

	// ==========================================
	// SLOT: LOGIC.COMPARE
	// ==========================================
	eng.Register("logic.compare", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var v1, v2 interface{}
		var op string
		target := "compare_result"

		for _, c := range node.Children {
			val := parseNodeValue(c, scope)
			if c.Name == "v1" {
				v1 = val
			}
			if c.Name == "v2" {
				v2 = val
			}
			if c.Name == "op" {
				op = coerce.ToString(val)
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		res := false
		f1, err1 := coerce.ToFloat64(v1)
		f2, err2 := coerce.ToFloat64(v2)

		if err1 == nil && err2 == nil {
			switch op {
			case "==":
				res = (f1 == f2)
			case "!=":
				res = (f1 != f2)
			case ">":
				res = (f1 > f2)
			case "<":
				res = (f1 < f2)
			case ">=":
				res = (f1 >= f2)
			case "<=":
				res = (f1 <= f2)
			}
		} else {
			s1 := coerce.ToString(v1)
			s2 := coerce.ToString(v2)
			switch op {
			case "==":
				res = (s1 == s2)
			case "!=":
				res = (s1 != s2)
			}
		}

		scope.Set(target, res)
		return nil
	}, engine.SlotMeta{
		Description: "Compare two values.",
		Example:     "logic.compare\n  v1: $age\n  op: '>'\n  v2: 17",
		Inputs: map[string]engine.InputMeta{
			"v1": {Description: "First value", Required: true},
			"v2": {Description: "Second value", Required: true},
			"op": {Description: "Comparison operator", Required: true},
			"as": {Description: "Variable to store result (Default: compare_result)", Required: false},
		},
	})

	// ==========================================
	// SLOT: TRY / CATCH
	// ==========================================
	eng.Register("try", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var doNode, catchNode *engine.Node
		errVar := "error"

		for _, c := range node.Children {
			if c.Name == "do" {
				doNode = c
			}
			if c.Name == "catch" {
				catchNode = c
			}
			if c.Name == "as" {
				errVar = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		if doNode != nil {
			for _, child := range doNode.Children {
				if err := eng.Execute(ctx, child, scope); err != nil {
					// DON'T catch control flow errors (return, break, continue)
					if errors.Is(err, ErrReturn) || errors.Is(err, ErrBreak) || errors.Is(err, ErrContinue) {
						return err
					}
					if catchNode != nil {
						scope.Set(errVar, err.Error())
						return eng.Execute(ctx, catchNode, scope)
					}
					return err
				}
			}
		}
		return nil
	}, engine.SlotMeta{
		Description: "Handle errors using a try-catch block.",
		Example:     "try {\n  do: { ... }\n  catch: { ... }\n}",
		Inputs: map[string]engine.InputMeta{
			"as":    {Description: "Variable name for error message (Default: 'error')", Required: false},
			"do":    {Description: "Main code block to execute", Required: false},
			"catch": {Description: "Error handling code block", Required: false},
		},
	})

	// ==========================================
	// SLOT: FOR LOOP
	// ==========================================
	// ==========================================
	handlerFor := func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var raw interface{}
		valStr, _ := node.Value.(string)
		if strings.Count(valStr, ";") == 2 {
			raw = valStr
		} else {
			raw = resolveValue(node.Value, scope)
		}
		rawStr := coerce.ToString(raw)

		var doNode *engine.Node
		for _, c := range node.Children {
			if c.Name == "do" {
				doNode = c
				break
			}
		}

		// A. C-Style For Loop (e.g. "$i = 0; $i < 10; $i++")
		if strings.Count(rawStr, ";") == 2 {
			parts := strings.Split(rawStr, ";")
			initStr := strings.TrimSpace(parts[0])
			condStr := strings.TrimSpace(parts[1])
			updStr := strings.TrimSpace(parts[2])

			// 1. INIT
			var loopVar string
			if strings.Contains(initStr, "=") {
				initParts := strings.SplitN(initStr, "=", 2)
				loopVar = strings.TrimPrefix(strings.TrimSpace(initParts[0]), "$")
				initVal, _ := strconv.Atoi(strings.TrimSpace(initParts[1]))
				scope.Set(loopVar, initVal)
			}

			// LOOP
			if doNode != nil {
				maxLoop := 10000
				for i := 0; i < maxLoop; i++ {
					// Check context cancellation/timeout every iteration
					select {
					case <-ctx.Done():
						return fmt.Errorf("for loop cancelled: %v", ctx.Err())
					default:
					}

					// 2. CONDITION
					if !evalSimpleCondition(condStr, scope) {
						break
					}

					// DO BODY
					if err := eng.Execute(ctx, doNode, scope); err != nil {
						if errors.Is(err, ErrBreak) || strings.Contains(err.Error(), "break") {
							return nil
						}
						if errors.Is(err, ErrContinue) || strings.Contains(err.Error(), "continue") {
							goto Update
						}
						return err
					}

				Update:
					// 3. UPDATE
					if strings.Contains(updStr, "++") {
						vRaw, _ := scope.Get(loopVar)
						vInt, _ := coerce.ToInt(vRaw)
						scope.Set(loopVar, vInt+1)
					} else if strings.Contains(updStr, "--") {
						vRaw, _ := scope.Get(loopVar)
						vInt, _ := coerce.ToInt(vRaw)
						scope.Set(loopVar, vInt-1)
					}
				}
			}
			return nil
		}

		// B. Foreach Loop
		sourceList, err := coerce.ToSlice(raw)
		if err != nil {
			return nil // Empty loop if data invalid
		}

		var itemName string = "item"
		for _, c := range node.Children {
			if c.Name == "as" {
				itemName = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		if doNode != nil {
			count := len(sourceList)
			parentLoop, hasParentLoop := scope.Get("loop")

			for i, item := range sourceList {
				scope.Set(itemName, item)
				loopData := map[string]interface{}{
					"index":     i,
					"iteration": i + 1,
					"remaining": count - (i + 1),
					"count":     count,
					"first":     (i == 0),
					"last":      (i == count-1),
					"even":      ((i+1)%2 == 0),
					"odd":       ((i+1)%2 != 0),
				}
				if hasParentLoop {
					loopData["parent"] = parentLoop
				}
				scope.Set("loop", loopData)

				if err := eng.Execute(ctx, doNode, scope); err != nil {
					if errors.Is(err, ErrBreak) || strings.Contains(err.Error(), "break") {
						goto EndForeach
					}
					if errors.Is(err, ErrContinue) || strings.Contains(err.Error(), "continue") {
						continue
					}
					return err
				}
			}
		EndForeach:
			if hasParentLoop {
				scope.Set("loop", parentLoop)
			} else {
				scope.Set("loop", nil)
			}
		}
		return nil
	}

	eng.Register("for", handlerFor, engine.SlotMeta{
		Description: "Iterate (loop) over a list or array.",
		Example:     "for: $list\n  as: $item\n  do: ...",
		Inputs: map[string]engine.InputMeta{
			"as": {Description: "Variable name for current element (Default: 'item')", Required: false},
			"do": {Description: "Code block to repeat", Required: false},
		},
		RequiredBlocks: []string{"do"},
	})
	eng.Register("foreach", handlerFor, engine.SlotMeta{Example: "foreach: $list { as: $item ... }"}) // Alias

	// ==========================================
	// SLOT: CTX TIMEOUT
	// ==========================================
	eng.Register("ctx.timeout", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var durationStr string
		var doNode *engine.Node

		// Get duration from main value (ctx.timeout: "5s")
		if node.Value != nil {
			durationStr = coerce.ToString(resolveValue(node.Value, scope))
		}

		for _, c := range node.Children {
			if c.Name == "duration" {
				durationStr = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "do" {
				doNode = c
			}
		}

		if durationStr == "" {
			return fmt.Errorf("ctx.timeout: duration is required")
		}

		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			return fmt.Errorf("ctx.timeout: invalid duration '%s'", durationStr)
		}

		timeoutCtx, cancel := context.WithTimeout(ctx, duration)
		defer cancel()

		if doNode != nil {
			for _, child := range doNode.Children {
				if err := eng.Execute(timeoutCtx, child, scope); err != nil {
					if timeoutCtx.Err() == context.DeadlineExceeded {
						return fmt.Errorf("execution timed out after %s", durationStr)
					}
					return err
				}
			}
		}
		return nil
	}, engine.SlotMeta{
		Description: "Limit execution time of a code block.",
		Example:     "ctx.timeout: '5s' {\n  do: { ... }\n}",
		Inputs: map[string]engine.InputMeta{
			"duration": {Description: "Timeout duration (e.g., '5s', '1m')", Required: false},
			"do":       {Description: "Code block to execute", Required: false},
		},
	})

	// ==========================================
	// SLOT: WHILE LOOP / LOOP
	// ==========================================
	handlerWhile := func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		condRaw := coerce.ToString(node.Value)
		maxLoop := 10000

		for i := 0; i < maxLoop; i++ {
			// Check context cancellation/timeout every iteration
			select {
			case <-ctx.Done():
				return fmt.Errorf("while loop cancelled: %v", ctx.Err())
			default:
			}

			if !evalSimpleCondition(condRaw, scope) {
				break
			}

			// Find 'do' block or use the node itself if it has no 'do' child but has children
			var childrenToExec []*engine.Node
			for _, c := range node.Children {
				if c.Name == "do" {
					childrenToExec = c.Children
					break
				}
			}

			// Fallback: execute all children if no 'do' block found
			if childrenToExec == nil {
				childrenToExec = node.Children
			}

			for _, child := range childrenToExec {
				if err := eng.Execute(ctx, child, scope); err != nil {
					if errors.Is(err, ErrBreak) || strings.Contains(err.Error(), "break") {
						goto EndWhile
					}
					if errors.Is(err, ErrContinue) || strings.Contains(err.Error(), "continue") {
						break
					}
					return err
				}
			}
		}
	EndWhile:
		return nil
	}

	eng.Register("while", handlerWhile, engine.SlotMeta{
		Description:    "While loop",
		RequiredBlocks: []string{"do"},
		Inputs: map[string]engine.InputMeta{
			"do": {Description: "Code block to execute"},
		},
	})
	eng.Register("loop", handlerWhile, engine.SlotMeta{
		Description:    "While loop",
		RequiredBlocks: []string{"do"},
		Inputs: map[string]engine.InputMeta{
			"do": {Description: "Code block to execute"},
		},
	})

	// ==========================================
	// SLOT: BREAK & CONTINUE (Support Conditional: break: $i == 5)
	// ==========================================
	eng.Register("break", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		if node.Value != nil {
			expr := coerce.ToString(node.Value)
			if !evalSimpleCondition(expr, scope) {
				return nil // Don't break if condition not met
			}
		}
		return ErrBreak
	}, engine.SlotMeta{Description: "Force stop. Supports conditional: `break: $i == 5`"})

	eng.Register("continue", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		if node.Value != nil {
			expr := coerce.ToString(node.Value)
			if !evalSimpleCondition(expr, scope) {
				return nil // Don't continue if condition not met
			}
		}
		return ErrContinue
	}, engine.SlotMeta{Description: "Continue to next iteration. Supports conditional: `continue: $i % 2 == 0`"})

	// ==========================================
	// SLOT: DD & DUMP (Laravel Style)
	// ==========================================
	eng.Register("dump", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		val := resolveValue(node.Value, scope)
		fmt.Printf("[DUMP]: %+v (Type: %T)\n", val, val)
		return nil
	}, engine.SlotMeta{Description: "Dump variable to console without stopping execution."})

	eng.Register("dd", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		val := resolveValue(node.Value, scope)
		fmt.Printf("[DD]: %+v (Type: %T)\n", val, val)
		return fmt.Errorf("DD HALT: execution stopped by dd")
	}, engine.SlotMeta{Description: "Dump and Die. Display variable content and stop script immediately."})

	// ==========================================
	// SLOT: SWITCH
	// ==========================================
	eng.Register("switch", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		val := resolveValue(node.Value, scope)

		for _, child := range node.Children {
			if child.Name == "case" {
				caseVal := resolveValue(child.Value, scope)
				if val == caseVal {
					err := eng.Execute(ctx, child, scope)
					if err != nil && (errors.Is(err, ErrBreak) || strings.Contains(err.Error(), "break")) {
						return nil
					}
					return err
				}
			} else if child.Name == "default" {
				err := eng.Execute(ctx, child, scope)
				if err != nil && (errors.Is(err, ErrBreak) || strings.Contains(err.Error(), "break")) {
					return nil
				}
				return err
			}
		}
		return nil
	}, engine.SlotMeta{
		Description: "Conditional branching (Switch Case).",
		Inputs: map[string]engine.InputMeta{
			"case":    {Description: "Case value to check"},
			"default": {Description: "Default block if no case matches"},
		},
	})

	// ==========================================
	// SLOT: ISSET, EMPTY, UNLESS
	// ==========================================
	eng.Register("isset", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		val := resolveValue(node.Value, scope)
		if val != nil {
			for _, child := range node.Children {
				if child.Name == "do" {
					return eng.Execute(ctx, child, scope)
				}
			}
		}
		return nil
	}, engine.SlotMeta{
		Description: "Execute block if variable is set/defined.",
		Inputs: map[string]engine.InputMeta{
			"do": {Description: "Code block to execute"},
		},
	})

	eng.Register("empty", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		val := resolveValue(node.Value, scope)
		isEmpty := false
		if val == nil {
			isEmpty = true
		} else {
			str := coerce.ToString(val)
			if str == "" {
				isEmpty = true
			} else {
				switch v := val.(type) {
				case []interface{}:
					isEmpty = len(v) == 0
				case []string:
					isEmpty = len(v) == 0
				case map[string]interface{}:
					isEmpty = len(v) == 0
				}
			}
		}

		if isEmpty {
			for _, child := range node.Children {
				if child.Name == "do" {
					return eng.Execute(ctx, child, scope)
				}
			}
		}
		return nil
	}, engine.SlotMeta{
		Description: "Execute block if variable is empty (null, '', or empty array).",
		Inputs: map[string]engine.InputMeta{
			"do": {Description: "Code block to execute"},
		},
	})

	eng.Register("unless", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		val := resolveValue(node.Value, scope)
		boolVal, _ := coerce.ToBool(val)
		if !boolVal {
			for _, child := range node.Children {
				if child.Name == "do" {
					return eng.Execute(ctx, child, scope)
				}
			}
		}
		return nil
	}, engine.SlotMeta{
		Description: "Reverse of IF. Execute block if condition is FALSE.",
		Inputs: map[string]engine.InputMeta{
			"do": {Description: "Code block to execute"},
		},
	})

	// ==========================================
	// SLOT: AUTH & GUEST (Check Login)
	// ==========================================
	eng.Register("auth", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		user, found := scope.Get("user")
		isAuth := found && user != nil
		if !isAuth {
			authObjRaw, foundAuth := scope.Get("auth")
			if foundAuth && authObjRaw != nil {
				isAuth = true
			}
		}

		if isAuth {
			for _, child := range node.Children {
				if child.Name == "do" {
					return eng.Execute(ctx, child, scope)
				}
			}
		}
		return nil
	}, engine.SlotMeta{
		Description: "Execute block if user is logged in.",
		Inputs: map[string]engine.InputMeta{
			"do": {Description: "Code block to execute"},
		},
	})

	eng.Register("guest", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		user, found := scope.Get("user")
		isAuth := found && user != nil
		if !isAuth {
			authObjRaw, foundAuth := scope.Get("auth")
			if foundAuth && authObjRaw != nil {
				isAuth = true
			}
		}

		if !isAuth {
			for _, child := range node.Children {
				if child.Name == "do" {
					return eng.Execute(ctx, child, scope)
				}
			}
		}
		return nil
	}, engine.SlotMeta{
		Description: "Execute block if user is NOT logged in (guest).",
		Inputs: map[string]engine.InputMeta{
			"do": {Description: "Code block to execute"},
		},
	})

	eng.Register("auth.user", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		target := "user"
		if node.Value != nil {
			target = strings.TrimPrefix(coerce.ToString(node.Value), "$")
		}
		for _, c := range node.Children {
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}
		user, found := scope.Get("user")
		if !found || user == nil {
			authObj, ok := scope.Get("auth")
			if ok && authObj != nil {
				user = authObj
			}
		}
		scope.Set(target, user)
		return nil
	}, engine.SlotMeta{Description: "Retrieve current logged-in user data."})

	eng.Register("auth.check", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		target := "is_auth"
		if node.Value != nil {
			target = strings.TrimPrefix(coerce.ToString(node.Value), "$")
		}
		for _, c := range node.Children {
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}
		user, found := scope.Get("user")
		isAuth := found && user != nil
		if !isAuth {
			authObj, ok := scope.Get("auth")
			isAuth = ok && authObj != nil
		}

		// Fallback: Check and VERIFY cookies if not in scope
		if !isAuth {
			reqVal := ctx.Value("httpRequest")
			if req, ok := reqVal.(*http.Request); ok {
				var tokenString string
				if cookie, err := req.Cookie("auth_token"); err == nil {
					tokenString = cookie.Value
				} else if cookie, err := req.Cookie("token"); err == nil {
					tokenString = cookie.Value
				}

				// VERIFY the token, not just check existence
				if tokenString != "" {
					jwtSecret := os.Getenv("JWT_SECRET")
					if jwtSecret == "" {
						jwtSecret = "ZENOLANG_DEMO_SECRET_KEY_2026"
					}

					token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
						return []byte(jwtSecret), nil
					})

					// Only set isAuth to true if token is VALID
					if err == nil && token.Valid {
						isAuth = true
					}
				}
			}
		}
		scope.Set(target, isAuth)
		return nil
	}, engine.SlotMeta{Description: "Check if user is logged in (returns boolean)."})

	// ==========================================
	// SLOT: CAN & CANNOT (RBAC)
	// ==========================================
	eng.Register("can", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		ability := coerce.ToString(node.Value)
		canFunc, ok := scope.Get("can")
		if !ok {
			return nil
		}

		var resource interface{}
		for _, child := range node.Children {
			if child.Name == "resource" {
				varName := strings.TrimPrefix(coerce.ToString(child.Value), "$")
				resource, _ = scope.Get(varName)
				break
			}
		}

		callback, ok := canFunc.(func(string, interface{}) bool)
		if !ok {
			return nil
		}

		if callback(ability, resource) {
			for _, child := range node.Children {
				if child.Name == "do" {
					return eng.Execute(ctx, child, scope)
				}
			}
		}
		return nil
	}, engine.SlotMeta{
		Description: "Execute block if user has specific permission (ability).",
		Inputs: map[string]engine.InputMeta{
			"resource": {Description: "Object to check permission for"},
			"do":       {Description: "Code block to execute"},
		},
	})

	eng.Register("cannot", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		ability := coerce.ToString(node.Value)
		canFunc, ok := scope.Get("can")

		var resource interface{}
		for _, child := range node.Children {
			if child.Name == "resource" {
				varName := strings.TrimPrefix(coerce.ToString(child.Value), "$")
				resource, _ = scope.Get(varName)
				break
			}
		}

		if !ok {
			// No callback, assume cannot
			for _, child := range node.Children {
				if child.Name == "do" {
					return eng.Execute(ctx, child, scope)
				}
			}
			return nil
		}

		callback, ok := canFunc.(func(string, interface{}) bool)
		if !ok {
			return nil
		}

		if !callback(ability, resource) {
			for _, child := range node.Children {
				if child.Name == "do" {
					return eng.Execute(ctx, child, scope)
				}
			}
		}
		return nil
	}, engine.SlotMeta{
		Description: "Menjalankan blok jika user TIDAK memiliki izin (ability).",
		Inputs: map[string]engine.InputMeta{
			"resource": {Description: "Objek yang dicek izinnya"},
			"do":       {Description: "Blok kode yang dijalankan"},
		},
	})

	// ==========================================
	// SLOT: JSON OUTPUT
	// ==========================================
	eng.Register("json", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		val := resolveValue(node.Value, scope)
		w, ok := ctx.Value("httpWriter").(http.ResponseWriter)
		if !ok {
			return nil
		}
		bytes, err := json.Marshal(val)
		if err != nil {
			return err
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
		return nil
	}, engine.SlotMeta{Description: "Mengeluarkan nilai sebagai JSON langsung ke HTTP response."})

	// ==========================================
	// SLOT: FORELSE
	// ==========================================
	eng.Register("forelse", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		val := resolveValue(node.Value, scope)
		list, _ := coerce.ToSlice(val)

		if len(list) == 0 {
			for _, child := range node.Children {
				if child.Name == "empty" {
					return eng.Execute(ctx, child, scope)
				}
			}
		} else {
			// Synthesize a "for" node
			forNode := &engine.Node{Name: "for", Value: node.Value}
			for _, child := range node.Children {
				if child.Name == "as" || child.Name == "do" {
					forNode.Children = append(forNode.Children, child)
				}
			}
			return eng.Execute(ctx, forNode, scope)
		}
		return nil
	}, engine.SlotMeta{
		Description: "Perulangan list dengan blok cadangan jika list kosong.",
		Inputs: map[string]engine.InputMeta{
			"as":    {Description: "Alias variabel item"},
			"do":    {Description: "Blok yang diulang"},
			"empty": {Description: "Blok jika data kosong"},
		},
	})

	// ==========================================
	// SLOT: VALIDATION ERROR
	// ==========================================
	eng.Register("error", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		fieldName := coerce.ToString(node.Value)
		errorsRaw, ok := scope.Get("errors")
		if !ok {
			return nil
		}
		errs, ok := errorsRaw.(map[string][]string)
		if !ok {
			return nil
		}
		msgs, exists := errs[fieldName]
		if !exists || len(msgs) == 0 {
			return nil
		}

		scope.Set("message", msgs[0])
		for _, child := range node.Children {
			if child.Name == "do" {
				return eng.Execute(ctx, child, scope)
			}
		}
		return nil
	}, engine.SlotMeta{
		Description: "Menampilkan pesan error validasi untuk field tertentu.",
		Inputs: map[string]engine.InputMeta{
			"do": {Description: "Blok kode yang dijalankan"},
		},
	})
}

func evalSimpleCondition(expr string, scope *engine.Scope) bool {
	var op string
	var left, right string

	ops := []string{"<=", ">=", "==", "!=", "<", ">"}
	for _, o := range ops {
		if strings.Contains(expr, o) {
			op = o
			parts := strings.SplitN(expr, o, 2)
			left = strings.TrimSpace(parts[0])
			right = strings.TrimSpace(parts[1])
			break
		}
	}

	if op == "" {
		return false
	}

	leftVal := resolveExpressionValue(left, scope)
	rightVal := resolveExpressionValue(right, scope)

	lInt, errL := coerce.ToInt(leftVal)
	rInt, errR := coerce.ToInt(rightVal)

	isNumeric := (errL == nil && errR == nil)

	switch op {
	case "<":
		if isNumeric {
			return lInt < rInt
		}
		return coerce.ToString(leftVal) < coerce.ToString(rightVal)
	case ">":
		if isNumeric {
			return lInt > rInt
		}
		return coerce.ToString(leftVal) > coerce.ToString(rightVal)
	case "<=":
		if isNumeric {
			return lInt <= rInt
		}
		return coerce.ToString(leftVal) <= coerce.ToString(rightVal)
	case ">=":
		if isNumeric {
			return lInt >= rInt
		}
		return coerce.ToString(leftVal) >= coerce.ToString(rightVal)
	case "==":
		if isNumeric {
			return lInt == rInt
		}
		return coerce.ToString(leftVal) == coerce.ToString(rightVal)
	case "!=":
		if isNumeric {
			return lInt != rInt
		}
		return coerce.ToString(leftVal) != coerce.ToString(rightVal)
	}
	return false
}

func resolveExpressionValue(s string, scope *engine.Scope) interface{} {
	if strings.HasPrefix(s, "$") {
		v, _ := scope.Get(strings.TrimPrefix(s, "$"))
		return v
	}
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return s
}
