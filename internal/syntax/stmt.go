package syntax

type Stmt interface {
	stmt()

	Locatable
}

type ExprStmt struct {
	Located

	Expr Expr
}

type PrintStmt struct {
	Located

	Expr Expr
}

type VarStmt struct {
	Located

	Name        Token
	Type        Expr
	Initializer Expr
}

type BlockStmt struct {
	Located

	Stmts []Stmt
}

type IfStmt struct {
	Located

	Condition  Expr
	ThenBranch []Stmt
	ElseBranch []Stmt
}

type WhileStmt struct {
	Located

	Condition Expr
	Block     []Stmt
}

type ForNumStmt struct {
	Located

	Var   Token
	Start Expr
	Limit Expr
	Step  Expr
	Block []Stmt
}

type ReturnStmt struct {
	Located

	Value Expr
}

type BreakStmt struct {
	Located
}

type ContinueStmt struct {
	Located
}

type TypeStmt struct {
	Located

	Name  Token
	Value Expr
}

func (s ExprStmt) stmt()     {}
func (s PrintStmt) stmt()    {}
func (s VarStmt) stmt()      {}
func (s BlockStmt) stmt()    {}
func (s IfStmt) stmt()       {}
func (s WhileStmt) stmt()    {}
func (s ReturnStmt) stmt()   {}
func (s BreakStmt) stmt()    {}
func (s ContinueStmt) stmt() {}
func (s ForNumStmt) stmt()   {}
func (s TypeStmt) stmt()     {}

////////////////////////////////////////////////////////////////////////////////

func NewExprStmt(expr Expr) ExprStmt {
	return ExprStmt{
		Located: NewLocated(expr.Location()),
		Expr:    expr,
	}
}

func NewPrintStmt(expr Expr, loc Location) PrintStmt {
	return PrintStmt{
		Located: NewLocated(loc),
		Expr:    expr,
	}
}

func NewVarStmt(
	name Token,
	valueType Expr,
	initializer Expr,
	loc Location,
) VarStmt {
	return VarStmt{
		Located: NewLocated(loc),

		Name:        name,
		Type:        valueType,
		Initializer: initializer,
	}
}

func NewBlockStmt(stmts []Stmt, loc Location) BlockStmt {
	return BlockStmt{
		Located: NewLocated(loc),

		Stmts: stmts,
	}
}

func NewIfStmt(
	condition Expr,
	thenBranch []Stmt,
	elseBranch []Stmt,
	loc Location,
) IfStmt {
	return IfStmt{
		Located: NewLocated(loc),

		Condition:  condition,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch,
	}
}

func NewWhileStmt(condition Expr, block []Stmt, loc Location) WhileStmt {
	return WhileStmt{
		Located: NewLocated(loc),

		Condition: condition,
		Block:     block,
	}
}

func NewReturnStmtAt(value Expr, loc Location) ReturnStmt {
	return ReturnStmt{
		Located: NewLocated(loc),

		Value: value,
	}
}

func NewBreakStmt(loc Location) BreakStmt {
	return BreakStmt{
		Located: NewLocated(loc),
	}
}

func NewContinueStmt(loc Location) ContinueStmt {
	return ContinueStmt{
		Located: NewLocated(loc),
	}
}

func NewForNumStmt(
	varName Token,
	start Expr,
	limit Expr,
	step Expr,
	block []Stmt,
	loc Location,
) ForNumStmt {
	return ForNumStmt{
		Located: NewLocated(loc),

		Var:   varName,
		Start: start,
		Limit: limit,
		Step:  step,
		Block: block,
	}
}

func NewTypeStmt(name Token, value Expr, loc Location) TypeStmt {
	return TypeStmt{
		Located: NewLocated(loc),

		Name:  name,
		Value: value,
	}
}
