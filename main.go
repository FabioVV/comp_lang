package main

import (
	"fmt"
	"github/FabioVV/comp_lang/compiler"
	lexer "github/FabioVV/comp_lang/lexer"
	object "github/FabioVV/comp_lang/object"
	parser "github/FabioVV/comp_lang/parser"
	repl "github/FabioVV/comp_lang/repl"
	"github/FabioVV/comp_lang/vm"
	"io"
	"os"
	"path/filepath"
)

func printParseErrors(out io.Writer, errors []*object.Error) {
	if len(errors) > 0 {
		for _, msg := range errors {
			io.WriteString(out, "parser errors:\n")
			io.WriteString(out, "\t"+msg.Inspect()+"\n")
		}
	}
}

func printCompilerError(out io.Writer, _error *object.Error) {
	io.WriteString(out, "compilation failed:\n")
	io.WriteString(out, "\t"+_error.Inspect()+"\n")
}

func main() {

	if len(os.Args) < 2 {
		repl.Start(os.Stdin, os.Stdout)
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
		fmt.Printf("momo-pre-pre-alpha - failed to open file: %s\n", err)
		return
	}

	defer file.Close()

	// constants := []object.Object{}
	// globals := make([]object.Object, vm.GLOBALSSIZE)
	symbolTable := compiler.NewSymbolTable()

	for i, v := range object.Builtins {
		symbolTable.DefineBuiltin(v.Name, i)
	}

	l := lexer.New(file, file.Name())
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		printParseErrors(os.Stdout, p.Errors())
		return
	}

	comp := compiler.New()
	err_obj := comp.Compile(program)

	if err_obj != nil {
		printCompilerError(os.Stdout, err_obj)
		return
	}

	code := comp.Bytecode()

	machine := vm.NewVM(code)
	err = machine.Run()

	if err != nil {
		fmt.Fprintf(os.Stdout, "executing bytecode failed:\n %s\n", err)
		return
	}

	lastPopped := machine.LastPoppedStackElement()
	io.WriteString(os.Stdout, lastPopped.Inspect())
	io.WriteString(os.Stdout, "\n")

}
