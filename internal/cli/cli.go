// Package cli Defines the parsing of the CLI: parses the arguments,
// options and flags passed to the program.
package cli

import (
	"errors"
	"fmt"
	"os"
	"slices"
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

func GetOptions() (Options, error) {
	args := os.Args[1:]

	if len(args) == 0 {
		return Options{}, nil // Should run interactively
	}

	opts := Options{
		ShowHelp:    slices.Contains(args, "-h") || slices.Contains(args, "--help"),
		ShowVersion: slices.Contains(args, "-v") || slices.Contains(args, "--version"),
	}

	if opts.ShowHelp || opts.ShowVersion {
		return opts, nil
	}

	errs := []error{}

	for _, arg := range args {
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

	if opts.SourceFilePath != nil && !strings.HasSuffix(*opts.SourceFilePath, ".setly") {
		errs = append(errs, fmt.Errorf("the source file should have the \".setly\" file extension"))
	}

	if len(errs) > 0 {
		return Options{}, errors.Join(errs...)
	}

	return opts, nil
}
