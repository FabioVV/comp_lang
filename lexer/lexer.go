package Lexer

import (
	"bufio"
	"fmt"
	Object "github/FabioVV/interp_lang/object"
	h "github/FabioVV/interp_lang/syshelpers"
	Token "github/FabioVV/interp_lang/token"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Lexer struct {
	Filename string
	Input    *bufio.Reader
	Pos      h.Position
	errors   []*Object.Error
}

// Creates our lexer. Initializes the line and column position at 1, our input as a *bufio.Reader and filename
func New(reader io.Reader, Filename string) *Lexer {
	return &Lexer{Input: bufio.NewReader(reader), Pos: h.Position{Line: 1, Column: 1}, Filename: Filename}
}

func newLexerError(format string, pos h.Position, filename string, a ...interface{}) *Object.Error {
	return &Object.Error{
		Message:  fmt.Sprintf(format, a...),
		Filename: filename,
		Line:     pos.Line,
		Column:   pos.Column,
	}
}

func (l *Lexer) Errors() []*Object.Error {
	return l.errors
}

// Backup rewinds the lexer by one rune.
func (l *Lexer) Backup() *Object.Error {

	err := l.Input.UnreadRune()

	if err != nil {
		return newLexerError("Error: not able to unread last character %w", l.Pos, l.Filename, err)
	}

	l.Pos.Column--
	return nil
}

// Peek advances the lexer by one rune.
// It still consumes it, if needed you have to backup manually using l.Backup()
func (l *Lexer) peek() (rune, *Object.Error) {

	peek, _, err := l.Input.ReadRune()

	if err != nil {
		return 0, newLexerError("Error: failed to peek next character %w", l.Pos, l.Filename, err)

	}

	return peek, nil
}

// if the lexer does not match any tokens inside of the switch, it defaults to numbers or letters or ilegal.
// if it finds a letter as defined in unicode.IsLetter, this functions reads the complete indentifier, if the next token is not letter anymore,
// it returns the identifier read
func (l *Lexer) ReadIdentifier() (string, *Object.Error) {

	var literal strings.Builder

	for {
		r, _, err := l.Input.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", newLexerError("Error: failed to read rune %w", l.Pos, l.Filename, err)
		}

		l.Pos.Column++

		if unicode.IsLetter(r) || r == '_' {
			literal.WriteRune(r)

		} else {
			l.Backup()
			break

		}
	}
	return literal.String(), nil

}

// if the lexer finds a ", this functions reads until the next " is encountered, if it reaches EOF the string that was read until that point is returned
func (l *Lexer) readString() (string, bool) {

	var str strings.Builder

	for {
		r, _, err := l.Input.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", false
		}

		l.Pos.Column++

		if r != '"' && r != '\x00' {
			str.WriteRune(r)
		} else {
			break

		}
	}

	return str.String(), true

}

// Read`s all the content between /*  */ (multiLine comment)
func (l *Lexer) readMultiLineComment() (string, *Object.Error) {
	var str strings.Builder

	for {
		r, _, err := l.Input.ReadRune()
		if err != nil {
			if err == io.EOF {
				return str.String(), nil
			}
		}

		if r == '\n' {
			l.Pos.Line++
			l.Pos.Column = 0
		} else {
			l.Pos.Column++
		}

		if r != '*' {
			str.WriteRune(r)

		} else {
			peekChar, err := l.peek()

			if err != nil {
				return "", err
			}

			if peekChar == '/' {
				return str.String(), nil
			} else {
				l.Backup()
			}

			continue
		}

	}
}

// Read`s all the content from // until a new Line is found \n (single Line comment)
func (l *Lexer) readComment() (string, *Object.Error) {
	var str strings.Builder

	for {
		r, _, err := l.Input.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", newLexerError("Error: failed to read rune %w", l.Pos, l.Filename, err)
		}

		l.Pos.Column++

		if r != '\n' {
			str.WriteRune(r)

		} else {
			l.Pos.Column = 0
			l.Pos.Line++
			break

		}

	}
	return str.String(), nil
}

