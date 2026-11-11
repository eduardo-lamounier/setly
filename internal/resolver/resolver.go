// Package resolver Defines and implements functions for resolving expressions
// and statements
package resolver

import (
	"github.com/eduardo-lamounier/setly/internal/interpreter"
	"github.com/eduardo-lamounier/setly/internal/syntax"
	"github.com/eduardo-lamounier/setly/internal/util/stack"
)

type resolver struct {
	Source string
	Scopes stack.Stack[map[string]struct{}]
}

func (r *resolver) beginScope() {
	r.Scopes.Push(make(map[string]struct{}))
}

func (r *resolver) endScope() {
	r.Scopes.Pop()
}

func (r *resolver) declare(name syntax.Token) {
	if !r.Scopes.Empty() {
		scope := r.Scopes.Top()
		scope[name.Lexeme] = struct{}{}
	}
}

func (r *resolver) resolveLocal(expr syntax.Expr, name syntax.Token) {
	for i, scope := range r.Scopes.TopToBottom() {
		if _, ok := scope[name.Lexeme]; ok {
			interpreter.Resolve(expr, r.Scopes.Count()-i)
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
			r.declare(param.Name)
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
			r.declare(stmt.Name)
			break
		}

		// A function should be able to reference itself to achieve recursion,
		// so it is declared before resolving its body:
		switch expr := stmt.Initializer.(type) {
		case syntax.FunctionExpr:
			r.declare(stmt.Name)
			r.resolveExpr(expr)
		default:
			r.resolveExpr(expr)
			r.declare(stmt.Name)
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

