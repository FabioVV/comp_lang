package Evaluator

import (
	"fmt"
	Ast "github/FabioVV/comp_lang/ast"
	Object "github/FabioVV/comp_lang/object"
	Token "github/FabioVV/comp_lang/token"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type IncludeTracker struct {
	includedFiles map[string]string
}

var (
	LOAD_TRACKER = &IncludeTracker{
		includedFiles: make(map[string]string),
	}
)

func (it *IncludeTracker) IsIncluded(filename string) bool {
	_, ok := it.includedFiles[filename]
	return ok
}

func (it *IncludeTracker) MarkIncluded(filename string) {
	it.includedFiles[filename] = "included"
}

var libraries = map[string]map[string]*Object.Lib{
	// "math": math.Math,
	// Add other libraries here...
}

func loadLib(libName string, node *Ast.LoadExpression, env *Object.Enviroment) Object.Object {

	libPath := filepath.Join("lib", libName, libName+".go")

	if _, err := os.Stat(libPath); err == nil {

		if lib, ok := libraries[libName]; ok {
			pairs := make(map[Object.HashKey]Object.HashPair)

			for fn_name, fn := range lib {

				var key Object.Object = &Object.String{Value: fn_name}

				hashKey, _ := key.(Object.Hashable)

				hashed := hashKey.HashKey()

				switch obj := fn.Fn.(type) {

				case func(Token.Token, ...Object.Object) Object.Object:
					pairs[hashed] = Object.HashPair{Key: key, Value: fn}

				case *Object.Float:
					pairs[hashed] = Object.HashPair{Key: key, Value: obj}

				default:
					return nil
				}

			}
			env.Set(libName, &Object.Hash{Pairs: pairs})
			return nil
		}

	} else if os.IsNotExist(err) {
		return NewError("Unknown library : %s", node.Token, node.File)

	} else if os.IsPermission(err) {
		return NewError("Permission denied trying to open file : %s", node.Token, node.File)

	}

	// Used the << to diferentiate the error
	return NewError("Unknown library : %s", node.Token, node.File)
}

func NewError(format string, token Token.Token, a ...interface{}) *Object.Error {
	return &Object.Error{
		Message:  fmt.Sprintf(format, a...),
		Filename: token.Filename,
		Line:     token.Pos.Line,
		Column:   token.Pos.Column,
	}
}

func NewWarning(format string, token Token.Token, a ...interface{}) *Object.Warning {
	return &Object.Warning{
		Message:  fmt.Sprintf(format, a...),
		Filename: token.Filename,
		Line:     token.Pos.Line,
		Column:   token.Pos.Column,
	}
}

func returnType(format string, a ...interface{}) *Object.String {
	return &Object.String{Value: fmt.Sprintf(format, a...)}
}

func isError(obj Object.Object) bool {

	if obj != nil {
		return obj.Type() == Object.ERROR_OBJ
	}

	return false
}

func nativeBoolToBooleanObject(input bool) *Object.Boolean {
	if input {
		return &Object.TRUE
	}
	return &Object.FALSE
}

func evalProgram(program *Ast.Program, env *Object.Enviroment) Object.Object {

	var result Object.Object

	for _, statement := range program.Statements {
		result = Eval(statement, env)

		switch result := result.(type) {
		case *Object.ReturnValue:
			return result.Value
		case *Object.Error:
			return result
		}
	}

	return result
}

func evalBangOPeratorExpression(right Object.Object) Object.Object {
	switch right {

	case &Object.TRUE:
		return &Object.FALSE
	case &Object.FALSE:
		return &Object.TRUE
	case &Object.NULL:
		return &Object.TRUE
	default:
		return &Object.FALSE
	}

}

func evalMinusPrefixOperatorExpression(right Object.Object, node *Ast.PrefixExpression) Object.Object {

	if right.Type() != Object.INTEGER_OBJ && right.Type() != Object.FLOAT_OBJ {
		return NewError("unknow operator : -%s", node.Token, right.Type())

	}

	switch right := right.(type) {
	case *Object.Integer:
		return &Object.Integer{Value: -right.Value}

	case *Object.Float:
		return &Object.Float{Value: -right.Value}
	default:
		return NewError("type not supported for '-' prefix : -%s", node.Token, right.Type())

	}

}

func evalPrefixExpression(operator string, right Object.Object, node *Ast.PrefixExpression) Object.Object {

	switch operator {
	case "-":
		return evalMinusPrefixOperatorExpression(right, node)
	case "!":
		return evalBangOPeratorExpression(right)

	default:
		return NewError("unknow operator : %s%s", node.Token, operator, right.Type())
	}

}

func evalFloatInfixExpression(operator string, left Object.Object, right Object.Object, node *Ast.InfixExpression) Object.Object {
	leftVal := left.(*Object.Float).Value
	rightVal := right.(*Object.Float).Value

	switch operator {
	case "+":
		return &Object.Float{Value: leftVal + rightVal}

	case "-":
		return &Object.Float{Value: leftVal - rightVal}

	case "*":
		return &Object.Float{Value: leftVal * rightVal}

	case "/":
		return &Object.Float{Value: leftVal / rightVal}

	// case "%": EDGE CASE
	// 	return &Object.Float{Value: leftVal % rightVal}

	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)

	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)

	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)

	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)

	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)

	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)

	default:

		return NewError("unknow operator : %s %s %s", node.Token, left.Type(), operator, right.Type())
	}
}

