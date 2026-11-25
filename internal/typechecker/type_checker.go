// Package typechecker Checks whether the operations and values used in the
// generated AST are coherent with the language's type system.
package typechecker

import (
	"github.com/eduardo-lamounier/setly/internal/diagnostics"
	"github.com/eduardo-lamounier/setly/internal/syntax"
	"github.com/eduardo-lamounier/setly/internal/types"
)

type checker struct {
	HadError    bool
	Diagnostics []diagnostics.Diagnostic
}

type typeError struct{}

func (c *checker) inferExpr(expr syntax.Expr) (typ types.Type, err *typeError) {
	switch expr := expr.(type) {
	case syntax.LiteralExpr:
	case syntax.UnaryExpr:
	case syntax.BinaryExpr:
	case syntax.GroupingExpr:
	case syntax.VarExpr:
	case syntax.CallExpr:
	case syntax.FuncExpr:
	default:
		panic("Not implemented expression kind to type check.")
	}

	panic("Type check for expressions not yet implemented.")
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
	types map[syntax.TypeExpr]int,
) (bool, []diagnostics.Diagnostic) {
	c := checker{
		Source: source,
		Types:  types,
	}

	for _, stmt := range stmts {
		_ = c.checkStmt(stmt)
	}

	return c.HadError, c.Diagnostics
}
