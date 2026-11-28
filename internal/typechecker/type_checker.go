// Package typechecker Checks whether the operations and values used in the
// generated AST are coherent with the language's type system.
package typechecker

import (
	"fmt"

	"github.com/eduardo-lamounier/setly/internal/diagnostics"
	"github.com/eduardo-lamounier/setly/internal/syntax"
	"github.com/eduardo-lamounier/setly/internal/types"
)

type checker struct {
	Source  string
	Context *context
	Vars    map[any]int
	Types   map[syntax.TypeExpr]int

	HadError    bool
	Diagnostics []diagnostics.Diagnostic
}

func (c *checker) reportAt(message string, loc syntax.Location) {
	c.Diagnostics = append(
		c.Diagnostics,
		diagnostics.New(message, loc, c.Source),
	)
	c.HadError = true
}

type typeError struct{}

// Converts a [syntax.TypeExpr] type representation to a [types.Type] type
// representation.
func (c checker) elaborateType(typeExpr syntax.TypeExpr) (typ types.Type) {
	switch typeExpr := typeExpr.(type) {
	case syntax.NamedTypeExpr:
		typ = c.Context.accessTypeAt(typeExpr.Name.Lexeme, c.Types[typeExpr])
	case syntax.UnionTypeExpr:
		unionTypes := make([]types.Type, len(typeExpr.Types))
		for i, unionTypeExpr := range typeExpr.Types {
			unionTypes[i] = c.elaborateType(unionTypeExpr)
		}

		typ = types.NewUnion(unionTypes...)
	case syntax.FuncTypeExpr:
		params := make([]types.Type, len(typeExpr.ParamsType))
		for i, paramTypeExpr := range typeExpr.ParamsType {
			params[i] = c.elaborateType(paramTypeExpr)
		}

		ret := c.elaborateType(typeExpr.ReturnType)

		typ = types.NewFunc(params, ret)
	}

	return
}

func (c *checker) inferExpr(expr syntax.Expr) (typ types.Type, err *typeError) {
	switch expr := expr.(type) {
	case syntax.LiteralExpr:
		typ = types.FromLiteral(expr.Value())
	case syntax.UnaryExpr:
		typ, err = c.inferExpr(expr.Operand)

		if expr.Operator.Kind == syntax.TokenNot {
			typ = types.Bool{}
		}
	case syntax.BinaryExpr:
		var l types.Type
		l, err = c.inferExpr(expr.Left)
		if err != nil {
			return
		}

		var r types.Type
		r, err = c.inferExpr(expr.Right)
		if err != nil {
			return
		}

		typ = l

		switch expr.Operator.Kind {
		case syntax.TokenOr, syntax.TokenAnd:
			if !l.Equals(r) {
				typ = types.NewUnion(l, r)
			}
		case syntax.TokenPlus,
			syntax.TokenMinus,
			syntax.TokenStar,
			syntax.TokenSlash:

			if !l.Equals(types.Number{}) && !l.Equals(types.String{}) {
				c.reportAt(
					fmt.Sprintf(
						"Operator '%s' not defined for %s.", expr.Operator.Lexeme, l,
					),
					expr.Left.Location(),
				)
				err = &typeError{}
				return
			}

			if !l.Equals(r) {
				c.reportAt(
					fmt.Sprintf(
						"Type mismatch (%s and %s) for operator '%s'.",
						l, r, expr.Operator.Lexeme,
					),
					expr.Location(),
				)
				err = &typeError{}
				return
			}
		default:
			panic(
				fmt.Sprintf(
					"Not implemented binary operator to type check: '%s'.",
					expr.Operator.Lexeme,
				),
			)
		}
	case syntax.GroupingExpr:
		typ, err = c.inferExpr(expr.Inner)
	case syntax.VarExpr:
		typ = c.Context.accessVarAt(expr.Name.Lexeme, c.Vars[expr])
	case syntax.CallExpr:
		var calleeType types.Type
		calleeType, err = c.inferExpr(expr.Callee)
		if err != nil {
			return
		}

		switch calleeType := calleeType.(type) {
		case types.Func:
			for i, argument := range expr.Arguments {
				var argType types.Type
				argType, err = c.inferExpr(argument)

				if !argType.Equals(calleeType.ParamsType[i]) {
					c.reportAt(
						fmt.Sprintf(
							"Cannot pass argument of type %s where %s is expected.",
							argType, calleeType.ParamsType[i],
						),
						expr.Location(),
					)
					err = &typeError{}
				}
			}

			typ = calleeType.ReturnType
		default:
			c.reportAt(
				fmt.Sprintf("Cannot call %s.", calleeType), expr.Callee.Location(),
			)
			err = &typeError{}
		}
	case syntax.FuncExpr:
		returnType := c.elaborateType(expr.ReturnType)
		paramsType := make([]types.Type, len(expr.Parameters))

		for _, param := range expr.Parameters {
			paramType := c.elaborateType(param.Type)
			c.Context.defineVar(param.Name.Lexeme, paramType)
		}

		typ = types.NewFunc(paramsType, returnType)
	default:
		panic("Not implemented expression kind to type check.")
	}

	return
}

