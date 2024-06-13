package Ast

import (
	"bytes"
	Token "github/FabioVV/comp_lang/token"
	"strings"
)

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

type BlockStatement struct {
	Token      Token.Token
	Statements []Statement
}

type Identifier struct {
	Token Token.Token
	Value string
}

type Boolean struct {
	Token Token.Token
	Value bool
}

type LoadExpression struct {
	Token Token.Token
	File  Expression
}

type TypeDef struct {
	Token Token.Token // The 'typedef' token
	Name  *Identifier
	Pairs map[string]Expression
}

type VarStatement struct {
	Token Token.Token
	Name  *Identifier
	Value Expression
}

// Point a = {}
type TypeDefStatement struct {
	Token Token.Token
	Name  *Identifier
	Value Expression
}

type ReturnStatement struct {
	Token       Token.Token
	ReturnValue Expression
}

type BreakStatement struct {
	Token Token.Token
}

type ContinueStatement struct {
	Token Token.Token
}

type ExpressionStatement struct {
	Token      Token.Token
	Expression Expression
}

type IntegerLiteral struct {
	Token Token.Token
	Value int64
}

type FloatLiteral struct {
	Token Token.Token
	Value float64
}

type StringLiteral struct {
	Token Token.Token
	Value string
}

type IFexpression struct {
	Token       Token.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

type FORexpression struct {
	Token         Token.Token     // FOR
	LoopVariable  *VarStatement   //(VAR  I = 0)
	LoopCondition Expression      // (i < number)
	LoopStep      Expression      //(I - 1)
	Body          *BlockStatement //{...}

}

type LoopExpression struct {
	Token Token.Token     // LOOP
	Body  *BlockStatement //{...}
}

type MultiLineComment struct {
	Token Token.Token
	Value string
}

type Comment struct {
	Token Token.Token
	Value string
}

type FunctionLiteral struct {
	Token      Token.Token // The 'fn' token
	Parameters []*Identifier
	Body       *BlockStatement
}

type FunctionStatement struct {
	Token      Token.Token // The 'fn' token
	Name       *Identifier
	Parameters []*Identifier
	Body       *BlockStatement
}

type ArrayLiteral struct {
	Token    Token.Token
	Elements []Expression
}

type HashLiteral struct {
	Token Token.Token // the '{' token
	Pairs map[Expression]Expression
}

type IndexExpression struct {
	Token Token.Token // The [ token
	Left  Expression
	Index Expression
}

type AssignExpression struct {
	Token Token.Token // The = token
	Left  *Identifier
	Value Expression
}

type CompoundAssignExpression struct {
	Token Token.Token // The += or -= or /= or *= token
	Left  *Identifier
	Value Expression
}

type AssignIndexExpression struct {
	Token Token.Token // The = token
	Index Expression
	Left  Expression
	Value Expression
}

type PrefixExpression struct {
	Token    Token.Token
	Operator string
	Right    Expression
}

type InfixExpression struct {
	Token    Token.Token // The operator token, e.g. *
	Right    Expression
	Operator string
	Left     Expression
}

type IncDecExpression struct {
	Token      Token.Token //  ++ or --
	Identifier *Identifier
}

type CallExpression struct {
	Token     Token.Token // The '(' token
	Function  Expression  // Identifier or FunctionLiteral
	Arguments []Expression
}

func (ide *IncDecExpression) expressionNode()      {}
func (ide *IncDecExpression) TokenLiteral() string { return ide.Token.Literal }
func (ide *IncDecExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ide.Identifier.String())
	out.WriteString(ide.TokenLiteral())
	out.WriteString(")")

	return out.String()
}

func (c *ContinueStatement) statementNode()       {}
func (c *ContinueStatement) TokenLiteral() string { return c.Token.Literal }
func (c *ContinueStatement) String() string       { return c.Token.Literal }

func (bs *BreakStatement) statementNode()       {}
func (bs *BreakStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BreakStatement) String() string       { return bs.Token.Literal }

func (le *LoadExpression) expressionNode()      {}
func (le *LoadExpression) TokenLiteral() string { return le.Token.Literal }
func (le *LoadExpression) String() string       { return le.Token.Literal }

func (mc *MultiLineComment) expressionNode()      {}
func (mc *MultiLineComment) TokenLiteral() string { return mc.Token.Literal }
func (mc *MultiLineComment) String() string       { return mc.Token.Literal }

func (c *Comment) expressionNode()      {}
func (c *Comment) TokenLiteral() string { return c.Token.Literal }
func (c *Comment) String() string       { return c.Token.Literal }

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }
func (b *Boolean) String() string       { return b.Token.Literal }

func (ls *VarStatement) statementNode()       {}
func (ls *VarStatement) TokenLiteral() string { return ls.Token.Literal }

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return sl.Token.Literal }

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString("(")

	return out.String()
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(ie.Operator + "")
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	return out.String()
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {

	var out bytes.Buffer

	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

func (ie *IFexpression) expressionNode()      {}
func (ie *IFexpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IFexpression) String() string {

	var out bytes.Buffer

	out.WriteString("if")
	out.WriteString(ie.Condition.String())
	out.WriteString(" ")
	out.WriteString(ie.Consequence.String())

	if ie.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(ie.Alternative.String())
	}

	return out.String()
}

