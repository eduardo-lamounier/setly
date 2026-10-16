// Package parser Defines and implements functions for converting a sequence
// of tokens to a simple single-expression
package parser

import (
	"github.com/eduardo-lamounier/setly/internal/diagnostics"
	"github.com/eduardo-lamounier/setly/internal/syntax"
)

type parser struct {
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

func (p *parser) expr() (syntax.Expr, bool, *diagnostics.Diagnostic) {
	return nil, false, nil
}

func Parse(tokens []syntax.Token) (syntax.Expr, bool, *diagnostics.Diagnostic) {
	p := parser{
		Tokens: tokens,
		Pos:    0,

		Exprs: []syntax.Expr{},
	}

	return p.expr()
}
