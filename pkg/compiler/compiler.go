package compiler

import (
	"fmt"
	"strconv"
	"strings"
	"zeno/pkg/engine"
	"zeno/pkg/engine/vm"
	"zeno/pkg/utils/coerce"
)

type Local struct {
	Name  string
	Depth int
}

type Compiler struct {
	chunk      *vm.Chunk
	locals     []Local
	scopeDepth int
}

func NewCompiler() *Compiler {
	return &Compiler{
		chunk: &vm.Chunk{
			Code:      []byte{},
			Constants: []vm.Value{},
		},
		locals:     []Local{},
		scopeDepth: 0,
	}
}

func NewFunctionCompiler() *Compiler {
	return NewCompiler()
}

func (c *Compiler) Compile(node *engine.Node) (*vm.Chunk, error) {
	err := c.compileNode(node)
	if err != nil {
		return nil, err
	}
	// Implicit return nil for main chunk
	c.emitByte(byte(vm.OpNil))
	c.emitByte(byte(vm.OpReturn))

	// Transfer local names for VM sync
	c.chunk.LocalNames = make([]string, len(c.locals))
	for i, l := range c.locals {
		c.chunk.LocalNames[i] = l.Name
	}

	return c.chunk, nil
}

func (c *Compiler) compileNode(node *engine.Node) error {
	// Simple Arithmetic Example: "1 + 2" atau "$x: 10"
	if node.Name == "stop" {
		c.emitByte(byte(vm.OpStop))
		return nil
	}

	// If it's a	// Variable Assignment: $x: 10
	if strings.HasPrefix(node.Name, "$") {
		varName := node.Name[1:]
		// Evaluate value
		if node.Value != nil {
			if s, ok := node.Value.(string); ok && strings.Contains(s, " ") && !strings.HasPrefix(s, "\x00") {
				// Likely an expression
				if err := c.compileExpression(s); err != nil {
					return err
				}
			} else {
				if err := c.compileValue(node.Value); err != nil {
					return err
				}
			}
		} else if len(node.Children) > 0 {
			// [FIX] Check if it's a Function Call Assignment ($res: call: foo)
			if len(node.Children) == 1 && node.Children[0].Name == "call" {
				// Compile the call instruction. Results will be on stack.
				if err := c.compileNode(node.Children[0]); err != nil {
					return err
				}
			} else {
				// Map literal
				for _, child := range node.Children {
					// Push key
					c.emitConstantOperand(vm.OpConstant, c.addConstant(vm.NewString(child.Name)))
					// Push value (Force evaluation as data, not instruction)
					if err := c.compileNodeAsValue(child); err != nil {
						return err
					}
				}
				c.emitByte(byte(vm.OpMakeMap))
				c.emitByte(byte(len(node.Children)))
			}
		}

		idx := c.resolveLocal(varName)
		if idx == -1 {
			// Automatic local allocation for $ variables
			c.locals = append(c.locals, Local{Name: varName, Depth: c.scopeDepth})
			idx = len(c.locals) - 1
		}

		c.emitByte(byte(vm.OpSetLocal))
		c.emitByte(byte(idx))
		c.emitByte(byte(vm.OpPop))
		return nil
	}

	// [NEW] Anonymous node or Map literal if not assignment
	if node.Name == "" && len(node.Children) > 0 {
		for _, child := range node.Children {
			// Push key
			c.emitConstantOperand(vm.OpConstant, c.addConstant(vm.NewString(child.Name)))
			// Push value
			if err := c.compileNodeAsValue(child); err != nil {
				return err
			}
		}
		c.emitByte(byte(vm.OpMakeMap))
		c.emitByte(byte(len(node.Children)))
		return nil
	}

	// [NEW] Function Definition: fn: myFunc
	if node.Name == "fn" {
		funcName := coerce.ToString(node.Value)
		// ALLOW ANONYMOUS: if funcName == "" { ... }

		// Create separate compiler for function body
		fnCompiler := NewCompiler()
		fnCompiler.scopeDepth = c.scopeDepth + 1

		// Compile Body (Children)
		for _, child := range node.Children {
			// Special handling for 'params' node to define argument order?
			// For now, assume implicit locals or no args.
			// Actually, let's assume standard execution flow.
			err := fnCompiler.compileNode(child)
			if err != nil {
				return err
			}
		}
		// Finalize body with Implicit Return Nil
		fnCompiler.emitByte(byte(vm.OpNil))
		fnCompiler.emitByte(byte(vm.OpReturn))

		// Sync locals
		fnCompiler.chunk.LocalNames = make([]string, len(fnCompiler.locals))
		for i, l := range fnCompiler.locals {
			fnCompiler.chunk.LocalNames[i] = l.Name
		}

		// Emit Function Constant in CURRENT chunk
		fnChunk := fnCompiler.chunk
		c.emitConstantOperand(vm.OpConstant, c.addConstant(vm.NewFunction(fnChunk)))

		// Store in Global Variable ONLY if named
		if funcName != "" {
			c.emitConstantOperand(vm.OpSetGlobal, c.addConstant(vm.NewString(funcName)))
		}

		return nil
	}

	// [NEW] Return Statement: return value
	if node.Name == "return" {
		if node.Value != nil {
			// Compile value (e.g., return 10 or return "foo")
			// Use compileExpression to handle expressions like return 1 + 2
			s := coerce.ToString(node.Value)
			if err := c.compileExpression(s); err != nil {
				return err
			}
		} else if len(node.Children) > 0 {
			// Handle return with children as expression (e.g. return { ... })
			// Use compileNodeAsValue logic on first child?
			// Simplification: Assume single child expression for now
			if err := c.compileNodeAsValue(node.Children[0]); err != nil {
				return err
			}
		} else {
			// return (void) -> return nil
			c.emitByte(byte(vm.OpNil))
		}
		c.emitByte(byte(vm.OpReturn))
		return nil
	}

	// [NEW] Function Call: call: myFunc
	if node.Name == "call" {
		funcName := coerce.ToString(node.Value)
		if funcName == "" {
			return fmt.Errorf("call node must have a function name value")
		}

		// 1. Push Function (Get from Global)
		c.emitConstantOperand(vm.OpGetGlobal, c.addConstant(vm.NewString(funcName)))

		// 2. Compile arguments (Children)
		argCount := 0
		for _, child := range node.Children {
			// Compile child value and push to stack
			// Supports: arg: 10 or just raw values
			if err := c.compileNodeAsValue(child); err != nil {
				return err
			}
			argCount++
		}

		// 3. Emit Call
		c.emitByte(byte(vm.OpCall))
		c.emitByte(byte(argCount))

		return nil
	}

	// If it's an expression like "1 + 2" or "string" + "concat"
	if node.Name == "" && node.Value != nil {
		valStr := fmt.Sprintf("%v", node.Value)
		// Check if it looks like an expression (contains operators)
		// Or just delegate to compileExpression which handles simple values too.
		// However, compileExpression does trim space, which we want for operators but
		// compileValue (called by compileExpression) now handles quotes intelligently.
		if err := c.compileExpression(valStr); err != nil {
			return err
		}
		return nil
	}

	// If it's an "if" statement
	if node.Name == "if" {
		// 1. Compile Expression (Condition)
		if err := c.compileExpression(coerce.ToString(node.Value)); err != nil {
			return err
		}

		// 2. Jump if False to Else or End
		jumpIfFalsePos := c.emitJump(vm.OpJumpIfFalse)

		// 3. Compile "then" block
		for _, child := range node.Children {
			if child.Name == "then" {
				for _, subchild := range child.Children {
					if err := c.compileNode(subchild); err != nil {
						return err
					}
				}
				break
			}
		}

		// 4. Jump over "else" block (if exists)
		jumpOverElsePos := c.emitJump(vm.OpJump)

		// 5. Backpatch JumpIfFalse
		c.patchJump(jumpIfFalsePos)

		// 6. Compile "else" block (if exists)
		for _, child := range node.Children {
			if child.Name == "else" {
				for _, subchild := range child.Children {
					if err := c.compileNode(subchild); err != nil {
						return err
					}
				}
				break
			}
		}

		// 7. Backpatch JumpOverElse
		c.patchJump(jumpOverElsePos)
		return nil
	}

	// Optimization: If node has a list value and children, it's likely an iteration
	if node.Value != nil && node.Name != "" && !strings.HasPrefix(node.Name, "$") && node.Name != "then" && node.Name != "else" {
		if slice, ok := node.Value.([]interface{}); ok && len(node.Children) > 0 {
			// Reserve Hidden Index and Iterable as Locals to avoid collisions
			indexLocalIdx := len(c.locals)
			c.locals = append(c.locals, Local{Name: fmt.Sprintf("iter_idx_%d", indexLocalIdx), Depth: c.scopeDepth})
			listLocalIdx := len(c.locals)
			c.locals = append(c.locals, Local{Name: fmt.Sprintf("iter_list_%d", listLocalIdx), Depth: c.scopeDepth})

			// 1. Push Initial Hidden Index (0)
			c.emitConstantOperand(vm.OpConstant, c.addConstant(vm.NewNumber(0)))

			// 2. Push Iterable
			c.emitConstantOperand(vm.OpConstant, c.addConstant(vm.NewValue(slice)))

			// 3. Mark Loop Start
			loopStart := len(c.chunk.Code)

			// 4. OpIterNext (Jumps to End if done)
			// OpIterNext expects [iterable] at peek(0) and [index] at peek(1) relative to CURRENT SP
			exitJump := c.emitJump(vm.OpIterNext)

			// 5. Loop Body
			c.emitByte(byte(vm.OpPop)) // Pop the boolean 'true' from IterNext

			// The 'item' is now at the top of the stack (pushed by OpIterNext).
			// We assign it to a local named 'item' so we can resolve it.
			itemIdx := c.resolveLocal("item")
			if itemIdx == -1 {
				c.locals = append(c.locals, Local{Name: "item", Depth: c.scopeDepth})
				itemIdx = len(c.locals) - 1
			}
			// Important: OpSetLocal DOES NOT pop.
			c.emitByte(byte(vm.OpSetLocal))
			c.emitByte(byte(itemIdx))

			for _, child := range node.Children {
				if err := c.compileNode(child); err != nil {
					return err
				}
			}

			// Pop the item before looping back to maintain stack balance
			c.emitByte(byte(vm.OpPop))

			// 6. Loop back
			c.emitByte(byte(vm.OpLoop))
			offset := len(c.chunk.Code) - loopStart + 2
			c.emitByte(byte((offset >> 8) & 0xff))
			c.emitByte(byte(offset & 0xff))

			// 7. Cleanup and Exit
			c.patchJump(exitJump)
			c.emitByte(byte(vm.OpIterEnd))
			return nil
		}
	}

	// [NEW] ZIP Iteration: items, tags: [[1, 2], ["a", "b"]]
	// For now, we handle single list iteration perfectly. Zip will be refined later.

	// If it's a regular slot call (e.g., http.response)
	if node.Name != "" && !strings.HasPrefix(node.Name, "$") && node.Name != "root" && node.Name != "do" && node.Name != "else" && node.Name != "then" && node.Name != "try" && node.Name != "catch" {
		argCount := len(node.Children)

		// [NEW] Handle implicit value (log: "msg")
		if node.Value != nil {
			// Push implicit key
			c.emitConstantOperand(vm.OpConstant, c.addConstant(vm.NewString("__value__")))
			// Push value
			// Compile value logic handles string/number/etc
			if err := c.compileValue(node.Value); err != nil {
				return err
			}
			argCount++
		}

		// Compile children as named arguments
		for _, child := range node.Children {
			// Push Name
			c.emitConstantOperand(vm.OpConstant, c.addConstant(vm.NewString(child.Name)))
			// Push Value
			if err := c.compileNodeAsValue(child); err != nil {
				return err
			}
		}

		c.emitCallSlot(node.Name, argCount)
		return nil
	}

	// Default: Compile children (Container node)
	for i := 0; i < len(node.Children); i++ {
		child := node.Children[i]

		// Check for try/catch sequence
		if child.Name == "try" {
			var catchNode *engine.Node
			if i+1 < len(node.Children) && node.Children[i+1].Name == "catch" {
				catchNode = node.Children[i+1]
			}

			if catchNode != nil {
				// Compile Try+Catch sequence

				// 1. Emit OpTry
				jumpToCatch := c.emitJump(vm.OpTry)

				// 2. Compile Try Body (the 'try' node itself is a container for body)
				// We need to compile ITS children.
				// We cannot call compileNode(child) because compileNode handles "container" logic
				// but here we are IN the container logic of the PARENT.
				// If we call compileNode(child), it will run the loop inside `try`.
				// That is correct. `try` is just a container of statements.
				if err := c.compileNode(child); err != nil {
					return err
				}

				// 3. Emit OpEndTry (Success path)
				c.emitByte(byte(vm.OpEndTry))

				// 4. Jump Over Catch (Success path)
				jumpOverCatch := c.emitJump(vm.OpJump)

				// 5. Patch OpTry (Target = Start of Catch)
				c.patchJump(jumpToCatch)

				// 6. Compile Catch Block
				// Make sure to handle error variable
				varName := ""
				if catchNode.Value != nil {
					v := coerce.ToString(catchNode.Value)
					v = strings.TrimSpace(v)
					if strings.HasPrefix(v, "$") {
						varName = v[1:]
					}
				}

				if varName != "" {
					idx := c.resolveLocal(varName)
					if idx == -1 {
						c.locals = append(c.locals, Local{Name: varName, Depth: c.scopeDepth})
						idx = len(c.locals) - 1
					}
					c.emitByte(byte(vm.OpSetLocal))
					c.emitByte(byte(idx))
					c.emitByte(byte(vm.OpPop))
				} else {
					c.emitByte(byte(vm.OpPop)) // Consume error
				}

				// Compile catch body
				if err := c.compileNode(catchNode); err != nil {
					return err
				}

				// 7. Patch JumpOverCatch
				c.patchJump(jumpOverCatch)

				i++ // Skip catch node
				continue
			}
		}

		if err := c.compileNode(child); err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiler) compileExpression(expr string) error {
	expr = strings.TrimSpace(expr)

	// Strip outer parentheses: ($x == 10) -> $x == 10
	for strings.HasPrefix(expr, "(") && strings.HasSuffix(expr, ")") {
		expr = strings.TrimSpace(expr[1 : len(expr)-1])
	}

	if expr == "" {
		return nil
	}

	// [NEW] Handle Logical NOT: "!$var"
	if strings.HasPrefix(expr, "!") {
		inner := strings.TrimSpace(expr[1:])
		if err := c.compileExpression(inner); err != nil {
			return err
		}
		c.emitByte(byte(vm.OpLogicalNot))
		return nil
	}

	// Multi-part expressions (Binary ops)
	// Order determines precedence (Lowest binding power first to split at top level)
	ops := []string{"||", "&&", "==", "!=", ">=", "<=", ">", "<", "+", "-", "*", "/"}
	for _, op := range ops {
		leftStr, rightStr, found := splitExpression(expr, op)
		if found {
			if err := c.compileExpression(leftStr); err != nil {
				return err
			}
			if err := c.compileExpression(rightStr); err != nil {
				return err
			}

			switch op {
			case "||":
				c.emitByte(byte(vm.OpLogicalOr))
			case "&&":
				c.emitByte(byte(vm.OpLogicalAnd))
			case "==":
				c.emitByte(byte(vm.OpEqual))
			case "!=":
				c.emitByte(byte(vm.OpNotEqual))
			case ">=":
				c.emitByte(byte(vm.OpGreaterEqual))
			case "<=":
				c.emitByte(byte(vm.OpLessEqual))
			case ">":
				c.emitByte(byte(vm.OpGreater))
			case "<":
				c.emitByte(byte(vm.OpLess))
			case "+":
				c.emitByte(byte(vm.OpAdd))
			case "-":
				c.emitByte(byte(vm.OpSubtract))
			case "*":
				c.emitByte(byte(vm.OpMultiply))
			case "/":
				c.emitByte(byte(vm.OpDivide))
			}
			return nil
		}
	}

	// Default: Single value truthiness
	return c.compileValue(expr)
}

func (c *Compiler) emitJump(op vm.OpCode) int {
	c.emitByte(byte(op))
	c.emitByte(0xff) // Placeholder for 16-bit offset
	c.emitByte(0xff)
	return len(c.chunk.Code) - 2
}

func (c *Compiler) patchJump(pos int) {
	// Calculate offset from instruction after jump to current end
	offset := len(c.chunk.Code) - pos - 2
	c.chunk.Code[pos] = byte((offset >> 8) & 0xff)
	c.chunk.Code[pos+1] = byte(offset & 0xff)
}

func (c *Compiler) resolveLocal(name string) int {
	for i := len(c.locals) - 1; i >= 0; i-- {
		if c.locals[i].Name == name {
			return i
		}
	}
	return -1
}

func (c *Compiler) compileValue(v interface{}) error {
	if v == nil {
		c.emitByte(byte(vm.OpNil))
		return nil
	}
	if b, ok := v.(bool); ok {
		if b {
			c.emitByte(byte(vm.OpTrue))
		} else {
			c.emitByte(byte(vm.OpFalse))
		}
		return nil
	}
	if n, ok := v.(float64); ok {
		c.emitConstantOperand(vm.OpConstant, c.addConstant(vm.NewNumber(n)))
		return nil
	}
	if n, ok := v.(int); ok {
		c.emitConstantOperand(vm.OpConstant, c.addConstant(vm.NewNumber(float64(n))))
		return nil
	}

	if s, ok := v.(string); ok {
		// [NEW] Raw String Literal support (flagged by Parser with \x00 prefix)
		if strings.HasPrefix(s, "\x00") {
			s = s[1:] // Strip prefix, preserve EVERYTHING else (including spaces)
		} else {
			// Check if it looks like a string literal first BEFORE trimming
			// This preserves " hello "
			args := strings.TrimSpace(s)
			if len(args) >= 2 && ((args[0] == '"' && args[len(args)-1] == '"') || (args[0] == '\'' && args[len(args)-1] == '\'')) {
				s = args
			} else {
				// Normal expression: Trim spaces
				s = strings.TrimSpace(s)
			}

			if !strings.HasPrefix(s, "[") || !strings.HasSuffix(s, "]") {
				s = strings.TrimSuffix(s, ",")
				// Only trim again if it's not a quoted string
				if !((strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"")) || (strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'"))) {
					s = strings.TrimSpace(s)
				}
			}
		}

		// Variable reference?
		if strings.HasPrefix(s, "$") {
			path := strings.Split(s[1:], ".")
			rootName := path[0]

			if idx := c.resolveLocal(rootName); idx != -1 {
				c.emitByte(byte(vm.OpGetLocal))
				c.emitByte(byte(idx))
			} else {
				c.emitConstantOperand(vm.OpGetGlobal, c.addConstant(vm.NewString(rootName)))
			}

			// Handle nested properties: .0, .name, etc.
			for i := 1; i < len(path); i++ {
				c.emitConstantOperand(vm.OpAccessProperty, c.addConstant(vm.NewString(path[i])))
			}
			return nil
		}

		// [NEW] List Literal: ["a", "b"]
		if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
			content := s[1 : len(s)-1]
			if content == "" {
				c.emitByte(byte(vm.OpMakeList))
				c.emitByte(0)
				return nil
			}
			parts := strings.Split(content, ",")
			for _, p := range parts {
				if err := c.compileValue(strings.TrimSpace(p)); err != nil {
					return err
				}
			}
			c.emitByte(byte(vm.OpMakeList))
			c.emitByte(byte(len(parts)))
			return nil
		}

		// [NEW] Boolean and Nil Literals
		if s == "true" {
			c.emitByte(byte(vm.OpTrue))
			return nil
		}
		if s == "false" {
			c.emitByte(byte(vm.OpFalse))
			return nil
		}
		if s == "nil" || s == "null" {
			c.emitByte(byte(vm.OpNil))
			return nil
		}

		// Number?
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			c.emitConstantOperand(vm.OpConstant, c.addConstant(vm.NewNumber(f)))
			return nil
		}
		// String literal (Strip quotes)
		if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) {
			s = s[1 : len(s)-1]
		}
		c.emitConstantOperand(vm.OpConstant, c.addConstant(vm.NewString(s)))
		return nil
	}
	// Fallback raw values
	c.emitConstantOperand(vm.OpConstant, c.addConstant(vm.NewValue(v)))
	return nil
}

