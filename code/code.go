package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Instructions []byte
type Opcode byte

const (
	Opconstant Opcode = iota
	OpAdd
)

type Definition struct {
	Name          string
	OperandWidths []int
}

/*
The new opcode is called OpAdd and tells the VM to pop the two topmost elements off the
stack, add them together and push the result back on to the stack. In contrast to OpConstant,
it doesn’t have any operands. It’s simply one byte, a single opcode
*/
var defs = map[Opcode]*Definition{
	Opconstant: {"Opconstant", []int{2}},
	OpAdd:      {"OpAdd", []int{}},
}

func LookupOp(op byte) (*Definition, error) {
	def, ok := defs[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)

	}
	return def, nil

}

func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}

func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		}

		offset += width
	}
	return operands, offset
}

func Make(op Opcode, operands ...int) []byte {
	def, ok := defs[op]

	if !ok {
		return []byte{}
	}

	instructionLength := 1
	for _, w := range def.OperandWidths {
		instructionLength += w
	}

	instruction := make([]byte, instructionLength)
	instruction[0] = byte(op)

	offset := 1

	for i, o := range operands {
		width := def.OperandWidths[i]

		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))

		}

		offset += width
	}

	return instruction
}

func (ins Instructions) MiniDisassembler() string {
	var out bytes.Buffer

	i := 0
	for i <= len(ins) {
		def, err := LookupOp(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "Error: %s\n", err)
			continue
		}
		operands, read := ReadOperands(def, ins[i+1:])
		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))
		i += 1 + read
	}

	return out.String()
}

func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)

	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n", len(operands), operandCount)
	}

	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	}

	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}
