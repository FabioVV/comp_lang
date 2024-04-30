package Evaluator

import (
	Object "github/FabioVV/interp_lang/object"
	Token "github/FabioVV/interp_lang/token"
)

// These are more general builtins that may have more than one use
// EX: remove() works both for a Array and a Hash
var builtins = map[string]*Object.Builtin{

	// TODO:
	"type": {
		Fn: func(token Token.Token, args ...Object.Object) Object.Object {
			if len(args) != 1 {
				return NewError("wrong number of arguments for TYPE. got=%d, want=1", token, len(args))
			}

			return returnType("%s", args[0].Type())

		},
	},

	"len": {
		Fn: func(token Token.Token, args ...Object.Object) Object.Object {
			if len(args) != 1 {
				return NewError("wrong number of arguments for len. got=%d, want=1", token, len(args))
			}

			switch arg := args[0].(type) {
			case *Object.String:
				return &Object.Integer{Value: int64(len(arg.Value))}

			case *Object.Array:
				return &Object.Integer{Value: int64(len(arg.Elements))}

			case *Object.Hash:
				return &Object.Integer{Value: int64(len(arg.Pairs))}

			default:
				return NewError("argument to 'len' not supported, got=%s", token, args[0].Type())
			}
		},
	},
	"remove": {
		Fn: func(token Token.Token, args ...Object.Object) Object.Object {
			if len(args) != 2 {
				return NewError("wrong number of arguments for 'remove'. got=%d, want=2", token, len(args))

			}

			switch arg := args[0].(type) {

			case *Object.Hash:

				val_to_remove := args[1]

				for key, pair := range arg.Pairs {

					if pair.Key.Inspect() == val_to_remove.Inspect() {

						delete(arg.Pairs, key)

						break
					}
				}
			case *Object.Array:

				val_to_remove := args[1]
				for i, val := range arg.Elements {

					if val.Inspect() == val_to_remove.Inspect() {

						arg.Elements = append(arg.Elements[0:i], arg.Elements[i+1])

						break
					}
				}

			default:
				return NewError("argument to 'clear' must be ARRAY or HASHABLE, got %s", token,
					args[0].Type())

			}

			return &Object.NULL
		},
	},
	"clear": {
		Fn: func(token Token.Token, args ...Object.Object) Object.Object {
			if len(args) != 1 {
				return NewError("wrong number of arguments for 'clear'. got=%d, want=1", token, len(args))

			}

			switch arg := args[0].(type) {

			case *Object.Hash:
				for hk := range arg.Pairs {
					delete(arg.Pairs, hk)
				}

			case *Object.Array:
				arg.Elements = make([]Object.Object, 0)

			default:
				return NewError("argument to 'clear' must be ARRAY or HASHABLE, got %s", token,
					args[0].Type())

			}

			return &Object.NULL
		},
	},
	"empty": {
		Fn: func(token Token.Token, args ...Object.Object) Object.Object {
			if len(args) != 1 {
				return NewError("wrong number of arguments for 'empty'. got=%d, want=1", token, len(args))
			}

			switch arg := args[0].(type) {

			case *Object.Hash:

				len_hash := len(arg.Pairs)

				if len_hash == 0 {
					return &Object.TRUE
				}

				return &Object.FALSE

			case *Object.Array:
				len_array := len(arg.Elements)

				if len_array == 0 {
					return &Object.TRUE
				}

				return &Object.FALSE

			case *Object.String:

				len_string := len(arg.Value)

				if len_string == 0 {
					return &Object.TRUE
				}

				return &Object.FALSE

			default:
				return NewError("argument to 'empty' must be ARRAY, HASHABLE or STRING, got %s", token,
					args[0].Type())

			}

		},
	},
}
