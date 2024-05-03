package repl

import (
	"bufio"
	"fmt"
	"github/FabioVV/interp_lang/compiler"
	Lexer "github/FabioVV/interp_lang/lexer"
	Object "github/FabioVV/interp_lang/object"
	Parser "github/FabioVV/interp_lang/parser"
	"github/FabioVV/interp_lang/vm"
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

		comp := compiler.New()
		err := comp.Compile(program)

		if err != nil {
			fmt.Fprintf(out, "compilation failed:\n %s\n", err)
			continue
		}

		machine := vm.NewVM(comp.Bytecode())
		err = machine.Run()

		if err != nil {
			fmt.Fprintf(out, "executing bytecode failed:\n %s\n", err)
			continue
		}

		stackTop := machine.StackTop()
		io.WriteString(out, stackTop.Inspect())
		io.WriteString(out, "\n")

	}
}
