package Evaluator

import (
	"fmt"
	object "github/FabioVV/comp_lang/object"
	token "github/FabioVV/comp_lang/token"
)

var stdout_builtins = map[string]*object.Builtin{
	/*
		Functions for OUTPUT
	*/
	"puts": {
		Fn: func(token token.Token, args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}

			return &object.NULL
		},
	},
	"print": {
		Fn: func(token token.Token, args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Print(arg.Inspect())
			}

			return &object.NULL
		},
	},
}
