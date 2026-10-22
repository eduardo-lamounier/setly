// Package setly Is responsible for handling source processing and execution
package setly

import (
	"fmt"

	"github.com/eduardo-lamounier/setly/internal/diagnostics"
	"github.com/eduardo-lamounier/setly/internal/interpreter"
	"github.com/eduardo-lamounier/setly/internal/lexer"
	"github.com/eduardo-lamounier/setly/internal/parser"
)

func Run(source string) error {
	tokens, hadError, ds := lexer.Scan(source)

	if hadError {
		reportAll(ds)
		return fmt.Errorf("errors were found in the source code")
	}

	stmts, hadError, ds := parser.Parse(source, tokens)

	if hadError {
		reportAll(ds)
		return fmt.Errorf("errors were found in the source code")
	}

	return interpreter.Interpret(stmts)
}

func reportAll(ds []diagnostics.Diagnostic) {
	for _, d := range ds {
		d.Report()
	}
}
