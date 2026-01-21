package vm

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"zeno/pkg/engine"
)

const StackMax = 256
const FramesMax = 64

// CallFrame represents a single function execution context.
type CallFrame struct {
	chunk *Chunk
	ip    int // Instruction Pointer for this frame
	base  int // Index in the global stack where this frame's locals begin
}

// VM is the bytecode execution engine.
type VM struct {
	frames     [FramesMax]CallFrame
	frameCount int

	stack [StackMax]Value
	sp    int // Stack Pointer
}

func NewVM() *VM {
	return &VM{}
}

// Chunk stores a sequence of bytecode and constants.
type Chunk struct {
	Code       []byte
	Constants  []Value
	LocalNames []string
}

// Serialize writes the chunk to a binary stream.
func (c *Chunk) Serialize(w io.Writer) error {
	// 1. Magic + Version
	if _, err := w.Write([]byte("ZBC1")); err != nil {
		return err
	}

	// 2. Code Size + Data
	if err := binary.Write(w, binary.LittleEndian, uint32(len(c.Code))); err != nil {
		return err
	}
	if _, err := w.Write(c.Code); err != nil {
		return err
	}

	// 3. Constants Size + Data
	if err := binary.Write(w, binary.LittleEndian, uint32(len(c.Constants))); err != nil {
		return err
	}
	for _, v := range c.Constants {
		if err := v.Serialize(w); err != nil {
			return err
		}
	}

	// 4. LocalNames Size + Data
	if err := binary.Write(w, binary.LittleEndian, uint32(len(c.LocalNames))); err != nil {
		return err
	}
	for _, name := range c.LocalNames {
		if err := writeString(w, name); err != nil {
			return err
		}
	}

	return nil
}

// Deserialize reads a chunk from a binary stream.
func DeserializeChunk(r io.Reader) (*Chunk, error) {
	magic := make([]byte, 4)
	if _, err := io.ReadFull(r, magic); err != nil {
		return nil, err
	}
	if string(magic) != "ZBC1" {
		return nil, fmt.Errorf("invalid magic number")
	}

	c := &Chunk{}

	// 1. Code
	var codeLen uint32
	if err := binary.Read(r, binary.LittleEndian, &codeLen); err != nil {
		return nil, err
	}
	c.Code = make([]byte, codeLen)
	if _, err := io.ReadFull(r, c.Code); err != nil {
		return nil, err
	}

	// 2. Constants
	var constLen uint32
	if err := binary.Read(r, binary.LittleEndian, &constLen); err != nil {
		return nil, err
	}
	c.Constants = make([]Value, constLen)
	for i := uint32(0); i < constLen; i++ {
		v, err := DeserializeValue(r)
		if err != nil {
			return nil, err
		}
		c.Constants[i] = v
	}

	// 3. LocalNames
	var localLen uint32
	if err := binary.Read(r, binary.LittleEndian, &localLen); err != nil {
		return nil, err
	}
	c.LocalNames = make([]string, localLen)
	for i := uint32(0); i < localLen; i++ {
		name, err := readString(r)
		if err != nil {
			return nil, err
		}
		c.LocalNames[i] = name
	}

	return c, nil
}

func writeString(w io.Writer, s string) error {
	if err := binary.Write(w, binary.LittleEndian, uint32(len(s))); err != nil {
		return err
	}
	_, err := w.Write([]byte(s))
	return err
}

func readString(r io.Reader) (string, error) {
	var l uint32
	if err := binary.Read(r, binary.LittleEndian, &l); err != nil {
		return "", err
	}
	buf := make([]byte, l)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

// SaveToFile saves the chunk to a file.
func (c *Chunk) SaveToFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return c.Serialize(f)
}

// LoadFromFile loads a chunk from a file.
func LoadFromFile(filename string) (*Chunk, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return DeserializeChunk(f)
}

func (vm *VM) push(val Value) {
	vm.stack[vm.sp] = val
	vm.sp++
}

func (vm *VM) pop() Value {
	vm.sp--
	return vm.stack[vm.sp]
}

func (vm *VM) peek(distance int) Value {
	return vm.stack[vm.sp-1-distance]
}

