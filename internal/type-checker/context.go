package typechecker

import (
	"fmt"

	"github.com/eduardo-lamounier/setly/internal/types"
)

type context struct {
	enclosing *context
	types     map[string]types.Type
}

// Should not be accesed directly. Use [globalContext].
var global *context

func globalContext() *context {
	if global == nil {
		global = newContext(nil)

		global.define("number", types.Number{})
		global.define("string", types.String{})
		global.define("bool", types.Bool{})
		global.define("nil", types.Nil{})
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

func (c *context) define(name string, typ types.Type) {
	c.types[name] = typ
}

func (c context) access(name string) types.Type {
	if typ, ok := c.types[name]; ok {
		return typ
	}

	panic(fmt.Sprintf("Undefined type '%s'.", name))
}

func (c context) accessAt(name string, distance int) types.Type {
	return c.ancestor(distance).access(name)
}
