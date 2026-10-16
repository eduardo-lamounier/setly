package syntax

import "fmt"

type Expr interface {
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
	LiteralToken Token
}

////////////////////////////////////////////////////////////////////////////////

func (e LiteralExpr) Literal() any {
	return e.LiteralToken.Literal
}
