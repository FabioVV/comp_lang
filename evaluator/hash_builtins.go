package Evaluator

import (
	Object "github/FabioVV/interp_lang/object"
	Token "github/FabioVV/interp_lang/token"
)

var hash_builtins = map[string]*Object.Builtin{
	"update": {
		Fn: func(token Token.Token, args ...Object.Object) Object.Object {
			if len(args) != 2 {
				return NewError("wrong number of arguments for 'update'. got=%d, want=2", token, len(args))

			}

			if args[0].Type() != Object.HASH_OBJ {
				return NewError("argument to 'update' must be HASHABLE, got %s", token,
					args[0].Type())
			}

			hash_new := args[1].(*Object.Hash)
			hash_old := args[0].(*Object.Hash)

			for key, value := range hash_new.Pairs {
				hash_old.Pairs[key] = value
			}

			return &Object.NULL
		},
	},
	"keys": {
		Fn: func(token Token.Token, args ...Object.Object) Object.Object {
			if len(args) != 1 {
				return NewError("wrong number of arguments for 'keys'. got=%d, want=1", token, len(args))

			}

			if args[0].Type() != Object.HASH_OBJ {
				return NewError("argument to 'keys' must be HASHABLE, got %s", token,
					args[0].Type())
			}

			hash := args[0].(*Object.Hash)

			if len(hash.Pairs) == 0 {
				return &Object.NULL
			}

			var key_array = &Object.Array{}

			for _, pair := range hash.Pairs {

				var temp = &Object.String{
					Value: pair.Key.Inspect(),
				}

				key_array.Elements = append(key_array.Elements, temp)
			}

			return key_array
		},
	},
	"values": {
		Fn: func(token Token.Token, args ...Object.Object) Object.Object {
			if len(args) != 1 {
				return NewError("wrong number of arguments for 'values'. got=%d, want=1", token, len(args))

			}

			if args[0].Type() != Object.HASH_OBJ {
				return NewError("argument to 'values' must be HASHABLE, got %s", token,
					args[0].Type())
			}

			hash := args[0].(*Object.Hash)

			if len(hash.Pairs) == 0 {
				return &Object.NULL
			}

			var val_array = &Object.Array{}

			for _, pair := range hash.Pairs {

				var temp = &Object.String{
					Value: pair.Value.Inspect(),
				}

				val_array.Elements = append(val_array.Elements, temp)
			}

			return val_array
		},
	},
}