func (vm *VM) frame() *CallFrame {
	return &vm.frames[vm.frameCount-1]
}

func (vm *VM) syncLocals(scope *engine.Scope) {
	frame := vm.frame()
	for i, name := range frame.chunk.LocalNames {
		stackIdx := frame.base + i
		scope.Set(name, vm.stack[stackIdx].ToNative())
	}
}

func (vm *VM) pushFrame(chunk *Chunk, base int) {
	frame := &vm.frames[vm.frameCount]
	frame.chunk = chunk
	frame.ip = 0
	frame.base = base
	vm.frameCount++
}

func (vm *VM) Run(ctx context.Context, chunk *Chunk, scope *engine.Scope) error {
	// Root Frame
	vm.frameCount = 0
	vm.sp = 0
	vm.pushFrame(chunk, 0)
	// Reserve space for locals
	vm.sp = len(chunk.LocalNames)

	for {
		instruction := OpCode(vm.readByte())
		switch instruction {
		case OpReturn:
			vm.syncLocals(scope)
			vm.frameCount--
			if vm.frameCount == 0 {
				return nil
			}
			// When returning from a function, we usually pop the call frame
			// and continue in the previous one.
			// Internal function return logic will be refined in OpCall implementation.

		case OpConstant:
			constant := vm.readConstant()
			vm.push(constant)

		case OpNil:
			vm.push(NewNil())
		case OpTrue:
			vm.push(NewBool(true))
		case OpFalse:
			vm.push(NewBool(false))
		// OpPop is now later

		case OpGetGlobal:
			name := vm.readConstant().AsPtr.(string)
			val, ok := scope.Get(name)
			if ok {
				vm.push(NewValue(val))
			} else {
				vm.push(NewNil())
			}

		case OpSetGlobal:
			name := vm.readConstant().AsPtr.(string)
			val := vm.pop()
			scope.Set(name, val.ToNative())

		case OpAdd:
			b := vm.pop()
			a := vm.pop()
			vm.push(NewNumber(a.AsNum + b.AsNum))

		case OpSubtract:
			b := vm.pop()
			a := vm.pop()
			vm.push(NewNumber(a.AsNum - b.AsNum))

		case OpCallSlot:
			slotName := vm.readConstant().AsPtr.(string)
			argCount := int(vm.readByte())

			// 1. Get Engine from context
			eng, ok := ctx.Value("engine").(*engine.Engine)
			if !ok {
				return fmt.Errorf("engine not found in context")
			}

			// 2. Lookup Handler
			handler, exists := eng.Registry[slotName]
			if !exists {
				return fmt.Errorf("slot not found: %s", slotName)
			}

			// 3. Prepare Mock Node for arguments (Stack to Node bridge)
			mockNode := &engine.Node{Name: slotName}
			if argCount > 0 {
				mockNode.Children = make([]*engine.Node, argCount)
				// Pop in reverse order because they were pushed in original order
				for i := argCount - 1; i >= 0; i-- {
					val := vm.pop()
					nameVal := vm.pop()
					name := nameVal.AsPtr.(string)

					mockNode.Children[i] = &engine.Node{
						Name:   name,
						Value:  val.ToNative(),
						Parent: mockNode,
					}
				}
			}

			// [NEW] Sync locals before calling native slot to ensure consistency
			vm.syncLocals(scope)

			// 4. Execution
			err := handler(ctx, mockNode, scope)
			if err != nil {
				return err
			}

		case OpEqual:
			b := vm.pop()
			a := vm.pop()
			vm.push(NewBool(a.ToNative() == b.ToNative()))

		case OpNotEqual:
			b := vm.pop()
			a := vm.pop()
			vm.push(NewBool(a.ToNative() != b.ToNative()))

		case OpGreater:
			b := vm.pop()
			a := vm.pop()
			vm.push(NewBool(a.AsNum > b.AsNum))

		case OpGreaterEqual:
			b := vm.pop()
			a := vm.pop()
			vm.push(NewBool(a.AsNum >= b.AsNum))

		case OpLess:
			b := vm.pop()
			a := vm.pop()
			vm.push(NewBool(a.AsNum < b.AsNum))

		case OpLessEqual:
			b := vm.pop()
			a := vm.pop()
			vm.push(NewBool(a.AsNum <= b.AsNum))

		case OpCall:
			argCount := int(vm.readByte())
			fnVal := vm.stack[vm.sp-1-argCount]
			if fnVal.Type != ValFunction {
				return fmt.Errorf("can only call functions, got %v", fnVal.Type)
			}
			fnChunk := fnVal.AsPtr.(*Chunk)
			// Base of new frame is the first argument's position
			// The function value itself stays on stack below base (to be cleaned up on return)
			vm.pushFrame(fnChunk, vm.sp-argCount)
			// Reserve space for locals
			vm.sp = vm.frame().base + len(fnChunk.LocalNames)

		case OpGetLocal:
			index := vm.readByte()
			stackIdx := vm.frame().base + int(index)
			val := vm.stack[stackIdx]
			vm.push(val)

		case OpSetLocal:
			index := vm.readByte()
			val := vm.peek(0)
			stackIdx := vm.frame().base + int(index)
			vm.stack[stackIdx] = val
			// Ensure sp covers the local slots
			if stackIdx >= vm.sp {
				vm.sp = stackIdx + 1
			}

		case OpJump:
			offset := vm.readShort()
			vm.frame().ip += int(offset)

		case OpJumpIfFalse:
			offset := vm.readShort()
			condition := vm.pop()
			if !vm.isTruthy(condition) {
				vm.frame().ip += int(offset)
			}

		case OpLoop:
			offset := vm.readShort()
			vm.frame().ip -= int(offset)

		case OpIterNext:
			offset := vm.readShort()
			iterable := vm.peek(0)
			// Hidden iteration index is at peek(1)
			indexVal := vm.peek(1)
			index := int(indexVal.AsNum)

			var nextVal Value
			var hasNext bool

			switch iterable.Type {
			case ValObject:
				if slice, ok := iterable.AsPtr.([]interface{}); ok {
					if index < len(slice) {
						nextVal = NewValue(slice[index])
						hasNext = true
						// Increment hidden index
						vm.stack[vm.sp-2] = NewNumber(float64(index + 1))
					}
				} else if m, ok := iterable.AsPtr.(map[string]interface{}); ok {
					// Optimized Map Iteration: convert to slice ONLY once (at index 0)
					// or better: store the keys as a hidden object at peek(2).
					// For now, let's just avoid the allocation if index > 0
					keys := make([]string, 0, len(m))
					for k := range m {
						keys = append(keys, k)
					}
					if index < len(keys) {
						key := keys[index]
						nextVal = NewValue(m[key])
						hasNext = true
						vm.stack[vm.sp-2] = NewNumber(float64(index + 1))
					}
				}
			default:
				fmt.Printf("DEBUG: OpIterNext unsupported type: %v\n", iterable.Type)
			}

			if hasNext {
				vm.push(nextVal)
				vm.push(NewBool(true))
			} else {
				vm.push(NewNil()) // Placeholder
				vm.push(NewBool(false))
				vm.frame().ip += int(offset)
			}

		case OpIterEnd:
			// Pop boolean status, nextVal, iterable, and hidden index
			// these were pushed AFTER the locals.
			vm.pop() // status
			vm.pop() // nextVal
			vm.pop() // iterable
			vm.pop() // index

		case OpPop:
			vm.pop()

		default:
			return fmt.Errorf("unsupported opcode: %d", instruction)
		}
	}
}

func (vm *VM) readByte() byte {
	frame := vm.frame()
	b := frame.chunk.Code[frame.ip]
	frame.ip++
	return b
}

func (vm *VM) readShort() uint16 {
	frame := vm.frame()
	frame.ip += 2
	return uint16(frame.chunk.Code[frame.ip-2])<<8 | uint16(frame.chunk.Code[frame.ip-1])
}

func (vm *VM) readConstant() Value {
	index := vm.readByte()
	return vm.frame().chunk.Constants[index]
}

func (vm *VM) isTruthy(v Value) bool {
	switch v.Type {
	case ValNil:
		return false
	case ValBool:
		return v.AsNum > 0
	case ValNumber:
		return v.AsNum != 0
	case ValString:
		return v.AsPtr.(string) != ""
	default:
		return true
	}
}
