package syntax

import "fmt"

type Expr interface {
	expr()

	Line() int
}

type UnaryExpr struct {
	Operator Token
	Operand  Expr
}

type BinaryExpr struct {
	Operator Token
	Operands [2]Expr
}

type GroupingExpr struct {
	Inner Expr
}

type LiteralExpr struct {
	line int

	Literal any
}

type VarExpr struct {
	Name Token
}

type AssignExpr struct {
	Name  Token
	Value Expr
}

func (e UnaryExpr) expr()    {}
func (e BinaryExpr) expr()   {}
func (e GroupingExpr) expr() {}
func (e LiteralExpr) expr()  {}
func (e VarExpr) expr()      {}
func (e AssignExpr) expr()   {}

func (e UnaryExpr) Line() int {
	return e.Operator.Line()
}

func (e BinaryExpr) Line() int {
	return e.Operands[0].Line()
}

func (e GroupingExpr) Line() int {
	return e.Inner.Line()
}

func (e LiteralExpr) Line() int {
	return e.line
}

func (e VarExpr) Line() int {
	return e.Name.Line()
}

func (e AssignExpr) Line() int {
	return e.Name.Line()
}

////////////////////////////////////////////////////////////////////////////////

func NewUnaryExpr(operator Token, operand Expr) UnaryExpr {
	return UnaryExpr{
		Operator: operator,
		Operand:  operand,
	}
}

func NewBinaryExpr(operator Token, left Expr, right Expr) BinaryExpr {
	return BinaryExpr{
		Operator: operator,
		Operands: [...]Expr{left, right},
	}
}

func NewGroupingExpr(inner Expr) GroupingExpr {
	return GroupingExpr{
		Inner: inner,
	}
}

func NewLiteralExprFromToken(literal Token) (expr LiteralExpr) {
	expr.line = literal.line

	switch literal.Kind {
	case TokenString, TokenNumber:
		expr.Literal = literal.Literal
		return
	case TokenNil:
		expr.Literal = nil
		return
	case TokenFalse:
		expr.Literal = false
		return
	case TokenTrue:
		expr.Literal = true
		return
	default:
		panic(fmt.Sprint("Non-literal token ", literal, " passed to the constructor of LiteralExpr."))
	}
}

func NewLiteralExprAt(literal any, line int) (expr LiteralExpr) {
	return LiteralExpr{
		Literal: literal,
		line:    line,
	}
}

func NewVarExpr(name Token) VarExpr {
	return VarExpr{
		Name: name,
	}
}

func NewAssignExpr(name Token, value Expr) AssignExpr {
	return AssignExpr{
		Name:  name,
		Value: value,
	}
}
