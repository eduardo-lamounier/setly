package syntax

import "fmt"

type Expr interface {
	fmt.Stringer

	Loc() int
	Line() int
	Col() int
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

func (e UnaryExpr) Loc() int {
	switch e.Operator.Kind {
	case TokenNot, TokenMinus, TokenPlus:
		return e.Operator.Loc
	default:
		panic("Some unary operator not handled when calculating expression location.")
	}
}

func (e UnaryExpr) Line() int {
	switch e.Operator.Kind {
	case TokenNot, TokenMinus, TokenPlus:
		return e.Operator.Line
	default:
		panic("Some unary operator not handled when calculating expression line.")
	}
}

func (e UnaryExpr) Col() int {
	switch e.Operator.Kind {
	case TokenNot, TokenMinus, TokenPlus:
		return e.Operator.Line
	default:
		panic("Some unary operator not handled when calculating expression column.")
	}
}

func (e BinaryExpr) Loc() int {
	return e.Operands[0].Loc()
}

func (e BinaryExpr) Line() int {
	return e.Operands[0].Line()
}

func (e BinaryExpr) Col() int {
	return e.Operands[0].Col()
}

func (e GroupingExpr) Loc() int {
	return e.Inner.Loc()
}

func (e GroupingExpr) Line() int {
	return e.Inner.Line()
}

func (e GroupingExpr) Col() int {
	return e.Inner.Col()
}

func (e LiteralExpr) Loc() int {
	return e.LiteralToken.Loc
}

func (e LiteralExpr) Line() int {
	return e.LiteralToken.Line
}

func (e LiteralExpr) Col() int {
	return e.LiteralToken.Col
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
