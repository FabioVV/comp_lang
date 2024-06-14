package Object

// import (
// 	token "github/FabioVV/comp_lang/token"
// )

// Here’s finally the how of executing built-in functions.
// We take the arguments from the stack (without removing them yet) and pass them to the
// object.BuiltinFunction that’s contained in the *object.Builtin’s Fn field. That’s the central
// part, the execution of the built-in function itself.
// After that, we decrease the vm.sp to take the arguments and the function we just executed off
// the stack. As per our calling convention, doing that is the duty of the VM.
// Once the stack is cleaned up, we check whether the result of the call is nil or not. If it’s not nil,
// we push the result on to the stack; but if it is, we push vm.Null. That’s the bring-your-own-null
// strategy at work again.

func GetBuiltinByName(name string) *Builtin {
	for _, def := range Builtins {
		if def.Name == name {
			return def.Builtin
		}
	}
	return nil
}

var Builtins = []struct {
	Name    string
	Builtin *Builtin
}{
	{
		// &Builtin{Fn: func(token token.Token, args ...Object) Object {

		"len",
		&Builtin{Fn: func(args ...Object) Object {
			return nil
		},
		},
	},
}
