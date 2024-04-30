package Object

import (
	"bytes"
	"fmt"
	Ast "github/FabioVV/interp_lang/ast"
	Token "github/FabioVV/interp_lang/token"
	"hash/fnv"
	"strings"
)

type ObjectType string
type BuiltInFunction func(Token Token.Token, args ...Object) Object
type LibFunction interface{}

const (
	INTEGER_OBJ      = "INTEGER"
	FLOAT_OBJ        = "FLOAT"
	STRING_OBJ       = "STRING"
	ARRAY_OBJ        = "ARRAY"
	HASH_OBJ         = "HASH"
	BOOLEAN_OBJ      = "BOOLEAN"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	BREAK_OBJ        = "BREAK"
	CONTINUE_OBJ     = "CONTINUE"
	ERROR_OBJ        = "ERROR"
	WARNING_OBJ      = "WARNING"
	TYPE_OBJ         = "TYPE_OBJ"
	TYPE_DEF_OBJ     = "TYPEDEF"
	FUNCTION_OBJ     = "FUNCTION"
	BUILTIN_OBJ      = "BUILTIN"
	LIB_OBJ          = "LIB_FN"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Error struct {
	Message  string
	Filename string
	Line     int
	Column   int
}

type Warning struct {
	Message  string
	Filename string
	Line     int
	Column   int
}

type Integer struct {
	Value int64
}

type Float struct {
	Value float64
}

type String struct {
	Value string
}

type Boolean struct {
	Value bool
}

type Array struct {
	Elements []Object
}

type HashPair struct {
	Key   Object
	Value Object
}

type TypeDef struct {
	Name       string
	Attributes *Hash
}

type HashKey struct {
	Type  ObjectType
	Value uint64
}

type Hash struct {
	Pairs map[HashKey]HashPair
}

type Hashable interface {
	HashKey() HashKey
}

type Type struct {
	Type_obj string
}

type Function struct {
	Parameters []*Ast.Identifier
	Body       *Ast.BlockStatement
	Env        *Enviroment
}

type Builtin struct {
	Fn BuiltInFunction
}

type Lib struct {
	Fn LibFunction
}

type ReturnValue struct {
	Value Object
}

type BreakValue struct{}
type ContinueValue struct{}
type Null struct{}

var (
	NULL     = Null{}
	TRUE     = Boolean{Value: true}
	FALSE    = Boolean{Value: false}
	BREAK    = BreakValue{}
	CONTINUE = ContinueValue{}
)

/*

Granted, there’s one more thing we could do before moving on: we could optimize the perfor-
mance of the HashKey() methods by caching their return values, but that sounds like a nice
exercise for the performance-minded reader.


There is still a possibility, albeit a small one, that different Strings with different Values result
in the same hash. That happens when the hash/fnv package generates the same integer for
different values, an event called a hash collision. Chances that we experience it are low, but it
should be noted that there are well-known techniques such as “separate chaining” and “open
addressing” to work around the problem. Implementing one of these mitigations is outside of
this book’s scope, but certainly a nice exercise for the curious reader.*/

func (b *Boolean) HashKey() HashKey {
	var value uint64

	if b.Value {
		value = 1
	} else {
		value = 0
	}

	return HashKey{Type: b.Type(), Value: value}
}

func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

func (s *String) HashKey() HashKey {

	h := fnv.New64a()
	h.Write([]byte(s.Value))

	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) Type() ObjectType { return INTEGER_OBJ }

func (i *Float) Inspect() string  { return fmt.Sprintf("%f", i.Value) }
func (i *Float) Type() ObjectType { return FLOAT_OBJ }

func (s *String) formatEscapeSequence() string {
	var builder strings.Builder

	for i := 0; i < len(s.Value); i++ {
		if s.Value[i] == '\\' && i+1 < len(s.Value) {
			switch s.Value[i+1] {
			case 'n':
				builder.WriteByte('\n')
			case 't':
				builder.WriteByte('\t')
			case 'r':
				builder.WriteByte('\r')
			case 'f':
				builder.WriteByte('\f')
			case 'v':
				builder.WriteByte('\v')
			case '\\':
				builder.WriteByte('\\')
			default:
				builder.WriteByte(s.Value[i+1])

			}
			i++
		} else {
			builder.WriteByte(s.Value[i])

		}
	}
	return builder.String()
}
func (s *String) Inspect() string {

	return s.formatEscapeSequence()
}
func (s *String) Type() ObjectType { return STRING_OBJ }

func (ao *Array) Type() ObjectType { return ARRAY_OBJ }
func (ao *Array) Inspect() string {
	var out bytes.Buffer

	elements := []string{}

	for _, e := range ao.Elements {
		elements = append(elements, e.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }

func (n *Null) Inspect() string  { return "null" }
func (n *Null) Type() ObjectType { return NULL_OBJ }

func (f *Function) Inspect() string {

	var out bytes.Buffer

	params := []string{}

	for _, p := range f.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")

	return out.String()

}
func (f *Function) Type() ObjectType { return FUNCTION_OBJ }

func (f *TypeDef) Inspect() string {

	var out bytes.Buffer

	params := []string{}

	for _, p := range f.Attributes.Pairs {
		params = append(params, p.Key.Inspect())
	}

	out.WriteString("typedef ")
	out.WriteString("{\n")
	out.WriteString(strings.Join(params, "\n"))
	out.WriteString("\n}")

	return out.String()

}
func (f *TypeDef) Type() ObjectType { return TYPE_DEF_OBJ }

func (b *Builtin) Inspect() string  { return "builtin function" }
func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }

func (l *Lib) Inspect() string  { return "library function" }
func (l *Lib) Type() ObjectType { return LIB_OBJ }

func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }
func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }

func (b *BreakValue) Inspect() string  { return BREAK_OBJ }
func (b *BreakValue) Type() ObjectType { return BREAK_OBJ }

func (c *ContinueValue) Inspect() string  { return "continue" }
func (c *ContinueValue) Type() ObjectType { return CONTINUE_OBJ }

func (h *Hash) Type() ObjectType { return HASH_OBJ }
func (h *Hash) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s",
			pair.Key.Inspect(), pair.Value.Inspect()))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}

func (t *Type) Inspect() string  { return t.Type_obj }
func (t *Type) Type() ObjectType { return ERROR_OBJ }

func (e *Warning) Inspect() string {

	if e.Filename == "" {
		e.Filename = "Unknow"
	}

	formattedError := fmt.Sprintf("Warning: %s", e.Message+"\n")
	formattedError += fmt.Sprintf(" Location: '%s', line %d, column %d", e.Filename, e.Line, e.Column)

	return formattedError

}
func (e *Warning) Type() ObjectType { return WARNING_OBJ }

func (e *Error) Inspect() string {

	formattedError := fmt.Sprintf("ERROR: %s", e.Message+"\n")
	formattedError += fmt.Sprintf(" Location: '%s', line %d, column %d", e.Filename, e.Line, e.Column)

	return formattedError

}
func (e *Error) Type() ObjectType { return ERROR_OBJ }

/*
As you can see, object.Error is really, really simple. It only wraps a string that serves as error
message. In a production-ready interpreter we’d want to attach a >>>>>>stack trace<<<<<< to such error
objects, add the line and column numbers of its origin and provide more than just a message.
That’s not so hard to do, provided that line and column numbers are attached to the tokens by
the lexer. Since our lexer doesn’t do that, to keep things simple, we only use an error message,
which still serves us a great deal by giving us some feedback and stopping execution.
*/
