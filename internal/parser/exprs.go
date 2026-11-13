package parser

import "github.com/eduardo-lamounier/setly/internal/syntax"

// NOTE: ***********************************************************************
// The notation used to define the grammar rules is documented in
// 'parser.go' in this package.
// *****************************************************************************

// functionTypeExpr ::= "function" "(" paramsTypeList ")" [ "->" typeExpr ] .
//
// paramsTypeList ::= [ typeExpr { "," typeExpr } [ "," ] ] .
func (p *parser) functionTypeExpr() (expr syntax.Expr, err *parsingError) {
	loc := p.prev().Location()

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

			// A comma after the last parameter is optional:
			if !p.match(syntax.TokenComma) || p.check(syntax.TokenParenR) {
				break
			}
		}
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

	expr = syntax.NewFunctionTypeExpr(paramsType, returnType, loc)
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

// paramsList ::= [ parameter { "," parameter } [ "," ] ] .
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

			// A comma after the last parameter is optional:
			if !p.match(syntax.TokenComma) || p.check(syntax.TokenParenR) {
				break
			}
		}
	}

	return
}

// primaryExpr ::= functionExpr | varExpr | literalExpr | groupingExpr .
//
// functionExpr ::= "function" "(" paramsList ")" [ "->" typeExpr ] blockStmt .
// varExpr      ::= ID .
// literalExpr  ::= "nil" | "false" | "true" | STRING | NUMBER .
// groupingExpr ::= "(" expr ")" .
func (p *parser) primaryExpr() (expr syntax.Expr, err *parsingError) {
	if p.match(syntax.TokenFunction) {
		loc := p.prev().Location()

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

		expr = syntax.NewFunctionExpr(nil, parameters, returnType, body, loc)
		return
	}

	if p.match(syntax.TokenNil, syntax.TokenFalse, syntax.TokenTrue,
		syntax.TokenNumber, syntax.TokenString) {

		expr = syntax.NewLiteralExpr(p.prev())
		return
	}

	if p.match(syntax.TokenParenL) {
		loc := p.prev().Location()

		expr, err = p.expr()
		if err != nil {
			return
		}

		if !p.match(syntax.TokenParenR) {
			p.report("Expected ')', got '%s'.")
			err = &parsingError{}
			return
		}

		expr = syntax.NewGroupingExpr(expr, loc)
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
// argumentsList ::= [ expr { "," expr } [ "," ] ] .
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

				// A comma after the last parameter is optional:
				if !p.match(syntax.TokenComma) || p.check(syntax.TokenParenR) {
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

		expr = syntax.NewUnaryExpr(operator, right, operator.Location())
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

		expr = syntax.NewBinaryExpr(operator, expr, right, expr.Location())
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

		expr = syntax.NewBinaryExpr(operator, expr, right, expr.Location())
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

		expr = syntax.NewBinaryExpr(operator, expr, right, expr.Location())
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

		expr = syntax.NewBinaryExpr(operator, expr, right, expr.Location())
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

		expr = syntax.NewUnaryExpr(operator, expr, operator.Location())
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

		expr = syntax.NewBinaryExpr(operator, expr, right, expr.Location())
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

		expr = syntax.NewBinaryExpr(operator, expr, right, expr.Location())
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
