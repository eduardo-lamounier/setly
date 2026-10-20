// Package runtime TODO: Add package description
package runtime

import (
	"fmt"
	"os"

	"github.com/eduardo-lamounier/setly/internal/syntax"
)

type Error struct {
	Line    int
	Message string
}

func (e Error) Error() string {
	return fmt.Sprintf("%s [line %d]", e.Message, e.Line)
}

func (e Error) Report() {
	fmt.Fprint(os.Stderr, e)
}

func FromToken(message string, token syntax.Token) Error {
	return Error{
		Message: message,
		Line:    token.Line,
	}
}
