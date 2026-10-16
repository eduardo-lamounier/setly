// Package setly Is responsible for handling source processing and execution
package setly

import (
	"fmt"

	"github.com/eduardo-lamounier/setly/internal/diagnostics"
	"github.com/eduardo-lamounier/setly/internal/lexer"
	"github.com/eduardo-lamounier/setly/internal/parser"
	"github.com/eduardo-lamounier/setly/internal/syntax"
)

func Run(source string) error {
	tokens, hadError, ds := lexer.Scan(source)

	if hadError {
		reportAll(ds)
		return fmt.Errorf("errors were found in the source code")
	}

	expr, hadError, d := parser.Parse(source, tokens)

	fmt.Println("Expression:")
	if expr != nil {
		syntax.ShowExpr(expr)
	}

	if hadError {
		d.Report()
	}

	return fmt.Errorf("running not yet implemented")
}

func reportAll(ds []diagnostics.Diagnostic) {
	for _, d := range ds {
		d.Report()
	}
}
