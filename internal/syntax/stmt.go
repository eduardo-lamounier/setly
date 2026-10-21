package syntax

type Stmt interface{}

type ExprStmt struct {
	expr Expr
}

type PrintStmt struct {
	expr Expr
}