func evalIntegerInfixExpression(operator string, left Object.Object, right Object.Object, node *Ast.InfixExpression) Object.Object {
	leftVal := left.(*Object.Integer).Value
	rightVal := right.(*Object.Integer).Value

	switch operator {
	case "+":
		return &Object.Integer{Value: leftVal + rightVal}

	case "-":
		return &Object.Integer{Value: leftVal - rightVal}

	case "*":
		return &Object.Integer{Value: leftVal * rightVal}

	case "/":
		return &Object.Integer{Value: leftVal / rightVal}

	case "%":
		return &Object.Integer{Value: leftVal % rightVal}

	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)

	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)

	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)

	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)

	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)

	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)

	default:
		return NewError("unknow operator : %s %s %s", node.Token, left.Type(), operator, right.Type())

	}
}

/*
The first thing here is the check for the correct operator. If it’s the supported + we unwrap the
string objects and construct a new string that’s a concatenation of both operands.
If we want to support more operators for strings this is the place where to add them. Also, if
we want to support comparison of strings with the == and != we’d need to add this here too.
Pointer comparison doesn’t work for strings, at least not in the way we want it to: with strings
we want to compare values and not pointers
*/

func evalStringInfixExpression(operator string, left Object.Object, right Object.Object, node *Ast.InfixExpression) Object.Object {
	leftVal := left.(*Object.String).Value
	rightVal := right.(*Object.String).Value

	switch operator {
	case "+":
		var str strings.Builder

		str.WriteString(leftVal)
		str.WriteString(rightVal)

		return &Object.String{Value: str.String()}

	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)

	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)

	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)

	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)

	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)

	default:
		return NewError("unknow operator : %s %s %s", node.Token, left.Type(), operator, right.Type())

	}

}

func multiplyString(str string, repeatCount int64) Object.Object {
	var new_str strings.Builder

	for i := int64(0); i < repeatCount; i++ {
		new_str.WriteString(str)
	}

	return &Object.String{Value: new_str.String()}
}

