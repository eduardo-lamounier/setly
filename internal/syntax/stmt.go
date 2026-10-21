package syntax

type Stmt interface{}

type ExprStmt struct {
	Expr Expr
}

type PrintStmt struct {
	Newline bool
	Expr    Expr
}
