package compiler

import (
	Ast "github/FabioVV/interp_lang/ast"
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

func (c *Compiler) Compile(node Ast.Node) error {
	return nil
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}