func evalInfixExpression(operator string, left Object.Object, right Object.Object, node *Ast.InfixExpression) Object.Object {

	switch {
	case operator == "=":
		return &Object.NULL

	case left.Type() == Object.INTEGER_OBJ && right.Type() == Object.STRING_OBJ:
		repeatStr := right.(*Object.String).Value
		repeatCount := left.(*Object.Integer).Value

		return multiplyString(repeatStr, repeatCount)

	case left.Type() == Object.STRING_OBJ && right.Type() == Object.INTEGER_OBJ:
		repeatStr := left.(*Object.String).Value
		repeatCount := right.(*Object.Integer).Value

		return multiplyString(repeatStr, repeatCount)

	case left.Type() == Object.INTEGER_OBJ && right.Type() == Object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right, node)

	case left.Type() == Object.FLOAT_OBJ && right.Type() == Object.FLOAT_OBJ:
		return evalFloatInfixExpression(operator, left, right, node)

	case left.Type() == Object.STRING_OBJ && (right.Type() == Object.STRING_OBJ || right.Type() == Object.TYPE_OBJ):
		return evalStringInfixExpression(operator, left, right, node)

	case operator == "==":
		return nativeBoolToBooleanObject(left == right)

	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)

		// TODO : MAKE THESE && || AVAILABLE FOR INTEGERS AND STRINGS
	case operator == "&&":
		leftBool, ok := left.(*Object.Boolean)
		if !ok {
			return NewError("Expected boolean, got %s", node.Token, left.Type())
		}
		rightBool, ok := right.(*Object.Boolean)
		if !ok {
			return NewError("Expected boolean, got %s", node.Token, right.Type())
		}
		return nativeBoolToBooleanObject(leftBool.Value && rightBool.Value)

	case operator == "||":
		leftBool, ok := left.(*Object.Boolean)
		if !ok {
			return NewError("Expected boolean, got %s", node.Token, left.Type())
		}
		rightBool, ok := right.(*Object.Boolean)
		if !ok {
			return NewError("Expected boolean, got %s", node.Token, right.Type())
		}
		return nativeBoolToBooleanObject(leftBool.Value || rightBool.Value)

	case left.Type() != right.Type():
		return NewError("type mismatch : %s %s %s", node.Token, left.Type(), operator, right.Type())

	default:
		return NewError("unknow operator : %s %s %s", node.Token, left.Type(), operator, right.Type())

	}
}

func isTruthy(obj Object.Object) bool {
	switch obj {
	case &Object.NULL:
		return false
	case &Object.TRUE:
		return true
	case &Object.FALSE:
		return false
	default:
		return true
	}
}

func evalLOOPexpression(ie *Ast.LoopExpression, env *Object.Enviroment) Object.Object {
start:

	evalBody := Eval(ie.Body, env)

	if isError(evalBody) {
		return evalBody

	} else if evalBody != nil && evalBody.Inspect() == Object.BREAK_OBJ {
		return &Object.NULL

	} else if evalBody != nil && evalBody.Inspect() == Object.CONTINUE_OBJ {
		goto start

	}

	goto start
}

func evalFORexpression(ie *Ast.FORexpression, env *Object.Enviroment) Object.Object {

start:
	LoopCondition := Eval(ie.LoopCondition, env)
	if isError(LoopCondition) {
		return LoopCondition
	}

	if isTruthy(LoopCondition) {

		evalBody := Eval(ie.Body, env)

		if isError(evalBody) {
			return evalBody

		} else if evalBody != nil && evalBody.Inspect() == Object.BREAK_OBJ {
			return &Object.NULL

		} else if evalBody != nil && evalBody.Inspect() == Object.CONTINUE_OBJ {
			goto start

		}

		LoopStep := Eval(ie.LoopStep, env)

		if isError(LoopStep) {
			return LoopStep
		}

		env.Set(ie.LoopVariable.Name.String(), LoopStep)

		goto start

	} else {
		return &Object.NULL

	}

}

func evalIFexpression(ie *Ast.IFexpression, env *Object.Enviroment) Object.Object {
	condition := Eval(ie.Condition, env)

	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env)

	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)

	} else {
		return &Object.NULL
	}
}

func evalBlockStatement(block *Ast.BlockStatement, env *Object.Enviroment) Object.Object {
	var result Object.Object

	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if result != nil {
			rt := result.Type()

			if rt == Object.RETURN_VALUE_OBJ || rt == Object.ERROR_OBJ {
				return result

			} else if rt == Object.BREAK_OBJ {

				return &Object.BREAK

			}
		}
	}

	return result
}

