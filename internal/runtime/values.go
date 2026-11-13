package runtime

import (
	"fmt"
	"strings"

	"github.com/eduardo-lamounier/setly/internal/syntax"
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

type Function struct {
	Declaration syntax.FunctionExpr
	Closure     *Environment
}

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

func (f Function) String() string {
	params := make([]string, f.Arity())

	for i, parameter := range f.Declaration.Parameters {
		params[i] = fmt.Sprintf(
			"%s %s",
			parameter.Name.Lexeme, typeExprToString(parameter.Type),
		)
	}

	if f.Declaration.ReturnType != nil {
		return fmt.Sprintf("function(%s)", strings.Join(params, ","))
	}

	return fmt.Sprintf(
		"function(%s) %s",
		strings.Join(params, ","), typeExprToString(f.Declaration.ReturnType),
	)
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

func (f Function) Stringify() string {
	if f.Declaration.Name != nil {
		return "<function " + f.Declaration.Name.Lexeme + ">"
	}

	return "<function>"
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

// =============================================================================

func (f Function) Arity() int {
	return len(f.Declaration.Parameters)
}

func typeExprToString(typeExpr syntax.Expr) string {
	switch typeExpr := typeExpr.(type) {
	case syntax.VariantTypeExpr:
		var types []string
		for _, _type := range typeExpr.Types {
			types = append(types, typeExprToString(_type))
		}
		return strings.Join(types, " | ")
	case syntax.NamedTypeExpr:
		return typeExpr.Name.Lexeme
	case syntax.FunctionTypeExpr:
		var params []string
		for _, param := range typeExpr.ParamsType {
			params = append(params, typeExprToString(param))
		}

		return fmt.Sprintf("function(%s) %s", strings.Join(params, ", "), typeExprToString(typeExpr.ReturnType))
	default:
		panic("Not implemented type expression to print.")
	}
}
