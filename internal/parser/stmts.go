package parser

import "github.com/eduardo-lamounier/setly/internal/syntax"

// NOTE: ***********************************************************************
// The notation used to define the grammar rules is documented in
// 'parser.go' in this package.
// *****************************************************************************

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
