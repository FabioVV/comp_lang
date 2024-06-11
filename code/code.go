package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Instructions []byte
type Opcode byte
type Definition struct {
	Name          string
	OperandWidths []int
}

const (
	Opconstant Opcode = iota
	OpAdd
	// pop the topmost element off the stack
	OpPop
	// subtraction
	OpSub
	// multiplication
	OpMul
	// division
	OpDiv

	// true and false literals
	OpTrue
	OpFalse

	// comparison operators
	OpEqual
	OpNotEqual
	OpGreaterThan

	// prefix operators
	OpMinus
	OpBang

	// Conditional jumping
	OpJumpNotTruthy
	OpJump

	OpNull

	// Value binding to names A.K.A variables
	OpGetGlobal
	OpSetGlobal

	OpArray
	OpHash
	OpIndex

	// Fn invoking :>> random_function()
	OpCall
	OpReturnValue
	OpReturn
)

/*
The new opcode is called OpAdd and tells the VM to pop the two topmost elements off the
stack, add them together and push the result back on to the stack. In contrast to OpConstant,
it doesn’t have any operands. It’s simply one byte, a single opcode
*/

/*
You might be wondering why there is no opcode for <. If we have OpGreaterThan, shouldn’t we
have an OpLessThan, too? That’s a valid question, because we could add OpLessThan and that
would be fine, but I want to show something that’s possible with compilation and not with
interpretation: reordering of code.
The expression 3 < 5 can be reordered to 5 > 3 without changing its result. And because it can
be reordered, that’s what our compiler is going to do. It will take every less-than expression
and reorder it to emit the greater-than version instead. That way we keep the instruction set
small, the loop of our VM tighter and learn about the things we can do with compilation
*/
var defs = map[Opcode]*Definition{
	Opconstant:      {"Opconstant", []int{2}},
	OpJump:          {"OpBang", []int{2}},
	OpJumpNotTruthy: {"OpBang", []int{2}},
	OpAdd:           {"OpAdd", []int{}},
	OpPop:           {"OpPop", []int{}},
	OpSub:           {"OpSub", []int{}},
	OpMul:           {"OpMul", []int{}},
	OpDiv:           {"OpDiv", []int{}},
	OpTrue:          {"OpTrue", []int{}},
	OpFalse:         {"OpFalse", []int{}},
	OpEqual:         {"OpEqual", []int{}},
	OpNotEqual:      {"OpNotEqual", []int{}},
	OpGreaterThan:   {"OpGreaterThan", []int{}},
	OpMinus:         {"OpMinus", []int{}},
	OpBang:          {"OpBang", []int{}},
	OpNull:          {"OpNull", []int{}},
	OpGetGlobal:     {"OpSetGlobal", []int{2}},
	OpSetGlobal:     {"OpSetGlobal", []int{2}},
	OpArray:         {"OpArray", []int{2}},
	OpHash:          {"OpHash", []int{2}},
	OpIndex:         {"OpIndex", []int{}},
	OpCall:          {"OpCall", []int{}},
	OpReturnValue:   {"OpReturnValue", []int{}},
	OpReturn:        {"OpReturn", []int{}},
}

func LookupOp(op byte) (*Definition, error) {
	def, ok := defs[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)

	}
	return def, nil

}

// Decodes the operands
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
