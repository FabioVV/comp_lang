package compiler

import (
	Ast "github/FabioVV/interp_lang/ast"
	"github/FabioVV/interp_lang/code"
	Code "github/FabioVV/interp_lang/code"
	Object "github/FabioVV/interp_lang/object"
)

type Compiler struct {
	instructions Code.Instructions
	constants    []Object.Object
}

type Bytecode struct {
	Instructions Code.Instructions
	Constants    []Object.Object
}

func New() *Compiler {
	return &Compiler{
		instructions: Code.Instructions{},
		constants:    []Object.Object{},
	}
}

// Add constant to the constants pool and return its index (position in the pool)
func (c *Compiler) addConstant(obj Object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

// add a new instruction to the instructions slice and return the position where the current
// instruction starts
func (c *Compiler) addInstruction(instruction []byte) int {
	posNewInstruction := len(c.instructions)

	c.instructions = append(c.instructions, instruction...)
	return posNewInstruction
}

// We use code.make inside of emitInstruction to generate the instruction
// we then add it to instructions slice and return its starting position
func (c *Compiler) emitInstruction(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)

	return pos
}

func (c *Compiler) Compile(node Ast.Node) error {
	switch node := node.(type) {

	case *Ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)

			if err != nil {
				return err
			}
		}

	case *Ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}

	case *Ast.InfixExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Right)
		if err != nil {
			return err
		}

	case *Ast.IntegerLiteral:
		integer := &Object.Integer{Value: node.Value}
		c.emitInstruction(code.Opconstant, c.addConstant(integer))

	}

	return nil

}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}
