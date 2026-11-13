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

/* *** Grammar notation ********************************************************

The grammar uses a variation of EBNF notation.

- `::=` defines a production.
- `|` separates alternatives.
- `[ ... ]` denotes an optional expression.
- `{ ... }` denotes zero or more repetitions.
- `( ... )` groups expressions.
- `"..."` denotes literal tokens.
- `.` terminates a production.
*******************************************************************************/

// varDecl ::= "let" ID ":" typeExpr [ ":=" expr ] .
func (p *parser) varDeclaration() (stmt syntax.Stmt, err *parsingError) {
	loc := p.prev().Location()
	varName := p.peek()

	if !p.expectWithMsg(syntax.TokenID, "Expected variable name.") {
		err = &parsingError{}
		return
	}

	if p.expect(syntax.TokenColon) == nil {
		err = &parsingError{}
		return
	}

	var typeName syntax.Expr
	typeName, err = p.typeExpr()
	if err != nil {
		return
	}

	var initializer syntax.Expr = nil
	if p.match(syntax.TokenColonEqual) {
		initializer, err = p.expr()
		if err != nil {
			return
		}
	}

	stmt = syntax.NewVarStmt(varName, typeName, initializer, loc)
	return
}

// functionDecl ::= "function" ID "(" paramsList ")" [ "->" typeExpr ] blockStmt .
func (p *parser) functionDeclaration() (stmt syntax.Stmt, err *parsingError) {
	loc := p.prev().Location()

	if !p.expectWithMsg(syntax.TokenID, "Expected function name.") {
		err = &parsingError{}
		return
	}

	name := p.prev()

	if p.expect(syntax.TokenParenL) == nil {
		err = &parsingError{}
		return
	}

	var parameters []struct {
		Name syntax.Token
		Type syntax.Expr
	}
	parameters, err = p.paramsList()
	if err != nil {
		return
	}

	if p.expect(syntax.TokenParenR) == nil {
		err = &parsingError{}
		return
	}

	var returnType syntax.Expr = nil

	if p.match(syntax.TokenArrowR) {
		returnType, err = p.typeExpr()
		if err != nil {
			return
		}
	}

	if !p.expectWithMsg(syntax.TokenBraceL, "Expected '{' before function body.") {
		err = &parsingError{}
		return
	}

	var body []syntax.Stmt
	body, err = p.blockStmt()
	if err != nil {
		return
	}

	parametersType := []syntax.Expr{}
	for _, param := range parameters {
		parametersType = append(parametersType, param.Type)
	}

	stmt = syntax.NewVarStmt(
		name,
		syntax.NewFunctionTypeExpr(parametersType, returnType, loc),
		syntax.NewFunctionExpr(&name, parameters, returnType, body, loc),
		loc,
	)
	return
}

// typeDecl ::= "type" ID ":=" typeExpr .
func (p *parser) typeDeclaration() (stmt syntax.Stmt, err *parsingError) {
	loc := p.prev().Location()

	if !p.expectWithMsg(syntax.TokenID, "Expected the type's name.") {
		err = &parsingError{}
		return
	}

	name := p.prev()

	if p.expect(syntax.TokenColonEqual) == nil {
		err = &parsingError{}
		return
	}

	var typeExpr syntax.Expr
	typeExpr, err = p.typeExpr()
	if err != nil {
		return
	}

	stmt = syntax.NewTypeStmt(name, typeExpr, loc)
	return
}

// declStmt ::= varDecl | functionDecl | stmt .
//
// A declaration statement should also have a terminator right after, but this
// check is done by the function calling 'declaration()'.
func (p *parser) declaration() (stmt syntax.Stmt, err *parsingError) {
	switch {
	case p.match(syntax.TokenLet):
		stmt, err = p.varDeclaration()
		if err != nil {
			p.synchronize()
			return
		}

		return
	case p.match(syntax.TokenFunction):
		stmt, err = p.functionDeclaration()
		if err != nil {
			p.synchronize()
			return
		}

		return
	case p.match(syntax.TokenType):
		stmt, err = p.typeDeclaration()
		if err != nil {
			p.synchronize()
			return
		}

		return
	default:
		stmt, err = p.stmt()
		if err != nil {
			p.synchronize()
			return
		}

		return
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
