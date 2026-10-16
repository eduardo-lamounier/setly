package syntax

import "fmt"

type Expr interface {
	print()
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

func (e UnaryExpr) print() {
	fmt.Printf("%s[ ", e.Operator.Lexeme)
	e.Operand.print()
	fmt.Print(" ]")
}

func (e BinaryExpr) print() {
	fmt.Printf("%s[ ", e.Operator.Lexeme)
	e.Operands[0].print()
	fmt.Print(", ")
	e.Operands[1].print()
	fmt.Print(" ]")
}

func (e GroupingExpr) print() {
	fmt.Print("( ")
	e.Inner.print()
	fmt.Print(" )")
}

func (e LiteralExpr) print() {
	fmt.Print(e.LiteralToken.Literal)
}

func ShowExpr(e Expr) {
	e.print()
	fmt.Println()
}

////////////////////////////////////////////////////////////////////////////////

func (e LiteralExpr) Literal() any {
	return e.LiteralToken.Literal
}

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

func NewLiteralExpr(literal Token) LiteralExpr {
	switch literal.Kind {
	case TokenNil, TokenFalse, TokenTrue, TokenString, TokenNumber:
		return LiteralExpr{
			LiteralToken: literal,
		}
	default:
		panic(fmt.Sprint("Non-literal token ", literal, " passed to the constructor of LiteralExpr."))
	}
}
