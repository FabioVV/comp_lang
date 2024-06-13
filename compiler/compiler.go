package compiler

import (
	"fmt"
	ast "github/FabioVV/comp_lang/ast"
	"github/FabioVV/comp_lang/code"
	object "github/FabioVV/comp_lang/object"
	"sort"
)

type EmittedInstruction struct {
	Opcode code.Opcode
	Pos    int
}

type CompilationScope struct {
	instructions        code.Instructions
	lastInstruction     EmittedInstruction
	PreviousInstruction EmittedInstruction
}

type Compiler struct {
	constants []object.Object

	symbolTable *SymbolTable

	scopes     []CompilationScope
	scopeIndex int
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

func New() *Compiler {
	//Main scope of our compilation
	mainScope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		PreviousInstruction: EmittedInstruction{},
	}

	return &Compiler{
		constants:   []object.Object{},
		symbolTable: NewSymbolTable(),
		scopes:      []CompilationScope{mainScope},
		scopeIndex:  0,
	}
}

func NewWithState(s *SymbolTable, constants []object.Object) *Compiler {
	compiler := New()
	compiler.symbolTable = s
	compiler.constants = constants
	return compiler
}

// Add constant to the constants pool and return its index (position in the pool)
func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) CurrentInstructions() code.Instructions {
	return c.scopes[c.scopeIndex].instructions
}

// add a new instruction to the instructions slice and return the position where the current
// instruction starts
func (c *Compiler) addInstruction(instruction []byte) int {
	posNewInstruction := len(c.CurrentInstructions())
	updatedInstructions := append(c.CurrentInstructions(), instruction...)

	c.scopes[c.scopeIndex].instructions = updatedInstructions

	return posNewInstruction
}

// We need this method to keep track of the last two emitted instructions
func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	previous := c.scopes[c.scopeIndex].lastInstruction
	last := EmittedInstruction{Opcode: op, Pos: pos}

	c.scopes[c.scopeIndex].PreviousInstruction = previous
	c.scopes[c.scopeIndex].lastInstruction = last
}

// We use code.make inside of emitInstruction to generate the instruction
// we then add it to instructions slice and return its starting position
func (c *Compiler) emitInstruction(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setLastInstruction(op, pos)

	return pos
}

func (c *Compiler) lastInstructionIs(op code.Opcode) bool {
	if len(c.CurrentInstructions()) == 0 {
		return false
	}

	return c.scopes[c.scopeIndex].lastInstruction.Opcode == op
}

func (c *Compiler) removeLastPop() {
	last := c.scopes[c.scopeIndex].lastInstruction
	previous := c.scopes[c.scopeIndex].PreviousInstruction

	old := c.CurrentInstructions()
	new := old[:last.Pos]

	c.scopes[c.scopeIndex].instructions = new
	c.scopes[c.scopeIndex].lastInstruction = previous

}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {

	ins := c.CurrentInstructions()

	for i := 0; i < len(newInstruction); i++ {
		ins[pos+1] = newInstruction[i]
	}
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := code.Opcode(c.CurrentInstructions()[opPos])
	newInstruction := code.Make(op, operand)

	c.replaceInstruction(opPos, newInstruction)
}

func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		PreviousInstruction: EmittedInstruction{},
	}

	c.scopes = append(c.scopes, scope)
	c.scopeIndex++
}

func (c *Compiler) leaveScope() code.Instructions {
	instructions := c.CurrentInstructions()

	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--

	return instructions
}

func (c *Compiler) replaceLastPopWithReturn() {
	lastPos := c.scopes[c.scopeIndex].lastInstruction.Pos
	c.replaceInstruction(lastPos, code.Make(code.OpReturnValue))

	c.scopes[c.scopeIndex].lastInstruction.Opcode = code.OpReturnValue
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {

	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)

			if err != nil {
				return err
			}
		}

	case *ast.ReturnStatement:
		if err := c.Compile(node.ReturnValue); err != nil {
			return err
		}

		c.emitInstruction(code.OpReturnValue)

	case *ast.CallExpression:

		if err := c.Compile(node.Function); err != nil {
			return err
		}
		c.emitInstruction(code.OpCall)

	case *ast.FunctionLiteral:
		c.enterScope()

		if err := c.Compile(node.Body); err != nil {
			return err
		}

		if c.lastInstructionIs(code.OpPop) {
			c.replaceLastPopWithReturn()
		}

		if !c.lastInstructionIs(code.OpReturnValue) {
			c.emitInstruction(code.OpReturn)
		}

		instructions := c.leaveScope()

		compiledFunction := &object.CompiledFunction{Instructions: instructions}

		c.emitInstruction(code.Opconstant, c.addConstant(compiledFunction))

	case *ast.VarStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		symbol := c.symbolTable.Define(node.Name.Value)
		c.emitInstruction(code.OpSetGlobal, symbol.Index)

	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			// Compile time error! Very cool.
			return fmt.Errorf("undefined variable %s", node.Value)
		}

		c.emitInstruction(code.OpGetGlobal, symbol.Index)

	case *ast.ExpressionStatement:
		if err := c.Compile(node.Expression); err != nil {
			return err

		} else {
			c.emitInstruction(code.OpPop)

		}

	case *ast.PrefixExpression:
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

	case *ast.InfixExpression:
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

	case *ast.BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	case *ast.IFexpression:

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

		if c.lastInstructionIs(code.OpPop) {
			c.removeLastPop()
		}

		jumpPos := c.emitInstruction(code.OpJump, 9999)
		afterConsequencePos := len(c.CurrentInstructions())
		c.changeOperand(jumpNotTruthyPos, afterConsequencePos)

		if node.Alternative == nil {
			c.emitInstruction(code.OpNull)

		} else {
			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}

			if c.lastInstructionIs(code.OpPop) {
				c.removeLastPop()
			}

		}

		afterAlternativePos := len(c.CurrentInstructions())
		c.changeOperand(jumpPos, afterAlternativePos)

	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emitInstruction(code.Opconstant, c.addConstant(integer))

	case *ast.FloatLiteral:
		float := &object.Float{Value: node.Value}
		c.emitInstruction(code.Opconstant, c.addConstant(float))

	case *ast.StringLiteral:
		str := &object.String{Value: node.Value}
		c.emitInstruction(code.Opconstant, c.addConstant(str))

	case *ast.ArrayLiteral:
		for _, el := range node.Elements {
			if err := c.Compile(el); err != nil {
				return err
			}
		}

		c.emitInstruction(code.OpArray, len(node.Elements))

	case *ast.HashLiteral:
		keys := []ast.Expression{}

		for k := range node.Pairs {
			keys = append(keys, k)
		}

		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[i].String()
		})

		for _, k := range keys {
			if err := c.Compile(k); err != nil {
				return err
			}

			if err := c.Compile(node.Pairs[k]); err != nil {
				return err
			}
		}

		c.emitInstruction(code.OpHash, len(node.Pairs))

	case *ast.IndexExpression:

		if err := c.Compile(node.Left); err != nil {
			return err
		}

		if err := c.Compile(node.Index); err != nil {
			return err
		}

		c.emitInstruction(code.OpIndex)

	case *ast.Boolean:
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
		Instructions: c.CurrentInstructions(),
		Constants:    c.constants,
	}
}
