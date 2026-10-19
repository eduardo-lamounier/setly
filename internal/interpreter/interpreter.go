// Package interpreter Defines and implements functions for evaluating
// a simple expression
package interpreter

import (
	"github.com/eduardo-lamounier/setly/internal/diagnostics"
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

func Evaluate(expr syntax.Expr) (any, bool, *diagnostics.Diagnostic) {
	switch expr := expr.(type) {
	case syntax.UnaryExpr:
		op, hadError, d := Evaluate(expr.Operand)

		if hadError {
			return nil, true, d
		}

		switch expr.Operator.Kind {
		case syntax.TokenNot:
			return !isTruthy(op), false, nil
		case syntax.TokenMinus:
			return -op.(float64), false, nil
		case syntax.TokenPlus:
			return op.(float64), false, nil
		default:
			panic("Some unary operator not implemented.")
		}
	case syntax.BinaryExpr:
		l, hadError, d := Evaluate(expr.Operands[0])

		if hadError {
			return nil, true, d
		}

		r, hadError, d := Evaluate(expr.Operands[1])

		if hadError {
			return nil, false, d
		}

		switch expr.Operator.Kind {
		case syntax.TokenPlus:
			switch l.(type) {
			case float64:
				return l.(float64) + r.(float64), false, nil
			case string:
				return l.(string) + r.(string), false, nil
			}
		case syntax.TokenMinus:
			return l.(float64) - r.(float64), false, nil
		case syntax.TokenStar:
			return l.(float64) * r.(float64), false, nil
		case syntax.TokenSlash:
			// TODO: Handle division by zero with runtime error
			return l.(float64) / r.(float64), false, nil
		case syntax.TokenGreater:
			return l.(float64) > r.(float64), false, nil
		case syntax.TokenLess:
			return l.(float64) < r.(float64), false, nil
		case syntax.TokenGreaterEqual:
			return l.(float64) >= r.(float64), false, nil
		case syntax.TokenLessEqual:
			return l.(float64) <= r.(float64), false, nil
		case syntax.TokenEqual:
			return l == r, false, nil
		case syntax.TokenTildeEqual:
			return l != r, false, nil
		}

		panic("Some binary operator not implemented.")
	case syntax.LiteralExpr:
		return expr.Literal(), false, nil
	case syntax.GroupingExpr:
		return Evaluate(expr.Inner)
	default:
		panic("Not implemented expr kind to evaluate.")
	}
}
