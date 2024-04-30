package Parser

import (
	"fmt"
	Ast "github/FabioVV/interp_lang/ast"
	Lexer "github/FabioVV/interp_lang/lexer"
	Object "github/FabioVV/interp_lang/object"
	h "github/FabioVV/interp_lang/syshelpers"
	Token "github/FabioVV/interp_lang/token"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	LOGICAL     // && or ||
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	PERIOD      // .
	CALL        // myFN(X)
	INDEX       // ARRAY[INDEX]
)

// PRECEDENCE TABLE
var precedences = map[Token.TokenType]int{
	Token.AND:      LOGICAL,
	Token.OR:       LOGICAL,
	Token.EQ:       EQUALS,
	Token.NOT_EQ:   EQUALS,
	Token.LT:       LESSGREATER,
	Token.GT:       LESSGREATER,
	Token.GT_OR_EQ: LESSGREATER,
	Token.LT_OR_EQ: LESSGREATER,
	Token.PLUS:     SUM,
	Token.MINUS:    SUM,
	Token.SLASH:    PRODUCT,
	Token.ASTERISK: PRODUCT,
	Token.MODULUS:  PRODUCT,
	Token.LPAREN:   CALL,
	Token.LBRACKET: INDEX,
	Token.PERIOD:   INDEX,
}

type prefixParseFN func() Ast.Expression
type infixParseFN func(Ast.Expression) Ast.Expression

type Parser struct {
	l *Lexer.Lexer

	curToken  Token.Token
	peekToken Token.Token

	pos h.Position

	prefixParseFNS map[Token.TokenType]prefixParseFN
	infixParseFNS  map[Token.TokenType]infixParseFN

	errors []*Object.Error
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken

	p.pos, p.peekToken = p.l.NextToken()

}

func (p *Parser) Errors() []*Object.Error {
	return p.errors
}

func newError(format string, token Token.Token, a ...interface{}) *Object.Error {
	return &Object.Error{
		Message:  fmt.Sprintf(format, a...),
		Filename: token.Filename,
		Line:     token.Pos.Line,
		Column:   token.Pos.Column,
	}
}

func (p *Parser) noPrefixParseFnError(t Token.TokenType) {
	msg := newError("no prefix parse function for %s found", p.curToken, t)

	p.errors = append(p.errors, msg)
}

func (p *Parser) registerPreFix(tokenType Token.TokenType, fn prefixParseFN) {
	p.prefixParseFNS[tokenType] = fn
}
func (p *Parser) registerInfix(tokenType Token.TokenType, fn infixParseFN) {
	p.infixParseFNS[tokenType] = fn

}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) peekError(t Token.TokenType) {
	msg := newError("expected %s, got %s instead", p.curToken, t, p.peekToken.Type)

	p.errors = append(p.errors, msg)
}

func (p *Parser) currentError(t Token.TokenType) {
	msg := newError("expected current to be %s, got %s instead", p.curToken, t, p.peekToken.Type)

	p.errors = append(p.errors, msg)
}

func (p *Parser) curTokenIs(t Token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t Token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t Token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) expectCurrent(t Token.TokenType) bool {
	if p.curTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.currentError(t)
		return false
	}
}

func (p *Parser) isSemicolonOptional() bool {
	optSemicolonTokens := []Token.TokenType{
		Token.RBRACE, Token.TYPEDEF, Token.LOOP, Token.RETURN, Token.BREAK, Token.CONTINUE, Token.COMMENT, Token.MULTILINE_COMMENT,
	}

	for _, t := range optSemicolonTokens {
		if p.curTokenIs(t) {
			return true
		}
	}

	return false
}

