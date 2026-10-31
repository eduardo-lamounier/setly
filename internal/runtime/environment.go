package runtime

import (
	"fmt"
)

type Environment struct {
	enclosing *Environment
	values    map[string]Value
}

var GlobalEnvironment = NewEnvironment(nil)

func NewEnvironment(enclosing *Environment) Environment {
	return Environment{
		enclosing: enclosing,
		values:    make(map[string]Value),
	}
}

func (e *Environment) Define(name string, value Value) {
	e.values[name] = value
}

func (e *Environment) Assign(name string, value Value) {
	if _, ok := e.values[name]; !ok {
		e.enclosing.Assign(name, value)
		return
	}

	e.values[name] = value
}

func (e *Environment) Access(name string) Value {
	if v, ok := e.values[name]; ok {
		return v
	}

	if e.enclosing != nil {
		return e.enclosing.Access(name)
	}

	panic(fmt.Sprintf("Undefined variable %s.", name))
}
