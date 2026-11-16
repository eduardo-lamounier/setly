package syntax

import "fmt"

type Expr interface {
	expr()

	Locatable
}

type UnaryExpr struct {
	Located

	Operator Token
	Operand  Expr
}

type BinaryExpr struct {
	Located

	Operator Token
	Left     Expr
	Right    Expr
}

type GroupingExpr struct {
	Located

	Inner Expr
}

type LiteralExpr struct {
	Located

	Literal Token
}

type VarExpr struct {
	Located

	Name Token
}

type FuncExpr struct {
	Located

	Name       *Token
	Parameters []Parameter
	ReturnType TypeExpr
	Body       []Stmt
}

type CallExpr struct {
	Located

	Callee    Expr
	Arguments []Expr
}

func (e UnaryExpr) expr()    {}
func (e BinaryExpr) expr()   {}
func (e GroupingExpr) expr() {}
func (e LiteralExpr) expr()  {}
func (e VarExpr) expr()      {}
func (e FuncExpr) expr()     {}
func (e CallExpr) expr()     {}

////////////////////////////////////////////////////////////////////////////////

func (e LiteralExpr) Value() any {
	return e.Literal.LiteralValue
}

func NewUnaryExpr(operator Token, operand Expr, loc Location) UnaryExpr {
	return UnaryExpr{
		Located: NewLocated(loc),

		Operator: operator,
		Operand:  operand,
	}
}

func NewBinaryExpr(
	operator Token,
	left Expr,
	right Expr,
	loc Location,
) BinaryExpr {
	return BinaryExpr{
		Located: NewLocated(loc),

		Operator: operator,
		Left:     left,
		Right:    right,
	}
}

func NewGroupingExpr(inner Expr, loc Location) GroupingExpr {
	return GroupingExpr{
		Located: NewLocated(loc),

		Inner: inner,
	}
}

func NewLiteralExpr(literal Token) (expr LiteralExpr) {
	expr.loc = LocationFromToken(literal)
	newLiteral := literal

	switch literal.Kind {
	case TokenString, TokenNumber:
	case TokenNil:
		newLiteral.LiteralValue = nil
	case TokenFalse:
		newLiteral.LiteralValue = false
	case TokenTrue:
		newLiteral.LiteralValue = true
	default:
		panic(fmt.Sprint("Non-literal token ", literal, " passed to the constructor of LiteralExpr."))
	}

	expr.Literal = newLiteral
	return
}

func NewVarExpr(name Token) VarExpr {
	return VarExpr{
		Located: NewLocated(LocationFromToken(name)),

		Name: name,
	}
}

func NewFuncExpr(
	name *Token,
	parameters []Parameter,
	returnType TypeExpr,
	body []Stmt,
	loc Location,
) FuncExpr {
	return FuncExpr{
		Located: NewLocated(loc),

		Name:       name,
		Parameters: parameters,
		ReturnType: returnType,
		Body:       body,
	}
}

func NewCallExpr(callee Expr, arguments []Expr) CallExpr {
	return CallExpr{
		Located: NewLocated(callee.Location()),

		Callee:    callee,
		Arguments: arguments,
	}
}
