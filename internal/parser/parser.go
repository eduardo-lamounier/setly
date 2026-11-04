// Package parser Defines and implements functions for converting a sequence
// of tokens to a simple single-expression
package parser

import (
	"fmt"
	"slices"

	"github.com/eduardo-lamounier/setly/internal/diagnostics"
	"github.com/eduardo-lamounier/setly/internal/syntax"
)

type parser struct {
	Source string
	Tokens []syntax.Token
	Pos    int

	HadError    bool
	Diagnostics []diagnostics.Diagnostic
	Stmts       []syntax.Stmt
}

func (p parser) atEnd() bool {
	return p.Tokens[p.Pos].Kind == syntax.TokenEOF
}

func (p *parser) advance() {
	if !p.atEnd() {
		p.Pos++

		// Skips newlines
		for ; p.peek().Kind == syntax.TokenNewline; p.Pos++ {
		}
	}
}

func (p parser) peek() syntax.Token {
	return p.Tokens[p.Pos]
}

func (p parser) prev() syntax.Token {
	return p.Tokens[p.Pos-1]
}

func (p parser) check(expected syntax.TokenKind) bool {
	if p.atEnd() {
		return false
	}

	return p.peek().Kind == expected
}

func (p *parser) report(message string) {
	p.reportAt(p.peek(), message)
}

func (p *parser) reportAt(token syntax.Token, message string) {
	p.HadError = true
	p.Diagnostics = append(p.Diagnostics, diagnostics.Diagnostic{
		Line:    token.Line(),
		Col:     token.Col(),
		Pos:     token.Pos(),
		Message: message,
		Source:  p.Source,
	})
}

func (p *parser) match(expected ...syntax.TokenKind) bool {
	if p.atEnd() {
		return false
	}

	if !slices.Contains(expected, p.peek().Kind) {
		return false
	}

	p.advance()
	return true
}

func (p *parser) expect(expected syntax.TokenKind) bool {
	if expected == syntax.TokenID || expected == syntax.TokenString || expected == syntax.TokenNumber {
		panic("Could not expect a literal token.")
	}

	if p.match(expected) {
		return true
	}

	p.report(fmt.Sprintf("Expected '%s', got '%s'.", expected.GetLexeme(), p.peek().Lexeme))
	return false
}

func (p *parser) expectWithMsg(expected syntax.TokenKind, message string) bool {
	if p.match(expected) {
		return true
	}

	p.report(message)
	return false
}

func (p *parser) emitStmt(stmt syntax.Stmt) {
	p.Stmts = append(p.Stmts, stmt)
}

func (p *parser) paramsList() (
	parameters []struct {
		Name syntax.Token
		Type syntax.Token
	},
	hadError bool,
) {
	if !p.check(syntax.TokenParenR) {
		for {
			var parameter struct {
				Name syntax.Token
				Type syntax.Token
			}

			if !p.expectWithMsg(syntax.TokenID, "Expected parameter name.") {
				hadError = true
				return
			}

			parameter.Name = p.prev()

			if !p.expectWithMsg(syntax.TokenID, "Expected parameter type.") {
				hadError = true
				return
			}

			parameter.Type = p.prev()

			parameters = append(parameters, parameter)

			if !p.match(syntax.TokenComma) {
				break
			}
		}
	}

	return
}

func (p *parser) primary() (syntax.Expr, bool) {
	if p.match(syntax.TokenFunction) {
		line := p.prev().Line()

		if !p.expect(syntax.TokenParenL) {
			return nil, true
		}

		parameters, hadError := p.paramsList()

		if hadError {
			return nil, true
		}

		if !p.expect(syntax.TokenParenR) {
			hadError = true
		}

		if !p.expectWithMsg(syntax.TokenBraceL, "Expected '{' before function body.") {
			return nil, true
		}

		body, hadError := p.blockStmt()

		if hadError {
			return nil, true
		}

		return syntax.NewFunctionExpr(parameters, body, line), false
	}

	if p.match(syntax.TokenNil, syntax.TokenFalse, syntax.TokenTrue,
		syntax.TokenNumber, syntax.TokenString) {

		return syntax.NewLiteralExprFromToken(p.prev()), false
	}

	if p.match(syntax.TokenParenL) {
		expr, hadError := p.expr()

		if hadError {
			return nil, true
		}

		if !p.match(syntax.TokenParenR) {
			p.report("Expected ')', got '%s'.")
			return nil, true
		}

		return syntax.NewGroupingExpr(expr), false
	}

	if p.match(syntax.TokenID) {
		return syntax.NewVarExpr(p.prev()), false
	}

	p.report("Expected expression.")
	return nil, true
}

