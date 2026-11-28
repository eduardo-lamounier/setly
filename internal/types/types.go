// Package types Defines the structures used to represent types, as well as
// operations on those types.
package types

import (
	"fmt"
	"reflect"
	"strings"
)

type Type interface {
	typ()

	Equals(other Type) bool

	fmt.Stringer
}

type Number struct{}

type String struct{}

type Bool struct{}

type Nil struct{}

type Func struct {
	ParamsType []Type
	ReturnType Type
}

type Union struct {
	Types []Type
}

func (t Number) typ() {}
func (t String) typ() {}
func (t Bool) typ()   {}
func (t Nil) typ()    {}
func (t Func) typ()   {}
func (t Union) typ()  {}

func (t Number) Equals(other Type) bool {
	_, ok := other.(Number)
	return ok
}

func (t String) Equals(other Type) bool {
	_, ok := other.(String)
	return ok
}

func (t Bool) Equals(other Type) bool {
	_, ok := other.(Bool)
	return ok
}

func (t Nil) Equals(other Type) bool {
	_, ok := other.(Nil)
	return ok
}

func (t Func) Equals(other Type) bool {
	switch other := other.(type) {
	case Func:
		sameParameters := func() bool {
			for i := range t.ParamsType {
				if !t.ParamsType[i].Equals(other.ParamsType[i]) {
					return false
				}
			}

			return true
		}

		if t.ReturnType == nil || other.ReturnType == nil {
			if t.ReturnType != other.ReturnType {
				return false
			}

			// Both have no return type
			return sameParameters()
		}

		if !t.ReturnType.Equals(other.ReturnType) {
			return false
		}

		if len(t.ParamsType) != len(other.ParamsType) {
			return false
		}

		return sameParameters()
	default:
		return false
	}
}

func (t Union) Equals(other Type) bool {
	switch other := other.(type) {
	case Union:
		if len(t.Types) != len(other.Types) {
			return false
		}

		// Both unions should contain the same types to be considered equal.
		// For instance 'number | string' is equal to 'string | number'. Since the
		// order doesn't matter.

		count := make(map[reflect.Type]int)
		for _, typ := range t.Types {
			count[reflect.TypeOf(typ)]++
		}

		for _, typ := range other.Types {
			t := reflect.TypeOf(typ)

			if count[t] == 0 {
				return false
			}

			count[t]--
		}

		return true
	default:
		return false
	}
}

func (t Number) String() string {
	return "number"
}

func (t String) String() string {
	return "string"
}

func (t Bool) String() string {
	return "bool"
}

func (t Nil) String() string {
	return "nil"
}

func (t Func) String() string {
	params := make([]string, len(t.ParamsType))
	for i, t := range t.ParamsType {
		params[i] = fmt.Sprint(t)
	}

	return fmt.Sprintf("func(%s) -> %s", strings.Join(params, ", "), t.ReturnType)
}

func (t Union) String() string {
	types := make([]string, len(t.Types))
	for i, t := range t.Types {
		types[i] = fmt.Sprint(t)
	}

	return strings.Join(types, " | ")
}

////////////////////////////////////////////////////////////////////////////////

func NewFunc(params []Type, ret Type) Func {
	return Func{
		ParamsType: params,
		ReturnType: ret,
	}
}

func NewUnion(types ...Type) Union {
	return Union{
		Types: types,
	}
}

func FromLiteral(literal any) Type {
	switch literal.(type) {
	case float64:
		return Number{}
	case string:
		return String{}
	case bool:
		return Bool{}
	case nil:
		return Nil{}
	}

	panic("Invalid literal type to get a supported type.")
}
