// Package runtime TODO: Add package description
package runtime

import (
	"fmt"
	"os"

	"github.com/eduardo-lamounier/setly/internal/syntax"
)

type Error struct {
	Token   syntax.Token
	Message string
}

func (e Error) Error() string {
	return fmt.Sprintf("%s [line %d]", e.Message, e.Token.Line)
}

func (e Error) Report() {
	fmt.Fprint(os.Stderr, e)
}