func (c *checker) checkBlock(
	stmts []syntax.Stmt, ctx *context,
) (err *typeError) {
	prev := c.Context

	c.Context = ctx
	defer func() { c.Context = prev }()

	for _, stmt := range stmts {
		if e := c.checkStmt(stmt); err == nil {
			err = e
		}
	}

	return
}

func (c *checker) checkStmt(stmt syntax.Stmt) (err *typeError) {
	switch stmt := stmt.(type) {
	case syntax.ExprStmt:
		_, err = c.inferExpr(stmt.Expr)
	case syntax.PrintStmt:
		_, err = c.inferExpr(stmt.Expr)
	case syntax.VarStmt:
		typ := c.elaborateType(stmt.Type)

		// Just like in the resolver, a function must be able to refer to itself:
		switch typ.(type) {
		case types.Func:
			c.Context.defineVar(stmt.Name.Lexeme, typ)
		}

		var got types.Type
		got, err = c.inferExpr(stmt.Initializer)
		if err != nil {
			return
		}

		if !typ.Equals(got) {
			c.reportAt(
				fmt.Sprintf(
					"Cannot initialize variable of type %s with value of type %s.",
					typ, got,
				),
				stmt.Initializer.Location(),
			)
			err = &typeError{}
			return
		}

		// A normal variable can't be able to refer to itself, so it needs to be
		// defined after type checking the initializer:
		switch typ.(type) {
		case types.Func:
		default:
			c.Context.defineVar(stmt.Name.Lexeme, typ)
		}
	case syntax.AssignStmt:
		typ := c.Context.accessVarAt(stmt.Name.Lexeme, c.Vars[stmt])

		var got types.Type
		got, err = c.inferExpr(stmt.Value)

		if !typ.Equals(got) {
			c.reportAt(
				fmt.Sprintf(
					"Cannot assign value of type %s to variable of type %s.",
					got, typ,
				),
				stmt.Value.Location(),
			)
			err = &typeError{}
			return
		}
	case syntax.BlockStmt:
		err = c.checkBlock(stmt.Stmts, newContext(c.Context))
	case syntax.IfStmt:
		_, err = c.inferExpr(stmt.Condition)

		e := c.checkBlock(stmt.ThenBranch, newContext(c.Context))
		if err == nil {
			err = e
		}

		e = c.checkBlock(stmt.ElseBranch, newContext(c.Context))
		if err == nil {
			err = e
		}
	case syntax.WhileStmt:
		_, err = c.inferExpr(stmt.Condition)

		e := c.checkBlock(stmt.Block, newContext(c.Context))
		if err == nil {
			err = e
		}
	case syntax.ForNumStmt:
		var startType types.Type
		startType, err = c.inferExpr(stmt.Start)

		if !startType.Equals(types.Number{}) {
			c.reportAt(
				fmt.Sprintf(
					"Expected start value to be a %s, but got value of type %s instead.",
					types.Number{}, startType,
				),
				stmt.Start.Location(),
			)
			err = &typeError{}
			return
		}

		limitType, e := c.inferExpr(stmt.Limit)
		if err == nil {
			err = e
		}

		if !limitType.Equals(types.Number{}) {
			c.reportAt(
				fmt.Sprintf(
					"Expected limit value to be a %s, but got value of type %s instead.",
					types.Number{}, limitType,
				),
				stmt.Limit.Location(),
			)
			err = &typeError{}
			return
		}

		var stepType types.Type
		stepType, e = c.inferExpr(stmt.Step)
		if err == nil {
			err = e
		}

		if !stepType.Equals(types.Number{}) {
			c.reportAt(
				fmt.Sprintf(
					"Expected step value to be a %s, but got value of type %s instead.",
					types.Number{}, stepType,
				),
				stmt.Step.Location(),
			)
			err = &typeError{}
			return
		}

		e = c.checkBlock(stmt.Block, newContext(c.Context))
		if err == nil {
			err = e
		}
	case syntax.ReturnStmt:
		// TODO: Check for the right return type. Since the resolver emitted
		// no errors, we can be sure any return statement is inside a function.
		_, err = c.inferExpr(stmt.Value)
	case syntax.ContinueStmt, syntax.BreakStmt:
	case syntax.TypeStmt:
		c.Context.defineType(stmt.Name.Lexeme, c.elaborateType(stmt.Value))
	default:
		panic("Not implemented statement kind to type check.")
	}

	return
}

func Check(
	source string,
	stmts []syntax.Stmt,
	vars map[any]int,
	types map[syntax.TypeExpr]int,
) (bool, []diagnostics.Diagnostic) {
	c := checker{
		Source:  source,
		Context: globalContext(),
		Vars:    vars,
		Types:   types,
	}

	for _, stmt := range stmts {
		_ = c.checkStmt(stmt)
	}

	return c.HadError, c.Diagnostics
}