// Called at the start of nextToken.
// Removes all whitespace as defined in unicode.IsSpace from the current file being read
func (l *Lexer) skipWhitespace() {
	for {
		r, _, err := l.Input.ReadRune()
		if err != nil {
			return
		}

		if r == '\n' {
			l.Pos.Line++
			l.Pos.Column = 0
		} else {

			l.Pos.Column += utf8.RuneLen(r)
		}

		if !unicode.IsSpace(r) {
			l.Input.UnreadRune() // Put back the non-space character
			break
		}

	}
}

// This functions receives a string of valid digits and then reads the next token, if the token is present in the set, returns it
func (l *Lexer) accept(valid string) (string, error) {

	var literal string

	r, _, err := l.Input.ReadRune()

	if err != nil {
		if err == io.EOF {
			return literal, nil
		}
	}

	l.Pos.Column++

	if strings.IndexRune(valid, r) >= 0 {
		literal += string(r)
		return literal, nil

	} else {
		l.Backup()
	}

	return "", nil
}

// This functions receives a string of valid digits and then keeps consuming runes if they they are present in the valid set
// EX: l.acceptRun("0123456789") -> if called, it's going to keep reading the next runes until one that is not in the 0-9 is read
func (l *Lexer) acceptRun(valid string) (string, error) {
	var literal string

	for {
		r, _, err := l.Input.ReadRune()
		if err != nil {
			if err == io.EOF {
				return literal, nil
			}
		}

		l.Pos.Column++

		if strings.IndexRune(valid, r) >= 0 {
			literal += string(r)
		} else {
			l.Backup()
			break
		}
	}

	return literal, nil
}

// Peaks the next rune and returns a bool indicating if it is a letter or number
// * Based on the unicode.IsLetter and unicode.IsDigit
func (l *Lexer) isAlphanumeric() (*Object.Error, bool) {
	r, err := l.peek()

	if err != nil {
		return err, false
	}

	l.Backup()
	return nil, unicode.IsLetter(r) || unicode.IsDigit(r)
}

// Returns a new Token
func newToken(tokenType Token.TokenType, Filename string, Line int, Column int, ch rune) Token.Token {
	return Token.Token{Type: tokenType, Pos: h.Position{Line: Line, Column: Column}, Filename: Filename, Literal: string(ch)}
}

