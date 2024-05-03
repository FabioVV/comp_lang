package vm

import (
	"fmt"
	"github/FabioVV/interp_lang/code"
	"github/FabioVV/interp_lang/compiler"
	Object "github/FabioVV/interp_lang/object"
)

const STACKSIZE = 2048

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

func (vm *VM) StackTop() Object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
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

		case code.OpAdd:
			//Since its a + operation, it does not matter which operand comes first
			right := vm.pop()
			left := vm.pop()

			rightVal := right.(*Object.Integer).Value
			leftVal := left.(*Object.Integer).Value

			vm.push(&Object.Integer{Value: rightVal + leftVal})
		}
	}

	return nil

}
