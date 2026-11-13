package syntax

type Location struct {
	Pos  int
	Line int
	Col  int
}

type Located struct {
	loc Location
}

type Locatable interface {
	Location() Location
}

func (l Located) Location() Location {
	return l.loc
}

func (l Located) Pos() int {
	return l.loc.Pos
}

func (l Located) Line() int {
	return l.loc.Line
}

func (l Located) Col() int {
	return l.loc.Col
}

func NewLocation(pos int, line int, col int) Location {
	return Location{pos, line, col}
}

func NewLocated(loc Location) Located {
	return Located{loc}
}

func LocationFromToken(tok Token) Location {
	return Location{tok.Pos(), tok.Line(), tok.Col()}
}
