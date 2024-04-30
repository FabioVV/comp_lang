package Object

/*
Structs ENVIROMENT
*/
type StructEnviroment struct {
	store map[string]Object
	outer *StructEnviroment
}

func NewEnclosedStructEnviroment(outer *StructEnviroment) *StructEnviroment {
	env := NewStructEnviroment()
	env.outer = outer
	return env
}

func NewStructEnviroment() *StructEnviroment {
	s := make(map[string]Object)

	return &StructEnviroment{store: s, outer: nil}
}

func (es *StructEnviroment) GetStruct(name string) (Object, bool) {
	obj, ok := es.store[name]

	if !ok && es.outer != nil {
		obj, ok = es.outer.GetStruct(name)
	}
	return obj, ok
}

func (es *StructEnviroment) SetStruct(name string, val Object) Object {
	es.store[name] = val
	return val
}

/*
Structs ENVIROMENT
*/

/*
GENERAL ENVIROMENT
*/

type Enviroment struct {
	store map[string]Object
	outer *Enviroment
}

func NewEnclosedEnviroment(outer *Enviroment) *Enviroment {
	env := NewEnviroment()
	env.outer = outer
	return env
}

func NewEnviroment() *Enviroment {
	s := make(map[string]Object)

	return &Enviroment{store: s, outer: nil}
}

func (e *Enviroment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]

	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}

	return obj, ok
}

func (e *Enviroment) Set(name string, val Object) Object {

	e.store[name] = val
	return val
}

/*
GENERAL ENVIROMENT
*/
