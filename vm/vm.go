package vm

import (
	"fmt"
	"github/FabioVV/comp_lang/code"
	"github/FabioVV/comp_lang/compiler"
	object "github/FabioVV/comp_lang/object"
	token "github/FabioVV/comp_lang/token"
)

const STACKSIZE int = 2048
const GLOBALSSIZE int = 65536
const MAXFRAMES int = 1024

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}

// var Null = &object.Null{}
var Null = &object.NULL

// The momo virtual machine. Hell yeah.
type VM struct {
	constants []object.Object
	globals   []object.Object
	stack     []object.Object

	frames      []*Frame
	framesIndex int

	sp int // stackpointer. Always points to the next value. Top of stack is stack[sp-1]

}

func (v *VM) newVMError(format string, token token.Token, a ...interface{}) *object.Error {
	return &object.Error{
		Message:  fmt.Sprintf(format, a...),
		Filename: token.Filename,
		Line:     token.Pos.Line,
		Column:   token.Pos.Column,
	}
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}

func NewVM(bytecode *compiler.Bytecode) *VM {

	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainClosure := &object.Closure{Fn: mainFn}
	mainFrame := NewFrame(mainClosure, 0)

	frames := make([]*Frame, MAXFRAMES)
	frames[0] = mainFrame

	return &VM{
		constants:   bytecode.Constants,
		stack:       make([]object.Object, STACKSIZE),
		globals:     make([]object.Object, GLOBALSSIZE),
		frames:      frames,
		framesIndex: 1,
		sp:          0,
	}
}

func NewWithGlobalsStore(bytecode *compiler.Bytecode, s []object.Object) *VM {
	vm := NewVM(bytecode)
	vm.globals = s
	return vm
}

/*
	func (vm *VM) StackTop() object.Object {
		if vm.sp == 0 {
			return nil
		}
		return vm.stack[vm.sp-1]
	}
*/

func (vm *VM) LastPoppedStackElement() object.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) push(obj object.Object) error {
	if vm.sp >= STACKSIZE {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = obj
	vm.sp++

	return nil
}

/*
We first take the element from the top of the stack, located at vm.sp-1, and put it on the
side. Then we decrement vm.sp, allowing the location of element that was just popped off being
overwritten eventually.
*/
func (vm *VM) pop() object.Object {
	obj := vm.stack[vm.sp-1]
	vm.sp--
	return obj
}

func (vm *VM) execBinaryIntOp(op code.Opcode, left object.Object, right object.Object) error {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch op {
	case code.OpAdd:
		return vm.push(&object.Integer{Value: leftVal + rightVal})

	case code.OpSub:
		return vm.push(&object.Integer{Value: leftVal - rightVal})

	case code.OpMul:
		return vm.push(&object.Integer{Value: leftVal * rightVal})

	case code.OpDiv:
		return vm.push(&object.Integer{Value: leftVal / rightVal})

	default:
		return fmt.Errorf("unknow integer operator -> %d", op)
	}
}

func (vm *VM) execBinaryFltOp(op code.Opcode, left object.Object, right object.Object) error {
	var leftVal float64
	var rightVal float64

	switch left := left.(type) {
	case *object.Float:
		leftVal = left.Value
	case *object.Integer:
		leftVal = float64(left.Value)
	}

	switch right := right.(type) {
	case *object.Float:
		rightVal = right.Value
	case *object.Integer:
		rightVal = float64(right.Value)
	}

	switch op {
	case code.OpAdd:
		return vm.push(&object.Float{Value: leftVal + rightVal})

	case code.OpSub:
		return vm.push(&object.Float{Value: leftVal - rightVal})

	case code.OpMul:
		return vm.push(&object.Float{Value: leftVal * rightVal})

	case code.OpDiv:
		return vm.push(&object.Float{Value: leftVal / rightVal})

	default:
		return fmt.Errorf("unknow floating point operator -> %d", op)
	}
}

func (vm *VM) execBinaryStrOp(op code.Opcode, left object.Object, right object.Object) error {
	if op != code.OpAdd {
		return fmt.Errorf("unknow string operator : %d", op)
	}

	leftValue := left.(*object.String).Value
	rightValue := right.(*object.String).Value

	return vm.push(&object.String{Value: leftValue + rightValue})

}

func (vm *VM) execBinaryOp(op code.Opcode) error {

	right := vm.pop()
	left := vm.pop()

	rightType := right.Type()
	LeftType := left.Type()

	switch {
	case LeftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ:
		return vm.execBinaryIntOp(op, left, right)

	case LeftType == object.STRING_OBJ && rightType == object.STRING_OBJ:
		return vm.execBinaryStrOp(op, left, right)

	case LeftType == object.FLOAT_OBJ && rightType == object.FLOAT_OBJ:
		return vm.execBinaryFltOp(op, left, right)

	case (LeftType == object.FLOAT_OBJ && rightType == object.INTEGER_OBJ) || (LeftType == object.INTEGER_OBJ && rightType == object.FLOAT_OBJ):
		return vm.execBinaryFltOp(op, left, right)

	default:
		return fmt.Errorf("unsupported typs for binary op -> %s %s", LeftType, rightType)

	}

}

func nativeBoolToBooleanObj(input bool) *object.Boolean {
	if input {
		return True
	}
	return False
}

func (vm *VM) execIntComparison(op code.Opcode, left object.Object, right object.Object) error {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObj(rightVal == leftVal))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObj(rightVal != leftVal))
	case code.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObj(leftVal > rightVal))
	default:
		return fmt.Errorf("unknown operator: %d", op)

	}
}

