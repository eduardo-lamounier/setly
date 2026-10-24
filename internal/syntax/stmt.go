package syntax

type Stmt interface{}

type ExprStmt struct {
	Expr Expr
}

type PrintStmt struct {
	Newline bool
	Expr    Expr
}

type VarStmt struct {
	Name        Token
	Type        Token
	Initializer Expr
}

func NewExprStmt(expr Expr) ExprStmt {
	return ExprStmt{
		Expr: expr,
	}
}

func NewPrintStmt(expr Expr, newline bool) PrintStmt {
	return PrintStmt{
		Newline: newline,
		Expr:    expr,
	}
}

func NewVarStmt(name Token, valueType Token, initializer Expr) VarStmt {
	return VarStmt{
		Name:        name,
		Type:        valueType,
		Initializer: initializer,
	}
}
