package emulator

// Instruction represents an actual instruction in
// the emulator's memory
type Instruction struct {
	OpType // Embed OpType

	Op8  byte   // 8-bit operand for 2-byte instructions
	Op16 uint16 // 16-bit operand for 3-byte instructions

	Address uint16 // Address location where this instruction got read, for debugging purposes
}

// DecodeInstruction will return the decoded instruction at the
// specified address
func DecodeInstruction(em *Emulator, addr uint16) Instruction {
	opcode := em.ReadMemory(addr)
	opType := NewOpType(opcode)
	result := Instruction{OpType: opType, Op8: 0, Op16: 0, Address: addr}
	switch opType.Size {
	case 2:
		result.Op8 = em.ReadMemory(addr + 1)
	case 3:
		result.Op16 = uint16(em.ReadMemory(addr+2)) << 8
		result.Op16 += uint16(em.ReadMemory(addr + 1))
	}

	return result
}

// DecodeCurrentInstruction will decode the instruction pointed
// to by the Program Counter (PC register)
func DecodeCurrentInstruction(em *Emulator) Instruction {
	return DecodeInstruction(em, em.CPU.PC)
}
