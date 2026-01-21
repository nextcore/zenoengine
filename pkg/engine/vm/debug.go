package vm

import (
	"fmt"
)

// DisassembleChunk prints all instructions in a chunk in a human-readable format.
func (c *Chunk) Disassemble(name string) {
	fmt.Printf("== %s ==\n", name)

	for offset := 0; offset < len(c.Code); {
		offset = c.DisassembleInstruction(offset)
	}
}

// DisassembleInstruction prints a single instruction and returns the next offset.
func (c *Chunk) DisassembleInstruction(offset int) int {
	fmt.Printf("%04d ", offset)
	instruction := OpCode(c.Code[offset])
	switch instruction {
	case OpReturn, OpNil, OpTrue, OpFalse, OpAdd, OpSubtract, OpMultiply, OpDivide, OpNegate,
		OpEqual, OpNotEqual, OpGreater, OpGreaterEqual, OpLess, OpLessEqual, OpIterEnd, OpPop,
		OpLogicalOr, OpLogicalAnd, OpLogicalNot, OpStop:
		return simpleInstruction(instruction.String(), offset)

	case OpConstant, OpGetGlobal, OpSetGlobal, OpCallSlot, OpAccessProperty:
		return constantInstruction(instruction.String(), c, offset)

	case OpGetLocal, OpSetLocal, OpCall, OpMakeMap, OpMakeList:
		return byteInstruction(instruction.String(), c, offset)

	case OpJump, OpJumpIfFalse, OpLoop, OpIterNext, OpTry:
		return jumpInstruction(instruction.String(), 1, c, offset)

	case OpEndTry:
		return simpleInstruction(instruction.String(), offset)

	default:
		fmt.Printf("Unknown opcode %d (%s)\n", byte(instruction), instruction.String())
		return offset + 1
	}
}

func simpleInstruction(name string, offset int) int {
	fmt.Printf("%-16s\n", name)
	return offset + 1
}

func constantInstruction(name string, chunk *Chunk, offset int) int {
	constant := chunk.Code[offset+1]
	fmt.Printf("%-16s %4d '", name, constant)
	fmt.Print(chunk.Constants[constant])
	fmt.Printf("'\n")
	return offset + 2
}

func byteInstruction(name string, chunk *Chunk, offset int) int {
	slot := chunk.Code[offset+1]
	fmt.Printf("%-16s %4d\n", name, slot)
	return offset + 2
}

func jumpInstruction(name string, sign int, chunk *Chunk, offset int) int {
	jump := uint16(chunk.Code[offset+1])<<8 | uint16(chunk.Code[offset+2])
	target := offset + 3 + int(jump)*sign
	if name == "OpLoop" {
		target = offset + 3 - int(jump)
	}
	fmt.Printf("%-16s %4d -> %04d\n", name, offset, target)
	return offset + 3
}
