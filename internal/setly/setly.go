// Package setly Is responsible for handling source processing and execution
package setly

import (
	"fmt"

	"github.com/eduardo-lamounier/setly/internal/diagnostics"
	"github.com/eduardo-lamounier/setly/internal/lexer"
)

func Run(source string) error {
	tokens, hadError, ds := lexer.Scan(source)

	fmt.Println("Tokens:")
	for _, token := range tokens {
		fmt.Println(token)
	}

	if hadError {
		reportAll(ds)
		return fmt.Errorf("errors were found in the source code")
	}

	return fmt.Errorf("running not yet implemented")
}

func reportAll(ds []diagnostics.Diagnostic) {
	for _, d := range ds {
		d.Report()
	}
}
