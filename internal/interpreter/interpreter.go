// Package interpreter Defines and implements functions for evaluating
// a simple expression
package interpreter

import (
	"fmt"

	"github.com/eduardo-lamounier/setly/internal/runtime"
	"github.com/eduardo-lamounier/setly/internal/syntax"
)

type interpreter struct {
	Stmts []syntax.Stmt
	Opts  Options
}

type Options struct {
	ShowExprStmtsResults bool
}

func (i interpreter) isTruthy(val runtime.Value) bool {
	switch val := val.(type) {
	case runtime.Nil:
		return false
	case runtime.Boolean:
		return bool(val)
	default:
		return true
	}
}

func (i interpreter) evaluate(expr syntax.Expr) (runtime.Value, *runtime.Error) {
	switch expr := expr.(type) {
	case syntax.UnaryExpr:
		op, err := i.evaluate(expr.Operand)
		if err != nil {
			return nil, err
		}

		switch expr.Operator.Kind {
		case syntax.TokenNot:
			return runtime.Boolean(!i.isTruthy(op)), nil
		case syntax.TokenMinus:
			return -op.(runtime.Number), nil
		case syntax.TokenPlus:
			return op.(runtime.Number), nil
		default:
			panic("Some unary operator not implemented.")
		}
	case syntax.BinaryExpr:
		l, err := i.evaluate(expr.Operands[0])
		if err != nil {
			return nil, err
		}

		r, err := i.evaluate(expr.Operands[1])
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
		return i.evaluate(expr.Inner)
	default:
		panic("Not implemented expr kind to evaluate.")
	}
}

func (i interpreter) execute(stmt syntax.Stmt) *runtime.Error {
	switch stmt := stmt.(type) {
	case syntax.PrintStmt:
		value, err := i.evaluate(stmt.Expr)
		if err != nil {
			return err
		}

		fmt.Print(value.Stringify())
		if stmt.Newline {
			fmt.Println()
		}
	case syntax.ExprStmt:
		value, err := i.evaluate(stmt.Expr)

		if err == nil && i.Opts.ShowExprStmtsResults {
			fmt.Println(value)
		}

		return err
	default:
		panic("Not implemented stmt kind to execute.")
	}

	return nil
}

func Interpret(stmts []syntax.Stmt, opts Options) *runtime.Error {
	i := interpreter{
		Stmts: stmts,
		Opts:  opts,
	}

	for _, stmt := range stmts {
		if err := i.execute(stmt); err != nil {
			return err
		}
	}

	return nil
}
