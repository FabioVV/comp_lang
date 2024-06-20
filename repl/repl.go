package repl

import (
	"bufio"
	"fmt"
	"github/FabioVV/comp_lang/compiler"
	Lexer "github/FabioVV/comp_lang/lexer"
	Object "github/FabioVV/comp_lang/object"
	Parser "github/FabioVV/comp_lang/parser"
	"github/FabioVV/comp_lang/vm"
	"io"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"time"
)

const PROMPT string = "!>> "

const DRAW string = ``

/*
read from the input source until encountering a newline, take
the just read line and pass it to an instance of our lexer and finally print all the tokens the lexer
gives us until we encounter EOF.
*/

func printParseErrors(out io.Writer, errors []*Object.Error) {
	if len(errors) > 0 {
		for _, msg := range errors {
			io.WriteString(out, "parser errors:\n")
			io.WriteString(out, "\t"+msg.Inspect()+"\n")
		}
	}
}

func printCompilerError(out io.Writer, _error *Object.Error) {
	io.WriteString(out, "compilation failed:\n")
	io.WriteString(out, "\t"+_error.Inspect()+"\n")
}

func ClearScreen() {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		return
	}
}

func Start(in io.Reader, out io.Writer) {
	ClearScreen()

	scanner := bufio.NewScanner(in)

	user, err := user.Current()
	username := user.Username
	currentTime := time.Now()
	hour := currentTime.Hour()
	platform := runtime.GOOS
	var message string

	if hour >= 18 || hour < 6 {
		message = "Good night! It's a bit late. You should get some sleep."

	} else if hour >= 6 && hour < 18 {
		message = "Good day! Happy coding."

	} else {
		message = "Happy coding evening!"

	}

	if err != nil {
		username = "Coder"
	}

	fmt.Printf("{Momo compiler pre-pre-alpha } : {%s} : {%s}\n", currentTime, platform)

	fmt.Printf("Hello %s \n", username)
	fmt.Printf("%s\n", message)

	fmt.Printf("Feel free to type in commands\n")

	constants := []Object.Object{}
	globals := make([]Object.Object, vm.GLOBALSSIZE)
	symbolTable := compiler.NewSymbolTable()

	for i, v := range Object.Builtins {
		symbolTable.DefineBuiltin(v.Name, i)
	}

	for {
		fmt.Fprintf(out, PROMPT)
		scanned := scanner.Scan()

		if !scanned {
			return
		}

		line := scanner.Text()
		reader := strings.NewReader(line)

		l := Lexer.New(reader, "<stdin>")
		p := Parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			printParseErrors(out, p.Errors())
			// continue, we dont want the REPL to exit on error
			continue
		}

		comp := compiler.NewWithState(symbolTable, constants)
		err_obj := comp.Compile(program)

		if err_obj != nil {
			printCompilerError(os.Stdout, err_obj)
			continue
		}

		code := comp.Bytecode()
		constants = code.Constants

		machine := vm.NewWithGlobalsStore(code, globals)
		mac_err := machine.Run()

		if mac_err != nil {
			fmt.Fprintf(out, "executing bytecode failed:\n %s\n", mac_err)
			continue
		}

		lastPopped := machine.LastPoppedStackElement()
		io.WriteString(out, lastPopped.Inspect())
		io.WriteString(out, "\n")

	}
}