func evalIdentifier(node *Ast.Identifier, env *Object.Enviroment) Object.Object {

	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin

	} else if stdout_builtin, ok := stdout_builtins[node.Value]; ok {
		return stdout_builtin

	} else if hash_builtin, ok := hash_builtins[node.Value]; ok {
		return hash_builtin

	} else if array_bultin, ok := array_builtings[node.Value]; ok {
		return array_bultin

	} else if stdin_builtin, ok := stdin_builtins[node.Value]; ok {
		return stdin_builtin
	}

	// if pkg, ok :=

	return NewError("indentifier not found : %s", node.Token, node.Value)
}

func evalExpression(exps []Ast.Expression, env *Object.Enviroment) []Object.Object {
	var result []Object.Object

	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []Object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

func unWrapRETURN(obj Object.Object) Object.Object {
	if returnValue, ok := obj.(*Object.ReturnValue); ok {
		return returnValue.Value
	}

	return obj
}

func extendFunction(fn *Object.Function, args []Object.Object) *Object.Enviroment {
	env := Object.NewEnclosedEnviroment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}

	return env
}

func applyFunction(fn Object.Object, args []Object.Object, node *Ast.CallExpression) Object.Object {

	switch fn := fn.(type) {

	case *Object.Function:
		extendedEnv := extendFunction(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unWrapRETURN(evaluated)

	case *Object.Builtin:
		return fn.Fn(node.Token, args...)

	case *Object.Lib:
		switch obj := fn.Fn.(type) {

		case func(Token.Token, ...Object.Object) Object.Object:
			return obj(node.Token, args...)

		case *Object.Float:
			return obj

		default:
			return nil
		}
		// return fn.Fn(node.Token, args...)

	default:
		return NewError("not a function: is [%s]", node.Token, fn.Type())

	}

}

func evalArrayIndexExpression(array Object.Object, index Object.Object, node *Ast.IndexExpression) Object.Object {
	arrayObj := array.(*Object.Array)

	idx := index.(*Object.Integer).Value
	max := int64(len(arrayObj.Elements) - 1)

	if idx < 0 || idx > max {
		return NewError("index out of bounds: [%d]", node.Token, idx)
	}

	return arrayObj.Elements[idx]
}

func evalHashIndexExpression(hash Object.Object, index Object.Object, node *Ast.IndexExpression) Object.Object {

	hashObject := hash.(*Object.Hash)

	key, ok := index.(Object.Hashable)

	if !ok {
		return NewError("unusable as hash key: %s", node.Token, index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]

	if !ok {
		return &Object.NULL
	}

	return pair.Value

}

func evalStringIndexExpression(str Object.Object, index Object.Object, node *Ast.IndexExpression) Object.Object {
	strObject := str.(*Object.String)

	idx := index.(*Object.Integer).Value

	max := int64(len(strObject.Value) - 1)

	if idx < 0 || idx > max {
		return NewError("index out of bounds: [%d] at STRING '%v' ", node.Token, idx, strObject.Value)
	}

	var runeAtIndex rune

	for i, r := range strObject.Value {
		if int64(i) == idx {
			runeAtIndex = r
			break
		}
	}

	var str_r = &Object.String{
		Value: string(runeAtIndex),
	}

	return str_r
}

func evalIndexExpression(left Object.Object, index Object.Object, node *Ast.IndexExpression) Object.Object {

	switch {
	case left.Type() == Object.ARRAY_OBJ && index.Type() == Object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index, node)

	case left.Type() == Object.HASH_OBJ:
		return evalHashIndexExpression(left, index, node)

	case left.Type() == Object.STRING_OBJ:
		return evalStringIndexExpression(left, index, node)

	default:
		return NewError("index operator not supported: %s", node.Token, left.Type())
	}

}

func evalHashLiteral(node *Ast.HashLiteral, env *Object.Enviroment) Object.Object {

	pairs := make(map[Object.HashKey]Object.HashPair)

	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, env)

		if isError(key) {
			return key
		}

		hashKey, ok := key.(Object.Hashable)
		if !ok {
			return NewError("unusable as hash key : %s", node.Token, key.Type())
		}

		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}

		hashed := hashKey.HashKey()
		pairs[hashed] = Object.HashPair{Key: key, Value: value}

	}

	return &Object.Hash{Pairs: pairs}
}

