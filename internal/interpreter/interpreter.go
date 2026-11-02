// Package interpreter Defines and implements functions for evaluating
// a simple expression
package interpreter

import (
	"fmt"

	"github.com/eduardo-lamounier/setly/internal/runtime"
	"github.com/eduardo-lamounier/setly/internal/syntax"
)

type interpreter struct {
	Environment *runtime.Environment
	Stmts       []syntax.Stmt
	Opts        Options
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

func (i *interpreter) executeBlock(
	stmts []syntax.Stmt, env *runtime.Environment,
) *runtime.Error {
	prev := i.Environment

	i.Environment = env
	defer func() { i.Environment = prev }()

	for _, stmt := range stmts {
		if err := i.execute(stmt); err != nil {
			return err
		}
	}

	return nil
}


func (i *interpreter) evaluate(expr syntax.Expr) (runtime.Value, *runtime.Error) {
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
		l, err := i.evaluate(expr.Left)
		if err != nil {
			return nil, err
		}

		// Returns early in logical expressions depending on what the left side
		// evaluates to:
		switch expr.Operator.Kind {
		case syntax.TokenOr:
			if i.isTruthy(l) {
				return l, nil
			}
		case syntax.TokenAnd:
			if !i.isTruthy(l) {
				return l, nil
			}
		}

		r, err := i.evaluate(expr.Right)
		if err != nil {
			return nil, err
		}

		switch expr.Operator.Kind {
		case syntax.TokenOr:
			// left side is "falsy" if we got here
			return r, nil
		case syntax.TokenAnd:
			// left side is "truthy" if we got here
			return r, nil
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
					Line:    expr.Line(),
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
		return runtime.ValueFromLiteral(expr.Literal), nil
	case syntax.GroupingExpr:
		return i.evaluate(expr.Inner)
	case syntax.VarExpr:
		return i.Environment.Access(expr.Name.Lexeme), nil
	case syntax.AssignExpr:
		value, err := i.evaluate(expr.Value)
		if err != nil {
			return nil, err
		}

		i.Environment.Assign(expr.Name.Lexeme, value)
		return i.Environment.Access(expr.Name.Lexeme), nil
	default:
		panic("Not implemented expr kind to evaluate.")
	}
}

func (i *interpreter) execute(stmt syntax.Stmt) *runtime.Error {
	if stmt == nil {
		panic("Tried running no statement.")
	}

	switch stmt := stmt.(type) {
	case syntax.PrintStmt:
		// Returns early if no expr was passed:
		if stmt.Expr == nil {
			fmt.Println()
			return nil
		}

		value, err := i.evaluate(stmt.Expr)
		if err != nil {
			return err
		}

		fmt.Println(value.Stringify())
	case syntax.ExprStmt:
		value, err := i.evaluate(stmt.Expr)

		if err == nil && i.Opts.ShowExprStmtsResults {
			fmt.Println(value)
		}

		return err
	case syntax.VarStmt:
		var value runtime.Value = runtime.Nil{}
		if stmt.Initializer != nil {
			var err *runtime.Error

			value, err = i.evaluate(stmt.Initializer)
			if err != nil {
				return err
			}
		}

		i.Environment.Define(stmt.Name.Lexeme, value)
		return nil
	case syntax.BlockStmt:
		return i.executeBlock(stmt.Stmts, runtime.NewEnvironment(i.Environment))
	case syntax.IfStmt:
		condition, err := i.evaluate(stmt.Condition)
		if err != nil {
			return err
		}

		if i.isTruthy(condition) {
			err = i.execute(stmt.ThenBranch)
		} else if stmt.ElseBranch != nil {
			err = i.execute(stmt.ElseBranch)
		}

		return err
	case syntax.WhileStmt:
		for {
			cond, err := i.evaluate(stmt.Condition)
			if err != nil {
				return err
			}

			if !i.isTruthy(cond) {
				break
			}

			err = i.execute(stmt.Block)
			if err != nil {
				return err
			}
		}
	default:
		panic("Not implemented stmt kind to execute.")
	}

	return nil
}

func Interpret(stmts []syntax.Stmt, opts Options) *runtime.Error {
	i := interpreter{
		Environment: runtime.GlobalEnvironment(),
		Stmts:       stmts,
		Opts:        opts,
	}

	for _, stmt := range stmts {
		if err := i.execute(stmt); err != nil {
			return err
		}
	}

	return nil
}
