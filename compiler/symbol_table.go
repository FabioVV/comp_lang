package compiler

type SymbolScope string

const (
	GLOBALSCOPE   SymbolScope = "GLOBAL"
	LOCALSCOPE    SymbolScope = "LOCAL"
	BUILTINSCOPE  SymbolScope = "BUILTIN"
	FREESCOPE     SymbolScope = "FREE" // For closures (fn(a){var f = fn(b){a+b} return f})
	FUNCTIONSCOPE SymbolScope = "FUNCTION"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	Outer *SymbolTable

	store          map[string]Symbol
	numDefinitions int

	FreeSymbols []Symbol
}

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	free := []Symbol{}

	return &SymbolTable{store: s, FreeSymbols: free}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.Outer = outer

	return s
}

func (s *SymbolTable) Define(name string) Symbol {
	symbol := Symbol{Name: name, Index: s.numDefinitions}

	if s.Outer == nil {
		symbol.Scope = GLOBALSCOPE
	} else {
		symbol.Scope = LOCALSCOPE
	}

	s.store[name] = symbol
	s.numDefinitions++

	return symbol
}

func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	obj, ok := s.store[name]

	if !ok && s.Outer != nil {
		obj, ok := s.Outer.Resolve(name)
		if !ok {
			return obj, ok
		}
		if obj.Scope == GLOBALSCOPE || obj.Scope == BUILTINSCOPE {
			return obj, ok
		}

		free := s.defineFree(obj)
		return free, true

	}

	return obj, ok
}

func (s *SymbolTable) DefineBuiltin(name string, index int) Symbol {
	symbol := Symbol{Name: name, Index: index, Scope: BUILTINSCOPE}
	s.store[name] = symbol

	return symbol
}

func (s *SymbolTable) defineFree(original Symbol) Symbol {
	s.FreeSymbols = append(s.FreeSymbols, original)

	symbol := Symbol{Name: original.Name, Index: len(s.FreeSymbols) - 1}
	symbol.Scope = FREESCOPE

	s.store[original.Name] = symbol
	return symbol
}

func (s *SymbolTable) DefineFunctionName(name string) Symbol {
	symbol := Symbol{Name: name, Index: 0, Scope: FUNCTIONSCOPE}
	s.store[name] = symbol

	return symbol
}