func (ie *FORexpression) expressionNode()      {}
func (ie *FORexpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *FORexpression) String() string {

	var out bytes.Buffer

	out.WriteString("for")
	out.WriteString(ie.LoopVariable.String())
	out.WriteString(" ")
	out.WriteString(ie.LoopCondition.String())
	out.WriteString(" ")
	out.WriteString(ie.LoopStep.String())
	out.WriteString(" ")
	out.WriteString(ie.Body.String())

	return out.String()
}

func (le *LoopExpression) expressionNode()      {}
func (le *LoopExpression) TokenLiteral() string { return le.Token.Literal }
func (le *LoopExpression) String() string {

	var out bytes.Buffer

	out.WriteString("loop")
	out.WriteString(" ")
	out.WriteString(le.Body.String())

	return out.String()
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fl.Body.String())

	return out.String()
}
func (td *TypeDef) expressionNode()      {}
func (td *TypeDef) TokenLiteral() string { return td.Token.Literal }
func (td *TypeDef) String() string {
	var out bytes.Buffer

	atts := []string{}

	for key, value := range td.Pairs {
		atts = append(atts, key+":"+value.String())
	}

	out.WriteString(td.TokenLiteral())
	out.WriteString(" ")
	out.WriteString(td.Name.String())
	out.WriteString("{\n")
	out.WriteString(strings.Join(atts, "\n"))
	out.WriteString("\n} ")

	return out.String()
}

func (fs *FunctionStatement) statementNode()       {}
func (fs *FunctionStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *FunctionStatement) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range fs.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(fs.TokenLiteral())
	out.WriteString(" ")
	out.WriteString(fs.Name.Value)
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fs.Body.String())

	return out.String()
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer

	elements := []string{}

	for _, el := range al.Elements {
		elements = append(elements, el.String())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

func (hl *HashLiteral) expressionNode()      {}
func (hl *HashLiteral) TokenLiteral() string { return hl.Token.Literal }
func (hl *HashLiteral) String() string {
	var out bytes.Buffer

	pairs := []string{}

	for key, value := range hl.Pairs {
		pairs = append(pairs, key.String()+":"+value.String())
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	var out bytes.Buffer

	/*
			It’s important to note that both Left and Index are just Expressions. Left is the object that’s
		being accessed and we’ve seen that it can be of any type: an identifier, an array literal, a
		function call. The same goes for Index. It can be any expression. Syntactically it doesn’t make
		a difference which one it is, but semantically it has to produce an integer.*/

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")

	return out.String()
}

func (ae *AssignExpression) expressionNode()      {}
func (ae *AssignExpression) TokenLiteral() string { return ae.Token.Literal }
func (ae *AssignExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ae.Left.String())
	out.WriteString(ae.Token.Literal)
	out.WriteString(ae.Value.String())
	out.WriteString(")")

	return out.String()
}

func (pae *CompoundAssignExpression) expressionNode()      {}
func (pae *CompoundAssignExpression) TokenLiteral() string { return pae.Token.Literal }
func (pae *CompoundAssignExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pae.Left.String())
	out.WriteString(pae.Token.Literal)
	out.WriteString(pae.Value.String())
	out.WriteString(")")

	return out.String()
}

func (aie *AssignIndexExpression) expressionNode()      {}
func (aie *AssignIndexExpression) TokenLiteral() string { return aie.Token.Literal }
func (aie *AssignIndexExpression) String() string {
	var out bytes.Buffer

	/*
			It’s important to note that both Left and Index are just Expressions. Left is the object that’s
		being accessed and we’ve seen that it can be of any type: an identifier, an array literal, a
		function call. The same goes for Index. It can be any expression. Syntactically it doesn’t make
		a difference which one it is, but semantically it has to produce an integer.*/

	out.WriteString("(")
	out.WriteString(aie.Left.String())
	out.WriteString("=")
	out.WriteString(aie.Value.String())
	out.WriteString(")")

	return out.String()
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {

	var out bytes.Buffer

	args := []string{}

	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

func (tds *TypeDefStatement) expressionNode()      {}
func (tds *TypeDefStatement) TokenLiteral() string { return tds.Token.Literal }
func (tds *TypeDefStatement) String() string {
	var out bytes.Buffer

	out.WriteString(tds.TokenLiteral() + " ")
	out.WriteString(tds.Name.String())
	out.WriteString(" = ")

	if tds.Value != nil {
		out.WriteString(tds.Value.String())

	}

	out.WriteString(";")
	return out.String()
}

func (ls *VarStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")

	if ls.Value != nil {
		out.WriteString(ls.Value.String())

	}

	out.WriteString(";")
	return out.String()

}

func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")

	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}

	out.WriteString(";")

	return out.String()
}

func (es *ExpressionStatement) String() string {

	if es.Expression != nil {
		return es.Expression.String()
	}

	return ""
}
