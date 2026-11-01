package runtime

import (
	"fmt"
	"strings"
)

type Value interface {
	value()

	fmt.Stringer

	Stringify() string
}

type Number float64

type String string

type Boolean bool

type Nil struct{}

type Function struct{}

func (n Number) value()   {}
func (s String) value()   {}
func (b Boolean) value()  {}
func (n Nil) value()      {}
func (f Function) value() {}

func (n Number) String() string {
	return n.Stringify()
}

func (s String) String() string {
	return fmt.Sprintf("\"%s\"", s.Stringify())
}

func (b Boolean) String() string {
	return b.Stringify()
}

func (n Nil) String() string {
	return n.Stringify()
}

func (n Number) Stringify() string {
	return strings.TrimSuffix(fmt.Sprint(float64(n)), ".0")
}

func (s String) Stringify() string {
	return string(s)
}

func (b Boolean) Stringify() string {
	if bool(b) {
		return "true"
	} else {
		return "false"
	}
}

func (n Nil) Stringify() string {
	return "nil"
}

func ValueFromLiteral(literal any) Value {
	switch value := literal.(type) {
	case float64:
		return Number(value)
	case string:
		return String(value)
	case bool:
		return Boolean(value)
	case nil:
		return Nil{}
	}

	panic("Invalid literal type to convert into a supported value.")
}
