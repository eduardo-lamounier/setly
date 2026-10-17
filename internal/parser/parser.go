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

	HadError bool
	Exprs    []syntax.Expr
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

func (p *parser) error(
	token syntax.Token, message string,
) *diagnostics.Diagnostic {
	return &diagnostics.Diagnostic{
		Line:    token.Line,
		Col:     token.Col,
		Loc:     token.Loc,
		Message: message,
		Source:  p.Source,
	}
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

func (p *parser) primary() (syntax.Expr, bool, *diagnostics.Diagnostic) {
	if p.match(syntax.TokenNil, syntax.TokenFalse, syntax.TokenTrue,
		syntax.TokenNumber, syntax.TokenString) {

		return syntax.NewLiteralExpr(p.prev()), false, nil
	}

	if p.match(syntax.TokenParenL) {
		expr, hadError, d := p.expr()

		if hadError {
			return nil, true, d
		}

		if !p.match(syntax.TokenParenR) {
			return nil, true, p.error(p.prev(),
				fmt.Sprintf("Expected ')', got '%s'.", p.peek().Lexeme))
		}

		return syntax.NewGroupingExpr(expr), false, nil
	}

	return nil, true, p.error(p.peek(), "Expected expression.")
}

func (p *parser) unary() (syntax.Expr, bool, *diagnostics.Diagnostic) {
	if p.match(syntax.TokenMinus, syntax.TokenPlus) {
		operator := p.prev()
		right, hadError, d := p.unary()

		if hadError {
			return nil, true, d
		}

		return syntax.NewUnaryExpr(operator, right), false, nil
	}

	return p.primary()
}

func (p *parser) factor() (syntax.Expr, bool, *diagnostics.Diagnostic) {
	expr, hadError, d := p.unary()

	if hadError {
		return nil, true, d
	}

	for p.match(syntax.TokenStar, syntax.TokenSlash) {
		operator := p.prev()
		right, hadError, d := p.unary()

		if hadError {
			return nil, true, d
		}

		expr = syntax.NewBinaryExpr(operator, expr, right)
	}

	return expr, false, nil
}

func (p *parser) term() (syntax.Expr, bool, *diagnostics.Diagnostic) {
	expr, hadError, d := p.factor()

	if hadError {
		return nil, true, d
	}

	for p.match(syntax.TokenPlus, syntax.TokenMinus) {
		operator := p.prev()
		right, hadError, d := p.factor()

		if hadError {
			return nil, true, d
		}

		expr = syntax.NewBinaryExpr(operator, expr, right)
	}

	return expr, false, nil
}

func (p *parser) comparison() (syntax.Expr, bool, *diagnostics.Diagnostic) {
	expr, hadError, d := p.term()

	if hadError {
		return nil, true, d
	}

	for p.match(syntax.TokenGreater, syntax.TokenLess,
		syntax.TokenGreaterEqual, syntax.TokenLessEqual) {

		operator := p.prev()
		right, hadError, d := p.term()

		if hadError {
			return nil, true, d
		}

		expr = syntax.NewBinaryExpr(operator, expr, right)
	}

	return expr, false, nil
}

func (p *parser) equality() (syntax.Expr, bool, *diagnostics.Diagnostic) {
	expr, hadError, d := p.comparison()

	if hadError {
		return nil, true, d
	}

	for p.match(syntax.TokenTildeEqual, syntax.TokenEqual) {
		operator := p.prev()
		right, hadError, d := p.comparison()

		if hadError {
			return nil, true, d
		}

		expr = syntax.NewBinaryExpr(operator, expr, right)
	}

	return expr, false, nil
}

func (p *parser) expr() (syntax.Expr, bool, *diagnostics.Diagnostic) {
	return p.equality()
}

func Parse(source string, tokens []syntax.Token) (
	syntax.Expr, bool, *diagnostics.Diagnostic,
) {
	p := parser{
		Source: source,
		Tokens: tokens,
		Pos:    0,

		Exprs: []syntax.Expr{},
	}

	return p.expr()
}