func evalTypeDefLiteral(node *Ast.TypeDef, env *Object.Enviroment) Object.Object {

	pairs := make(map[Object.HashKey]Object.HashPair)
	free_duplicates := make(map[string]bool)

	for key, _ := range node.Pairs {
		if _, ok := free_duplicates[key]; ok {
			return NewError("duplicate entry on typedef : %s", node.Token, key)

		} else {
			free_duplicates[key] = true
		}
	}

	for key, val := range node.Pairs {

		var op Object.Object = &Object.String{Value: key}

		hashKey, ok := op.(Object.Hashable)
		if !ok {
			return NewError("unusable as hash key : %s", node.Token, op.Type())
		}

		hashed := hashKey.HashKey()

		_val := Eval(val, env)

		if isError(_val) {
			return _val
		}

		pairs[hashed] = Object.HashPair{Key: op, Value: _val}
	}

	// This is how i can access them
	// for _, value := range pairs {

	// 	fmt.Println(value.Key.Type())
	// 	fmt.Println(value.Value.Type())

	// }

	env.Set(node.Name.Value, &Object.TypeDef{Name: node.Name.Value, Attributes: &Object.Hash{Pairs: pairs}})

	return &Object.NULL
}

func performCompoundAssignment(ie *Ast.CompoundAssignExpression, oldVal Object.Object, Val Object.Object, varName string, compound_exp string, env *Object.Enviroment) Object.Object {
	var op Object.Object

	switch oldVal := oldVal.(type) {
	case *Object.Integer:

		if compound_exp == "+=" {
			op = &Object.Integer{Value: oldVal.Value + Val.(*Object.Integer).Value}
			env.Set(varName, op)

		} else if compound_exp == "-=" {
			op = &Object.Integer{Value: oldVal.Value - Val.(*Object.Integer).Value}
			env.Set(varName, op)

		} else if compound_exp == "*=" {
			op = &Object.Integer{Value: oldVal.Value * Val.(*Object.Integer).Value}
			env.Set(varName, op)

		} else if compound_exp == "/=" {
			op = &Object.Integer{Value: oldVal.Value / Val.(*Object.Integer).Value}
			env.Set(varName, op)

		}
	case *Object.Float:

		if compound_exp == "+=" {
			op = &Object.Float{Value: oldVal.Value + Val.(*Object.Float).Value}
			env.Set(varName, op)

		} else if compound_exp == "-=" {
			op = &Object.Float{Value: oldVal.Value - Val.(*Object.Float).Value}
			env.Set(varName, op)

		} else if compound_exp == "*=" {
			op = &Object.Float{Value: oldVal.Value * Val.(*Object.Float).Value}
			env.Set(varName, op)

		} else if compound_exp == "/=" {
			op = &Object.Float{Value: oldVal.Value / Val.(*Object.Float).Value}
			env.Set(varName, op)

		}
	case *Object.String:
		if compound_exp == "+=" {
			op = &Object.String{Value: oldVal.Value + Val.(*Object.String).Value}

			env.Set(varName, op)

		} else {
			return NewError("operation not available for STRING : %s", ie.Token, compound_exp)

		}

	}

	return op
}

func evalCompoundAssignExpression(ie *Ast.CompoundAssignExpression, env *Object.Enviroment) Object.Object {
	exp := Eval(ie.Value, env)

	if isError(exp) {
		return exp
	}

	if old_val, ok := env.Get(ie.Left.String()); ok {

		if old_val.Type() != exp.Type() {
			return NewError("type mismatch : %s %s %s", ie.Token, old_val.Type(), ie.Token.Literal, exp.Type())

		}
		return performCompoundAssignment(ie, old_val, exp, ie.Left.String(), ie.Token.Literal, env)

	}
	return nil
}

