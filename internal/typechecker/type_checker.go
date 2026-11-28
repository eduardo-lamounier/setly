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
		name := expr.Name.Lexeme
		typ = c.Context.accessTypeAt(name, c.Vars[name])
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

func (c *checker) checkStmt(stmt syntax.Stmt) (err *typeError) {
	switch stmt := stmt.(type) {
	case syntax.ExprStmt:
	case syntax.PrintStmt:
	case syntax.VarStmt:
	case syntax.AssignStmt:
	case syntax.BlockStmt:
	case syntax.IfStmt:
	case syntax.WhileStmt:
	case syntax.ForNumStmt:
	case syntax.ReturnStmt:
	case syntax.BreakStmt:
	case syntax.ContinueStmt:
	case syntax.TypeStmt:
	default:
		panic("Not implemented statement kind to type check.")
	}

	panic("Type check for expressions not yet implemented.")
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
