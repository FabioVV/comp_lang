package Token

import h "github/FabioVV/comp_lang/syshelpers"

/*
We defined the TokenType type to be a string. That allows us to use many different values
as TokenTypes, which in turn allows us to distinguish between different types of tokens. Using
string also has the advantage of being easy to debug without a lot of boilerplate and helper
functions: we can just print a string.
*/
type TokenType string

type Token struct {
	Type     TokenType
	Pos      h.Position
	Filename string
	Literal  string
}

// Token's list
const (
	ILLEGAL           = "ILLEGAL"
	EOF               = "EOF"
	COMMENT           = "COMMENT"
	MULTILINE_COMMENT = "MULTILINE_COMMENT"

	// Identifiers + Literals
	IDENTIFIER = "IDENTIFIER" // Functions names, variables names etc
	INT        = "INT"        // 123
	STRING     = "STRING"     // "abc"
	FLOAT      = "FLOAT"      // 123.45

	// Operators
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	ASTERISK = "*"
	SLASH    = "/"
	MODULUS  = "%"

	LT       = "<"
	GT       = ">"
	EQ       = "=="
	NOT_EQ   = "!="
	LT_OR_EQ = "<="
	GT_OR_EQ = ">="

	BANG = "!"
	OR   = "||"
	AND  = "&&"

	PIPE = "|"

	INC = "++" // ? TODO: FIX
	DEC = "--" // ? TODO: FIX

	PLUS_ASSIGN  = "+="
	MINUS_ASSIGN = "-="
	MULT_ASSIGN  = "*="
	DIV_ASSIGN   = "/="

	// BIT_AND   = "&"  // TODO //  Bitwise AND.
	// BIT_OR    = "|"  // TODO //  Bitwise OR.
	// BIT_XOR   = "^"  // TODO //  Bitwise XOR (exclusive or).
	// BIT_SHL   = "<<" // TODO //  Bitwise left shift.
	// BIT_SHR   = ">>" // TODO //  Bitwise right shift.
	// BIT_CLEAR = "&^" // TODO //  Bit clear (AND NOT).

	AMPERSAND  = "&"
	OR_ASSIGN  = "|="
	AND_ASSIGN = "&="

	QUESTION_MARK = "?" // TODO
	POUND         = "#" // TODO

	// Delimiters
	COMMA     = ","
	SEMICOLON = ";"
	COLON     = ":"
	PERIOD    = "." // TODO

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"

	// Keywords
	FUNCTION           = "FUNCTION"
	FUNCTION_STATEMENT = "FUNCTION_STATEMENT"
	TYPEDEF            = "TYPEDEF"
	VAR                = "VAR"
	TRUE               = "TRUE"
	FALSE              = "FALSE"
	IF                 = "IF"
	FOR                = "FOR"
	LOOP               = "LOOP"
	ELSE               = "ELSE"
	RETURN             = "RETURN"
	BREAK              = "BREAK"
	CONTINUE           = "CONTINUE"
	LOAD               = "LOAD"
)

var keywords = map[string]TokenType{
	"fn":       FUNCTION,
	"typedef":  TYPEDEF,
	"var":      VAR,
	"true":     TRUE,
	"false":    FALSE,
	"if":       IF,
	"for":      FOR,
	"loop":     LOOP,
	"else":     ELSE,
	"return":   RETURN,
	"break":    BREAK,
	"continue": CONTINUE,
	"load":     LOAD,
}

func LookupIdentifier(ident string) TokenType {

	if KEYWORD, is_keyword := keywords[ident]; is_keyword {

		return KEYWORD
	}

	return IDENTIFIER
}
