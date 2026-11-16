package syntax

type TypeExpr interface {
	typeExpr()

	Locatable
}

type VariantTypeExpr struct {
	Located

	Types []TypeExpr
}

type NamedTypeExpr struct {
	Located

	Name Token
}

type FunctionTypeExpr struct {
	Located

	ParamsType []TypeExpr
	ReturnType TypeExpr
}

func (e VariantTypeExpr) typeExpr()  {}
func (e NamedTypeExpr) typeExpr()    {}
func (e FunctionTypeExpr) typeExpr() {}

////////////////////////////////////////////////////////////////////////////////

func NewVariantTypeExpr(types []TypeExpr) VariantTypeExpr {
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
	paramsType []TypeExpr,
	returnType TypeExpr,
	loc Location,
) FunctionTypeExpr {
	return FunctionTypeExpr{
		Located: NewLocated(loc),

		ParamsType: paramsType,
		ReturnType: returnType,
	}
}
