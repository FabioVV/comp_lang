package main

import (
	"fmt"
	Evaluator "github/FabioVV/interp_lang/evaluator"
	Lexer "github/FabioVV/interp_lang/lexer"
	Object "github/FabioVV/interp_lang/object"
	Parser "github/FabioVV/interp_lang/parser"
	Repl "github/FabioVV/interp_lang/repl"
	"io"
	"os"
	"path/filepath"
)

func printParseErrors(out io.Writer, errors []*Object.Error) {
	if len(errors) > 0 {
		for _, msg := range errors {
			io.WriteString(out, "parser errors:\n")
			io.WriteString(out, "\t"+msg.Inspect()+"\n")
		}
	}
}

func main() {

	if len(os.Args) < 2 {
		Repl.Start(os.Stdin, os.Stdout)
		return
	}

	var cwd, err = os.Getwd()
	var path string

	if err != nil {
		return
	}

	filePath := os.Args[1]

	_, err = os.Stat(filePath)

	if err != nil {
		if filePath == "-help" || filePath == "help" {
			fmt.Println("Usage: go run main.go <path-to-file>\nor\nUsage: go run main.go")
			fmt.Println("If executed without arguments it will start the REPL else it will execute the file")
			return
		}
	}

	if filepath.IsAbs(filePath) {
		path = filePath
	} else {
		path = filepath.Join(cwd, filePath)
	}

	if os.IsNotExist(err) {
		fmt.Printf("momo-pre-alpha - can't open file '%s'\nDoes the file exists? is the path correct?\n", path)
		return
	}

	file, err := os.Open(filePath)

	if err != nil {
		fmt.Printf("momo-pre-alpha - failed to open file: %s\n", err)
		return
	}

	defer file.Close()

	env := Object.NewEnviroment()

	l := Lexer.New(file, file.Name())
	p := Parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		printParseErrors(os.Stdout, p.Errors())
		return
	}

	evaluated := Evaluator.Eval(program, env)
	if evaluated != nil {
		io.WriteString(os.Stdout, evaluated.Inspect())
		io.WriteString(os.Stdout, "\n")
	}

}
