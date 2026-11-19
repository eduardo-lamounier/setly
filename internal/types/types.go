// Package types Defines the structures used to represent types, as well as
// operations on those types.
package types

import "reflect"

type Type interface {
	typ()
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
