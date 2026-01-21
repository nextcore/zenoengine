package vm

import (
	"fmt"
	"strconv"
	"strings"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

type Local struct {
	Name  string
	Depth int
}

type Compiler struct {
	chunk      *Chunk
	locals     []Local
	scopeDepth int
}

func NewCompiler() *Compiler {
	return &Compiler{
		chunk: &Chunk{
			Code:      []byte{},
			Constants: []Value{},
		},
		locals:     []Local{},
		scopeDepth: 0,
	}
}

func NewFunctionCompiler() *Compiler {
	return NewCompiler()
}

func (c *Compiler) Compile(node *engine.Node) (*Chunk, error) {
	err := c.compileNode(node)
	if err != nil {
		return nil, err
	}
	c.emitByte(byte(OpReturn))

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
		c.emitByte(byte(OpStop))
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
			// Map literal
			for _, child := range node.Children {
				// Push key
				c.emitByte(byte(OpConstant))
				c.emitByte(c.addConstant(NewString(child.Name)))
				// Push value (Force evaluation as data, not instruction)
				if err := c.compileNodeAsValue(child); err != nil {
					return err
				}
			}
			c.emitByte(byte(OpMakeMap))
			c.emitByte(byte(len(node.Children)))
		}

		idx := c.resolveLocal(varName)
		if idx == -1 {
			// Automatic local allocation for $ variables
			c.locals = append(c.locals, Local{Name: varName, Depth: c.scopeDepth})
			idx = len(c.locals) - 1
		}

		c.emitByte(byte(OpSetLocal))
		c.emitByte(byte(idx))
		c.emitByte(byte(OpPop))
		return nil
	}

	// [NEW] Anonymous node or Map literal if not assignment
	if node.Name == "" && len(node.Children) > 0 {
		for _, child := range node.Children {
			// Push key
			c.emitByte(byte(OpConstant))
			c.emitByte(c.addConstant(NewString(child.Name)))
			// Push value
			if err := c.compileNodeAsValue(child); err != nil {
				return err
			}
		}
		c.emitByte(byte(OpMakeMap))
		c.emitByte(byte(len(node.Children)))
		return nil
	}

	// Function Definition?
	if node.Name == "fn" {
		funcName := coerce.ToString(node.Value)
		// Compile children as a separate function chunk
		subCompiler := NewFunctionCompiler()
		// Function arguments handling (optional for now, ZenoLang uses dynamic scope)
		for _, child := range node.Children {
			if err := subCompiler.compileNode(child); err != nil {
				return err
			}
		}
		subCompiler.emitByte(byte(OpReturn))

		// Push the compiled function as a constant
		fnChunk := subCompiler.chunk
		// Sync local names for the sub-chunk
		fnChunk.LocalNames = make([]string, len(subCompiler.locals))
		for i, l := range subCompiler.locals {
			fnChunk.LocalNames[i] = l.Name
		}

		c.emitByte(byte(OpConstant))
		c.emitByte(c.addConstant(NewFunction(fnChunk)))

		// Set as global (for now, to match existing behavior)
		c.emitByte(byte(OpSetGlobal))
		c.emitByte(c.addConstant(NewString(funcName)))
		return nil
	}

	// Native Call?
	if node.Name == "call" {
		funcName := coerce.ToString(node.Value)
		// Get function from global
		c.emitByte(byte(OpGetGlobal))
		c.emitByte(c.addConstant(NewString(funcName)))
		// For now, native call doesn't pass explicit arguments through stack
		// (Dynamic scope is used)
		c.emitByte(byte(OpCall))
		c.emitByte(0) // 0 arguments
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
		jumpIfFalsePos := c.emitJump(OpJumpIfFalse)

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
		jumpOverElsePos := c.emitJump(OpJump)

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
			c.emitByte(byte(OpConstant))
			c.emitByte(c.addConstant(NewNumber(0)))

			// 2. Push Iterable
			c.emitByte(byte(OpConstant))
			c.emitByte(c.addConstant(NewObject(slice)))

			// 3. Mark Loop Start
			loopStart := len(c.chunk.Code)

			// 4. OpIterNext (Jumps to End if done)
			// OpIterNext expects [iterable] at peek(0) and [index] at peek(1) relative to CURRENT SP
			exitJump := c.emitJump(OpIterNext)

			// 5. Loop Body
			c.emitByte(byte(OpPop)) // Pop the boolean 'true' from IterNext

			// The 'item' is now at the top of the stack (pushed by OpIterNext).
			// We assign it to a local named 'item' so we can resolve it.
			itemIdx := c.resolveLocal("item")
			if itemIdx == -1 {
				c.locals = append(c.locals, Local{Name: "item", Depth: c.scopeDepth})
				itemIdx = len(c.locals) - 1
			}
			// Important: OpSetLocal DOES NOT pop.
			c.emitByte(byte(OpSetLocal))
			c.emitByte(byte(itemIdx))

			for _, child := range node.Children {
				if err := c.compileNode(child); err != nil {
					return err
				}
			}

			// Pop the item before looping back to maintain stack balance
			c.emitByte(byte(OpPop))

			// 6. Loop back
			c.emitByte(byte(OpLoop))
			offset := len(c.chunk.Code) - loopStart + 2
			c.emitByte(byte((offset >> 8) & 0xff))
			c.emitByte(byte(offset & 0xff))

			// 7. Cleanup and Exit
			c.patchJump(exitJump)
			c.emitByte(byte(OpIterEnd))
			return nil
		}
	}

	// [NEW] ZIP Iteration: items, tags: [[1, 2], ["a", "b"]]
	// For now, we handle single list iteration perfectly. Zip will be refined later.

	// If it's a regular slot call (e.g., http.response)
	if node.Name != "" && !strings.HasPrefix(node.Name, "$") && node.Name != "root" && node.Name != "else" && node.Name != "then" && node.Name != "try" && node.Name != "catch" {
		// Compile children as named arguments
		for _, child := range node.Children {
			// Push Name
			c.emitByte(byte(OpConstant))
			c.emitByte(c.addConstant(NewString(child.Name)))
			// Push Value
			if err := c.compileNodeAsValue(child); err != nil {
				return err
			}
		}

		c.emitByte(byte(OpCallSlot))
		c.emitByte(c.addConstant(NewString(node.Name)))
		c.emitByte(byte(len(node.Children))) // Argument count
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
				jumpToCatch := c.emitJump(OpTry)

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
				c.emitByte(byte(OpEndTry))

				// 4. Jump Over Catch (Success path)
				jumpOverCatch := c.emitJump(OpJump)

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
					c.emitByte(byte(OpSetLocal))
					c.emitByte(byte(idx))
					c.emitByte(byte(OpPop))
				} else {
					c.emitByte(byte(OpPop)) // Consume error
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
		c.emitByte(byte(OpLogicalNot))
		return nil
	}

	// Multi-part expressions (Binary ops)
	ops := []string{"||", "&&", "==", "!=", ">=", "<=", ">", "<", "+"}
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
				c.emitByte(byte(OpLogicalOr))
			case "&&":
				c.emitByte(byte(OpLogicalAnd))
			case "==":
				c.emitByte(byte(OpEqual))
			case "!=":
				c.emitByte(byte(OpNotEqual))
			case ">=":
				c.emitByte(byte(OpGreaterEqual))
			case "<=":
				c.emitByte(byte(OpLessEqual))
			case ">":
				c.emitByte(byte(OpGreater))
			case "<":
				c.emitByte(byte(OpLess))
			case "+":
				c.emitByte(byte(OpAdd))
			}
			return nil
		}
	}

	// Default: Single value truthiness
	return c.compileValue(expr)
}

