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

	// If it's a	// Variable Assignment: $x: 10
	if strings.HasPrefix(node.Name, "$") {
		varName := node.Name[1:]
		// Evaluate value
		if s, ok := node.Value.(string); ok && strings.Contains(s, " ") {
			// Likely an expression
			if err := c.compileExpression(s); err != nil {
				return err
			}
		} else {
			if err := c.compileValue(node.Value); err != nil {
				return err
			}
		}

		if idx := c.resolveLocal(varName); idx != -1 {
			c.emitByte(byte(OpSetLocal))
			c.emitByte(byte(idx))
			c.emitByte(byte(OpPop))
		} else {
			// Global assignment
			c.emitByte(byte(OpSetGlobal))
			c.emitByte(c.addConstant(NewString(varName)))
		}
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

	// If it's an expression like "1 + 2" (Current Zeno stores this in Value)
	if node.Value != nil {
		valStr := fmt.Sprintf("%v", node.Value)
		parts := strings.Fields(valStr)
		if len(parts) == 3 && parts[1] == "+" {
			// Very basic arithmetic parser
			v1, _ := strconv.ParseFloat(parts[0], 64)
			v2, _ := strconv.ParseFloat(parts[2], 64)

			c.emitByte(byte(OpConstant))
			c.emitByte(c.addConstant(NewNumber(v1)))

			c.emitByte(byte(OpConstant))
			c.emitByte(c.addConstant(NewNumber(v2)))

			c.emitByte(byte(OpAdd))
			return nil
		}
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
	if node.Name != "" && !strings.HasPrefix(node.Name, "$") && node.Name != "root" && node.Name != "else" && node.Name != "then" {
		// Compile children as named arguments
		for _, child := range node.Children {
			// Push Name
			c.emitByte(byte(OpConstant))
			c.emitByte(c.addConstant(NewString(child.Name)))
			// Push Value
			if err := c.compileValue(child.Value); err != nil {
				return err
			}
		}

		c.emitByte(byte(OpCallSlot))
		c.emitByte(c.addConstant(NewString(node.Name)))
		c.emitByte(byte(len(node.Children))) // Argument count
		return nil
	}

	// Default: Compile children (Container node)
	for _, child := range node.Children {
		if err := c.compileNode(child); err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiler) compileExpression(expr string) error {
	// Simple Expression Parser: $x == 10
	expr = strings.TrimSpace(expr)

	ops := []string{"==", "!=", ">=", "<=", ">", "<", "+"}
	for _, op := range ops {
		if strings.Contains(expr, op) {
			parts := strings.SplitN(expr, op, 2)
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])

			if err := c.compileValue(left); err != nil {
				return err
			}
			if err := c.compileValue(right); err != nil {
				return err
			}

			switch op {
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
		s = strings.TrimSpace(s)
		// Variable reference?
		if strings.HasPrefix(s, "$") {
			varName := s[1:]
			if idx := c.resolveLocal(varName); idx != -1 {
				c.emitByte(byte(OpGetLocal))
				c.emitByte(byte(idx))
			} else {
				c.emitByte(byte(OpGetGlobal))
				c.emitByte(c.addConstant(NewString(varName)))
			}
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
