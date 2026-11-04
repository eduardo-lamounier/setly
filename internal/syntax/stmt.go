package syntax

type Stmt interface {
	stmt()

	Line() int
}

type ExprStmt struct {
	Expr Expr
}

type PrintStmt struct {
	Expr Expr
}

type VarStmt struct {
	Name        Token
	Type        Token
	Initializer Expr
}

type BlockStmt struct {
	line int

	Stmts []Stmt
}

type IfStmt struct {
	Condition  Expr
	ThenBranch []Stmt
	ElseBranch []Stmt
}

type WhileStmt struct {
	Condition Expr
	Block     []Stmt
}

type ForNumStmt struct {
	line int

	Var   Token
	Start Expr
	Limit Expr
	Step  Expr
	Block []Stmt
}

type ReturnStmt struct {
	line int

	Value Expr
}

type BreakStmt struct {
	line int
}

type ContinueStmt struct {
	line int
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

func (s ExprStmt) Line() int {
	return s.Expr.Line()
}

func (s PrintStmt) Line() int {
	return s.Expr.Line()
}

func (s VarStmt) Line() int {
	return s.Name.Line()
}

func (s BlockStmt) Line() int {
	return s.line
}

func (s IfStmt) Line() int {
	return s.Condition.Line()
}

func (s WhileStmt) Line() int {
	return s.Condition.Line()
}

func (s ReturnStmt) Line() int {
	return s.line
}

func (s BreakStmt) Line() int {
	return s.line
}

func (s ContinueStmt) Line() int {
	return s.line
}

func (s ForNumStmt) Line() int {
	return s.line
}

////////////////////////////////////////////////////////////////////////////////

func NewExprStmt(expr Expr) ExprStmt {
	return ExprStmt{
		Expr: expr,
	}
}

func NewPrintStmt(expr Expr) PrintStmt {
	return PrintStmt{
		Expr: expr,
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

func NewIfStmt(condition Expr, thenBranch []Stmt, elseBranch []Stmt) IfStmt {
	return IfStmt{
		Condition:  condition,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch,
	}
}

func NewWhileStmt(condition Expr, block []Stmt) WhileStmt {
	return WhileStmt{
		Condition: condition,
		Block:     block,
	}
}

func NewReturnStmtAt(value Expr, line int) ReturnStmt {
	return ReturnStmt{
		line:  line,
		Value: value,
	}
}

func NewBreakStmt(line int) BreakStmt {
	return BreakStmt{
		line: line,
	}
}

func NewContinueStmt(line int) ContinueStmt {
	return ContinueStmt{
		line: line,
	}
}

func NewForNumStmt(
	varName Token, start Expr, limit Expr, step Expr, block []Stmt, line int,
) ForNumStmt {
	return ForNumStmt{
		line: line,

		Var:   varName,
		Start: start,
		Limit: limit,
		Step:  step,
		Block: block,
	}
}
