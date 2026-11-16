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

type FuncTypeExpr struct {
	Located

	ParamsType []TypeExpr
	ReturnType TypeExpr
}

func (e VariantTypeExpr) typeExpr() {}
func (e NamedTypeExpr) typeExpr()   {}
func (e FuncTypeExpr) typeExpr()    {}

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

func NewFuncTypeExpr(
	paramsType []TypeExpr,
	returnType TypeExpr,
	loc Location,
) FuncTypeExpr {
	return FuncTypeExpr{
		Located: NewLocated(loc),

		ParamsType: paramsType,
		ReturnType: returnType,
	}
}
