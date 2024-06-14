package vm

import (
	code "github/FabioVV/comp_lang/code"
	object "github/FabioVV/comp_lang/object"
)

type Frame struct {
	cl          *object.Closure
	ip          int
	basePointer int // The stack pointer before we execute a function
}

func NewFrame(cl *object.Closure, basePointer int) *Frame {
	return &Frame{cl: cl, ip: -1, basePointer: basePointer}
}

func (f *Frame) Instructions() code.Instructions {
	return f.cl.Fn.Instructions
}
