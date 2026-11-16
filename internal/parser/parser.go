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

type parsingError struct{}

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

func (p parser) next() syntax.Token {
	if p.Pos+1 >= len(p.Tokens) {
		return p.Tokens[len(p.Tokens)-1]
	}

	return p.Tokens[p.Pos+1]
}

func (p parser) check(expected syntax.TokenKind) bool {
	if p.atEnd() {
		return false
	}

	return p.peek().Kind == expected
}

func (p parser) checkNext(expected syntax.TokenKind) bool {
	return p.next().Kind == expected
}

func (p *parser) report(message string) {
	p.reportAt(p.peek(), message)
}

func (p *parser) reportAt(token syntax.Token, message string) {
	p.HadError = true
	p.Diagnostics = append(
		p.Diagnostics,
		diagnostics.New(message, token.Location(), p.Source),
	)
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

func (p *parser) expect(expected syntax.TokenKind) *syntax.Token {
	if expected == syntax.TokenID || expected == syntax.TokenString || expected == syntax.TokenNumber {
		panic("Could not expect a literal token.")
	}

	if p.match(expected) {
		tok := p.prev()
		return &tok
	}

	p.report(fmt.Sprintf("Expected '%s', got '%s'.", expected.GetLexeme(), p.peek().Lexeme))
	return nil
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

func (p *parser) synchronize() {
	for ; !p.atEnd(); p.advance() {
		if p.Pos > 0 && p.prev().Kind == syntax.TokenSemicolon {
			return
		}

		switch p.peek().Kind {
		case syntax.TokenFunc,
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

/***** Grammar notation ********************************************************

The grammar uses a variation of EBNF notation.

- `::=` defines a production.
- `|` separates alternatives.
- `[ ... ]` denotes an optional expression.
- `{ ... }` denotes zero or more repetitions.
- `( ... )` groups expressions.
- `"..."` denotes literal tokens.
- `.` terminates a production.
*******************************************************************************/

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
		pos := p.Pos
		stmt, err := p.declaration()
		if err != nil {
			// Ensures the parser moved forward
			if pos == p.Pos {
				p.advance()
			}

			continue
		}

		if stmt != nil {
			p.emitStmt(stmt)
		}

		if p.expect(syntax.TokenSemicolon) == nil {
			p.synchronize()
		}
	}

	return p.Stmts, p.HadError, p.Diagnostics
}
