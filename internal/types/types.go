// Package types Defines the structures used to represent types, as well as
// operations on those types.
package types

import "reflect"

type Type interface {
	typ()

	Equals(other Type) bool
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
		if !t.ReturnType.Equals(other.ReturnType) {
			return false
		}

		if len(t.ParamsType) != len(other.ParamsType) {
			return false
		}

		for i := range t.ParamsType {
			if !t.ParamsType[i].Equals(other.ParamsType[i]) {
				return false
			}
		}

		return true
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

////////////////////////////////////////////////////////////////////////////////

func NewNumber() Number {
	return Number{}
}

func NewString() String {
	return String{}
}

func NewBool() Bool {
	return Bool{}
}

func NewNil() Nil {
	return Nil{}
}

func NewFunc(params []Type, ret Type) Func {
	return Func{
		ParamsType: params,
		ReturnType: ret,
	}
}

func NewUnion(types []Type) Union {
	return Union{
		Types: types,
	}
}
