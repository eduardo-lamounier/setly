// Package setly Is responsible for handling source processing and execution
package setly

import (
	"fmt"

	"github.com/eduardo-lamounier/setly/internal/diagnostics"
	"github.com/eduardo-lamounier/setly/internal/interpreter"
	"github.com/eduardo-lamounier/setly/internal/lexer"
	"github.com/eduardo-lamounier/setly/internal/parser"
	"github.com/eduardo-lamounier/setly/internal/resolver"
	"github.com/eduardo-lamounier/setly/internal/typechecker"
)

const errorMessage = "errors were found in the source code"

func Run(source string, opts interpreter.Options) error {
	tokens, hadError, ds := lexer.Scan(source)

	if hadError {
		reportAll(ds)
		return fmt.Errorf(errorMessage)
	}

	stmts, hadError, ds := parser.Parse(source, tokens)

	if hadError {
		reportAll(ds)
		return fmt.Errorf(errorMessage)
	}

	vars, types, hadError, ds := resolver.Resolve(source, stmts)

	if hadError {
		reportAll(ds)
		return fmt.Errorf("errors were found in the source code")
	}

	hadError, ds = typechecker.Check(source, stmts, vars, types)

	if hadError {
		reportAll(ds)
		return fmt.Errorf(errorMessage)
	}

	if err := interpreter.Interpret(stmts, vars, opts); err != nil {
		return err
	}

	return nil
}

func reportAll(ds []diagnostics.Diagnostic) {
	for _, d := range ds {
		d.Report()
	}
}