func evalLoadStatement(node *Ast.LoadExpression, env *Object.Enviroment) Object.Object {
	var path string

	_node_file_path := Eval(node.File, env)
	if isError(_node_file_path) {
		return _node_file_path
	}

	if _node_file_path.Type() != Object.STRING_OBJ {
		return NewError("Path to #load must be a string (path of file) : %s", node.Token, node.File)
	}

	path = _node_file_path.Inspect()

	if LOAD_TRACKER.IsIncluded(path) {
		w := NewWarning("%s has already been loaded once, skipping load", node.Token, node.File)
		io.WriteString(os.Stdout, w.Inspect()+"\n")

	} else {

		if strings.HasSuffix(path, ".momo") {

			if !fileExists(path) {
				return NewError("File not found : %s", node.Token, node.File)

			}

			file, _ := os.Open(path)

			defer file.Close()

			// l := Lexer.New(file, path)
			// p := Parser.New(l)
			// program := p.ParseProgram()

			// if len(p.Errors()) != 0 {
			// 	printParseErrors(os.Stdout, p.Errors())
			// 	return nil
			// }

			// evaluated := Eval(program, env)
			// if evaluated != nil && evaluated.Inspect() != "null" {
			// 	io.WriteString(os.Stdout, evaluated.Inspect())
			// 	io.WriteString(os.Stdout, "\n")
			// }

		} else {

			lib_loaded := loadLib(path, node, env)

			if isError(lib_loaded) {
				return lib_loaded

			}

		}

		LOAD_TRACKER.MarkIncluded(path)

	}

	return nil
}

