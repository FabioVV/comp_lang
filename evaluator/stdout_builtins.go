package Evaluator

import (
	"fmt"
	Object "github/FabioVV/interp_lang/object"
	Token "github/FabioVV/interp_lang/token"
)

var stdout_builtins = map[string]*Object.Builtin{
	/*
		Functions for OUTPUT
	*/
	"puts": {
		Fn: func(token Token.Token, args ...Object.Object) Object.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}

			return &Object.NULL
		},
	},
	"print": {
		Fn: func(token Token.Token, args ...Object.Object) Object.Object {
			for _, arg := range args {
				fmt.Print(arg.Inspect())
			}

			return &Object.NULL
		},
	},
}
