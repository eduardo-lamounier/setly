// Package resolver Defines and implements functions for resolving expressions
// and statements
package resolver

import (
	"fmt"

	"github.com/eduardo-lamounier/setly/internal/diagnostics"
	"github.com/eduardo-lamounier/setly/internal/syntax"
	"github.com/eduardo-lamounier/setly/internal/types"
	"github.com/eduardo-lamounier/setly/internal/util/stack"
)

type resolver struct {
	Source string
	Scopes stack.Stack[map[string]struct{}]

	InsideFunc bool
	InsideLoop bool

	Locals      map[any]int
	Types       map[syntax.TypeExpr]int
	HadError    bool
	Diagnostics []diagnostics.Diagnostic
}

type resolvingError struct{}

func (r *resolver) reportAt(message string, loc syntax.Location) {
	r.HadError = true
	r.Diagnostics = append(
		r.Diagnostics,
		diagnostics.New(message, loc, r.Source),
	)
}

func (r *resolver) reportAtToken(message string, tok syntax.Token) {
	r.HadError = true
	r.Diagnostics = append(
		r.Diagnostics,
		diagnostics.New(message, tok.Location(), r.Source),
	)
}

func (r *resolver) beginScope() {
	r.Scopes.Push(make(map[string]struct{}))
}

func (r *resolver) endScope() {
	r.Scopes.Pop()
}

func (r *resolver) define(name syntax.Token) {
	if r.Scopes.Empty() {
		panic(
			fmt.Sprintf(
				"No scope to define '%s'. Where is the global scope?",
				name.Lexeme,
			),
		)
	}

	scope := r.Scopes.Top()
	scope[name.Lexeme] = struct{}{}
}

func (r *resolver) resolveLocal(
	astNode any, name syntax.Token,
) *resolvingError {
	for i, scope := range r.Scopes.TopToBottom() {
		if _, ok := scope[name.Lexeme]; ok {
			r.Locals[astNode] = r.Scopes.Count() - i - 1
			return nil
		}
	}

	r.reportAtToken(fmt.Sprintf("Undefined '%s'.", name.Lexeme), name)
	return &resolvingError{}
}

func (r *resolver) resolveExpr(expr syntax.Expr) (err *resolvingError) {
	switch expr := expr.(type) {
	case syntax.VarExpr:
		err = r.resolveLocal(expr, expr.Name)
	case syntax.UnaryExpr:
		r.resolveExpr(expr.Operand)
	case syntax.BinaryExpr:
		r.resolveExpr(expr.Left)
		r.resolveExpr(expr.Right)
	case syntax.GroupingExpr:
		r.resolveExpr(expr.Inner)
	case syntax.FuncExpr:
		r.beginScope()
		wereInsideFunction := r.InsideFunc
		wereInsideLoop := r.InsideLoop
		r.InsideFunc = true
		r.InsideLoop = false

		defer func() {
			r.endScope()
			r.InsideFunc = wereInsideFunction
			r.InsideLoop = wereInsideLoop
		}()

		seenNames := make(map[string]struct{})
		for _, param := range expr.Parameters {
			if _, seen := seenNames[param.Name.Lexeme]; seen {
				r.reportAtToken(
					fmt.Sprintf(
						"'%s' redeclarated in this parameter list.",
						param.Name.Lexeme,
					),
					param.Name,
				)
				err = &resolvingError{}
				break
			}

			r.define(param.Name)
			seenNames[param.Name.Lexeme] = struct{}{}
		}
		r.resolveAll(expr.Body)
	case syntax.CallExpr:
		r.resolveExpr(expr.Callee)

		for _, argument := range expr.Arguments {
			r.resolveExpr(argument)
		}
	default:
		panic("Not implemented expression kind to resolve.")
	}

	return
}

func (r *resolver) resolveType(
	typeExpr syntax.TypeExpr, name syntax.Token,
) *resolvingError {
	for i, scope := range r.Scopes.TopToBottom() {
		if _, ok := scope[name.Lexeme]; ok {
			r.Types[typeExpr] = r.Scopes.Count() - i - 1
			return nil
		}
	}

	r.reportAtToken(fmt.Sprintf("Undefined type '%s'.", name.Lexeme), name)
	return &resolvingError{}
}

// TODO: Implement
func (r *resolver) resolveTypeExpr(
	typeExpr syntax.TypeExpr,
) (typ types.Type, err *resolvingError) {
	switch typeExpr := typeExpr.(type) {
	case syntax.NamedTypeExpr:
	case syntax.UnionTypeExpr:
	case syntax.FuncTypeExpr:
	default:
		panic("Not implemented type expression kind to resolve.")
	}

	return
}

func (r *resolver) resolveStmt(stmt syntax.Stmt) (err *resolvingError) {
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
		case syntax.FuncExpr:
			r.define(stmt.Name)
			r.resolveExpr(expr)
		default:
			r.resolveExpr(expr)
			r.define(stmt.Name)
		}
	case syntax.AssignStmt:
		r.resolveExpr(stmt.Value)
		r.resolveLocal(stmt, stmt.Name)
	case syntax.ExprStmt:
		r.resolveExpr(stmt.Expr)
	case syntax.IfStmt:
		r.resolveExpr(stmt.Condition)
		r.resolveAll(stmt.ThenBranch)

		if stmt.ElseBranch != nil {
			r.resolveAll(stmt.ElseBranch)
		}
	case syntax.WhileStmt:
		wereInsideLoop := r.InsideLoop
		r.InsideLoop = true

		defer func() {
			r.InsideLoop = wereInsideLoop
		}()

		r.resolveExpr(stmt.Condition)
		r.resolveAll(stmt.Block)
	case syntax.ForNumStmt:
		wereInsideLoop := r.InsideLoop
		r.InsideLoop = true

		defer func() {
			r.InsideLoop = wereInsideLoop
		}()

		err = r.resolveLocal(stmt.Start, stmt.Var)
		r.resolveExpr(stmt.Limit)
		r.resolveExpr(stmt.Step)

		r.resolveAll(stmt.Block)
	case syntax.PrintStmt:
		r.resolveExpr(stmt.Expr)
	case syntax.ReturnStmt:
		if !r.InsideFunc {
			r.reportAt("Can't return outside a function.", stmt.Location())
			err = &resolvingError{}
			return
		}

		r.resolveExpr(stmt.Value)
	case syntax.BreakStmt:
		if !r.InsideLoop {
			r.reportAt("Can't break outside a loop.", stmt.Location())
			err = &resolvingError{}
			return
		}
	case syntax.ContinueStmt:
		if !r.InsideLoop {
			r.reportAt("Can't continue outside a loop.", stmt.Location())
			err = &resolvingError{}
			return
		}
	case syntax.TypeStmt:
		// TODO: Resolve type declarations statements?
	default:
		panic("Not implemented statement kind to resolve.")
	}

	return
}

func (r *resolver) resolveAll(stmts []syntax.Stmt) {
	for _, stmt := range stmts {
		r.resolveStmt(stmt)
	}
}

func Resolve(source string, stmts []syntax.Stmt) (
	locals map[any]int,
	hadError bool,
	ds []diagnostics.Diagnostic,
) {
	global := make(map[string]struct{})
	r := resolver{
		Source: source,
		Scopes: stack.New(global),
		Locals: make(map[any]int),
	}

	r.resolveAll(stmts)
	return r.Locals, r.HadError, r.Diagnostics
}
