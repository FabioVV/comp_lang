package Evaluator

import (
	"bufio"
	"fmt"
	Object "github/FabioVV/comp_lang/object"
	Token "github/FabioVV/comp_lang/token"
	"os"
)

var stdin_builtins = map[string]*Object.Builtin{
	"input": {
		Fn: func(token Token.Token, args ...Object.Object) Object.Object {
			if len(args) > 1 {
				return NewError("wrong number of arguments for 'input'. got=%d, want=1 or 0", token, len(args))
			}

			scanner := bufio.NewScanner(os.Stdin)
			scanner.Split(bufio.ScanLines)

			if len(args) == 1 && args[0].Type() == Object.STRING_OBJ {
				fmt.Print(args[0].Inspect())

			} else {
				return NewError("argument to 'input' must be STRING, got %s", token, args[0].Type())

			}

			var input string

			if scanner.Scan() {
				input = scanner.Text()

			} else if err := scanner.Err(); err != nil {
				return NewError("Error reading standard input for 'input'. error: %s", token, err)
			}

			return &Object.String{Value: input + "\n"}

		},
	},
}
