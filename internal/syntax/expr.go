package syntax

import "fmt"

type Expr interface {
	expr()

	fmt.Stringer
	locatable
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
	Location

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
	return fmt.Sprint(e.Literal)
}

func (e VarExpr) String() string {
	return fmt.Sprint(e.Name.Lexeme)
}

func (e AssignExpr) String() string {
	return fmt.Sprintf(":=[ %s, %s ]", e.Name, e.Value)
}

func (e UnaryExpr) Pos() int {
	switch e.Operator.Kind {
	case TokenNot, TokenMinus, TokenPlus:
		return e.Operator.Pos()
	default:
		panic("Some unary operator not handled when calculating expression location.")
	}
}

func (e UnaryExpr) Line() int {
	switch e.Operator.Kind {
	case TokenNot, TokenMinus, TokenPlus:
		return e.Operator.Line()
	default:
		panic("Some unary operator not handled when calculating expression line.")
	}
}

func (e UnaryExpr) Col() int {
	switch e.Operator.Kind {
	case TokenNot, TokenMinus, TokenPlus:
		return e.Operator.Line()
	default:
		panic("Some unary operator not handled when calculating expression column.")
	}
}

func (e BinaryExpr) Pos() int {
	return e.Operands[0].Pos()
}

func (e BinaryExpr) Line() int {
	return e.Operands[0].Line()
}

func (e BinaryExpr) Col() int {
	return e.Operands[0].Col()
}

func (e GroupingExpr) Pos() int {
	return e.Inner.Pos()
}

func (e GroupingExpr) Line() int {
	return e.Inner.Line()
}

func (e GroupingExpr) Col() int {
	return e.Inner.Col()
}

func (e LiteralExpr) Pos() int {
	return e.pos
}

func (e LiteralExpr) Line() int {
	return e.line
}

func (e LiteralExpr) Col() int {
	return e.col
}

func (e VarExpr) Pos() int {
	return e.Name.Pos()
}

func (e VarExpr) Line() int {
	return e.Name.Line()
}

func (e VarExpr) Col() int {
	return e.Name.Col()
}

func (e AssignExpr) Pos() int {
	return e.Name.Pos()
}

func (e AssignExpr) Line() int {
	return e.Name.Line()
}

func (e AssignExpr) Col() int {
	return e.Name.Col()
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
	expr.Location = literal.Location

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

func NewLiteralExprAt(literal any, loc Location) (expr LiteralExpr) {
	return LiteralExpr{
		Literal:  literal,
		Location: loc,
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