func evalAssignIndexExpression(ie *Ast.AssignIndexExpression, env *Object.Enviroment) Object.Object {
	left := Eval(ie.Left, env)

	if isError(left) {
		return left
	}

	idx := Eval(ie.Index, env)

	if isError(idx) {
		return idx
	}

	new_val := Eval(ie.Value, env)

	if isError(new_val) {
		return new_val
	}

	switch obj := left.(type) {
	case *Object.Array:
		idxx := idx.(*Object.Integer).Value
		max := int64(len(obj.Elements) - 1)

		if idxx < 0 || idxx > max {
			return NewError("index out of bounds: [%d]", ie.Token, idxx)
		}

		if _, ok := env.Get(ie.Left.String()); ok {
			obj.Elements[idxx] = new_val
			env.Set(ie.Left.String(), obj)

		} else {
			return NewError("indentifier not found : %s", ie.Token, ie.Value)

		}

	case *Object.Hash:

		if _, ok := env.Get(ie.Left.String()); ok {

			key, ok := idx.(Object.Hashable)

			if !ok {
				return NewError("unusable as hash key: %s", ie.Token, idx.Type())
			}

			pair, ok := obj.Pairs[key.HashKey()]

			if !ok {
				return &Object.NULL
			}

			pair.Value = new_val

			obj.Pairs[key.HashKey()] = pair

			env.Set(ie.Left.String(), obj)

		} else {
			return NewError("indentifier not found : %s", ie.Token, ie.Value)

		}
	case *Object.String:
		return NewError("STRING does not support item assignment : %s", ie.Token, ie.Value)

	}

	return &Object.NULL
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func printParseErrors(out io.Writer, errors []*Object.Error) {
	if len(errors) > 0 {
		for _, msg := range errors {
			io.WriteString(out, "parser errors:\n")
			io.WriteString(out, "\t"+msg.Inspect()+"\n")
		}
	}
}

func Eval(node Ast.Node, env *Object.Enviroment) Object.Object {

	switch node := node.(type) {

	case *Ast.Identifier:

		return evalIdentifier(node, env)

	case *Ast.Program:
		return evalProgram(node, env)

	case *Ast.PrefixExpression:
		/*After the first call to Eval here, right may be an *object.Integer or an *object.Boolean or
		maybe even NULL. We then take this right operand and pass it to evalPrefixExpression which
		checks if the operator is supported*/

		right := Eval(node.Right, env)

		if isError(right) {
			return right
		}

		return evalPrefixExpression(node.Operator, right, node)

	case *Ast.InfixExpression:
		left := Eval(node.Left, env)

		if isError(left) {
			return left
		}

		if node.Operator == "." {
			switch obj := left.(type) {

			case *Object.Hash:

				for _, pair := range obj.Pairs {
					if pair.Key.Inspect() == node.Right.String() {
						return pair.Value
					}

				}
				return NewError("Unknown attribute : %s", node.Token, node.Right.String())
			default:
				return NewError("%s has no attributes. var typeof %s", node.Token, node.Left.String(), obj.Type())

			}
		}

		right := Eval(node.Right, env)

		if isError(right) {
			return right
		}

		return evalInfixExpression(node.Operator, left, right, node)

	case *Ast.ExpressionStatement:
		return Eval(node.Expression, env)

	case *Ast.IntegerLiteral:
		return &Object.Integer{Value: node.Value}

	case *Ast.FloatLiteral:
		return &Object.Float{Value: node.Value}

	case *Ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)

	case *Ast.BlockStatement:
		return evalBlockStatement(node, env)

	case *Ast.FORexpression:
		val := Eval(node.LoopVariable, env)

		if isError(val) {
			return val
		}

		return evalFORexpression(node, env)

	case *Ast.LoopExpression:
		return evalLOOPexpression(node, env)

	case *Ast.VarStatement:

		val := Eval(node.Value, env)

		if isError(val) {
			return val
		}

		env.Set(node.Name.Value, val)

	case *Ast.IFexpression:

		return evalIFexpression(node, env)

	case *Ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)

		if isError(val) {
			return val
		}

		return &Object.ReturnValue{Value: val}

	case *Ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &Object.Function{Parameters: params, Env: env, Body: body}

	case *Ast.FunctionStatement:
		params := node.Parameters
		body := node.Body

		env.Set(node.Name.Value, &Object.Function{Parameters: params, Env: env, Body: body})

	case *Ast.CallExpression:
		function := Eval(node.Function, env)

		if isError(function) {
			return function
		}

		args := evalExpression(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(function, args, node)

	case *Ast.StringLiteral:
		return &Object.String{Value: node.Value}

	case *Ast.ArrayLiteral:
		elements := evalExpression(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}

		return &Object.Array{Elements: elements}

	case *Ast.IndexExpression:
		left := Eval(node.Left, env)

		if isError(left) {
			return left
		}

		index := Eval(node.Index, env)

		if isError(index) {
			return index
		}

		return evalIndexExpression(left, index, node)

	case *Ast.AssignIndexExpression:
		return evalAssignIndexExpression(node, env)

	case *Ast.AssignExpression:

		identifier := Eval(node.Left, env)
		if isError(identifier) {
			return identifier
		}

		exp := Eval(node.Value, env)

		if isError(exp) {
			return exp
		}

		env.Set(node.Left.String(), exp)

	case *Ast.CompoundAssignExpression:
		return evalCompoundAssignExpression(node, env)

	case *Ast.IncDecExpression:

		identifier := Eval(node.Identifier, env)
		if isError(identifier) {
			return identifier
		}

		var new_val Object.Object

		if node.Token.Literal == Token.INC {

			switch curr_val := identifier.(type) {
			case *Object.Integer:
				new_val = &Object.Integer{Value: curr_val.Value + 1}
				env.Set(node.Identifier.String(), new_val)

			case *Object.Float:
				new_val = &Object.Float{Value: curr_val.Value + 1.0}
				env.Set(node.Identifier.String(), new_val)
			}

		} else if node.Token.Literal == Token.DEC {

			switch curr_val := identifier.(type) {
			case *Object.Integer:
				new_val = &Object.Integer{Value: curr_val.Value - 1}
				env.Set(node.Identifier.String(), new_val)

			case *Object.Float:
				new_val = &Object.Float{Value: curr_val.Value - 1.0}
				env.Set(node.Identifier.String(), new_val)

			}
		}

		return new_val

	case *Ast.HashLiteral:
		return evalHashLiteral(node, env)

	case *Ast.TypeDef:
		return evalTypeDefLiteral(node, env)

	case *Ast.TypeDefStatement:
		fmt.Println(node.Name)

		// identifier := Eval(node.Value, env)
		// if isError(identifier) {
		// 	return identifier
		// }

	case *Ast.LoadExpression:
		return evalLoadStatement(node, env)

	case *Ast.BreakStatement:
		return &Object.BREAK

	case *Ast.ContinueStatement:
		return &Object.CONTINUE

	}

	return nil
}