func (p *parser) call() (expr syntax.Expr, hadError bool) {
	expr, hadError = p.primary()

	if hadError {
		return
	}

	for p.match(syntax.TokenParenL) {
		arguments := []syntax.Expr{}
		if !p.check(syntax.TokenParenR) {
			for {
				var e syntax.Expr
				e, hadError = p.expr()

				if hadError {
					return
				}

				arguments = append(arguments, e)

				if !p.match(syntax.TokenComma) {
					break
				}
			}
		}

		if !p.expect(syntax.TokenParenR) {
			hadError = true
			return
		}

		expr = syntax.NewCallExpr(expr, arguments)
	}

	return
}

func (p *parser) unary() (syntax.Expr, bool) {
	if p.match(syntax.TokenMinus, syntax.TokenPlus) {
		operator := p.prev()
		right, hadError := p.unary()

		if hadError {
			return nil, true
		}

		return syntax.NewUnaryExpr(operator, right), false
	}

	return p.call()
}

func (p *parser) factor() (syntax.Expr, bool) {
	expr, hadError := p.unary()

	if hadError {
		return nil, true
	}

	for p.match(syntax.TokenStar, syntax.TokenSlash) {
		operator := p.prev()
		right, hadError := p.unary()

		if hadError {
			return nil, true
		}

		expr = syntax.NewBinaryExpr(operator, expr, right)
	}

	return expr, false
}

func (p *parser) term() (syntax.Expr, bool) {
	expr, hadError := p.factor()

	if hadError {
		return nil, true
	}

	for p.match(syntax.TokenPlus, syntax.TokenMinus) {
		operator := p.prev()
		right, hadError := p.factor()

		if hadError {
			return nil, true
		}

		expr = syntax.NewBinaryExpr(operator, expr, right)
	}

	return expr, false
}

func (p *parser) comparison() (syntax.Expr, bool) {
	expr, hadError := p.term()

	if hadError {
		return nil, true
	}

	for p.match(syntax.TokenGreater, syntax.TokenLess,
		syntax.TokenGreaterEqual, syntax.TokenLessEqual) {

		operator := p.prev()
		right, hadError := p.term()

		if hadError {
			return nil, true
		}

		expr = syntax.NewBinaryExpr(operator, expr, right)
	}

	return expr, false
}

func (p *parser) equality() (syntax.Expr, bool) {
	expr, hadError := p.comparison()

	if hadError {
		return nil, true
	}

	for p.match(syntax.TokenTildeEqual, syntax.TokenEqual) {
		operator := p.prev()
		right, hadError := p.comparison()

		if hadError {
			return nil, true
		}

		expr = syntax.NewBinaryExpr(operator, expr, right)
	}

	return expr, false
}

func (p *parser) logicalNot() (syntax.Expr, bool) {
	if p.match(syntax.TokenNot) {
		operator := p.prev()
		expr, hadError := p.logicalNot()

		if hadError {
			return nil, true
		}

		return syntax.NewUnaryExpr(operator, expr), false
	}

	return p.equality()
}

func (p *parser) logicalAnd() (syntax.Expr, bool) {
	expr, hadError := p.logicalNot()

	if hadError {
		return nil, true
	}

	for p.match(syntax.TokenAnd) {
		operator := p.prev()
		right, hadError := p.logicalNot()

		if hadError {
			return nil, true
		}

		expr = syntax.NewBinaryExpr(operator, expr, right)
	}

	return expr, false
}

func (p *parser) logicalOr() (syntax.Expr, bool) {
	expr, hadError := p.logicalAnd()

	if hadError {
		return nil, true
	}

	for p.match(syntax.TokenOr) {
		operator := p.prev()
		right, hadError := p.logicalAnd()

		if hadError {
			return nil, true
		}

		expr = syntax.NewBinaryExpr(operator, expr, right)
	}

	return expr, false
}

func (p *parser) assignment() (syntax.Expr, bool) {
	expr, hadError := p.logicalOr()

	if hadError {
		return nil, true
	}

	if p.match(syntax.TokenColonEqual) {
		op := p.prev()
		value, hadError := p.assignment()

		if hadError {
			return nil, true
		}

		switch expr := expr.(type) {
		case syntax.VarExpr:
			return syntax.NewAssignExpr(expr.Name, value), false
		default:
			p.reportAt(op, "Invalid target to assign a value to.")
			return nil, true
		}
	}

	return expr, false
}

func (p *parser) expr() (syntax.Expr, bool) {
	return p.assignment()
}

