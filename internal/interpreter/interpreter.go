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

type completion interface {
	completion()
}

type normalCompletion struct{}

type returnCompletion struct {
	Value runtime.Value
}

type breakCompletion struct{}

type continueCompletion struct{}

func (c returnCompletion) completion()   {}
func (c normalCompletion) completion()   {}
func (c breakCompletion) completion()    {}
func (c continueCompletion) completion() {}

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
) (*runtime.Error, completion) {
	prev := i.Environment

	i.Environment = env
	defer func() { i.Environment = prev }()

	for _, stmt := range stmts {
		if err, compl := i.execute(stmt); err != nil {
			return err, nil
		} else {
			switch compl.(type) {
			case normalCompletion:
				continue
			default:
				return nil, compl
			}
		}
	}

	return nil, normalCompletion{}
}

func (i *interpreter) CallFunction(
	function runtime.Function, arguments []runtime.Value,
) (runtime.Value, *runtime.Error) {
	environment := runtime.NewEnvironment(function.Closure)

	for i, param := range function.Declaration.Parameters {
		environment.Define(param.Name.Lexeme, arguments[i])
	}

	err, ret := i.executeBlock(function.Declaration.Body, environment)

	if err != nil {
		return nil, err
	}

	switch ret := ret.(type) {
	case returnCompletion:
		return ret.Value, nil
	default:
		return runtime.Nil{}, nil
	}
}

func (i *interpreter) evaluate(
	expr syntax.Expr,
) (runtime.Value, *runtime.Error) {
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
	case syntax.FunctionExpr:
		return runtime.Function{
			Declaration: expr,
			Closure:     i.Environment,
		}, nil
	case syntax.CallExpr:
		callee, err := i.evaluate(expr.Callee)
		if err != nil {
			return nil, err
		}

		var arguments []runtime.Value
		for _, argument := range expr.Arguments {
			v, err := i.evaluate(argument)
			if err != nil {
				return nil, err
			}

			arguments = append(arguments, v)
		}

		function := callee.(runtime.Function)
		return i.CallFunction(function, arguments)
	default:
		panic("Not implemented expr kind to evaluate.")
	}
}

func (i *interpreter) execute(
	stmt syntax.Stmt,
) (*runtime.Error, completion) {
	if stmt == nil {
		panic("Tried running no statement.")
	}

	switch stmt := stmt.(type) {
	case syntax.PrintStmt:
		// Returns early if no expr was passed:
		if stmt.Expr == nil {
			fmt.Println()
			return nil, nil
		}

		value, err := i.evaluate(stmt.Expr)
		if err != nil {
			return err, nil
		}

		fmt.Println(value.Stringify())
	case syntax.ExprStmt:
		value, err := i.evaluate(stmt.Expr)
		if err != nil {
			return err, nil
		}

		if i.Opts.ShowExprStmtsResults {
			fmt.Println(value)
		}
	case syntax.VarStmt:
		var value runtime.Value = runtime.Nil{}
		if stmt.Initializer != nil {
			var err *runtime.Error

			value, err = i.evaluate(stmt.Initializer)
			if err != nil {
				return err, nil
			}
		}

		i.Environment.Define(stmt.Name.Lexeme, value)
	case syntax.BlockStmt:
		return i.executeBlock(stmt.Stmts, runtime.NewEnvironment(i.Environment))
	case syntax.IfStmt:
		condition, err := i.evaluate(stmt.Condition)
		if err != nil {
			return err, nil
		}

		var compl completion = normalCompletion{}

		if i.isTruthy(condition) {
			err, compl = i.executeBlock(
				stmt.ThenBranch,
				runtime.NewEnvironment(i.Environment),
			)
		} else if stmt.ElseBranch != nil {
			err, compl = i.executeBlock(
				stmt.ElseBranch,
				runtime.NewEnvironment(i.Environment),
			)
		}

		if err != nil {
			return err, nil
		}

		return nil, compl
	case syntax.WhileStmt:
		for {
			cond, err := i.evaluate(stmt.Condition)
			if err != nil {
				return err, nil
			}

			if !i.isTruthy(cond) {
				break
			}

			var compl completion
			err, compl = i.executeBlock(
				stmt.Block,
				runtime.NewEnvironment(i.Environment),
			)

			if err != nil {
				return err, nil
			}

			switch compl.(type) {
			case breakCompletion:
				return nil, nil
			case returnCompletion:
				return nil, compl
			}
		}
	case syntax.ForNumStmt:
		prev := i.Environment

		i.Environment = runtime.NewEnvironment(i.Environment)
		defer func() { i.Environment = prev }()

		start, err := i.evaluate(stmt.Start)
		if err != nil {
			return err, nil
		}

		i.Environment.Define(stmt.Var.Lexeme, start)

		limit, err := i.evaluate(stmt.Limit)
		if err != nil {
			return err, nil
		}

		step, err := i.evaluate(stmt.Step)
		if err != nil {
			return err, nil
		}

		isUpperLimit := limit.(runtime.Number) > start.(runtime.Number)

		// non-inclusive
		stop := func() bool {
			controlVar := i.Environment.Access(stmt.Var.Lexeme).(runtime.Number)
			if isUpperLimit {
				return controlVar >= limit.(runtime.Number)
			} else {
				return controlVar <= limit.(runtime.Number)
			}
		}

		for !stop() {
			controlVar := i.Environment.Access(stmt.Var.Lexeme).(runtime.Number)

			err, compl := i.executeBlock(
				stmt.Block, runtime.NewEnvironment(i.Environment),
			)

			if err != nil {
				return err, nil
			}

			switch compl.(type) {
			case breakCompletion:
				return nil, nil
			case returnCompletion:
				return nil, compl
			}

			i.Environment.Assign(stmt.Var.Lexeme, controlVar+step.(runtime.Number))
		}
	case syntax.ReturnStmt:
		if stmt.Value == nil {
			return nil, returnCompletion{Value: runtime.Nil{}}
		}

		v, err := i.evaluate(stmt.Value)
		if err != nil {
			return err, nil
		}

		return nil, returnCompletion{Value: v}
	case syntax.BreakStmt:
		return nil, breakCompletion{}
	case syntax.ContinueStmt:
		return nil, continueCompletion{}
	default:
		panic("Not implemented stmt kind to execute.")
	}

	return nil, normalCompletion{}
}

// TODO: Implement
func Resolve(expr syntax.Expr, depth int) {
}

func Interpret(stmts []syntax.Stmt, opts Options) *runtime.Error {
	i := interpreter{
		Environment: runtime.GlobalEnvironment(),
		Stmts:       stmts,
		Opts:        opts,
	}

	for _, stmt := range stmts {
		if err, _ := i.execute(stmt); err != nil {
			return err
		}
	}

	return nil
}
