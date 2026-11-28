// Package resolver Defines and implements functions for resolving expressions
// and statements
package resolver

import (
	"fmt"

	"github.com/eduardo-lamounier/setly/internal/diagnostics"
	"github.com/eduardo-lamounier/setly/internal/syntax"
	"github.com/eduardo-lamounier/setly/internal/util/stack"
)

type scope struct {
	Vars  map[string]struct{}
	Types map[string]struct{}
}

func newScope() scope {
	return scope{
		Vars:  make(map[string]struct{}),
		Types: make(map[string]struct{}),
	}
}

type resolver struct {
	Source string
	Scopes stack.Stack[scope]

	InsideFunc bool
	InsideLoop bool

	Vars        map[any]int
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
	r.reportAt(message, tok.Location())
}

func (r *resolver) beginScope() {
	r.Scopes.Push(newScope())
}

func (r *resolver) endScope() {
	r.Scopes.Pop()
}

func (r *resolver) defineVar(name string) {
	if r.Scopes.Empty() {
		panic(
			fmt.Sprintf(
				"No scope to define '%s'. Where is the global scope?",
				name,
			),
		)
	}

	scope := r.Scopes.Top()
	scope.Vars[name] = struct{}{}
}

func (r *resolver) defineType(name string) {
	if r.Scopes.Empty() {
		panic(
			fmt.Sprintf(
				"No scope to define '%s'. Where is the global scope?",
				name,
			),
		)
	}

	scope := r.Scopes.Top()
	scope.Types[name] = struct{}{}
}

func (r *resolver) resolveVar(
	astNode any, name syntax.Token,
) *resolvingError {
	for i, scope := range r.Scopes.TopToBottom() {
		if _, ok := scope.Vars[name.Lexeme]; ok {
			r.Vars[astNode] = r.Scopes.Count() - i - 1
			return nil
		}
	}

	r.reportAtToken(fmt.Sprintf("Undefined '%s'.", name.Lexeme), name)
	return &resolvingError{}
}

func (r *resolver) resolveExpr(expr syntax.Expr) (err *resolvingError) {
	switch expr := expr.(type) {
	case syntax.LiteralExpr:
	case syntax.VarExpr:
		err = r.resolveVar(expr, expr.Name)
	case syntax.UnaryExpr:
		err = r.resolveExpr(expr.Operand)
	case syntax.BinaryExpr:
		err = r.resolveExpr(expr.Left)
		e := r.resolveExpr(expr.Right)

		if err == nil {
			err = e
		}
	case syntax.GroupingExpr:
		err = r.resolveExpr(expr.Inner)
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

			r.defineVar(param.Name.Lexeme)
			seenNames[param.Name.Lexeme] = struct{}{}
		}
		e := r.resolveAll(expr.Body)

		if err == nil {
			err = e
		}
	case syntax.CallExpr:
		// TODO: Handle wrong amount of passed arguments
		err = r.resolveExpr(expr.Callee)

		for _, argument := range expr.Arguments {
			e := r.resolveExpr(argument)
			if err == nil {
				err = e
			}
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
		if _, ok := scope.Types[name.Lexeme]; ok {
			r.Types[typeExpr] = r.Scopes.Count() - i - 1
			return nil
		}
	}

	r.reportAtToken(fmt.Sprintf("Undefined type '%s'.", name.Lexeme), name)
	return &resolvingError{}
}

func (r *resolver) resolveTypeExpr(
	typeExpr syntax.TypeExpr,
) (err *resolvingError) {
	switch typeExpr := typeExpr.(type) {
	case syntax.NamedTypeExpr:
		err = r.resolveType(typeExpr, typeExpr.Name)
	case syntax.UnionTypeExpr:
		for _, typ := range typeExpr.Types {
			e := r.resolveTypeExpr(typ)

			if err == nil {
				err = e
			}
		}
	case syntax.FuncTypeExpr:
		for _, param := range typeExpr.ParamsType {
			e := r.resolveTypeExpr(param)

			if err == nil {
				err = e
			}
		}

		e := r.resolveTypeExpr(typeExpr.ReturnType)

		if err == nil {
			err = e
		}
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

		err = r.resolveAll(stmt.Stmts)
	case syntax.VarStmt:
		if stmt.Initializer == nil {
			r.defineVar(stmt.Name.Lexeme)
			break
		}

		// A function should be able to reference itself to achieve recursion,
		// so it is declared before resolving its body:
		switch expr := stmt.Initializer.(type) {
		case syntax.FuncExpr:
			r.defineVar(stmt.Name.Lexeme)
			err = r.resolveExpr(expr)
		default:
			err = r.resolveExpr(expr)
			r.defineVar(stmt.Name.Lexeme)
		}
	case syntax.AssignStmt:
		err = r.resolveExpr(stmt.Value)
		e := r.resolveVar(stmt, stmt.Name)

		if err == nil {
			err = e
		}
	case syntax.ExprStmt:
		err = r.resolveExpr(stmt.Expr)
	case syntax.IfStmt:
		err = r.resolveExpr(stmt.Condition)
		e := r.resolveAll(stmt.ThenBranch)

		if err == nil {
			err = e
		}

		if stmt.ElseBranch != nil {
			e := r.resolveAll(stmt.ElseBranch)

			if err == nil {
				err = e
			}
		}
	case syntax.WhileStmt:
		wereInsideLoop := r.InsideLoop
		r.InsideLoop = true

		defer func() {
			r.InsideLoop = wereInsideLoop
		}()

		err = r.resolveExpr(stmt.Condition)
		e := r.resolveAll(stmt.Block)

		if err == nil {
			err = e
		}
	case syntax.ForNumStmt:
		wereInsideLoop := r.InsideLoop
		r.InsideLoop = true

		defer func() {
			r.InsideLoop = wereInsideLoop
		}()

		err = r.resolveVar(stmt.Start, stmt.Var)

		e := r.resolveExpr(stmt.Limit)
		if err == nil {
			err = e
		}

		e = r.resolveExpr(stmt.Step)
		if err == nil {
			err = e
		}

		e = r.resolveAll(stmt.Block)
		if err == nil {
			err = e
		}
	case syntax.PrintStmt:
		err = r.resolveExpr(stmt.Expr)
	case syntax.ReturnStmt:
		if !r.InsideFunc {
			r.reportAt("Can't return outside a function.", stmt.Location())
			err = &resolvingError{}
			return
		}

		err = r.resolveExpr(stmt.Value)
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
		r.resolveTypeExpr(stmt.Value)
		r.defineType(stmt.Name.Lexeme)
	default:
		panic("Not implemented statement kind to resolve.")
	}

	return
}

func (r *resolver) resolveAll(stmts []syntax.Stmt) (err *resolvingError) {
	for _, stmt := range stmts {
		e := r.resolveStmt(stmt)

		if e != nil {
			err = e
		}
	}

	return
}

func Resolve(source string, stmts []syntax.Stmt) (
	vars map[any]int,
	types map[syntax.TypeExpr]int,
	hadError bool,
	ds []diagnostics.Diagnostic,
) {
	global := newScope()
	r := resolver{
		Source: source,
		Scopes: stack.New(global),
		Vars:   make(map[any]int),
		Types:  make(map[syntax.TypeExpr]int),
	}

	r.defineType("number")
	r.defineType("string")
	r.defineType("bool")
	r.defineType("nil")

	_ = r.resolveAll(stmts)
	return r.Vars, r.Types, r.HadError, r.Diagnostics
}
