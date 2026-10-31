package syntax

type Stmt interface {
	stmt()
}

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

type BlockStmt struct {
	Stmts []Stmt
}

type IfStmt struct {
	Condition  Expr
	ThenBranch Stmt
	ElseBranch Stmt
}

type LoopStmt struct {
	Initialization Stmt
	Condition      Expr
	Update         Stmt
	Block          Stmt
}

func (s ExprStmt) stmt()  {}
func (s PrintStmt) stmt() {}
func (s VarStmt) stmt()   {}
func (s BlockStmt) stmt() {}
func (s IfStmt) stmt()    {}
func (s LoopStmt) stmt()  {}

////////////////////////////////////////////////////////////////////////////////

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

func NewBlockStmt(stmts []Stmt) BlockStmt {
	return BlockStmt{
		Stmts: stmts,
	}
}

func NewIfStmt(condition Expr, thenBranch Stmt, elseBranch Stmt) IfStmt {
	return IfStmt{
		Condition:  condition,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch,
	}
}

func NewLoopStmt(
	initialization Stmt, condition Expr, update Stmt, block Stmt,
) LoopStmt {
	return LoopStmt{
		Initialization: initialization,
		Condition:      condition,
		Update:         update,
		Block:          block,
	}
}
