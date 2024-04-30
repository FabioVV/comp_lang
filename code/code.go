package code

import (
	"encoding/binary"
	"fmt"
)

type Instructions []byte
type Opcode byte

const (
	Opconstant Opcode = iota
)

type Definition struct {
	Name          string
	OperandWidths []int
}

var defs = map[Opcode]*Definition{
	Opconstant: {"Opconstant", []int{2}},
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
	return ""
}
