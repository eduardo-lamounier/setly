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

// functionTypeExpr ::= "function" "(" paramsTypeList ")" [ ":" typeExpr ] .
//
// paramsTypeList ::= [ typeExpr { "," typeExpr } ] .
func (p *parser) functionTypeExpr() (expr syntax.Expr, hadError bool) {
	line := p.peek().Line()

	if !p.expect(syntax.TokenParenL) {
		hadError = true
		return
	}

	paramsType := []syntax.Expr{}
	if !p.check(syntax.TokenParenR) {
		for {
			var paramType syntax.Expr
			paramType, hadError = p.typeExpr()

			if hadError {
				return
			}

			paramsType = append(paramsType, paramType)

			if !p.match(syntax.TokenComma) {
				break
			}
		}
	}

	if !p.expect(syntax.TokenParenR) {
		hadError = true
		return
	}

	var returnType syntax.Expr = nil
	if p.match(syntax.TokenColon) {
		returnType, hadError = p.typeExpr()

		if hadError {
			return
		}
	}

	expr = syntax.NewFunctionTypeExpr(paramsType, returnType, line)
	return
}

// literalTypeExpr ::= functionTypeExpr
func (p *parser) literalTypeExpr() (expr syntax.Expr, hadError bool) {
	switch {
	case p.match(syntax.TokenFunction):
		return p.functionTypeExpr()
	default:
		p.report("Expected a type.")
		hadError = true
		return
	}
}

// primaryTypeExpr ::= literalTypeExpr | namedTypeExpr .
//
// namedTypeExpr ::= ID | NIL .
func (p *parser) primaryTypeExpr() (expr syntax.Expr, hadError bool) {
	switch {
	case p.match(syntax.TokenID, syntax.TokenNil):
		expr = syntax.NewNamedTypeExpr(p.prev())
	default:
		return p.literalTypeExpr()
	}

	return
}

// unionTypeExpr ::= primaryTypeExpr { "|" primaryTypeExpr } .
func (p *parser) unionTypeExpr() (expr syntax.Expr, hadError bool) {
	expr, hadError = p.primaryTypeExpr()

	if hadError {
		return
	}

	if p.check(syntax.TokenPipe) {
		types := []syntax.Expr{expr}

		for p.match(syntax.TokenPipe) {
			var type_ syntax.Expr
			type_, hadError = p.primaryTypeExpr()

			if hadError {
				return
			}

			types = append(types, type_)
		}

		expr = syntax.NewVariantTypeExpr(types)
	}

	return
}

// typeExpr ::= unionTypeExpr .
func (p *parser) typeExpr() (syntax.Expr, bool) {
	return p.unionTypeExpr()
}

// paramsList ::= [ parameter { "," parameter } ] .
//
// parameter ::= ID ":" typeExpr .
func (p *parser) paramsList() (
	parameters []struct {
		Name syntax.Token
		Type syntax.Expr
	},
	hadError bool,
) {
	if !p.check(syntax.TokenParenR) {
		for {
			var parameter struct {
				Name syntax.Token
				Type syntax.Expr
			}

			if !p.expectWithMsg(syntax.TokenID, "Expected parameter name.") {
				hadError = true
				return
			}

			parameter.Name = p.prev()

			if !p.expect(syntax.TokenColon) {
				hadError = true
				return
			}

			parameter.Type, hadError = p.typeExpr()

			if hadError {
				return
			}

			parameters = append(parameters, parameter)

			if !p.match(syntax.TokenComma) {
				break
			}
		}
	}

	return
}

