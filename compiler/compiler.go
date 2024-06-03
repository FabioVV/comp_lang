package compiler

import (
	"fmt"
	Ast "github/FabioVV/interp_lang/ast"
	"github/FabioVV/interp_lang/code"
	Code "github/FabioVV/interp_lang/code"
	Object "github/FabioVV/interp_lang/object"
)

type EmittedInstruction struct {
	Opcode code.Opcode
	Pos    int
}

type Compiler struct {
	instructions Code.Instructions
	constants    []Object.Object

	LastInstruction     EmittedInstruction
	PreviousInstruction EmittedInstruction
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

// We need this method to keep track of the last two emitted instructions
func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	previous := c.LastInstruction
	last := EmittedInstruction{Opcode: op, Pos: pos}

	c.PreviousInstruction = previous
	c.LastInstruction = last
}

// We use code.make inside of emitInstruction to generate the instruction
// we then add it to instructions slice and return its starting position
func (c *Compiler) emitInstruction(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setLastInstruction(op, pos)

	return pos
}

func (c *Compiler) lastInstructionIsPop() bool {
	return c.LastInstruction.Opcode == code.OpPop
}

func (c *Compiler) removeLastPop() {
	c.instructions = c.instructions[:c.LastInstruction.Pos]
	c.LastInstruction = c.PreviousInstruction
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	for i := 0; i < len(newInstruction); i++ {
		c.instructions[pos+i] = newInstruction[i]
	}
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := code.Opcode(c.instructions[opPos])
	newInstruction := code.Make(op, operand)

	c.replaceInstruction(opPos, newInstruction)
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
		if err := c.Compile(node.Expression); err != nil {
			return err

		} else {
			c.emitInstruction(code.OpPop)

		}

	case *Ast.PrefixExpression:
		if err := c.Compile(node.Right); err != nil {
			return err
		}

		switch node.Operator {
		case "!":
			c.emitInstruction(code.OpBang)

		case "-":
			c.emitInstruction(code.OpMinus)

		default:
			return fmt.Errorf("unknown operator %s", node.Operator)

		}

	case *Ast.InfixExpression:
		/*
			What we did here is to turn < into a special case. We turn the order around and first compile
			node.Right and then node.Left in case the operator is <. After that we emit the OpGreaterThan
			opcode. We changed a less-than comparison into a greater-than comparison – while compiling.
		*/
		if node.Operator == "<" {
			err := c.Compile(node.Right)
			if err != nil {
				return err
			}

			err = c.Compile(node.Left)
			if err != nil {
				return err
			}
			c.emitInstruction(code.OpGreaterThan)
			return nil
		}

		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "+":
			c.emitInstruction(code.OpAdd)

		case "-":
			c.emitInstruction(code.OpSub)

		case "*":
			c.emitInstruction(code.OpMul)

		case "/":
			c.emitInstruction(code.OpDiv)

		case ">":
			c.emitInstruction(code.OpGreaterThan)

		case "==":
			c.emitInstruction(code.OpEqual)

		case "!=":
			c.emitInstruction(code.OpNotEqual)

		default:
			return fmt.Errorf("unknow operator %s", node.Operator)

		}

	case *Ast.BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	case *Ast.IFexpression:

		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		// Emit an OpJumpNotTruthy with a bogus value
		jumpNotTruthyPos := c.emitInstruction(code.OpJumpNotTruthy, 9999)

		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}

		if c.lastInstructionIsPop() {
			c.removeLastPop()
		}

		jumpPos := c.emitInstruction(code.OpJump, 9999)
		afterConsequencePos := len(c.instructions)
		c.changeOperand(jumpNotTruthyPos, afterConsequencePos)

		if node.Alternative == nil {
			c.emitInstruction(code.OpNull)

		} else {
			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}

			if c.lastInstructionIsPop() {
				c.removeLastPop()
			}

		}

		afterAlternativePos := len(c.instructions)
		c.changeOperand(jumpPos, afterAlternativePos)

	case *Ast.IntegerLiteral:
		integer := &Object.Integer{Value: node.Value}
		c.emitInstruction(code.Opconstant, c.addConstant(integer))

	case *Ast.FloatLiteral:
		float := &Object.Float{Value: node.Value}
		c.emitInstruction(code.Opconstant, c.addConstant(float))

	case *Ast.Boolean:
		if node.Value {
			c.emitInstruction(code.OpTrue)

		} else {
			c.emitInstruction(code.OpFalse)

		}

	}

	return nil

}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}
