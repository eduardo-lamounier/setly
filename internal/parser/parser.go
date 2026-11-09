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

// functionTypeExpr ::= "function" "(" paramsTypeList ")" [ ":" typeExpr ] .
//
// paramsTypeList ::= [ typeExpr { "," typeExpr } ] .
func (p *parser) functionTypeExpr() (expr syntax.Expr, err *parsingError) {
	line := p.peek().Line()

	if p.expect(syntax.TokenParenL) == nil {
		err = &parsingError{}
		return
	}

	paramsType := []syntax.Expr{}
	if !p.check(syntax.TokenParenR) {
		for {
			var paramType syntax.Expr
			paramType, err = p.typeExpr()
			if err != nil {
				return
			}

			paramsType = append(paramsType, paramType)

			if !p.match(syntax.TokenComma) {
				break
			}
		}
	}

	if p.expect(syntax.TokenParenR) == nil {
		err = &parsingError{}
		return
	}

	var returnType syntax.Expr = nil
	if p.match(syntax.TokenColon) {
		returnType, err = p.typeExpr()
		if err != nil {
			return
		}
	}

	expr = syntax.NewFunctionTypeExpr(paramsType, returnType, line)
	return
}

// literalTypeExpr ::= functionTypeExpr
func (p *parser) literalTypeExpr() (expr syntax.Expr, err *parsingError) {
	switch {
	case p.match(syntax.TokenFunction):
		return p.functionTypeExpr()
	default:
		p.report("Expected a type.")
		err = &parsingError{}
		return
	}
}

// primaryTypeExpr ::= literalTypeExpr | namedTypeExpr .
//
// namedTypeExpr ::= ID | NIL .
func (p *parser) primaryTypeExpr() (expr syntax.Expr, err *parsingError) {
	switch {
	case p.match(syntax.TokenID, syntax.TokenNil):
		expr = syntax.NewNamedTypeExpr(p.prev())
	default:
		return p.literalTypeExpr()
	}

	return
}

// unionTypeExpr ::= primaryTypeExpr { "|" primaryTypeExpr } .
func (p *parser) unionTypeExpr() (expr syntax.Expr, err *parsingError) {
	expr, err = p.primaryTypeExpr()
	if err != nil {
		return
	}

	if p.check(syntax.TokenPipe) {
		types := []syntax.Expr{expr}

		for p.match(syntax.TokenPipe) {
			var type_ syntax.Expr
			type_, err = p.primaryTypeExpr()
			if err != nil {
				return
			}

			types = append(types, type_)
		}

		expr = syntax.NewVariantTypeExpr(types)
	}

	return
}

