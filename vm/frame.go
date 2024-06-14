package vm

import (
	code "github/FabioVV/comp_lang/code"
	object "github/FabioVV/comp_lang/object"
)

type Frame struct {
	fn          *object.CompiledFunction
	ip          int
	basePointer int // The stack pointer before we execute a function
}

func NewFrame(fn *object.CompiledFunction, basePointer int) *Frame {
	return &Frame{fn: fn, ip: -1, basePointer: basePointer}
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}
