// Package interpreter Defines and implements functions for evaluating
// a simple expression
package interpreter

import (
	"github.com/eduardo-lamounier/setly/internal/runtime"
	"github.com/eduardo-lamounier/setly/internal/syntax"
)

func isTruthy(val any) bool {
	switch val := val.(type) {
	case nil:
		return false
	case bool:
		return val
	default:
		return true
	}
}

func Evaluate(expr syntax.Expr) (any, *runtime.Error) {
	switch expr := expr.(type) {
	case syntax.UnaryExpr:
		op, err := Evaluate(expr.Operand)
		if err != nil {
			return nil, err
		}

		switch expr.Operator.Kind {
		case syntax.TokenNot:
			return !isTruthy(op), nil
		case syntax.TokenMinus:
			return -op.(float64), nil
		case syntax.TokenPlus:
			return op.(float64), nil
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
			case float64:
				return l.(float64) + r.(float64), nil
			case string:
				return l.(string) + r.(string), nil
			}
		case syntax.TokenMinus:
			return l.(float64) - r.(float64), nil
		case syntax.TokenStar:
			return l.(float64) * r.(float64), nil
		case syntax.TokenSlash:
			if r.(float64) == 0 {
				return nil, &runtime.Error{
					Line:    expr.Operands[0].Line(),
					Message: "Attempted division by zero.",
				}
			}

			return l.(float64) / r.(float64), nil
		case syntax.TokenGreater:
			return l.(float64) > r.(float64), nil
		case syntax.TokenLess:
			return l.(float64) < r.(float64), nil
		case syntax.TokenGreaterEqual:
			return l.(float64) >= r.(float64), nil
		case syntax.TokenLessEqual:
			return l.(float64) <= r.(float64), nil
		case syntax.TokenEqual:
			return l == r, nil
		case syntax.TokenTildeEqual:
			return l != r, nil
		}

		panic("Some binary operator not implemented.")
	case syntax.LiteralExpr:
		return expr.Literal(), nil
	case syntax.GroupingExpr:
		return Evaluate(expr.Inner)
	default:
		panic("Not implemented expr kind to evaluate.")
	}
}