// primaryExpr ::= functionExpr | varExpr | literalExpr | groupingExpr .
//
// functionExpr ::= "function" "(" paramsList ")" [ ":" typeExpr ] blockStmt .
// varExpr      ::= ID .
// literalExpr  ::= "nil" | "false" | "true" | STRING | NUMBER .
// groupingExpr ::= "(" expr ")" .
func (p *parser) primaryExpr() (syntax.Expr, bool) {
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

		var returnType syntax.Expr = nil

		if p.match(syntax.TokenColon) {
			returnType, hadError = p.typeExpr()

			if hadError {
				return nil, true
			}
		}

		if !p.expectWithMsg(syntax.TokenBraceL, "Expected '{' before function body.") {
			return nil, true
		}

		body, hadError := p.blockStmt()

		if hadError {
			return nil, true
		}

		return syntax.NewFunctionExpr(nil, parameters, returnType, body, line), false
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

// callExpr ::= primary { "(" argumentsList ")" } .
//
// argumentsList ::= [ expr { "," expr } ] .
func (p *parser) callExpr() (expr syntax.Expr, hadError bool) {
	expr, hadError = p.primaryExpr()

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

// unaryExpr ::= callExpr | ( ( "-" | "+" ) unaryExpr ) .
func (p *parser) unaryExpr() (syntax.Expr, bool) {
	if p.match(syntax.TokenMinus, syntax.TokenPlus) {
		operator := p.prev()
		right, hadError := p.unaryExpr()

		if hadError {
			return nil, true
		}

		return syntax.NewUnaryExpr(operator, right), false
	}

	return p.callExpr()
}

// factorExpr ::= unaryExpr { ( "*" | "/" ) unaryExpr } .
func (p *parser) factorExpr() (syntax.Expr, bool) {
	expr, hadError := p.unaryExpr()

	if hadError {
		return nil, true
	}

	for p.match(syntax.TokenStar, syntax.TokenSlash) {
		operator := p.prev()
		right, hadError := p.unaryExpr()

		if hadError {
			return nil, true
		}

		expr = syntax.NewBinaryExpr(operator, expr, right)
	}

	return expr, false
}

// termExpr ::= factorExpr { ( "+" | "-" ) factorExpr } .
func (p *parser) termExpr() (syntax.Expr, bool) {
	expr, hadError := p.factorExpr()

	if hadError {
		return nil, true
	}

	for p.match(syntax.TokenPlus, syntax.TokenMinus) {
		operator := p.prev()
		right, hadError := p.factorExpr()

		if hadError {
			return nil, true
		}

		expr = syntax.NewBinaryExpr(operator, expr, right)
	}

	return expr, false
}

// comparisonExpr ::= termExpr { ( ">" | "<" | ">=" | "<=" ) termExpr } .
func (p *parser) comparisonExpr() (syntax.Expr, bool) {
	expr, hadError := p.termExpr()

	if hadError {
		return nil, true
	}

	for p.match(syntax.TokenGreater, syntax.TokenLess,
		syntax.TokenGreaterEqual, syntax.TokenLessEqual) {

		operator := p.prev()
		right, hadError := p.termExpr()

		if hadError {
			return nil, true
		}

		expr = syntax.NewBinaryExpr(operator, expr, right)
	}

	return expr, false
}

// equalityExpr ::= comparisonExpr { ( "~=" | "=" ) comparisonExpr } .
func (p *parser) equalityExpr() (syntax.Expr, bool) {
	expr, hadError := p.comparisonExpr()

	if hadError {
		return nil, true
	}

	for p.match(syntax.TokenTildeEqual, syntax.TokenEqual) {
		operator := p.prev()
		right, hadError := p.comparisonExpr()

		if hadError {
			return nil, true
		}

		expr = syntax.NewBinaryExpr(operator, expr, right)
	}

	return expr, false
}

// logicalNotExpr ::= equalityExpr | ( "not" logicalNotExpr ) .
func (p *parser) logicalNotExpr() (syntax.Expr, bool) {
	if p.match(syntax.TokenNot) {
		operator := p.prev()
		expr, hadError := p.logicalNotExpr()

		if hadError {
			return nil, true
		}

		return syntax.NewUnaryExpr(operator, expr), false
	}

	return p.equalityExpr()
}

// logicalAndExpr ::= logicalNotExpr { "and" logicalNotExpr } .
func (p *parser) logicalAndExpr() (syntax.Expr, bool) {
	expr, hadError := p.logicalNotExpr()

	if hadError {
		return nil, true
	}

	for p.match(syntax.TokenAnd) {
		operator := p.prev()
		right, hadError := p.logicalNotExpr()

		if hadError {
			return nil, true
		}

		expr = syntax.NewBinaryExpr(operator, expr, right)
	}

	return expr, false
}

// logicalOrExpr ::= logicalAndExpr { "or" logicalAndExpr } .
func (p *parser) logicalOrExpr() (syntax.Expr, bool) {
	expr, hadError := p.logicalAndExpr()

	if hadError {
		return nil, true
	}

	for p.match(syntax.TokenOr) {
		operator := p.prev()
		right, hadError := p.logicalAndExpr()

		if hadError {
			return nil, true
		}

		expr = syntax.NewBinaryExpr(operator, expr, right)
	}

	return expr, false
}

// assignmentExpr ::= logicalOrExpr [ ":=" assignmentExpr ] .
func (p *parser) assignmentExpr() (syntax.Expr, bool) {
	expr, hadError := p.logicalOrExpr()

	if hadError {
		return nil, true
	}

	if p.match(syntax.TokenColonEqual) {
		op := p.prev()
		value, hadError := p.assignmentExpr()

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

// expr ::= assignmentExpr .
func (p *parser) expr() (syntax.Expr, bool) {
	return p.assignmentExpr()
}

// printStmt ::= "print" "(" [ expr ] ")" .
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

// exprStmt ::= expr .
func (p *parser) exprStmt() (syntax.Stmt, bool) {
	expr, hadError := p.expr()

	if hadError {
		return nil, true
	}

	return syntax.NewExprStmt(expr), false
}

// blockStmt ::= "{" { declStmt ";" } "}" .
//
// Do note the "{" should be consumed by the calling function.
func (p *parser) blockStmt() (stmts []syntax.Stmt, hadError bool) {
	for !p.check(syntax.TokenBraceR) && !p.atEnd() {
		pos := p.Pos
		decl, hadErrorInDecl := p.declaration()

		if hadErrorInDecl {
			hadError = true

			// Ensures the parser moved forward
			if pos == p.Pos {
				p.advance()
			}
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

// ifStmt ::= "if" expr blockStmt { "elif" expr blockStmt } [ "else" blockStmt ] .
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

// whileStmt ::= "while" expr blockStmt .
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

// forNumStmt ::= "for" ID ":=" expr "," expr [ "," expr ] blockStmt .
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

// repeatStmt ::= "repeat" blockStmt [ "until" expr ] .
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

// returnStmt ::= "return" [ expr ] .
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

// breakStmt ::= "break" .
func (p *parser) breakStmt() (stmt syntax.Stmt, hadError bool) {
	return syntax.NewBreakStmt(p.prev().Line()), false
}

// continueStmt ::= "continue" .
func (p *parser) continueStmt() (stmt syntax.Stmt, hadError bool) {
	return syntax.NewContinueStmt(p.prev().Line()), false
}

// stmt ::= printStmt | blockStmt | ifStmt | whileStmt | forNumStmt | repeatStmt
// | returnStmt | breakStmt | continueStmt | exprStmt .
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

// varDecl ::= "let" ID ":" typeExpr [ ":=" expr ] .
func (p *parser) varDeclaration() (syntax.Stmt, bool) {
	varName := p.peek()
	if !p.expectWithMsg(syntax.TokenID, "Expected variable name.") {
		return nil, true
	}

	if !p.expect(syntax.TokenColon) {
		return nil, true
	}

	typeName, hadError := p.typeExpr()

	if hadError {
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

// functionDecl ::= "function" ID "(" paramsList ")" [ ":" typeExpr ] blockStmt .
func (p *parser) functionDeclaration() (stmt syntax.Stmt, hadError bool) {
	line := p.peek().Line()

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
		return
	}

	var returnType syntax.Expr = nil

	if p.match(syntax.TokenColon) {
		returnType, hadError = p.typeExpr()

		if hadError {
			return
		}
	}

	if !p.expectWithMsg(syntax.TokenBraceL, "Expected '{' before function body.") {
		hadError = true
		return
	}

	body, hadError := p.blockStmt()

	if hadError {
		return
	}

	parametersType := []syntax.Expr{}
	for _, param := range parameters {
		parametersType = append(parametersType, param.Type)
	}

	stmt = syntax.NewVarStmt(
		name,
		syntax.NewFunctionTypeExpr(parametersType, returnType, line),
		syntax.NewFunctionExpr(&name, parameters, returnType, body, line),
	)
	return
}

// typeDecl ::= "type" ID ":=" typeExpr .
func (p *parser) typeDeclaration() (stmt syntax.Stmt, hadError bool) {
	line := p.peek().Line()

	if !p.expectWithMsg(syntax.TokenID, "Expected the type's name.") {
		hadError = true
		return
	}

	name := p.prev()

	if !p.expect(syntax.TokenColonEqual) {
		hadError = true
		return
	}

	expr, hadError := p.typeExpr()

	if hadError {
		return
	}

	stmt = syntax.NewTypeStmt(name, expr, line)
	return
}

// declStmt ::= varDecl | functionDecl | stmt .
//
// A declaration statement should also have a terminator right after, but this
// check is done by the function calling 'declaration()'.
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
	case p.match(syntax.TokenType):
		stmt, hadError := p.typeDeclaration()

		if hadError {
			p.synchronize()
			return nil, false
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
		pos := p.Pos
		stmt, hadError := p.declaration()

		if hadError {
			// Ensures the parser moved forward
			if pos == p.Pos {
				p.advance()
			}

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
