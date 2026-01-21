package vm

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strconv"
)

const StackMax = 256
const FramesMax = 64

// CallFrame represents a single function execution context.
type CallFrame struct {
	chunk *Chunk
	ip    int // Instruction Pointer for this frame
	base  int // Index in the global stack where this frame's locals begin
}

// CatchFrame represents a try/catch block handler.
type CatchFrame struct {
	Ip         int // Instruction Pointer to jump to (Catch block)
	Sp         int // Stack Pointer to restore
	FrameCount int // Call Frame count to restore
}

// VM is the bytecode execution engine.
//
// OWNERSHIP: VM does NOT own Chunk, ExternalCallHandler, or ScopeInterface.
//
//	It only borrows them during Run().
//
// THREAD-SAFETY: NOT thread-safe. Use one VM instance per goroutine.
// LIFECYCLE: Can be reused for multiple Run() calls after each completes.
type VM struct {
	frames     [FramesMax]CallFrame
	frameCount int

	stack [StackMax]Value
	sp    int // Stack Pointer

	catchFrames []CatchFrame

	// External dependencies (injected, not owned)
	externalHandler ExternalCallHandler
	scope           ScopeInterface
}

// NewVM creates a new VM instance with given dependencies.
//
// PRECONDITION: handler and scope must be non-nil
// POSTCONDITION: VM is ready for Run() calls
func NewVM(handler ExternalCallHandler, scope ScopeInterface) *VM {
	if handler == nil {
		panic("NewVM: handler must not be nil")
	}
	if scope == nil {
		panic("NewVM: scope must not be nil")
	}

	return &VM{
		catchFrames:     make([]CatchFrame, 0),
		externalHandler: handler,
		scope:           scope,
	}
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

func (vm *VM) syncLocals() {
	frame := vm.frame()
	for i, name := range frame.chunk.LocalNames {
		stackIdx := frame.base + i
		vm.scope.Set(name, vm.stack[stackIdx].ToNative())
	}
}

func (vm *VM) pushFrame(chunk *Chunk, base int) {
	frame := &vm.frames[vm.frameCount]
	frame.chunk = chunk
	frame.ip = 0
	frame.base = base
	vm.frameCount++
}

func (vm *VM) Run(chunk *Chunk) error {
	// Root Frame
	vm.frameCount = 0
	vm.sp = 0
	vm.pushFrame(chunk, 0)
	// Reserve space for locals
	vm.sp = len(chunk.LocalNames)

	for {
		instruction := OpCode(vm.readByte())
		var err error

		switch instruction {
		case OpReturn:
			vm.syncLocals()
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
			nameVal := vm.readConstant()
			name, ok := nameVal.AsString()
			if !ok {
				err = fmt.Errorf("OpGetGlobal: expected string constant")
				if vm.handleError(err) {
					continue
				}
				return err
			}
			val, ok := vm.scope.Get(name)
			if ok {
				vm.push(NewValue(val))
			} else {
				vm.push(NewNil())
			}

		case OpSetGlobal:
			nameVal := vm.readConstant()
			name, ok := nameVal.AsString()
			if !ok {
				err = fmt.Errorf("OpSetGlobal: expected string constant")
				if vm.handleError(err) {
					continue
				}
				return err
			}
			val := vm.pop()
			vm.scope.Set(name, val.ToNative())

		case OpAdd:
			b := vm.pop()
			a := vm.pop()

			// String Concatenation?
			if a.Type == ValString || b.Type == ValString {
				strA, _ := a.AsString()
				strB, _ := b.AsString()
				// If one is not string, use String() method to convert
				if a.Type != ValString {
					strA = a.String()
				}
				if b.Type != ValString {
					strB = b.String()
				}
				vm.push(NewString(strA + strB))
			} else {
				// Numeric Addition
				numA, _ := a.AsNumber()
				numB, _ := b.AsNumber()
				vm.push(NewNumber(numA + numB))
			}

		case OpSubtract:
			b := vm.pop()
			a := vm.pop()
			numA, _ := a.AsNumber()
			numB, _ := b.AsNumber()
			vm.push(NewNumber(numA - numB))

		case OpCallSlot:
			nameVal := vm.readConstant()
			slotName, ok := nameVal.AsString()
			if !ok {
				err = fmt.Errorf("OpCallSlot: expected string slot name")
				if vm.handleError(err) {
					continue
				}
				return err
			}
			argCount := int(vm.readByte())

			// Collect arguments from stack into map
			args := make(map[string]interface{}, argCount)
			for i := argCount - 1; i >= 0; i-- {
				val := vm.pop()
				argNameVal := vm.pop()
				argName, ok := argNameVal.AsString()
				if !ok {
					err = fmt.Errorf("OpCallSlot: argument name must be string")
					if vm.handleError(err) {
						continue
					}
					return err
				}
				args[argName] = val.ToNative()
			}

			// Sync locals before external call
			vm.syncLocals()

			// Call external handler
			_, err = vm.externalHandler.Call(slotName, args)
			if err != nil {
				if vm.handleError(err) {
					continue
				}
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
			numA, _ := a.AsNumber()
			numB, _ := b.AsNumber()
			vm.push(NewBool(numA > numB))

		case OpGreaterEqual:
			b := vm.pop()
			a := vm.pop()
			numA, _ := a.AsNumber()
			numB, _ := b.AsNumber()
			vm.push(NewBool(numA >= numB))

		case OpLess:
			b := vm.pop()
			a := vm.pop()
			numA, _ := a.AsNumber()
			numB, _ := b.AsNumber()
			vm.push(NewBool(numA < numB))

		case OpLessEqual:
			b := vm.pop()
			a := vm.pop()
			numA, _ := a.AsNumber()
			numB, _ := b.AsNumber()
			vm.push(NewBool(numA <= numB))

		case OpCall:
			argCount := int(vm.readByte())
			fnVal := vm.stack[vm.sp-1-argCount]
			if fnVal.Type != ValFunction {
				err = fmt.Errorf("can only call functions, got %v", fnVal.Type)
				if vm.handleError(err) {
					continue
				}
				return err
			}
			fnChunk, _ := fnVal.AsFunction()
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
			indexFloat, _ := indexVal.AsNumber()
			index := int(indexFloat)

			var nextVal Value
			var hasNext bool

			switch iterable.Type {
			case ValList:
				if list, ok := iterable.AsList(); ok {
					if index < len(list) {
						nextVal = list[index]
						hasNext = true
						// Increment hidden index
						vm.stack[vm.sp-2] = NewNumber(float64(index + 1))
					}
				}
			case ValMap:
				if m, ok := iterable.AsMap(); ok {
					// Inefficient map iteration (re-keys every step) but works for now.
					// Implementation plan suggested re-keying or creating separate iterator object.
					// Keeping simple for Phase 2: extract keys every time.
					keys := make([]string, 0, len(m))
					for k := range m {
						keys = append(keys, k)
					}
					if index < len(keys) {
						key := keys[index]
						nextVal = m[key]
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

		case OpLogicalOr:
			b := vm.pop()
			a := vm.pop()
			vm.push(NewBool(vm.isTruthy(a) || vm.isTruthy(b)))

		case OpLogicalAnd:
			b := vm.pop()
			a := vm.pop()
			vm.push(NewBool(vm.isTruthy(a) && vm.isTruthy(b)))

		case OpLogicalNot:
			a := vm.pop()
			vm.push(NewBool(!vm.isTruthy(a)))

		case OpTry:
			offset := vm.readShort()
			vm.catchFrames = append(vm.catchFrames, CatchFrame{
				Ip:         vm.frame().ip + int(offset),
				Sp:         vm.sp,
				FrameCount: vm.frameCount,
			})

		case OpEndTry:
			if len(vm.catchFrames) > 0 {
				vm.catchFrames = vm.catchFrames[:len(vm.catchFrames)-1]
			}

		case OpStop:
			vm.syncLocals()
			return nil

		case OpAccessProperty:
			nameVal := vm.readConstant()
			name, ok := nameVal.AsString()
			if !ok {
				err = fmt.Errorf("OpAccessProperty: expected string property name")
				if vm.handleError(err) {
					continue
				}
				return err
			}
			obj := vm.pop()

			var res Value
			switch obj.Type {
			case ValMap:
				if m, ok := obj.AsMap(); ok {
					if v, ok := m[name]; ok {
						res = v
					}
				}
			case ValList:
				if list, ok := obj.AsList(); ok {
					// Numeric index?
					if idx, err := strconv.Atoi(name); err == nil {
						if idx >= 0 && idx < len(list) {
							res = list[idx]
						}
					}
				}
			}
			vm.push(res)

		case OpMakeMap:
			count := int(vm.readByte())
			m := make(map[string]Value)
			for i := 0; i < count; i++ {
				val := vm.pop()
				keyVal := vm.pop()

				key, _ := keyVal.AsString() // Assumes compiler pushed string key
				m[key] = val
			}
			vm.push(NewMap(m))

		case OpMakeList:
			count := int(vm.readByte())
			slice := make([]Value, count)
			for i := count - 1; i >= 0; i-- {
				slice[i] = vm.pop()
			}
			vm.push(NewList(slice))

		case OpPop:
			vm.pop()

		default:
			err = fmt.Errorf("unsupported opcode: %d", instruction)
			if vm.handleError(err) {
				continue
			}
			return err
		}
	}
}

// handleError attempts to recover from an error using catch frames.
// Returns true if recovered (jumped to catch block), false if fatal.
func (vm *VM) handleError(err error) bool {
	if len(vm.catchFrames) > 0 {
		// Recover
		catch := vm.catchFrames[len(vm.catchFrames)-1]
		vm.catchFrames = vm.catchFrames[:len(vm.catchFrames)-1] // Pop handler

		vm.frameCount = catch.FrameCount
		vm.sp = catch.Sp

		// Push error object
		vm.push(NewString(err.Error()))

		vm.frame().ip = catch.Ip
		// Continue execution from catch block
		return true
	}
	return false
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
		b, _ := v.AsBool()
		return b
	case ValNumber:
		n, _ := v.AsNumber()
		return n != 0
	case ValString:
		s, _ := v.AsString()
		return s != ""
	default:
		return true
	}
}
