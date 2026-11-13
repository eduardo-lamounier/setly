package syntax

import "fmt"

func MakeNumberLiteral(v float64, loc Location) LiteralExpr {
	return LiteralExpr{
		Literal: NewLiteralToken(TokenNumber, fmt.Sprint(v), v, loc),
	}
}

func MakeStringLiteral(v string, loc Location) LiteralExpr {
	return LiteralExpr{
		Literal: NewLiteralToken(TokenString, `"`+v+`"`, v, loc),
	}
}

func MakeBooleanLiteral(v bool, loc Location) LiteralExpr {
	tk := TokenFalse
	if v {
		tk = TokenTrue
	}

	return LiteralExpr{
		Literal: NewLiteralToken(tk, fmt.Sprint(v), v, loc),
	}
}

func MakeNilLiteral(loc Location) LiteralExpr {
	return LiteralExpr{
		Literal: NewLiteralToken(TokenNil, "nil", nil, loc),
	}
}
