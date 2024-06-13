package vm

import (
	"fmt"
	"github/FabioVV/comp_lang/code"
	"github/FabioVV/comp_lang/compiler"
	Object "github/FabioVV/comp_lang/object"
)

const STACKSIZE int = 2048
const GLOBALSSIZE int = 65536
const MAXFRAMES int = 1024

var True = &Object.Boolean{Value: true}
var False = &Object.Boolean{Value: false}

// var Null = &object.Null{}
var Null = &Object.NULL

// The momo virtual machine. Hell yeah.
type VM struct {
	constants []Object.Object
	globals   []Object.Object
	stack     []Object.Object

	frames      []*Frame
	framesIndex int

	sp int // stackpointer. Always points to the next value. Top of stack is stack[sp-1]

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

	mainFn := &Object.CompiledFunction{Instructions: bytecode.Instructions}
	mainFrame := NewFrame(mainFn)

	frames := make([]*Frame, MAXFRAMES)
	frames[0] = mainFrame

	return &VM{
		constants:   bytecode.Constants,
		stack:       make([]Object.Object, STACKSIZE),
		globals:     make([]Object.Object, GLOBALSSIZE),
		frames:      frames,
		framesIndex: 1,
		sp:          0,
	}
}

func NewWithGlobalsStore(bytecode *compiler.Bytecode, s []Object.Object) *VM {
	vm := NewVM(bytecode)
	vm.globals = s
	return vm
}

/*
	func (vm *VM) StackTop() Object.Object {
		if vm.sp == 0 {
			return nil
		}
		return vm.stack[vm.sp-1]
	}
*/

func (vm *VM) LastPoppedStackElement() Object.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) push(obj Object.Object) error {
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
func (vm *VM) pop() Object.Object {
	obj := vm.stack[vm.sp-1]
	vm.sp--
	return obj
}

func (vm *VM) execBinaryIntOp(op code.Opcode, left Object.Object, right Object.Object) error {
	leftVal := left.(*Object.Integer).Value
	rightVal := right.(*Object.Integer).Value

	switch op {
	case code.OpAdd:
		return vm.push(&Object.Integer{Value: leftVal + rightVal})

	case code.OpSub:
		return vm.push(&Object.Integer{Value: leftVal - rightVal})

	case code.OpMul:
		return vm.push(&Object.Integer{Value: leftVal * rightVal})

	case code.OpDiv:
		return vm.push(&Object.Integer{Value: leftVal / rightVal})

	default:
		return fmt.Errorf("unknow integer operator -> %d", op)
	}
}

func (vm *VM) execBinaryFltOp(op code.Opcode, left Object.Object, right Object.Object) error {
	leftVal := left.(*Object.Float).Value
	rightVal := right.(*Object.Float).Value

	switch op {
	case code.OpAdd:
		return vm.push(&Object.Float{Value: leftVal + rightVal})

	case code.OpSub:
		return vm.push(&Object.Float{Value: leftVal - rightVal})

	case code.OpMul:
		return vm.push(&Object.Float{Value: leftVal * rightVal})

	case code.OpDiv:
		return vm.push(&Object.Float{Value: leftVal / rightVal})

	default:
		return fmt.Errorf("unknow floating point operator -> %d", op)
	}
}

func (vm *VM) execBinaryStrOp(op code.Opcode, left Object.Object, right Object.Object) error {
	if op != code.OpAdd {
		return fmt.Errorf("unknow string operator : %d", op)
	}

	leftValue := left.(*Object.String).Value
	rightValue := right.(*Object.String).Value

	return vm.push(&Object.String{Value: leftValue + rightValue})

}

func (vm *VM) execBinaryOp(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	rightType := right.Type()
	LeftType := left.Type()

	switch {
	case LeftType == Object.INTEGER_OBJ && rightType == Object.INTEGER_OBJ:
		return vm.execBinaryIntOp(op, left, right)

	case LeftType == Object.STRING_OBJ && rightType == Object.STRING_OBJ:
		return vm.execBinaryStrOp(op, left, right)

	case LeftType == Object.FLOAT_OBJ && rightType == Object.FLOAT_OBJ:
		return vm.execBinaryFltOp(op, left, right)

	default:
		return fmt.Errorf("unsupported typs for binary op -> %s %s", LeftType, rightType)

	}

}

func nativeBoolToBooleanObj(input bool) *Object.Boolean {
	if input {
		return True
	}
	return False
}

func (vm *VM) execIntComparison(op code.Opcode, left Object.Object, right Object.Object) error {
	leftVal := left.(*Object.Integer).Value
	rightVal := right.(*Object.Integer).Value

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

	if left.Type() == Object.INTEGER_OBJ && right.Type() == Object.INTEGER_OBJ {
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

	if operand.Type() != Object.INTEGER_OBJ {
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}

	val := operand.(*Object.Integer).Value
	return vm.push(&Object.Integer{Value: -val})

}

func isTruthy(obj Object.Object) bool {
	switch obj := obj.(type) {
	case *Object.Boolean:
		return obj.Value

	case *Object.Null:
		return false

	default:
		return true
	}
}

func (vm *VM) buildArray(startIndex int, endIndex int) Object.Object {
	elements := make([]Object.Object, endIndex-startIndex)

	for i := startIndex; i < endIndex; i++ {
		elements[i-startIndex] = vm.stack[i]

	}

	return &Object.Array{Elements: elements}
}

func (vm *VM) buildHash(startIndex int, endIndex int) (Object.Object, error) {
	hashedPairs := make(map[Object.HashKey]Object.HashPair)

	for i := startIndex; i < endIndex; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]

		pair := Object.HashPair{Key: key, Value: value}

		hashKey, ok := key.(Object.Hashable)

		if !ok {
			return nil, fmt.Errorf("unusable as hash key : %s", key.Type())
		}

		hashedPairs[hashKey.HashKey()] = pair
	}

	return &Object.Hash{Pairs: hashedPairs}, nil
}

func (vm *VM) execArrayIndex(left Object.Object, index Object.Object) error {
	array := left.(*Object.Array)
	i := index.(*Object.Integer).Value

	max := int64(len(array.Elements) - 1)

	if i < 0 || i > max {
		return vm.push(Null)
	}

	return vm.push(array.Elements[i])
}

func (vm *VM) execHashIndex(left Object.Object, index Object.Object) error {
	hash := left.(*Object.Hash)

	key, ok := index.(Object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key : %s", index.Type())
	}

	pair, ok := hash.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Null)
	}

	return vm.push(pair.Value)
}

func (vm *VM) execIndexExpression(left Object.Object, index Object.Object) error {
	switch {
	case left.Type() == Object.ARRAY_OBJ && index.Type() == Object.INTEGER_OBJ:
		return vm.execArrayIndex(left, index)

	case left.Type() == Object.HASH_OBJ:
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
		case code.Opconstant:
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

		case code.OpSetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2

			vm.globals[globalIndex] = vm.pop()

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
			vm.popFrame()
			vm.pop()

			if err := vm.push(Null); err != nil {
				return err
			}

		case code.OpReturnValue:
			returnValue := vm.pop()

			vm.popFrame()
			vm.pop()

			if err := vm.push(returnValue); err != nil {
				return err
			}

		case code.OpCall:
			fn, ok := vm.stack[vm.sp-1].(*Object.CompiledFunction)
			if !ok {
				return fmt.Errorf("calling non function")
			}

			frame := NewFrame(fn)
			vm.pushFrame(frame)
		}
	}

	return nil

}