func (c *Compiler) emitJump(op OpCode) int {
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
		c.emitByte(byte(OpNil))
		return nil
	}
	if b, ok := v.(bool); ok {
		if b {
			c.emitByte(byte(OpTrue))
		} else {
			c.emitByte(byte(OpFalse))
		}
		return nil
	}
	if n, ok := v.(float64); ok {
		c.emitByte(byte(OpConstant))
		c.emitByte(c.addConstant(NewNumber(n)))
		return nil
	}
	if n, ok := v.(int); ok {
		c.emitByte(byte(OpConstant))
		c.emitByte(c.addConstant(NewNumber(float64(n))))
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
		// Variable reference?
		if strings.HasPrefix(s, "$") {
			path := strings.Split(s[1:], ".")
			rootName := path[0]

			if idx := c.resolveLocal(rootName); idx != -1 {
				c.emitByte(byte(OpGetLocal))
				c.emitByte(byte(idx))
			} else {
				c.emitByte(byte(OpGetGlobal))
				c.emitByte(c.addConstant(NewString(rootName)))
			}

			// Handle nested properties: .0, .name, etc.
			for i := 1; i < len(path); i++ {
				c.emitByte(byte(OpAccessProperty))
				c.emitByte(c.addConstant(NewString(path[i])))
			}
			return nil
		}

		// [NEW] List Literal: ["a", "b"]
		if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
			content := s[1 : len(s)-1]
			if content == "" {
				c.emitByte(byte(OpMakeList))
				c.emitByte(0)
				return nil
			}
			parts := strings.Split(content, ",")
			for _, p := range parts {
				if err := c.compileValue(strings.TrimSpace(p)); err != nil {
					return err
				}
			}
			c.emitByte(byte(OpMakeList))
			c.emitByte(byte(len(parts)))
			return nil
		}

		// [NEW] Boolean and Nil Literals
		if s == "true" {
			c.emitByte(byte(OpTrue))
			return nil
		}
		if s == "false" {
			c.emitByte(byte(OpFalse))
			return nil
		}
		if s == "nil" || s == "null" {
			c.emitByte(byte(OpNil))
			return nil
		}

		// Number?
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			c.emitByte(byte(OpConstant))
			c.emitByte(c.addConstant(NewNumber(f)))
			return nil
		}
		// String literal (Strip quotes)
		if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) {
			s = s[1 : len(s)-1]
		}
		c.emitByte(byte(OpConstant))
		c.emitByte(c.addConstant(NewString(s)))
		return nil
	}
	// Fallback raw values
	c.emitByte(byte(OpConstant))
	c.emitByte(c.addConstant(NewObject(v)))
	return nil
}

