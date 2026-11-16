// Package cli Defines the parsing of the CLI: parses the arguments,
// options and flags passed to the program.
package cli

import (
	"errors"
	"fmt"
	"strings"
)

const HelpMessage = `usage: setly [-h | --help] [-v | --version] [source file]

Flags:
  [-h | --help]    Shows this help message.
  [-v | --version] Shows the installed program's version.`

type Options struct {
	ShowHelp    bool
	ShowVersion bool

	SourceFilePath *string
}

func GetOptions(args []string) (opts Options, err error) {
	if len(args) == 0 {
		return // Should run interactively
	}

	errs := []error{}

	for _, arg := range args {
		switch arg {
		case "--help", "-h":
			opts.ShowHelp = true
		case "--version", "-v":
			opts.ShowVersion = true
		default:
			if len(arg) >= 2 && strings.HasPrefix(arg, "-") {
				errs = append(errs, fmt.Errorf("unknown flag/option '%s'", arg))
				continue
			}

			if opts.SourceFilePath != nil {
				errs = append(
					errs,
					fmt.Errorf("too many arguments passed to the program: expected only one main source file"),
				)
				continue
			}

			opts.SourceFilePath = &arg
		}
	}

	if opts.SourceFilePath != nil && !strings.HasSuffix(*opts.SourceFilePath, ".setly") {
		errs = append(errs, fmt.Errorf("the source file should have the \".setly\" file extension"))
	}

	if len(errs) > 0 {
		err = errors.Join(errs...)
		return
	}

	return
}
