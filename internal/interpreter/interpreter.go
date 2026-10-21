// Package interpreter Defines and implements functions for evaluating
// a simple expression
package interpreter

import (
	"github.com/eduardo-lamounier/setly/internal/runtime"
	"github.com/eduardo-lamounier/setly/internal/syntax"
)

func isTruthy(val runtime.Value) bool {
	switch val := val.(type) {
	case runtime.Nil:
		return false
	case runtime.Boolean:
		return bool(val)
	default:
		return true
	}
}

func Evaluate(expr syntax.Expr) (runtime.Value, *runtime.Error) {
	switch expr := expr.(type) {
	case syntax.UnaryExpr:
		op, err := Evaluate(expr.Operand)
		if err != nil {
			return nil, err
		}

		switch expr.Operator.Kind {
		case syntax.TokenNot:
			return runtime.Boolean(!isTruthy(op)), nil
		case syntax.TokenMinus:
			return -op.(runtime.Number), nil
		case syntax.TokenPlus:
			return op.(runtime.Number), nil
		default:
			panic("Some unary operator not implemented.")
		}
	case syntax.BinaryExpr:
		l, err := Evaluate(expr.Operands[0])
		if err != nil {
			return nil, err
		}

		r, err := Evaluate(expr.Operands[1])
		if err != nil {
			return nil, err
		}

		switch expr.Operator.Kind {
		case syntax.TokenPlus:
			switch l.(type) {
			case runtime.Number:
				return l.(runtime.Number) + r.(runtime.Number), nil
			case runtime.String:
				return l.(runtime.String) + r.(runtime.String), nil
			}
		case syntax.TokenMinus:
			return l.(runtime.Number) - r.(runtime.Number), nil
		case syntax.TokenStar:
			return l.(runtime.Number) * r.(runtime.Number), nil
		case syntax.TokenSlash:
			if r.(runtime.Number) == 0 {
				return nil, &runtime.Error{
					Line:    expr.Operands[0].Line(),
					Message: "Attempted division by zero.",
				}
			}

			return l.(runtime.Number) / r.(runtime.Number), nil
		case syntax.TokenGreater:
			return runtime.Boolean(l.(runtime.Number) > r.(runtime.Number)), nil
		case syntax.TokenLess:
			return runtime.Boolean(l.(runtime.Number) < r.(runtime.Number)), nil
		case syntax.TokenGreaterEqual:
			return runtime.Boolean(l.(runtime.Number) >= r.(runtime.Number)), nil
		case syntax.TokenLessEqual:
			return runtime.Boolean(l.(runtime.Number) <= r.(runtime.Number)), nil
		case syntax.TokenEqual:
			return runtime.Boolean(l == r), nil
		case syntax.TokenTildeEqual:
			return runtime.Boolean(l != r), nil
		}

		panic("Some binary operator not implemented.")
	case syntax.LiteralExpr:
		return runtime.ValueFromLiteral(expr.Literal()), nil
	case syntax.GroupingExpr:
		return Evaluate(expr.Inner)
	default:
		panic("Not implemented expr kind to evaluate.")
	}
}
