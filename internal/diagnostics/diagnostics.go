// Package diagnostics TODO: Add package description
package diagnostics

import (
	"fmt"
	"math"
	"os"
)

type Diagnostic struct {
	Line    int
	Col     int
	Loc     int
	Message string
	Source  string
}

func (d Diagnostic) Report() {
	lineBegin := 0
	for i := d.Loc; i > 0; i-- {
		if d.Source[i-1] == '\n' {
			lineBegin = i
			break
		}
	}

	lineEnding := len(d.Source)
	for i := d.Loc; i < len(d.Source)-1; i++ {
		if d.Source[i] == '\n' {
			lineEnding = i
			break
		}
	}

	line := d.Source[lineBegin:lineEnding]

	fmt.Fprintf(os.Stderr, "%d:%d: %s\n", d.Line, d.Col, d.Message)
	fmt.Fprintf(os.Stderr, "%d | %s\n", d.Line, line)

	offset := int(math.Log10(float64(d.Line))+1) + 3
	fmt.Fprintf(os.Stderr, "%*s\n", offset+d.Loc-lineBegin+len("^ here!"), "^ here!")
}