// typeExpr ::= unionTypeExpr .
func (p *parser) typeExpr() (syntax.Expr, *parsingError) {
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
	err *parsingError,
) {
	if !p.check(syntax.TokenParenR) {
		for {
			var parameter struct {
				Name syntax.Token
				Type syntax.Expr
			}

			if !p.expectWithMsg(syntax.TokenID, "Expected parameter name.") {
				err = &parsingError{}
				return
			}

			parameter.Name = p.prev()

			if p.expect(syntax.TokenColon) == nil {
				err = &parsingError{}
				return
			}

			parameter.Type, err = p.typeExpr()
			if err != nil {
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
func (p *parser) primaryExpr() (expr syntax.Expr, err *parsingError) {
	if p.match(syntax.TokenFunction) {
		line := p.prev().Line()

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
		}

		var returnType syntax.Expr = nil

		if p.match(syntax.TokenColon) {
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

		expr = syntax.NewFunctionExpr(nil, parameters, returnType, body, line)
		return
	}

	if p.match(syntax.TokenNil, syntax.TokenFalse, syntax.TokenTrue,
		syntax.TokenNumber, syntax.TokenString) {

		expr = syntax.NewLiteralExprFromToken(p.prev())
		return
	}

	if p.match(syntax.TokenParenL) {
		expr, err = p.expr()
		if err != nil {
			return
		}

		if !p.match(syntax.TokenParenR) {
			p.report("Expected ')', got '%s'.")
			err = &parsingError{}
			return
		}

		expr = syntax.NewGroupingExpr(expr)
		return
	}

	if p.match(syntax.TokenID) {
		expr = syntax.NewVarExpr(p.prev())
		return
	}

	p.report("Expected expression.")
	err = &parsingError{}
	return
}

// callExpr ::= primary { "(" argumentsList ")" } .
//
// argumentsList ::= [ expr { "," expr } ] .
func (p *parser) callExpr() (expr syntax.Expr, err *parsingError) {
	expr, err = p.primaryExpr()
	if err != nil {
		return
	}

	for p.match(syntax.TokenParenL) {
		arguments := []syntax.Expr{}
		if !p.check(syntax.TokenParenR) {
			for {
				var e syntax.Expr
				e, err = p.expr()
				if err != nil {
					return
				}

				arguments = append(arguments, e)

				if !p.match(syntax.TokenComma) {
					break
				}
			}
		}

		if p.expect(syntax.TokenParenR) == nil {
			err = &parsingError{}
			return
		}

		expr = syntax.NewCallExpr(expr, arguments)
	}

	return
}

// unaryExpr ::= callExpr | ( ( "-" | "+" ) unaryExpr ) .
func (p *parser) unaryExpr() (expr syntax.Expr, err *parsingError) {
	if p.match(syntax.TokenMinus, syntax.TokenPlus) {
		operator := p.prev()

		var right syntax.Expr
		right, err = p.unaryExpr()
		if err != nil {
			return
		}

		expr = syntax.NewUnaryExpr(operator, right)
		return
	}

	return p.callExpr()
}

// factorExpr ::= unaryExpr { ( "*" | "/" ) unaryExpr } .
func (p *parser) factorExpr() (expr syntax.Expr, err *parsingError) {
	expr, err = p.unaryExpr()
	if err != nil {
		return
	}

	for p.match(syntax.TokenStar, syntax.TokenSlash) {
		operator := p.prev()

		var right syntax.Expr
		right, err = p.unaryExpr()
		if err != nil {
			return
		}

		expr = syntax.NewBinaryExpr(operator, expr, right)
	}

	return
}

// termExpr ::= factorExpr { ( "+" | "-" ) factorExpr } .
func (p *parser) termExpr() (expr syntax.Expr, err *parsingError) {
	expr, err = p.factorExpr()
	if err != nil {
		return
	}

	for p.match(syntax.TokenPlus, syntax.TokenMinus) {
		operator := p.prev()

		var right syntax.Expr
		right, err = p.factorExpr()
		if err != nil {
			return
		}

		expr = syntax.NewBinaryExpr(operator, expr, right)
	}

	return
}

// comparisonExpr ::= termExpr { ( ">" | "<" | ">=" | "<=" ) termExpr } .
func (p *parser) comparisonExpr() (expr syntax.Expr, err *parsingError) {
	expr, err = p.termExpr()
	if err != nil {
		return
	}

	for p.match(syntax.TokenGreater, syntax.TokenLess,
		syntax.TokenGreaterEqual, syntax.TokenLessEqual) {

		operator := p.prev()

		var right syntax.Expr
		right, err = p.termExpr()
		if err != nil {
			return
		}

		expr = syntax.NewBinaryExpr(operator, expr, right)
	}

	return
}

// equalityExpr ::= comparisonExpr { ( "~=" | "=" ) comparisonExpr } .
func (p *parser) equalityExpr() (expr syntax.Expr, err *parsingError) {
	expr, err = p.comparisonExpr()
	if err != nil {
		return
	}

	for p.match(syntax.TokenTildeEqual, syntax.TokenEqual) {
		operator := p.prev()

		var right syntax.Expr
		right, err = p.comparisonExpr()
		if err != nil {
			return
		}

		expr = syntax.NewBinaryExpr(operator, expr, right)
	}

	return
}

// logicalNotExpr ::= equalityExpr | ( "not" logicalNotExpr ) .
func (p *parser) logicalNotExpr() (expr syntax.Expr, err *parsingError) {
	if p.match(syntax.TokenNot) {
		operator := p.prev()
		expr, err = p.logicalNotExpr()
		if err != nil {
			return
		}

		expr = syntax.NewUnaryExpr(operator, expr)
		return
	}

	return p.equalityExpr()
}

// logicalAndExpr ::= logicalNotExpr { "and" logicalNotExpr } .
func (p *parser) logicalAndExpr() (expr syntax.Expr, err *parsingError) {
	expr, err = p.logicalNotExpr()

	if expr != nil {
		return
	}

	for p.match(syntax.TokenAnd) {
		operator := p.prev()

		var right syntax.Expr
		right, err = p.logicalNotExpr()
		if err != nil {
			return
		}

		expr = syntax.NewBinaryExpr(operator, expr, right)
	}

	return
}

// logicalOrExpr ::= logicalAndExpr { "or" logicalAndExpr } .
func (p *parser) logicalOrExpr() (expr syntax.Expr, err *parsingError) {
	expr, err = p.logicalAndExpr()
	if err != nil {
		return
	}

	for p.match(syntax.TokenOr) {
		operator := p.prev()

		var right syntax.Expr
		right, err = p.logicalAndExpr()
		if err != nil {
			return
		}

		expr = syntax.NewBinaryExpr(operator, expr, right)
	}

	return
}

// assignmentExpr ::= logicalOrExpr [ ":=" assignmentExpr ] .
func (p *parser) assignmentExpr() (expr syntax.Expr, err *parsingError) {
	expr, err = p.logicalOrExpr()
	if err != nil {
		return
	}

	if p.match(syntax.TokenColonEqual) {
		op := p.prev()

		var value syntax.Expr
		value, err = p.assignmentExpr()
		if err != nil {
			return
		}

		switch e := expr.(type) {
		case syntax.VarExpr:
			expr = syntax.NewAssignExpr(e.Name, value)
		default:
			p.reportAt(op, "Invalid target to assign a value to.")
			err = &parsingError{}
			return
		}
	}

	return
}

// expr ::= assignmentExpr .
func (p *parser) expr() (expr syntax.Expr, err *parsingError) {
	return p.assignmentExpr()
}

// printStmt ::= "print" "(" [ expr ] ")" .
func (p *parser) printStmt() (stmt syntax.Stmt, err *parsingError) {
	if p.expect(syntax.TokenParenL) == nil {
		err = &parsingError{}
		return
	}

	if p.match(syntax.TokenParenR) {
		stmt = syntax.NewPrintStmt(nil)
		return
	}

	var expr syntax.Expr
	expr, err = p.expr()
	if err != nil {
		return
	}

	if p.expect(syntax.TokenParenR) == nil {
		err = &parsingError{}
		return
	}

	stmt = syntax.NewPrintStmt(expr)
	return
}

// exprStmt ::= expr .
func (p *parser) exprStmt() (stmt syntax.Stmt, err *parsingError) {
	var expr syntax.Expr
	expr, err = p.expr()
	if err != nil {
		return
	}

	stmt = syntax.NewExprStmt(expr)
	return
}

// blockStmt ::= "{" { declStmt ";" } "}" .
//
// Do note the "{" should be consumed by the calling function.
func (p *parser) blockStmt() (stmts []syntax.Stmt, err *parsingError) {
	for !p.check(syntax.TokenBraceR) && !p.atEnd() {
		pos := p.Pos
		var errInDecl *parsingError
		decl, errInDecl := p.declaration()

		if errInDecl != nil {
			err = &parsingError{}

			// Ensures the parser moved forward
			if pos == p.Pos {
				p.advance()
			}
			continue
		}

		stmts = append(stmts, decl)

		if p.expect(syntax.TokenSemicolon) == nil {
			err = &parsingError{}
			continue
		}
	}

	if p.expect(syntax.TokenBraceR) == nil {
		err = &parsingError{}
		return
	}

	return
}

// ifStmt ::= "if" expr blockStmt { "elif" expr blockStmt } [ "else" blockStmt ] .
func (p *parser) ifStmt() (stmt syntax.Stmt, err *parsingError) {
	var cond syntax.Expr
	cond, err = p.expr()
	if err != nil {
		return
	}

	if p.expect(syntax.TokenBraceL) == nil {
		err = &parsingError{}
		return
	}

	var thenBranch []syntax.Stmt
	thenBranch, err = p.blockStmt()
	if err != nil {
		return
	}

	var elseBranch []syntax.Stmt = nil

	if p.match(syntax.TokenElif) {
		var elif syntax.Stmt
		elif, err = p.ifStmt()
		if err != nil {
			return
		}

		stmt = syntax.NewIfStmt(cond, thenBranch, []syntax.Stmt{elif})
		return
	}

	if p.match(syntax.TokenElse) {
		if p.expect(syntax.TokenBraceL) == nil {
			err = &parsingError{}
			return
		}

		elseBranch, err = p.blockStmt()
		if err != nil {
			return
		}
	}

	stmt = syntax.NewIfStmt(cond, thenBranch, elseBranch)
	return
}

// whileStmt ::= "while" expr blockStmt .
func (p *parser) whileStmt() (stmt syntax.Stmt, err *parsingError) {
	var cond syntax.Expr
	cond, err = p.expr()
	if err != nil {
		return
	}

	if p.expect(syntax.TokenBraceL) == nil {
		err = &parsingError{}
		return
	}

	var block []syntax.Stmt
	block, err = p.blockStmt()
	if err != nil {
		return
	}

	stmt = syntax.NewWhileStmt(cond, block)
	return
}

// forNumStmt ::= "for" ID ":=" expr "," expr [ "," expr ] blockStmt .
func (p *parser) forStmt() (stmt syntax.Stmt, err *parsingError) {
	line := p.peek().Line()

	if !p.expectWithMsg(syntax.TokenID, "Expected variable name.") {
		err = &parsingError{}
		return
	}

	varName := p.prev()

	if p.expect(syntax.TokenColonEqual) == nil {
		err = &parsingError{}
		return
	}

	var start syntax.Expr
	start, err = p.expr()
	if err != nil {
		return
	}

	if !p.expectWithMsg(syntax.TokenComma, "Expected limit value.") {
		err = &parsingError{}
		return
	}

	var limit syntax.Expr
	limit, err = p.expr()
	if err != nil {
		return
	}

	var step syntax.Expr = syntax.NewLiteralExprAt(float64(1), line)

	if p.match(syntax.TokenComma) {
		step, err = p.expr()
		if err != nil {
			return
		}
	}

	if p.expect(syntax.TokenBraceL) == nil {
		err = &parsingError{}
		return
	}

	var block []syntax.Stmt
	block, err = p.blockStmt()
	if err != nil {
		return
	}

	stmt = syntax.NewForNumStmt(varName, start, limit, step, block, line)
	return
}

// repeatStmt ::= "repeat" blockStmt [ "until" expr ] .
func (p *parser) repeatStmt() (stmt syntax.Stmt, err *parsingError) {
	if p.expect(syntax.TokenBraceL) == nil {
		err = &parsingError{}
		return
	}

	var block []syntax.Stmt
	block, err = p.blockStmt()
	if err != nil {
		return
	}

	line := p.peek().Line()

	var cond syntax.Expr = syntax.NewLiteralExprAt(false, line)

	if p.match(syntax.TokenUntil) {
		cond, err = p.expr()
		if err != nil {
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
func (p *parser) returnStmt() (stmt syntax.Stmt, err *parsingError) {
	line := p.prev().Line()

	if p.check(syntax.TokenSemicolon) {
		stmt = syntax.NewReturnStmtAt(nil, line)
		return
	}

	var value syntax.Expr
	value, err = p.expr()
	if err != nil {
		return
	}

	stmt = syntax.NewReturnStmtAt(value, line)
	return
}

// breakStmt ::= "break" .
func (p *parser) breakStmt() (stmt syntax.Stmt, err *parsingError) {
	stmt = syntax.NewBreakStmt(p.prev().Line())
	return
}

// continueStmt ::= "continue" .
func (p *parser) continueStmt() (stmt syntax.Stmt, err *parsingError) {
	stmt = syntax.NewContinueStmt(p.prev().Line())
	return
}

// stmt ::= printStmt | blockStmt | ifStmt | whileStmt | forNumStmt | repeatStmt
// | returnStmt | breakStmt | continueStmt | exprStmt .
func (p *parser) stmt() (stmt syntax.Stmt, err *parsingError) {
	switch {
	case p.match(syntax.TokenPrint):
		return p.printStmt()
	case p.match(syntax.TokenBraceL):
		var stmts []syntax.Stmt
		stmts, err = p.blockStmt()
		if err != nil {
			return
		}

		stmt = syntax.NewBlockStmt(stmts)
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

	return
}

// varDecl ::= "let" ID ":" typeExpr [ ":=" expr ] .
func (p *parser) varDeclaration() (stmt syntax.Stmt, err *parsingError) {
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

	stmt = syntax.NewVarStmt(varName, typeName, initializer)
	return
}

// functionDecl ::= "function" ID "(" paramsList ")" [ ":" typeExpr ] blockStmt .
func (p *parser) functionDeclaration() (stmt syntax.Stmt, err *parsingError) {
	line := p.peek().Line()

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

	if p.match(syntax.TokenColon) {
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
		syntax.NewFunctionTypeExpr(parametersType, returnType, line),
		syntax.NewFunctionExpr(&name, parameters, returnType, body, line),
	)
	return
}

// typeDecl ::= "type" ID ":=" typeExpr .
func (p *parser) typeDeclaration() (stmt syntax.Stmt, err *parsingError) {
	line := p.peek().Line()

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

	stmt = syntax.NewTypeStmt(name, typeExpr, line)
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