func (vm *VM) execComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return vm.execIntComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObj(right == left))

	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObj(right != left))

	default:
		return fmt.Errorf("unknow operator -> %d (%s %s)", op, left.Type(), right.Type())
	}
}

func (vm *VM) execBangOperator() error {
	operand := vm.pop()

	switch operand {
	case Null:
		return vm.push(True)

	case True:
		return vm.push(False)

	case False:
		return vm.push(True)

	default:
		return vm.push(False)

	}
}

func (vm *VM) execMinusOperator() error {
	operand := vm.pop()

	if operand.Type() != object.INTEGER_OBJ {
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}

	val := operand.(*object.Integer).Value
	return vm.push(&object.Integer{Value: -val})

}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value

	case *object.Null:
		return false

	default:
		return true
	}
}

func (vm *VM) buildArray(startIndex int, endIndex int) object.Object {
	elements := make([]object.Object, endIndex-startIndex)

	for i := startIndex; i < endIndex; i++ {
		elements[i-startIndex] = vm.stack[i]

	}

	return &object.Array{Elements: elements}
}

func (vm *VM) buildHash(startIndex int, endIndex int) (object.Object, error) {
	hashedPairs := make(map[object.HashKey]object.HashPair)

	for i := startIndex; i < endIndex; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]

		pair := object.HashPair{Key: key, Value: value}

		hashKey, ok := key.(object.Hashable)

		if !ok {
			return nil, fmt.Errorf("unusable as hash key : %s", key.Type())
		}

		hashedPairs[hashKey.HashKey()] = pair
	}

	return &object.Hash{Pairs: hashedPairs}, nil
}

func (vm *VM) execArrayIndex(left object.Object, index object.Object) error {
	array := left.(*object.Array)
	i := index.(*object.Integer).Value

	max := int64(len(array.Elements) - 1)

	if i < 0 || i > max {
		return vm.push(Null)
	}

	return vm.push(array.Elements[i])
}

func (vm *VM) execHashIndex(left object.Object, index object.Object) error {
	hash := left.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key : %s", index.Type())
	}

	pair, ok := hash.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Null)
	}

	return vm.push(pair.Value)
}

func (vm *VM) execIndexExpression(left object.Object, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return vm.execArrayIndex(left, index)

	case left.Type() == object.HASH_OBJ:
		return vm.execHashIndex(left, index)

	default:
		return fmt.Errorf("index operator not supported : %s", left.Type())
	}
}