func (c *Compiler) emitByte(b byte) {
	c.chunk.Code = append(c.chunk.Code, b)
}

func (c *Compiler) compileNodeAsValue(node *engine.Node) error {
	if node.Name == "fn" {
		return c.compileNode(node)
	}
	if node.Value != nil {
		return c.compileValue(node.Value)
	}
	if len(node.Children) > 0 {
		for _, child := range node.Children {
			c.emitConstantOperand(vm.OpConstant, c.addConstant(vm.NewString(child.Name)))
			if err := c.compileNodeAsValue(child); err != nil {
				return err
			}
		}
		c.emitByte(byte(vm.OpMakeMap))
		c.emitByte(byte(len(node.Children)))
		return nil
	}
	c.emitByte(byte(vm.OpNil))
	return nil
}

// splitExpression splits the expression s by the first occurrence of op,
// ignoring operators that are inside quotes (single or double).
// Returns left part, right part, and true if found.
func splitExpression(s, op string) (string, string, bool) {
	inQuote := false
	var quoteChar rune
	lastIdx := -1

	for i := 0; i < len(s); i++ {
		char := rune(s[i])
		if char == '"' || char == '\'' {
			if !inQuote {
				inQuote = true
				quoteChar = char
			} else if char == quoteChar {
				// Check for escaped quote?
				if i > 0 && s[i-1] == '\\' {
					// escaped
				} else {
					inQuote = false
				}
			}
		}

		if !inQuote {
			// Check if op matches here
			if strings.HasPrefix(s[i:], op) {
				lastIdx = i
			}
		}
	}

	if lastIdx != -1 {
		return s[:lastIdx], s[lastIdx+len(op):], true
	}

	return "", "", false
}

