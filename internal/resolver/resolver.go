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