// Turns on momo's virtual machine
func (vm *VM) Run() error {

	//ip =  instruction pointer
	var ip int
	var ins code.Instructions
	var op code.Opcode

	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {

		vm.currentFrame().ip++
		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()
		op = code.Opcode(ins[ip])

		switch op {
		case code.OpConstant:
			/*
				After decoding the operands, we must be careful to increment ip by the correct amount – the
				number of bytes we read to decode the operands. The result is that the next iteration of the
				loop starts with ip pointing to an opcode instead of an operand.
			*/
			constIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2

			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}

		case code.OpPop:
			vm.pop()

		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:

			if err := vm.execBinaryOp(op); err != nil {
				return err
			}

		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan:
			if err := vm.execComparison(op); err != nil {
				return err
			}

		case code.OpTrue:
			if err := vm.push(True); err != nil {
				return err
			}

		case code.OpFalse:
			if err := vm.push(False); err != nil {
				return err
			}

		case code.OpBang:
			if err := vm.execBangOperator(); err != nil {
				return err
			}

		case code.OpMinus:
			if err := vm.execMinusOperator(); err != nil {
				return err
			}

		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(ins[ip+1:]))
			// After that we manually
			// increase ip by two so we correctly skip over the two bytes of the operand in the next cycle.
			vm.currentFrame().ip += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				vm.currentFrame().ip = pos - 1

			}

		case code.OpJump:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip = pos - 1

		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2

			err := vm.push(vm.globals[globalIndex])

			if err != nil {
				return err
			}
		case code.OpGetLocal:
			localIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()

			err := vm.push(vm.stack[frame.basePointer+int(localIndex)])

			if err != nil {
				return err
			}
		case code.OpSetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2

			vm.globals[globalIndex] = vm.pop()

		case code.OpSetLocal:
			localIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()

			vm.stack[frame.basePointer+int(localIndex)] = vm.pop()

		case code.OpGetFree:
			freeIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			currentClosure := vm.currentFrame().cl

			if err := vm.push(currentClosure.Free[freeIndex]); err != nil {
				return err
			}

		case code.OpNull:
			err := vm.push(Null)
			if err != nil {
				return err
			}

		case code.OpArray:
			numElements := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			array := vm.buildArray(vm.sp-numElements, vm.sp)
			vm.sp = vm.sp - numElements

			if err := vm.push(array); err != nil {
				return err
			}

		case code.OpHash:
			numElements := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			hash, err := vm.buildHash(vm.sp-numElements, vm.sp)
			if err != nil {
				return err
			}

			vm.sp = vm.sp - numElements

			if err := vm.push(hash); err != nil {
				return err
			}

		case code.OpIndex:
			index := vm.pop() // Removes value from the top of the stack
			left := vm.pop()  // Removes the value that has become the top (before it was the the value just before it) -- 1+(what is this explanation)2+(what is happening)

			if err := vm.execIndexExpression(left, index); err != nil {
				return err
			}

		case code.OpReturn:
			frame := vm.popFrame()
			vm.sp = frame.basePointer + 1

			if err := vm.push(Null); err != nil {
				return err
			}

		case code.OpReturnValue:
			returnValue := vm.pop()

			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1

			if err := vm.push(returnValue); err != nil {
				return err
			}

		case code.OpCall:
			numArgs := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			if err := vm.executeCall(int(numArgs)); err != nil {
				return err
			}

		case code.OpGetBuiltin:
			builtingIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			def := object.Builtins[builtingIndex]

			if err := vm.push(def.Builtin); err != nil {
				return err
			}

		case code.OpClosure:
			constIndex := code.ReadUint16(ins[ip+1:])

			numFree := code.ReadUint8(ins[ip+3:])
			vm.currentFrame().ip += 3

			if err := vm.pushClosure(int(constIndex), int(numFree)); err != nil {
				return err
			}
		case code.OpCurrentClosure:
			current := vm.currentFrame().cl
			if err := vm.push(current); err != nil {
				return err
			}

		}
	}

	return nil

}

func (vm *VM) executeCall(numArgs int) error {
	calee := vm.stack[vm.sp-1-numArgs]

	switch calee := calee.(type) {
	case *object.Closure:
		return vm.callClosure(calee, numArgs)

	case *object.Builtin:
		return vm.callBuiltin(calee, numArgs)

	default:
		return fmt.Errorf("calling non-function and (non built-in)")
	}
}

func (vm *VM) callClosure(cl *object.Closure, numArgs int) error {

	if numArgs != cl.Fn.NumParameters {
		return fmt.Errorf("wrong number of arguments : want=%d got=%d", cl.Fn.NumParameters, numArgs)
	}

	frame := NewFrame(cl, vm.sp-numArgs)
	vm.pushFrame(frame)
	vm.sp = frame.basePointer + cl.Fn.NumLocals

	return nil
}

func (vm *VM) callBuiltin(builtin *object.Builtin, numArgs int) error {

	args := vm.stack[vm.sp-numArgs : vm.sp]

	result := builtin.Fn(args...)
	vm.sp = vm.sp - numArgs - 1

	if result != nil {
		vm.push(result)
	} else {
		vm.push(Null)
	}

	return nil
}

func (vm *VM) pushClosure(constIndex int, numFree int) error {
	constant := vm.constants[constIndex]
	function, ok := constant.(*object.CompiledFunction)

	if !ok {
		return fmt.Errorf("not a function %+v", constant)
	}

	free := make([]object.Object, numFree)
	for i := 0; i < numFree; i++ {
		free[i] = vm.stack[vm.sp-numFree+i]
	}

	vm.sp = vm.sp - numFree

	return vm.push(&object.Closure{Fn: function, Free: free})
}
