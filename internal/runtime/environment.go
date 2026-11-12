package runtime

import (
	"fmt"
)

type Environment struct {
	enclosing *Environment
	values    map[string]Value
}

var globalEnvironment *Environment

func GlobalEnvironment() *Environment {
	if globalEnvironment == nil {
		globalEnvironment = NewEnvironment(nil)
	}

	return globalEnvironment
}

func NewEnvironment(enclosing *Environment) *Environment {
	return &Environment{
		enclosing: enclosing,
		values:    make(map[string]Value),
	}
}

func (e *Environment) ancestor(distance int) *Environment {
	env := e
	for range distance {
		env = env.enclosing

		if env == nil {
			panic("Tried accesing an undefined environment. Invalid distance.")
		}
	}

	return env
}

func (e *Environment) Define(name string, value Value) {
	e.values[name] = value
}

func (e *Environment) Assign(name string, value Value) {
	e.values[name] = value
}

func (e *Environment) AssignAt(name string, value Value, distance int) {
	e.ancestor(distance).Assign(name, value)
}

func (e *Environment) Access(name string) Value {
	if v, ok := e.values[name]; ok {
		return v
	}

	panic(fmt.Sprintf("Undefined variable %s.", name))
}

func (e *Environment) AccessAt(name string, distance int) Value {
	return e.ancestor(distance).Access(name)
}