func (c *Compiler) emitByte(b byte) {
	c.chunk.Code = append(c.chunk.Code, b)
}

func (c *Compiler) addConstant(v Value) byte {
	c.chunk.Constants = append(c.chunk.Constants, v)
	return byte(len(c.chunk.Constants) - 1)
}

func (c *Compiler) compileNodeAsValue(node *engine.Node) error {
	if node.Value != nil {
		return c.compileValue(node.Value)
	}
	if len(node.Children) > 0 {
		for _, child := range node.Children {
			c.emitByte(byte(OpConstant))
			c.emitByte(c.addConstant(NewString(child.Name)))
			if err := c.compileNodeAsValue(child); err != nil {
				return err
			}
		}
		c.emitByte(byte(OpMakeMap))
		c.emitByte(byte(len(node.Children)))
		return nil
	}
	c.emitByte(byte(OpNil))
	return nil
}

// splitExpression splits the expression s by the first occurrence of op,
// ignoring operators that are inside quotes (single or double).
// Returns left part, right part, and true if found.
func splitExpression(s, op string) (string, string, bool) {
	inQuote := false
	var quoteChar rune

	for i := 0; i < len(s); i++ {
		char := rune(s[i])
		if char == '"' || char == '\'' {
			if !inQuote {
				inQuote = true
				quoteChar = char
			} else if char == quoteChar {
				// Check for escaped quote? For now assume simple strings
				if i > 0 && s[i-1] == '\\' {
					// escaped, continue
				} else {
					inQuote = false
				}
			}
		}

		if !inQuote {
			// Check if op matches here
			if strings.HasPrefix(s[i:], op) {
				return s[:i], s[i+len(op):], true
			}
		}
	}
	return "", "", false
}