func (p *parser) printStmt() (syntax.Stmt, bool) {
	if !p.expect(syntax.TokenParenL) {
		return nil, true
	}

	if p.match(syntax.TokenParenR) {
		return syntax.NewPrintStmt(nil), false
	}

	expr, hadError := p.expr()

	if hadError {
		return nil, true
	}

	if !p.expect(syntax.TokenParenR) {
		return nil, true
	}

	return syntax.NewPrintStmt(expr), false
}

func (p *parser) exprStmt() (syntax.Stmt, bool) {
	expr, hadError := p.expr()

	if hadError {
		return nil, true
	}

	return syntax.NewExprStmt(expr), false
}

func (p *parser) blockStmt() (stmts []syntax.Stmt, hadError bool) {
	for !p.check(syntax.TokenBraceR) && !p.atEnd() {
		decl, hadErrorInDecl := p.declaration()

		if hadErrorInDecl {
			hadError = true
			continue
		}

		stmts = append(stmts, decl)

		if !p.expect(syntax.TokenSemicolon) {
			hadError = true
			continue
		}
	}

	if !p.expect(syntax.TokenBraceR) {
		hadError = true
	}

	return
}

func (p *parser) ifStmt() (stmt syntax.Stmt, hadError bool) {
	cond, hadError := p.expr()

	if hadError {
		return
	}

	if !p.expect(syntax.TokenBraceL) {
		hadError = true
		return
	}

	thenBranch, hadError := p.blockStmt()

	if hadError {
		return
	}

	var elseBranch []syntax.Stmt = nil

	if p.match(syntax.TokenElif) {
		var elif syntax.Stmt
		elif, hadError = p.ifStmt()

		if hadError {
			return
		}

		stmt = syntax.NewIfStmt(cond, thenBranch, []syntax.Stmt{elif})
		return
	}

	if p.match(syntax.TokenElse) {
		if !p.expect(syntax.TokenBraceL) {
			hadError = true
			return
		}

		elseBranch, hadError = p.blockStmt()

		if hadError {
			return
		}
	}

	stmt = syntax.NewIfStmt(cond, thenBranch, elseBranch)
	return
}

func (p *parser) whileStmt() (stmt syntax.Stmt, hadError bool) {
	cond, hadError := p.expr()

	if hadError {
		return
	}

	if !p.expect(syntax.TokenBraceL) {
		hadError = true
		return
	}

	block, hadError := p.blockStmt()

	if hadError {
		return
	}

	stmt = syntax.NewWhileStmt(cond, block)
	return
}

func (p *parser) forStmt() (stmt syntax.Stmt, hadError bool) {
	line := p.peek().Line()

	if !p.expectWithMsg(syntax.TokenID, "Expected variable name.") {
		hadError = true
		return
	}

	varName := p.prev()

	if !p.expect(syntax.TokenColonEqual) {
		hadError = true
		return
	}

	start, hadError := p.expr()

	if hadError {
		return
	}

	if !p.expectWithMsg(syntax.TokenComma, "Expected limit value.") {
		hadError = true
		return
	}

	limit, hadError := p.expr()

	if hadError {
		return
	}

	var step syntax.Expr = syntax.NewLiteralExprAt(float64(1), line)

	if p.match(syntax.TokenComma) {
		step, hadError = p.expr()

		if hadError {
			return
		}
	}

	if !p.expect(syntax.TokenBraceL) {
		hadError = true
		return
	}

	block, hadError := p.blockStmt()

	if hadError {
		return
	}

	return syntax.NewForNumStmt(varName, start, limit, step, block, line), false
}

func (p *parser) repeatStmt() (stmt syntax.Stmt, hadError bool) {
	if !p.expect(syntax.TokenBraceL) {
		hadError = true
		return
	}

	block, hadError := p.blockStmt()

	if hadError {
		return
	}

	line := p.peek().Line()

	var cond syntax.Expr = syntax.NewLiteralExprAt(false, line)

	if p.match(syntax.TokenUntil) {
		cond, hadError = p.expr()

		if hadError {
			return
		}
	}

	block = append(
		block, syntax.NewIfStmt(
			cond,
			[]syntax.Stmt{syntax.NewBreakStmt(line)},
			nil,
		),
	)

	stmt = syntax.NewWhileStmt(
		syntax.NewLiteralExprAt(true, line),
		block,
	)
	return
}

func (p *parser) returnStmt() (stmt syntax.Stmt, hadError bool) {
	line := p.prev().Line()

	if p.check(syntax.TokenSemicolon) {
		stmt = syntax.NewReturnStmtAt(nil, line)
		return
	}

	var value syntax.Expr
	value, hadError = p.expr()

	if hadError {
		return
	}

	stmt = syntax.NewReturnStmtAt(value, line)
	return
}

