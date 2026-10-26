// Package runtime Defines runtime error, value types and the runtime's
// environment structure
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
	return fmt.Sprintf("setly: %s [line %d]", e.Message, e.Line)
}

func (e Error) Report() {
	fmt.Fprintln(os.Stderr, e)
}

func FromToken(message string, token syntax.Token) Error {
	return Error{
		Message: message,
		Line:    token.Line,
	}
}
