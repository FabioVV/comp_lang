package Tests

import (
	Lexer "github/FabioVV/interp_lang/lexer"
	Token "github/FabioVV/interp_lang/token"
	"strings"
	"testing"
)

func TestNextToken(t *testing.T) {

	var input string = `
	for(var i = 0; i < 5; var i = i - 1){
		puts()
	}
	"asd"
	/*

	teste
	
	*/
	
	`

	tests := []struct {
		expectedType    Token.TokenType
		expectedLiteral string
	}{
		{Token.FOR, "for"},
		{Token.LPAREN, "("},
		{Token.VAR, "var"},
		{Token.IDENTIFIER, "i"},
		{Token.ASSIGN, "="},
		{Token.INT, "0"},
		{Token.SEMICOLON, ";"},

		{Token.IDENTIFIER, "i"},
		{Token.LT, "<"},
		{Token.INT, "5"},
		{Token.SEMICOLON, ";"},

		{Token.VAR, "var"},
		{Token.IDENTIFIER, "i"},
		{Token.ASSIGN, "="},
		{Token.IDENTIFIER, "i"},
		{Token.MINUS, "-"},
		{Token.INT, "1"},
		{Token.RPAREN, ")"},
		{Token.LBRACE, "{"},
		{Token.IDENTIFIER, "puts"},
		{Token.LPAREN, "("},
		{Token.RPAREN, ")"},
		{Token.RBRACE, "}"},
		{Token.STRING, "asd"},
		{Token.MULTILINE_COMMENT, "teste"},

		{Token.EOF, ""},
	}

	reader := strings.NewReader(input)

	l := Lexer.New(reader, "Test")

	for i, tt := range tests {
		// SEE THIS LATER
		_, tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected:%q, got:%q", i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected:%q, got:%q", i, tt.expectedLiteral, tok.Literal)
		}
	}

}
