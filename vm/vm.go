package vm

import (
	"fmt"
	"github/FabioVV/interp_lang/code"
	"github/FabioVV/interp_lang/compiler"
	Object "github/FabioVV/interp_lang/object"
)

const STACKSIZE int = 2048

var True = &Object.Boolean{Value: true}
var False = &Object.Boolean{Value: false}

// The momo virtual machine. Hell yeah.
type VM struct {
	instructions code.Instructions
	constants    []Object.Object

	stack []Object.Object
	sp    int // stackpointer. Always points to the next value. Top of stack is stack[sp-1]

}

func NewVM(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,
		stack:        make([]Object.Object, STACKSIZE),
		sp:           0,
	}
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

func (vm *VM) execBinaryOp(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	rightType := right.Type()
	LeftType := left.Type()

	if LeftType == Object.INTEGER_OBJ && rightType == Object.INTEGER_OBJ {
		return vm.execBinaryIntOp(op, left, right)
	}

	return fmt.Errorf("unsupported typs for binary op -> %s %s", LeftType, rightType)
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

// Turns on momo's virtual machine
func (vm *VM) Run() error {

	//ip =  instruction pointer
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])

		switch op {
		case code.Opconstant:
			/*
				After decoding the operands, we must be careful to increment ip by the correct amount – the
				number of bytes we read to decode the operands. The result is that the next iteration of the
				loop starts with ip pointing to an opcode instead of an operand.
			*/
			constIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

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
		}
	}

	return nil

}
