package vm

import "fmt"

// OpCode represents a single instruction for the VM.
type OpCode byte

const (
	OpReturn OpCode = iota // Terminate execution and return

	// Constants and Literals
	OpConstant // Push a constant from the constant pool
	OpNil      // Push nil
	OpTrue     // Push true
	OpFalse    // Push false

	// Variables
	OpGetGlobal // Get variable from global scope
	OpSetGlobal // Set variable in global scope
	OpGetLocal  // Get variable from local stack/arena scope
	OpSetLocal  // Set variable in local stack/arena scope

	// Arithmetic
	OpAdd
	OpSubtract
	OpMultiply
	OpDivide
	OpNegate

	// Comparison
	OpEqual
	OpNotEqual
	OpGreater
	OpGreaterEqual
	OpLess
	OpLessEqual

	// Control Flow
	OpJump        // Jump forward
	OpJumpIfFalse // Jump if stack top is false
	OpLoop        // Jump backward
	OpCall        // Call internal function

	// Engine Specific
	OpCallSlot // Call a ZenoEngine Slot (e.g., http.get)

	// New Opcodes (Add to end to keep iota stable)
	OpPop
	OpIterNext
	OpIterEnd

	// Logical [NEW]
	OpLogicalOr
	OpLogicalAnd
	OpLogicalNot

	// Access [NEW]
	OpAccessProperty // [iterable] name -> [val]

	// Literals [NEW]
	OpMakeMap  // count -> [key1, val1, ..., keyN, valN] -> map
	OpMakeList // count -> [val1, ..., valN] -> list

	// Control [NEW]
	OpStop

	// Exception Handling [NEW]
	OpTry    // jump_offset -> (Push catch frame)
	OpEndTry // (Pop catch frame)
)

func (o OpCode) String() string {
	switch o {
	case OpReturn:
		return "OpReturn"
	case OpConstant:
		return "OpConstant"
	case OpNil:
		return "OpNil"
	case OpTrue:
		return "OpTrue"
	case OpFalse:
		return "OpFalse"
	case OpGetGlobal:
		return "OpGetGlobal"
	case OpSetGlobal:
		return "OpSetGlobal"
	case OpGetLocal:
		return "OpGetLocal"
	case OpSetLocal:
		return "OpSetLocal"
	case OpAdd:
		return "OpAdd"
	case OpSubtract:
		return "OpSubtract"
	case OpMultiply:
		return "OpMultiply"
	case OpDivide:
		return "OpDivide"
	case OpNegate:
		return "OpNegate"
	case OpEqual:
		return "OpEqual"
	case OpNotEqual:
		return "OpNotEqual"
	case OpGreater:
		return "OpGreater"
	case OpGreaterEqual:
		return "OpGreaterEqual"
	case OpLess:
		return "OpLess"
	case OpLessEqual:
		return "OpLessEqual"
	case OpJump:
		return "OpJump"
	case OpJumpIfFalse:
		return "OpJumpIfFalse"
	case OpLoop:
		return "OpLoop"
	case OpCall:
		return "OpCall"
	case OpCallSlot:
		return "OpCallSlot"
	case OpPop:
		return "OpPop"
	case OpIterNext:
		return "OpIterNext"
	case OpIterEnd:
		return "OpIterEnd"
	case OpLogicalOr:
		return "OpLogicalOr"
	case OpLogicalAnd:
		return "OpLogicalAnd"
	case OpLogicalNot:
		return "OpLogicalNot"
	case OpAccessProperty:
		return "OpAccessProperty"
	case OpMakeMap:
		return "OpMakeMap"
	case OpMakeList:
		return "OpMakeList"
	case OpStop:
		return "OpStop"
	case OpTry:
		return "OpTry"
	case OpEndTry:
		return "OpEndTry"
	default:
		return fmt.Sprintf("OpUnknown(%d)", byte(o))
	}
}
