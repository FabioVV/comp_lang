package Tests

import (
	Ast "github/FabioVV/comp_lang/ast"
	Token "github/FabioVV/comp_lang/token"
	"testing"
)

func TestString(t *testing.T) {
	program := &Ast.Program{
		Statements: []Ast.Statement{
			&Ast.VarStatement{
				Token: Token.Token{Type: Token.VAR, Literal: "var"},
				Name: &Ast.Identifier{
					Token: Token.Token{Type: Token.IDENTIFIER, Literal: "myVar"},
					Value: "myVar",
				},
				Value: &Ast.Identifier{
					Token: Token.Token{Type: Token.IDENTIFIER, Literal: "anotherVar"},
					Value: "anotherVar",
				},
			},
		},
	}

	if program.String() != "var myVar = anotherVar;" {
		t.Errorf("program.String() wrong, got=%q", program.String())
	}
}