func (p *parser) breakStmt() (stmt syntax.Stmt, hadError bool) {
	return syntax.NewBreakStmt(p.prev().Line()), false
}

func (p *parser) continueStmt() (stmt syntax.Stmt, hadError bool) {
	return syntax.NewContinueStmt(p.prev().Line()), false
}

func (p *parser) stmt() (syntax.Stmt, bool) {
	switch {
	case p.match(syntax.TokenPrint):
		return p.printStmt()
	case p.match(syntax.TokenBraceL):
		stmts, hadError := p.blockStmt()

		if hadError {
			return nil, true
		}

		return syntax.NewBlockStmt(stmts), false
	case p.match(syntax.TokenIf):
		return p.ifStmt()
	case p.match(syntax.TokenWhile):
		return p.whileStmt()
	case p.match(syntax.TokenFor):
		return p.forStmt()
	case p.match(syntax.TokenRepeat):
		return p.repeatStmt()
	case p.match(syntax.TokenReturn):
		return p.returnStmt()
	case p.match(syntax.TokenBreak):
		return p.breakStmt()
	case p.match(syntax.TokenContinue):
		return p.continueStmt()
	default:
		return p.exprStmt()
	}
}

func (p *parser) varDeclaration() (syntax.Stmt, bool) {
	varName := p.peek()
	if !p.expectWithMsg(syntax.TokenID, "Expected variable name.") {
		return nil, true
	}

	typeName := p.peek()
	if !p.expectWithMsg(syntax.TokenID, "Expected the variable's type.") {
		return nil, true
	}

	var initializer syntax.Expr = nil
	if p.match(syntax.TokenColonEqual) {
		var hadError bool
		initializer, hadError = p.expr()

		if hadError {
			return nil, true
		}
	}

	return syntax.NewVarStmt(varName, typeName, initializer), false
}

func (p *parser) functionDeclaration() (stmt syntax.Stmt, hadError bool) {
	if !p.expectWithMsg(syntax.TokenID, "Expected function name.") {
		hadError = true
		return
	}

	name := p.prev()

	if !p.expect(syntax.TokenParenL) {
		hadError = true
		return
	}

	parameters, hadError := p.paramsList()

	if hadError {
		return
	}

	if !p.expect(syntax.TokenParenR) {
		hadError = true
	}

	if !p.expectWithMsg(syntax.TokenBraceL, "Expected '{' before function body.") {
		hadError = true
		return
	}

	body, hadError := p.blockStmt()

	if hadError {
		return
	}

	stmt = syntax.NewVarStmt(
		name,
		syntax.Token{Lexeme: "number"},
		syntax.NewFunctionExpr(parameters, body, name.Line()),
	)
	return
}

func (p *parser) declaration() (syntax.Stmt, bool) {
	switch {
	case p.match(syntax.TokenLet):
		stmt, hadError := p.varDeclaration()

		if hadError {
			p.synchronize()
			return nil, true
		}

		return stmt, false
	case p.match(syntax.TokenFunction):
		stmt, hadError := p.functionDeclaration()

		if hadError {
			p.synchronize()
			return nil, true
		}

		return stmt, false
	default:
		stmt, hadError := p.stmt()

		if hadError {
			p.synchronize()
			return nil, true
		}

		return stmt, false
	}
}

func (p *parser) synchronize() {
	for ; !p.atEnd(); p.advance() {
		if p.Pos > 0 && p.prev().Kind == syntax.TokenSemicolon {
			return
		}

		switch p.peek().Kind {
		case syntax.TokenFunction,
			syntax.TokenLet,
			syntax.TokenFor,
			syntax.TokenWhile,
			syntax.TokenIf,
			syntax.TokenPrint,
			syntax.TokenReturn:
			return
		}
	}
}

func Parse(source string, tokens []syntax.Token) (
	[]syntax.Stmt, bool, []diagnostics.Diagnostic,
) {
	p := parser{
		Source: source,
		Tokens: tokens,
		Pos:    0,

		Diagnostics: []diagnostics.Diagnostic{},
		Stmts:       []syntax.Stmt{},
	}

	// Skips first newline tokens:
	for p.match(syntax.TokenNewline) {
	}

	// The rest of the newline tokens are skipped by 'p.advance()'

	for !p.atEnd() {
		stmt, hadError := p.declaration()

		if hadError {
			continue
		}

		if stmt != nil {
			p.emitStmt(stmt)
		}

		if !p.expect(syntax.TokenSemicolon) {
			p.synchronize()
		}
	}

	return p.Stmts, p.HadError, p.Diagnostics
}
