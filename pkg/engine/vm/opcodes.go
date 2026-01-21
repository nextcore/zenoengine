package vm

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
)