func (l *Lexer) NextToken() (h.Position, Token.Token) {
	var tok Token.Token

	l.skipWhitespace()

	for {

		r, _, err := l.Input.ReadRune()

		if err != nil {
			if err == io.EOF {
				tok.Literal = ""
				tok.Type = Token.EOF
			}

		}

		l.Pos.Column++

		switch r {

		case '#':
			peekChar, err := l.peek()

			if err != nil {
				tok = newToken(Token.ILLEGAL, l.Filename, l.Pos.Line, l.Pos.Column, r)
				break
			}

			if unicode.IsLetter(peekChar) {
				l.Backup()

				if peekChar == 'l' {
					// I just assume its #load
					if token_name, _ := l.acceptRun("load"); token_name == "load" {

						tok = Token.Token{Type: Token.LOAD, Pos: h.Position{Line: l.Pos.Line, Column: l.Pos.Column}, Filename: l.Filename, Literal: token_name}

					}
				}

			} else {
				l.Backup()
				tok = newToken(Token.POUND, l.Filename, l.Pos.Line, l.Pos.Column, r)

			}

		case '=':
			peekChar, err := l.peek()

			if err != nil {
				tok = newToken(Token.ILLEGAL, l.Filename, l.Pos.Line, l.Pos.Column, r)
				break
			}

			if peekChar == '=' {
				literal := string(r) + string(peekChar)
				tok = Token.Token{Type: Token.EQ, Pos: h.Position{Line: l.Pos.Line, Column: l.Pos.Column}, Filename: l.Filename, Literal: literal}

			} else {
				l.Backup()
				tok = newToken(Token.ASSIGN, l.Filename, l.Pos.Line, l.Pos.Column, r)

			}

		case '!':
			peekChar, err := l.peek()

			if err != nil {
				tok = newToken(Token.ILLEGAL, l.Filename, l.Pos.Line, l.Pos.Column, r)
				break
			}

			if peekChar == '=' {
				literal := string(r) + string(peekChar)
				tok = Token.Token{Type: Token.NOT_EQ, Pos: h.Position{Line: l.Pos.Line, Column: l.Pos.Column}, Filename: l.Filename, Literal: literal}

			} else {
				l.Backup()
				tok = newToken(Token.BANG, l.Filename, l.Pos.Line, l.Pos.Column, r)

			}

		case '/':
			peekChar, err := l.peek()

			if err != nil {
				tok = newToken(Token.ILLEGAL, l.Filename, l.Pos.Line, l.Pos.Column, r)
				break
			}

			if peekChar == '/' {
				literal, _ := l.readComment()
				tok = Token.Token{Type: Token.COMMENT, Pos: h.Position{Line: l.Pos.Line, Column: l.Pos.Column}, Filename: l.Filename, Literal: literal}

			} else if peekChar == '*' {
				literal, err := l.readMultiLineComment()

				if err != nil {
					tok = newToken(Token.ILLEGAL, l.Filename, l.Pos.Line, l.Pos.Column, r)
					break
				}

				tok = Token.Token{Type: Token.MULTILINE_COMMENT, Pos: h.Position{Line: l.Pos.Line, Column: l.Pos.Column}, Filename: l.Filename, Literal: literal}

			} else if peekChar == '=' {
				literal := string(r) + string(peekChar)
				tok = Token.Token{Type: Token.DIV_ASSIGN, Pos: h.Position{Line: l.Pos.Line, Column: l.Pos.Column}, Filename: l.Filename, Literal: literal}

			} else {
				l.Backup()
				tok = newToken(Token.SLASH, l.Filename, l.Pos.Line, l.Pos.Column, r)

			}

		case '+':
			peekChar, err := l.peek()

			if err != nil {
				tok = newToken(Token.ILLEGAL, l.Filename, l.Pos.Line, l.Pos.Column, r)
				break
			}

			if peekChar == '=' {
				literal := string(r) + string(peekChar)
				tok = Token.Token{Type: Token.PLUS_ASSIGN, Pos: h.Position{Line: l.Pos.Line, Column: l.Pos.Column}, Filename: l.Filename, Literal: literal}

			} else if peekChar == '+' {

				literal := string(r) + string(peekChar)
				tok = Token.Token{Type: Token.INC, Pos: h.Position{Line: l.Pos.Line, Column: l.Pos.Column}, Filename: l.Filename, Literal: literal}

			} else {
				l.Backup()
				tok = newToken(Token.PLUS, l.Filename, l.Pos.Line, l.Pos.Column, r)

			}

		case '-':
			peekChar, err := l.peek()

			if err != nil {
				tok = newToken(Token.ILLEGAL, l.Filename, l.Pos.Line, l.Pos.Column, r)
				break
			}

			if peekChar == '=' {
				literal := string(r) + string(peekChar)
				tok = Token.Token{Type: Token.MINUS_ASSIGN, Pos: h.Position{Line: l.Pos.Line, Column: l.Pos.Column}, Filename: l.Filename, Literal: literal}

			} else if peekChar == '-' {
				literal := string(r) + string(peekChar)
				tok = Token.Token{Type: Token.DEC, Pos: h.Position{Line: l.Pos.Line, Column: l.Pos.Column}, Filename: l.Filename, Literal: literal}

			} else {
				l.Backup()
				tok = newToken(Token.MINUS, l.Filename, l.Pos.Line, l.Pos.Column, r)

			}

		case '*':
			peekChar, err := l.peek()

			if err != nil {
				tok = newToken(Token.ILLEGAL, l.Filename, l.Pos.Line, l.Pos.Column, r)
				break
			}

			if peekChar == '=' {
				literal := string(r) + string(peekChar)
				tok = Token.Token{Type: Token.MULT_ASSIGN, Pos: h.Position{Line: l.Pos.Line, Column: l.Pos.Column}, Filename: l.Filename, Literal: literal}

			} else {
				l.Backup()
				tok = newToken(Token.ASTERISK, l.Filename, l.Pos.Line, l.Pos.Column, r)

			}

		case '%':
			tok = newToken(Token.MODULUS, l.Filename, l.Pos.Line, l.Pos.Column, r)

		case '<':

			peekChar, err := l.peek()

			if err != nil {
				tok = newToken(Token.ILLEGAL, l.Filename, l.Pos.Line, l.Pos.Column, r)
				break
			}

			if peekChar == '=' {
				literal := string(r) + string(peekChar)
				tok = Token.Token{Type: Token.LT_OR_EQ, Pos: h.Position{Line: l.Pos.Line, Column: l.Pos.Column}, Filename: l.Filename, Literal: literal}

			} else {
				l.Backup()
				tok = newToken(Token.LT, l.Filename, l.Pos.Line, l.Pos.Column, r)

			}

		case '>':
			peekChar, err := l.peek()

			if err != nil {
				tok = newToken(Token.ILLEGAL, l.Filename, l.Pos.Line, l.Pos.Column, r)
				break
			}

			if peekChar == '=' {
				literal := string(r) + string(peekChar)
				tok = Token.Token{Type: Token.GT_OR_EQ, Pos: h.Position{Line: l.Pos.Line, Column: l.Pos.Column}, Filename: l.Filename, Literal: literal}

			} else {
				l.Backup()
				tok = newToken(Token.GT, l.Filename, l.Pos.Line, l.Pos.Column, r)

			}

		case ';':
			tok = newToken(Token.SEMICOLON, l.Filename, l.Pos.Line, l.Pos.Column, r)

		case '(':
			tok = newToken(Token.LPAREN, l.Filename, l.Pos.Line, l.Pos.Column, r)

		case ')':
			tok = newToken(Token.RPAREN, l.Filename, l.Pos.Line, l.Pos.Column, r)

		case ',':
			tok = newToken(Token.COMMA, l.Filename, l.Pos.Line, l.Pos.Column, r)

		case '{':
			tok = newToken(Token.LBRACE, l.Filename, l.Pos.Line, l.Pos.Column, r)

		case '}':
			tok = newToken(Token.RBRACE, l.Filename, l.Pos.Line, l.Pos.Column, r)

		case ']':
			tok = newToken(Token.RBRACKET, l.Filename, l.Pos.Line, l.Pos.Column, r)

		case '[':
			tok = newToken(Token.LBRACKET, l.Filename, l.Pos.Line, l.Pos.Column, r)

		case ':':
			tok = newToken(Token.COLON, l.Filename, l.Pos.Line, l.Pos.Column, r)

		case '.':
			tok = newToken(Token.PERIOD, l.Filename, l.Pos.Line, l.Pos.Column, r)

		case '?':
			tok = newToken(Token.QUESTION_MARK, l.Filename, l.Pos.Line, l.Pos.Column, r)

		case '|':
			peekChar, _ := l.peek()

			if peekChar == '|' {
				literal := string(r) + string(peekChar)
				tok = Token.Token{Type: Token.OR, Pos: h.Position{Line: l.Pos.Line, Column: l.Pos.Column}, Filename: l.Filename, Literal: literal}

			} else if peekChar == '=' {
				literal := string(r) + string(peekChar)
				tok = Token.Token{Type: Token.OR_ASSIGN, Pos: h.Position{Line: l.Pos.Line, Column: l.Pos.Column}, Filename: l.Filename, Literal: literal}

			} else {
				l.Backup()
				tok = newToken(Token.PIPE, l.Filename, l.Pos.Line, l.Pos.Column, r)
			}

		case '&':
			peekChar, _ := l.peek()
			if peekChar == '&' {
				literal := string(r) + string(peekChar)
				tok = Token.Token{Type: Token.AND, Pos: h.Position{Line: l.Pos.Line, Column: l.Pos.Column}, Filename: l.Filename, Literal: literal}

			} else if peekChar == '=' {
				literal := string(r) + string(peekChar)
				tok = Token.Token{Type: Token.AND_ASSIGN, Pos: h.Position{Line: l.Pos.Line, Column: l.Pos.Column}, Filename: l.Filename, Literal: literal}

			} else {
				l.Backup()
				tok = newToken(Token.AMPERSAND, l.Filename, l.Pos.Line, l.Pos.Column, r)
			}

		case '"':
			tok.Type = Token.STRING
			tok.Literal, _ = l.readString()

		case 0:
			tok.Literal = ""
			tok.Type = Token.EOF

		case '\n':
			l.Pos.Line++
			l.Pos.Column = 0

		default:
			if unicode.IsLetter(r) {

				l.Backup()

				tok.Literal, _ = l.ReadIdentifier()

				tok.Type = Token.LookupIdentifier(tok.Literal)

				tok.Pos.Column = l.Pos.Column
				tok.Pos.Line = l.Pos.Line
				tok.Filename = l.Filename
				return l.Pos, tok

			} else if unicode.IsDigit(r) || r == '.' || r == 'x' || r == 'X' {

				l.Backup()

				// The full number read
				var literal string

				digits := "0123456789"

				if digit, _ := l.accept("0"); digit != "" {
					literal += digit
				}

				if hexDigit, _ := l.accept("xX"); hexDigit != "" {
					digits = "0123456789abcdefABCDEF"
					literal += hexDigit
				}

				if nums, _ := l.acceptRun(digits); nums != "" {
					literal += nums
				}

				if period, _ := l.accept("."); period != "" {
					literal += period
				}

				if nums_f, _ := l.acceptRun(digits); nums_f != "" {
					literal += nums_f
				}

				if exponent, _ := l.accept("eE"); exponent != "" {
					literal += exponent

					if sign, _ := l.accept("+-"); sign != "" {
						literal += sign
					}

					if exponentLiteral, err := l.acceptRun("0123456789"); err == nil {
						literal += exponentLiteral
					}
				}

				// Is it imaginary?
				if imaginary, err := l.accept("i"); err != nil && imaginary != "" {
					literal += imaginary
				}

				// This is the maximum length we can go and stil be a number,
				// If there is any alphanumeric chars after this then its an error
				//THANK YOU ROB PIKE
				if _, ok := l.isAlphanumeric(); ok {
					panic(fmt.Errorf("bad number syntax: %q", string(r)))
				}

				if strings.ContainsRune(literal, '.') || strings.ContainsRune(literal, 'e') || strings.ContainsRune(literal, 'E') {
					tok.Type = Token.FLOAT
					tok.Literal = literal
				} else if strings.ContainsRune(literal, 'x') || strings.ContainsRune(literal, 'X') {
					tok.Type = Token.INT
					tok.Literal = literal
				} else {
					tok.Type = Token.INT
					tok.Literal = literal
				}

				tok.Pos.Column = l.Pos.Column
				tok.Pos.Line = l.Pos.Line
				tok.Filename = l.Filename

				return l.Pos, tok

			} else {
				tok = newToken(Token.ILLEGAL, l.Filename, l.Pos.Line, l.Pos.Column, r)

			}

		}

		return l.Pos, tok

	}

}