func (p *Parser) parseVarStatement() *Ast.VarStatement {
	stmt := &Ast.VarStatement{Token: p.curToken}

	if !p.expectPeek(Token.IDENTIFIER) {
		return nil
	}

	stmt.Name = &Ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	if !p.expectPeek(Token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(Token.SEMICOLON) {
		p.nextToken()
	}

	return stmt

}

func (p *Parser) parseFNStatement() *Ast.FunctionStatement {
	stmt := &Ast.FunctionStatement{Token: p.curToken}

	if !p.expectPeek(Token.IDENTIFIER) {
		return nil
	}

	stmt.Name = &Ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	if !p.expectPeek(Token.LPAREN) {
		return nil
	}

	stmt.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(Token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt

}

func (p *Parser) parseReturnStatement() *Ast.ReturnStatement {
	stmt := &Ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(Token.SEMICOLON) {
		p.nextToken()
	}

	return stmt

}

func (p *Parser) parseBreakStatement() *Ast.BreakStatement {
	return &Ast.BreakStatement{Token: p.curToken}
}

func (p *Parser) parseContinueStatement() *Ast.ContinueStatement {
	return &Ast.ContinueStatement{Token: p.curToken}
}

func (p *Parser) parseExpression(precedence int) Ast.Expression {

	/*THE MAP IS A HASH, WE ARE ACCESSING THE P.CURTOKEN.TYPE POSITION*/
	/*if it returns nil, means there is no function for that current token type in the hash*/
	/*If there is, call it on leftEXp*/
	/*Same thing is done for infix*/

	prefix := p.prefixParseFNS[p.curToken.Type]

	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}

	leftExp := prefix()

	//peekPrecedence LEFT BINDING POWER
	for !p.peekTokenIs(Token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFNS[p.peekToken.Type]

		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseExpressionStatement() *Ast.ExpressionStatement {
	stmt := &Ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)

	/*we check for
	an optional semicolon. Yes, it’s optional. If the peekToken is a token.SEMICOLON, we advance so
	it’s the curToken. If it’s not there, that’s okay too, we don’t add an error to the parser if it’s
	not there. That’s because we want expression statements to have optional semicolons (which
	makes it easier to type something like 5 + 5 into the REPL later on).~
	*/
	if p.peekTokenIs(Token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseStatement() Ast.Statement {

	var stmt Ast.Statement

	switch p.curToken.Type {

	case Token.VAR:
		stmt = p.parseVarStatement()

	case Token.FUNCTION:
		stmt = p.parseFNStatement()

	case Token.RETURN:
		stmt = p.parseReturnStatement()

	case Token.BREAK:
		stmt = p.parseBreakStatement()

	case Token.CONTINUE:
		stmt = p.parseContinueStatement()

	default:
		stmt = p.parseExpressionStatement()

	}

	if !p.isSemicolonOptional() {
		if !p.curTokenIs(Token.SEMICOLON) {
			msg := newError("Syntax error - Expected semicolon (;) at the end of the statement %s", p.curToken, stmt.String())
			p.errors = append(p.errors, msg)

			return nil
		}
	} else {
		// Optionally skip semicolon after certain tokens where it's unnecessary but not an error
		if p.peekTokenIs(Token.SEMICOLON) {
			p.nextToken() // Consume the semicolon without generating an error
		}
	}

	return stmt
}

func (p *Parser) ParseProgram() *Ast.Program {
	program := &Ast.Program{}

	program.Statements = []Ast.Statement{}

	for !p.curTokenIs(Token.EOF) {
		stmt := p.parseStatement()

		if stmt != nil {
			program.Statements = append(program.Statements, stmt)

		}

		p.nextToken()
	}

	return program

}

func (p *Parser) parseCompoundExpression() Ast.Expression {
	exp := &Ast.CompoundAssignExpression{Left: &Ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}}

	p.nextToken() // Consumes the += or -= or /= or *= token
	exp.Token = p.curToken
	p.nextToken()

	exp.Value = p.parseExpression(LOWEST)

	return exp
}

func (p *Parser) parseIncDecExpression() Ast.Expression {
	exp := &Ast.IncDecExpression{Identifier: &Ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}}

	p.nextToken() // Consumes the ++ or -- token
	exp.Token = p.curToken

	return exp
}

func (p *Parser) parseIndentifier() Ast.Expression {

	if p.curTokenIs(Token.IDENTIFIER) && p.peekTokenIs(Token.IDENTIFIER) {
		// TODO: Find a better way of doing this.
		// Parsing of a struct typedef: typedef followed by a indentifier
		// ex:
		/*
			typedef Point = {
				X;
				Y;
			}

			Point a = {X:5, Y:5} <<<<<<<<<<< (currently parsing this)
		*/
		typef_statement := &Ast.TypeDefStatement{Token: p.curToken}

		p.nextToken()

		typef_statement.Name = &Ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

		if !p.expectPeek(Token.ASSIGN) {
			return nil
		}
		p.nextToken()

		typef_statement.Value = p.parseExpression(LOWEST)

		return typef_statement
	}

	switch p.peekToken.Literal {
	case Token.ASSIGN:
		exp := &Ast.AssignExpression{Left: &Ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}}

		p.nextToken() // Consumes the = token
		exp.Token = p.curToken
		p.nextToken()

		exp.Value = p.parseExpression(LOWEST)

		return exp

	case Token.INC, Token.DEC:
		return p.parseIncDecExpression()

	case Token.PLUS_ASSIGN, Token.MULT_ASSIGN, Token.MINUS_ASSIGN, Token.DIV_ASSIGN:
		return p.parseCompoundExpression()

	}

	return &Ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() Ast.Expression {
	lit := &Ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)

	if err != nil {
		msg := newError("Could not parse %q as integer", p.curToken, p.curToken.Literal)

		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseFloatLiteral() Ast.Expression {
	lit := &Ast.FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)

	if err != nil {
		msg := newError("Could not parse %q as float", p.curToken, p.curToken.Literal)

		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseBoolean() Ast.Expression {
	return &Ast.Boolean{Token: p.curToken, Value: p.curTokenIs(Token.TRUE)}
}

func (p *Parser) parseGroupedExpression() Ast.Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(Token.RPAREN) {
		return nil
	}

	return exp
}

/*

For token.BANG and token.MINUS we register the same method as prefixParseFn: the
newly created parsePrefixExpression. This method builds an AST node, in this case
*ast.PrefixExpression, just like the parsing functions we saw before. But then it does
something different: it actually advances our tokens by calling p.nextToken()!


*/

func (p *Parser) parsePrefixExpression() Ast.Expression {
	expression := &Ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left Ast.Expression) Ast.Expression {
	expression := &Ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression

	// dot? <exp>.<exp>

}

func (p *Parser) parseBlockStatement() *Ast.BlockStatement {
	block := &Ast.BlockStatement{Token: p.curToken}

	block.Statements = []Ast.Statement{}

	p.nextToken()
	for !p.curTokenIs(Token.RBRACE) && !p.curTokenIs(Token.EOF) {
		stmt := p.parseStatement()

		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}

		p.nextToken()
	}

	return block
}

/*
The whole part of this method is constructed in a way that allows an
optional else but doesn’t add a parser error if there is none. After we parse the consequence block-statement we check if the next token is a token.ELSE token. Remember, at the end of
parseBlockStatement we’re sitting on the } */

func (p *Parser) parseIFexpression() Ast.Expression {

	expression := &Ast.IFexpression{Token: p.curToken}

	if !p.expectPeek(Token.LPAREN) {
		return nil
	}

	p.nextToken()

	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(Token.RPAREN) {
		return nil
	}

	if !p.expectPeek(Token.LBRACE) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(Token.ELSE) {
		p.nextToken()

		if !p.expectPeek(Token.LBRACE) {
			return nil
		}

		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

func (p *Parser) parseLOOPexpression() Ast.Expression {
	expression := &Ast.LoopExpression{Token: p.curToken}

	if !p.expectPeek(Token.LBRACE) {
		return nil
	}

	expression.Body = p.parseBlockStatement()

	return expression
}

func (p *Parser) parseFORexpression() Ast.Expression {
	expression := &Ast.FORexpression{Token: p.curToken}

	if !p.expectPeek(Token.LPAREN) {
		return nil
	}

	p.nextToken()

	if !p.curTokenIs(Token.VAR) {
		return nil
	}

	expression.LoopVariable = p.parseVarStatement()

	p.nextToken()

	expression.LoopCondition = p.parseExpression(LOWEST)

	if !p.expectPeek(Token.SEMICOLON) {
		return nil
	}

	p.nextToken()

	switch p.peekToken.Literal {

	case Token.INC, Token.DEC:
		expression.LoopStep = p.parseIncDecExpression()

	case Token.PLUS_ASSIGN, Token.MULT_ASSIGN, Token.MINUS_ASSIGN, Token.DIV_ASSIGN:
		expression.LoopStep = p.parseCompoundExpression()

	default:
		expression.LoopStep = p.parseExpression(LOWEST)

	}

	//expression.LoopStep = p.parseExpression(LOWEST)

	if !p.peekTokenIs(Token.LBRACE) {
		p.nextToken()

	}

	p.nextToken()

	expression.Body = p.parseBlockStatement()

	return expression
}

func (p *Parser) parseFunctionParameters() []*Ast.Identifier {

	identifiers := []*Ast.Identifier{}

	if p.peekTokenIs(Token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	ident := &Ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(Token.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &Ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(Token.RPAREN) {
		return nil
	}

	return identifiers

}

func (p *Parser) parseFunctionLiteral() Ast.Expression {

	lit := &Ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(Token.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(Token.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit

}

func (p *Parser) parseTypedefLiteral() Ast.Expression {
	typedef_exp := &Ast.TypeDef{Token: p.curToken}

	if !p.peekTokenIs(Token.IDENTIFIER) {
		return nil
	}

	p.nextToken()

	typedef_exp.Name = &Ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	if !p.peekTokenIs(Token.LBRACE) {
		return nil
	}

	p.nextToken()

	typedef_exp.Pairs = make(map[string]Ast.Expression)

	if p.peekTokenIs(Token.RBRACE) {
		p.nextToken()
		return typedef_exp
	}

	p.nextToken()

	for p.peekTokenIs(Token.COLON) {

		key := p.parseStringLiteral()

		if !p.expectPeek(Token.COLON) {
			return nil
		}

		p.nextToken()

		value := p.parseExpression(LOWEST)

		p.nextToken()

		typedef_exp.Pairs[key.TokenLiteral()] = value

		if p.curToken.Literal == Token.RBRACE {
			continue
		}

		p.nextToken()

	}

	return typedef_exp
}

func (p *Parser) parseCallExpression(function Ast.Expression) Ast.Expression {

	exp := &Ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(Token.RPAREN)

	return exp
}

func (p *Parser) parseStringLiteral() Ast.Expression {
	return &Ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseExpressionList(end Token.TokenType) []Ast.Expression {
	list := []Ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(Token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) parseArrayLiteral() Ast.Expression {
	array := &Ast.ArrayLiteral{Token: p.curToken}

	array.Elements = p.parseExpressionList(Token.RBRACKET)

	return array
}

func (p *Parser) parseIndexExpression(left Ast.Expression) Ast.Expression {
	exp_index := &Ast.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()
	exp_index.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(Token.RBRACKET) {
		return nil
	}

	if p.peekTokenIs(Token.ASSIGN) {
		p.nextToken()

		exp := &Ast.AssignIndexExpression{Token: p.curToken, Left: left}
		exp.Index = exp_index.Index

		p.nextToken()

		exp.Value = p.parseExpression(LOWEST)
		return exp
	}

	return exp_index
}

func (p *Parser) parseHashLiteral() Ast.Expression {
	hash := &Ast.HashLiteral{Token: p.curToken}

	hash.Pairs = make(map[Ast.Expression]Ast.Expression)

	for !p.peekTokenIs(Token.RBRACE) {
		p.nextToken()

		key := p.parseExpression(LOWEST)

		if !p.expectPeek(Token.COLON) {
			return nil
		}

		p.nextToken()
		value := p.parseExpression(LOWEST)

		hash.Pairs[key] = value

		if !p.peekTokenIs(Token.RBRACE) && !p.expectPeek(Token.COMMA) {
			return nil
		}
	}

	if !p.expectPeek(Token.RBRACE) {
		return nil
	}

	return hash
}

func (p *Parser) parseLoadExpression() Ast.Expression {

	load_exp := &Ast.LoadExpression{Token: p.curToken}

	if !p.expectPeek(Token.STRING) {
		return nil
	}

	load_exp.File = p.parseStringLiteral()

	return load_exp

}

func (p *Parser) parseMultiLineComment() Ast.Expression {
	return &Ast.MultiLineComment{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseComment() Ast.Expression {
	return &Ast.Comment{Token: p.curToken, Value: p.curToken.Literal}
}

func New(l *Lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []*Object.Error{}}

	p.prefixParseFNS = make(map[Token.TokenType]prefixParseFN)
	p.infixParseFNS = make(map[Token.TokenType]infixParseFN)

	p.registerPreFix(Token.IDENTIFIER, p.parseIndentifier)

	p.registerPreFix(Token.INT, p.parseIntegerLiteral)
	p.registerPreFix(Token.BANG, p.parsePrefixExpression)
	p.registerPreFix(Token.MINUS, p.parsePrefixExpression)
	p.registerPreFix(Token.TRUE, p.parseBoolean)
	p.registerPreFix(Token.FALSE, p.parseBoolean)
	p.registerPreFix(Token.LPAREN, p.parseGroupedExpression)
	p.registerPreFix(Token.IF, p.parseIFexpression)

	// NEW
	p.registerPreFix(Token.FOR, p.parseFORexpression)
	p.registerPreFix(Token.LOOP, p.parseLOOPexpression)
	p.registerPreFix(Token.MULTILINE_COMMENT, p.parseMultiLineComment)
	p.registerPreFix(Token.COMMENT, p.parseComment)
	p.registerPreFix(Token.FLOAT, p.parseFloatLiteral)
	p.registerPreFix(Token.LOAD, p.parseLoadExpression)
	// NEW

	// FINISH THIS
	p.registerPreFix(Token.TYPEDEF, p.parseTypedefLiteral)
	// FINISH THIS

	// NEW
	p.registerPreFix(Token.FUNCTION, p.parseFunctionLiteral)
	p.registerPreFix(Token.STRING, p.parseStringLiteral)
	p.registerPreFix(Token.LBRACKET, p.parseArrayLiteral)
	p.registerPreFix(Token.LBRACE, p.parseHashLiteral)
	// NEW

	// NEW
	p.registerInfix(Token.MODULUS, p.parseInfixExpression)
	p.registerInfix(Token.AND, p.parseInfixExpression)
	p.registerInfix(Token.OR, p.parseInfixExpression)
	p.registerInfix(Token.GT_OR_EQ, p.parseInfixExpression)
	p.registerInfix(Token.LT_OR_EQ, p.parseInfixExpression)
	p.registerInfix(Token.PERIOD, p.parseInfixExpression)
	// NEW

	p.registerInfix(Token.PLUS, p.parseInfixExpression)
	p.registerInfix(Token.MINUS, p.parseInfixExpression)
	p.registerInfix(Token.SLASH, p.parseInfixExpression)
	p.registerInfix(Token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(Token.EQ, p.parseInfixExpression)

	p.registerInfix(Token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(Token.LT, p.parseInfixExpression)
	p.registerInfix(Token.GT, p.parseInfixExpression)

	p.registerInfix(Token.LPAREN, p.parseCallExpression)
	p.registerInfix(Token.LBRACKET, p.parseIndexExpression)

	// Read two tokens, so curToken and peekToken are both set

	p.nextToken()
	p.nextToken()

	return p
}
