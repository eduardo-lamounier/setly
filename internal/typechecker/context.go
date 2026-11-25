package typechecker

import (
	"fmt"

	"github.com/eduardo-lamounier/setly/internal/types"
)

type context struct {
	enclosing *context
	types     map[string]types.Type // Defined types
	vars      map[string]types.Type // Types associated with declared variables
}

// Should not be accesed directly. Use [globalContext].
var global *context

func globalContext() *context {
	if global == nil {
		global = newContext(nil)

		global.defineType("number", types.Number{})
		global.defineType("string", types.String{})
		global.defineType("bool", types.Bool{})
		global.defineType("nil", types.Nil{})
	}

	return global
}

func newContext(enclosing *context) *context {
	return &context{
		enclosing: enclosing,
		types:     make(map[string]types.Type),
	}
}

func (c *context) ancestor(distance int) *context {
	ctx := c
	for range distance {
		ctx = ctx.enclosing

		if ctx == nil {
			panic("Tried accessing an undefined context. Invalid distance.")
		}
	}

	return ctx
}

func (c *context) defineType(name string, typ types.Type) {
	c.types[name] = typ
}

func (c context) accessType(name string) types.Type {
	if typ, ok := c.types[name]; ok {
		return typ
	}

	panic(fmt.Sprintf("Undefined type '%s'.", name))
}

func (c context) accessTypeAt(name string, distance int) types.Type {
	return c.ancestor(distance).accessType(name)
}

func (c *context) defineVar(name string, typ types.Type) {
	c.vars[name] = typ
}

func (c context) accessVar(name string) types.Type {
	if typ, ok := c.vars[name]; ok {
		return typ
	}

	panic(fmt.Sprintf("Undefined '%s'.", name))
}

func (c context) accessVarAt(name string, distance int) types.Type {
	return c.ancestor(distance).accessType(name)
}
