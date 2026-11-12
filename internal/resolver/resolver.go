// Package resolver Defines and implements functions for resolving expressions
// and statements
package resolver

import (
	"github.com/eduardo-lamounier/setly/internal/syntax"
	"github.com/eduardo-lamounier/setly/internal/util/stack"
)

type resolver struct {
	Source string
	Scopes stack.Stack[map[string]struct{}]
	Locals map[syntax.Expr]int
}

func (r *resolver) beginScope() {
	r.Scopes.Push(make(map[string]struct{}))
}

func (r *resolver) endScope() {
	r.Scopes.Pop()
}

func (r *resolver) define(name syntax.Token) {
	if !r.Scopes.Empty() {
		scope := r.Scopes.Top()
		scope[name.Lexeme] = struct{}{}
	}
}

func (r *resolver) resolveLocal(expr syntax.Expr, name syntax.Token) {
	for i, scope := range r.Scopes.TopToBottom() {
		if _, ok := scope[name.Lexeme]; ok {
			r.Locals[expr] = r.Scopes.Count() - i
			return
		}
	}
}

func (r *resolver) resolveExpr(expr syntax.Expr) {
	switch expr := expr.(type) {
	case syntax.VarExpr:
		r.resolveLocal(expr, expr.Name)
	case syntax.AssignExpr:
		r.resolveExpr(expr.Value)
		r.resolveLocal(expr, expr.Name)
	case syntax.UnaryExpr:
		r.resolveExpr(expr.Operand)
	case syntax.BinaryExpr:
		r.resolveExpr(expr.Left)
		r.resolveExpr(expr.Right)
	case syntax.GroupingExpr:
		r.resolveExpr(expr.Inner)
	case syntax.FunctionExpr:
		r.beginScope()
		defer r.endScope()

		for _, param := range expr.Parameters {
			r.define(param.Name)
		}
		r.resolveAll(expr.Body)
	case syntax.CallExpr:
		r.resolveExpr(expr.Callee)

		for _, argument := range expr.Arguments {
			r.resolveExpr(argument)
		}

	// TODO: Resolve type expressions?
	case syntax.VariantTypeExpr:
	case syntax.NamedTypeExpr:
	case syntax.FunctionTypeExpr:
	}
}

func (r *resolver) resolveStmt(stmt syntax.Stmt) {
	switch stmt := stmt.(type) {
	case syntax.BlockStmt:
		r.beginScope()
		defer r.endScope()

		r.resolveAll(stmt.Stmts)
	case syntax.VarStmt:
		if stmt.Initializer == nil {
			r.define(stmt.Name)
			break
		}

		// A function should be able to reference itself to achieve recursion,
		// so it is declared before resolving its body:
		switch expr := stmt.Initializer.(type) {
		case syntax.FunctionExpr:
			r.define(stmt.Name)
			r.resolveExpr(expr)
		default:
			r.resolveExpr(expr)
			r.define(stmt.Name)
		}
	case syntax.ExprStmt:
		r.resolveExpr(stmt.Expr)
	case syntax.IfStmt:
		r.resolveExpr(stmt.Condition)
		r.resolveAll(stmt.ThenBranch)

		if stmt.ElseBranch != nil {
			r.resolveAll(stmt.ElseBranch)
		}
	case syntax.WhileStmt:
		r.resolveExpr(stmt.Condition)
		r.resolveAll(stmt.Block)
	case syntax.ForNumStmt:
		r.resolveLocal(stmt.Start, stmt.Var)
		r.resolveExpr(stmt.Limit)
		r.resolveExpr(stmt.Step)

		r.resolveAll(stmt.Block)
	case syntax.PrintStmt:
		r.resolveExpr(stmt.Expr)
	case syntax.ReturnStmt:
		r.resolveExpr(stmt.Value)
	case syntax.TypeStmt:
		// TODO: Resolve type declarations statements?
	}
}

func (r *resolver) resolveAll(stmts []syntax.Stmt) {
	for _, stmt := range stmts {
		r.resolveStmt(stmt)
	}
}

func Resolve(source string, stmts []syntax.Stmt) (locals map[syntax.Expr]int) {
	r := resolver{
		Scopes: stack.New[map[string]struct{}](),
		Locals: make(map[syntax.Expr]int),
	}

	r.resolveAll(stmts)
	locals = r.Locals
	return
}
