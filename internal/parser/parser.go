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

func (p *parser) expr() (syntax.Expr, bool, diagnostics.Diagnostic) {
	return nil, false, diagnostics.Diagnostic{}
}

func Parse(tokens []syntax.Token) (syntax.Expr, bool, []diagnostics.Diagnostic) {
	p := parser{
		Tokens: tokens,
		Pos:    0,

		Exprs: []syntax.Expr{},
	}

	return p.expr()
}
