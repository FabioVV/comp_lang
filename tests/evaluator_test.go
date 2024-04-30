package Tests

import (
	Evaluator "github/FabioVV/interp_lang/evaluator"
	Lexer "github/FabioVV/interp_lang/lexer"
	Object "github/FabioVV/interp_lang/object"
	Parser "github/FabioVV/interp_lang/parser"
	"io"
	"strings"

	"testing"
)

func testEval(reader io.Reader) Object.Object {
	l := Lexer.New(reader, "Test")
	p := Parser.New(l)
	program := p.ParseProgram()
	env := Object.NewEnviroment()

	return Evaluator.Eval(program, env)
}

func testIntegerObject(t *testing.T, obj Object.Object, expected int64) bool {
	result, ok := obj.(*Object.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%d, want=%d",
			result.Value, expected)
		return false
	}

	return true
}

func TestEvalIntegerExpression(t *testing.T) {

	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10 + 5", -5},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
		{"10 * 2", 20},
	}

	for _, tt := range tests {
		reader := strings.NewReader(tt.input)

		evaluated := testEval(reader)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func testBooleanObject(t *testing.T, obj Object.Object, expected bool) bool {
	result, ok := obj.(*Object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%t, want=%t",
			result.Value, expected)
		return false
	}
	return true
}

func TestEvalBooleanExpression(t *testing.T) {

	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
	}

	for _, tt := range tests {

		reader := strings.NewReader(tt.input)
		evaluated := testEval(reader)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestBangOPerator(t *testing.T) {

	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}

	for _, tt := range tests {
		reader := strings.NewReader(tt.input)

		evaluated := testEval(reader)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func testNullObject(t *testing.T, obj Object.Object) bool {
	if obj != &Object.NULL {
		t.Errorf("object is not NULL. got=%t (%+v)", obj, obj)
		return false
	}
	return true
}

func TestIfElseExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 } else { 20 }", 10},
	}

	for _, tt := range tests {
		reader := strings.NewReader(tt.input)

		evaluated := testEval(reader)
		integer, ok := tt.expected.(int)

		if ok {
			testIntegerObject(t, evaluated, int64(integer))

		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 * 5; 9;", 10},
		{"9; return 2 * 5; 9;", 10},
		{
			`
			if (10 > 1) {
				if (10 > 1) {
					return 10;
				}

				return 1;
			}`, 10,
		},
	}
	for _, tt := range tests {
		reader := strings.NewReader(tt.input)

		evaluated := testEval(reader)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestLetSt(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; let b = a; let c = a + b + 5; c;", 15},
	}

	for _, tt := range tests {
		reader := strings.NewReader(tt.input)

		testIntegerObject(t, testEval(reader), tt.expected)
	}
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"var identity = fn(x) { x; }; identity(5);", 5},
		{"var identity = fn(x) { return x; }; identity(5);", 5},
		{"var double = fn(x) { x * 2; }; double(5);", 10},
		{"var add = fn(x, y) { x + y; }; add(5, 5);", 10},
		{"var add = fn(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20},
		{"fn(x) { x; }(5)", 5},
	}
	for _, tt := range tests {
		reader := strings.NewReader(tt.input)

		testIntegerObject(t, testEval(reader), tt.expected)
	}
}

func TestStringConcatenation(t *testing.T) {
	input := `"Hello" + " " + "World!"`
	reader := strings.NewReader(input)

	evaluated := testEval(reader)
	str, ok := evaluated.(*Object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}
	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}
