package Evaluator

import (
	object "github/FabioVV/comp_lang/object"
	Token "github/FabioVV/comp_lang/token"
)

// Functions that are built-in for Arrays
var array_builtings = map[string]*object.Builtin{
	"sort": {
		Fn: func(token Token.Token, args ...object.Object) object.Object {
			if len(args) != 1 {
				return NewError("wrong number of arguments. got=%d, want=1", token,
					len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return NewError("argument to 'sort' must be ARRAY, got %s", token,
					args[0].Type())
			}

			arr := args[0].(*object.Array)

			var newElements = make([]int64, 0)

			for _, obj := range arr.Elements {

				switch temp := obj.(type) {
				case *object.Integer:
					newElements = append(newElements, temp.Value)

				default:
					return NewError("other found, ARRAY values must be INTEGER, got %s", token,
						temp.Type())
				}

			}

			quickSort(newElements, 0, int64(len(newElements))-1)

			arr.Elements = nil
			arr.Elements = []object.Object{}

			for _, obj := range newElements {
				var temp = &object.Integer{
					Value: obj,
				}

				arr.Elements = append(arr.Elements, temp)
			}

			return &object.NULL
		},
	},
	"first": {
		Fn: func(token Token.Token, args ...object.Object) object.Object {
			if len(args) != 1 {
				return NewError("wrong number of arguments. got=%d, want=1", token,
					len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return NewError("argument to 'first' must be ARRAY, got %s", token,
					args[0].Type())
			}

			arr := args[0].(*object.Array)

			if len(arr.Elements) > 0 {
				return arr.Elements[0]

			}

			return &object.NULL
		},
	},
	"last": {
		Fn: func(token Token.Token, args ...object.Object) object.Object {
			if len(args) != 1 {
				return NewError("wrong number of arguments for 'last'. got=%d, want=1", token, len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return NewError("argument to 'last' must be ARRAY, got %s", token,
					args[0].Type())
			}

			// CONVERT E ASSERT??
			arr := args[0].(*object.Array)
			length := len(arr.Elements)

			if length > 0 {
				return arr.Elements[length-1]
			}
			return &object.NULL
		},
	},
	"tail": {
		Fn: func(token Token.Token, args ...object.Object) object.Object {
			if len(args) != 1 {
				return NewError("wrong number of arguments for 'tail'. got=%d, want=1", token, len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return NewError("argument to 'tail' must be ARRAY, got %s", token,
					args[0].Type())
			}

			// // CONVERT E ASSERT??
			arr := args[0].(*object.Array)
			length := len(arr.Elements)

			newElements := make([]object.Object, length-1)
			copy(newElements, arr.Elements[1:length])
			return &object.Array{Elements: newElements}

		},
	},
	"push": {
		Fn: func(token Token.Token, args ...object.Object) object.Object {
			if len(args) != 2 {
				return NewError("wrong number of arguments for 'push'. got=%d, want=2", token, len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return NewError("argument to 'push' must be ARRAY, got %s", token,
					args[0].Type())
			}

			// CONVERT E ASSERT??
			arr := args[0].(*object.Array)

			arr.Elements = append(arr.Elements, args[1])

			return &object.NULL
		},
	},
	"pop": {
		Fn: func(token Token.Token, args ...object.Object) object.Object {
			if len(args) != 1 {
				return NewError("wrong number of arguments for 'pop'. got=%d, want=1", token, len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return NewError("argument to 'pop' must be ARRAY, got %s", token,
					args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)
			last_el := arr.Elements[length-1]

			if length > 0 {
				arr.Elements = arr.Elements[0 : length-1]

			}

			return last_el
		},
	},
	"shift": {
		Fn: func(token Token.Token, args ...object.Object) object.Object {
			if len(args) != 1 {
				return NewError("wrong number of arguments for 'shift'.  got=%d, want=1", token, len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return NewError("argument to 'shift' must be ARRAY, got %s", token,
					args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)
			first_el := arr.Elements[0]

			if length > 0 {
				arr.Elements = arr.Elements[1:length]

			}

			return first_el
		},
	},
}
