// Package parser Defines and implements functions for converting a sequence
// of tokens to a simple single-expression
package parser

import (
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
		Line:    token.Line,
		Col:     token.Col,
		Loc:     token.Loc,
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

func (p *parser) expect(expected syntax.TokenKind, message string) bool {
	if !p.match(expected) {
		p.report(message)
		p.advance()
		return false
	}

	return true
}

func (p *parser) emitStmt(stmt syntax.Stmt) {
	p.Stmts = append(p.Stmts, stmt)
}

func (p *parser) primary() (syntax.Expr, bool) {
	if p.match(syntax.TokenNil, syntax.TokenFalse, syntax.TokenTrue,
		syntax.TokenNumber, syntax.TokenString) {

		return syntax.NewLiteralExpr(p.prev()), false
	}

	if p.match(syntax.TokenParenL) {
		expr, hadError := p.expr()

		if hadError {
			return nil, true
		}

		if !p.match(syntax.TokenParenR) {
			p.reportAt(p.peek(), "Expected ')', got '%s'.")
			return nil, true
		}

		return syntax.NewGroupingExpr(expr), false
	}

	p.reportAt(p.peek(), "Expected expression.")
	return nil, true
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

	return p.primary()
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

func (p *parser) expr() (syntax.Expr, bool) {
	return p.logicalOr()
}

func (p *parser) printStmt() (syntax.Stmt, bool) {
	if p.prev().Kind != syntax.TokenPrint && p.prev().Kind != syntax.TokenPrintln {
		panic("Tried parsing print statement without a 'print' or 'println' token")
	}

	newline := p.prev().Kind == syntax.TokenPrintln

	if !p.expect(syntax.TokenParenL, "Expected opening parentheses.") {
		return nil, true
	}

	// only 'print()' is required to have received an expression:
	if p.match(syntax.TokenParenR) {
		if !newline {
			p.reportAt(p.prev(), "Expected expression to print")
			p.expect(syntax.TokenSemicolon, "Expected ';' after statement.")

			return nil, true
		}

		if !p.expect(syntax.TokenSemicolon, "Expected ';' after statement.") {
			return nil, true
		}

		return syntax.PrintStmt{
			Newline: true,
			Expr:    nil,
		}, false
	}

	expr, hadError := p.expr()

	if hadError {
		return nil, true
	}

	if !p.expect(syntax.TokenParenR, "Expected closing parentheses.") {
		return nil, true
	}

	if !p.expect(syntax.TokenSemicolon, "Expected ';' after statement.") {
		return nil, true
	}

	return syntax.PrintStmt{
		Newline: newline,
		Expr:    expr,
	}, false
}

func (p *parser) exprStmt() (syntax.Stmt, bool) {
	expr, hadError := p.expr()

	if hadError {
		return nil, true
	}

	if !p.match(syntax.TokenSemicolon) {
		p.report("Expected ';' after an expression.")
		return nil, true
	}

	return syntax.ExprStmt{
		Expr: expr,
	}, false
}

func (p *parser) stmt() (syntax.Stmt, bool) {
	if p.match(syntax.TokenPrint, syntax.TokenPrintln) {
		return p.printStmt()
	}

	return p.exprStmt()
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

	for !p.atEnd() {
		stmt, hadError := p.stmt()

		if !hadError {
			p.emitStmt(stmt)
		}
	}

	return p.Stmts, p.HadError, p.Diagnostics
}
