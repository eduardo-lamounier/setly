package parser

import "github.com/eduardo-lamounier/setly/internal/syntax"

// NOTE: ***********************************************************************
// The notation used to define the grammar rules is documented in
// 'parser.go' in this package.
// *****************************************************************************

// varDecl ::= "let" ID ":" typeExpr [ "=" expr ] .
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

	var typeName syntax.TypeExpr
	typeName, err = p.typeExpr()
	if err != nil {
		return
	}

	var initializer syntax.Expr = nil
	if p.match(syntax.TokenEqual) {
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

	var parameters []syntax.Parameter
	parameters, err = p.paramsList()
	if err != nil {
		return
	}

	if p.expect(syntax.TokenParenR) == nil {
		err = &parsingError{}
		return
	}

	var returnType syntax.TypeExpr = nil

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

	parametersType := []syntax.TypeExpr{}
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

// typeDecl ::= "type" ID "=" typeExpr .
func (p *parser) typeDeclaration() (stmt syntax.Stmt, err *parsingError) {
	loc := p.prev().Location()

	if !p.expectWithMsg(syntax.TokenID, "Expected the type's name.") {
		err = &parsingError{}
		return
	}

	name := p.prev()

	if p.expect(syntax.TokenEqual) == nil {
		err = &parsingError{}
		return
	}

	var typeExpr syntax.TypeExpr
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