func (c *Compiler) addConstant(val vm.Value) int {
	c.chunk.Constants = append(c.chunk.Constants, val)
	return len(c.chunk.Constants) - 1
}

// emitConstantOperand automatically chooses between Short and Long opcodes
func (c *Compiler) emitConstantOperand(op vm.OpCode, idx int) {
	if idx > 255 {
		// Use Long variant
		switch op {
		case vm.OpConstant:
			c.emitByte(byte(vm.OpConstantLong))
		case vm.OpGetGlobal:
			c.emitByte(byte(vm.OpGetGlobalLong))
		case vm.OpSetGlobal:
			c.emitByte(byte(vm.OpSetGlobalLong))
		case vm.OpCallSlot:
			// Expects explicit handling in emitCallSlot, but if called here:
			// This case might be invalid as OpCallSlot takes explicit ArgCount?
			// But if used generically:
			c.emitByte(byte(vm.OpCallSlotLong))
		case vm.OpAccessProperty:
			c.emitByte(byte(vm.OpAccessPropertyLong))
		default:
			// Fallback (or panic)
			panic(fmt.Sprintf("No Long variant for opcode %v", op))
		}
		// Emit 16-bit index (Big Endian)
		c.emitByte(byte((idx >> 8) & 0xff))
		c.emitByte(byte(idx & 0xff))
	} else {
		// Use Short variant
		c.emitByte(byte(op))
		c.emitByte(byte(idx))
	}
}

func (c *Compiler) emitCallSlot(name string, argCount int) {
	idx := c.addConstant(vm.NewString(name))
	if idx > 255 {
		c.emitByte(byte(vm.OpCallSlotLong))
		c.emitByte(byte((idx >> 8) & 0xff))
		c.emitByte(byte(idx & 0xff))
	} else {
		c.emitByte(byte(vm.OpCallSlot))
		c.emitByte(byte(idx))
	}
	c.emitByte(byte(argCount))
}
