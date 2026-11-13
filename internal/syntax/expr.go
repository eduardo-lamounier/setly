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

type AssignExpr struct {
	Located

	Name  Token
	Value Expr
}

type FunctionExpr struct {
	Located

	Name       *Token
	Parameters []struct {
		Name Token
		Type Expr
	}
	ReturnType Expr
	Body       []Stmt
}

type CallExpr struct {
	Located

	Callee    Expr
	Arguments []Expr
}

type VariantTypeExpr struct {
	Located

	Types []Expr
}

type NamedTypeExpr struct {
	Located

	Name Token
}

type FunctionTypeExpr struct {
	Located

	teste func() int

	ParamsType []Expr
	ReturnType Expr
}

func (e UnaryExpr) expr()        {}
func (e BinaryExpr) expr()       {}
func (e GroupingExpr) expr()     {}
func (e LiteralExpr) expr()      {}
func (e VarExpr) expr()          {}
func (e AssignExpr) expr()       {}
func (e FunctionExpr) expr()     {}
func (e CallExpr) expr()         {}
func (e VariantTypeExpr) expr()  {}
func (e NamedTypeExpr) expr()    {}
func (e FunctionTypeExpr) expr() {}

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

func NewAssignExpr(name Token, value Expr) AssignExpr {
	return AssignExpr{
		Located: NewLocated(LocationFromToken(name)),

		Name:  name,
		Value: value,
	}
}

func NewFunctionExpr(
	name *Token,
	parameters []struct {
		Name Token
		Type Expr
	},
	returnType Expr,
	body []Stmt,
	loc Location,
) FunctionExpr {
	return FunctionExpr{
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

func NewVariantTypeExpr(types []Expr) VariantTypeExpr {
	return VariantTypeExpr{
		Located: NewLocated(types[0].Location()),

		Types: types,
	}
}

func NewNamedTypeExpr(name Token) NamedTypeExpr {
	return NamedTypeExpr{
		Located: NewLocated(LocationFromToken(name)),

		Name: name,
	}
}

func NewFunctionTypeExpr(
	paramsType []Expr,
	returnType Expr,
	loc Location,
) FunctionTypeExpr {
	return FunctionTypeExpr{
		Located: NewLocated(loc),

		ParamsType: paramsType,
		ReturnType: returnType,
	}
}
