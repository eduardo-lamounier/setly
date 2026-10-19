package syntax

import "fmt"

type Expr interface {
	fmt.Stringer
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

func (e UnaryExpr) String() string {
	return fmt.Sprintf("%s[ %s ]", e.Operator.Lexeme, e.Operand)
}

func (e BinaryExpr) String() string {
	return fmt.Sprintf("%s[ %s, %s ]", e.Operator.Lexeme,
		e.Operands[0], e.Operands[1])
}

func (e GroupingExpr) String() string {
	return fmt.Sprint("( ", e.Inner, " )")
}

func (e LiteralExpr) String() string {
	return fmt.Sprint(e.LiteralToken.Literal)
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
	case TokenString, TokenNumber:
		return LiteralExpr{
			LiteralToken: literal,
		}
	case TokenNil:
		literal.Literal = nil
		return LiteralExpr{
			LiteralToken: literal,
		}
	case TokenFalse:
		literal.Literal = false
		return LiteralExpr{
			LiteralToken: literal,
		}
	case TokenTrue:
		literal.Literal = true
		return LiteralExpr{
			LiteralToken: literal,
		}
	default:
		panic(fmt.Sprint("Non-literal token ", literal, " passed to the constructor of LiteralExpr."))
	}
}
