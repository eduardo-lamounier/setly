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

const HelpMessage = `usage: setly [-h | --help] [-v | --version] [source files]

Flags:
  [-h | --help]    Shows this help message.
  [-v | --version] Shows the installed program's version.`

type Options struct {
	ShowHelp    bool
	ShowVersion bool

	SourceFiles []string
}

func GetOptions() (Options, error) {
	args := os.Args[1:]

	if len(args) == 0 {
		return Options{}, fmt.Errorf("no arguments, flags, or options were passed to the program")
	}

	opts := Options{
		ShowHelp:    slices.Contains(args, "-h") || slices.Contains(args, "--help"),
		ShowVersion: slices.Contains(args, "-v") || slices.Contains(args, "--version"),
	}

	toIgnore := map[string]struct{}{
		"-h": {}, "--help": {},
		"-v": {}, "--version": {},
	}

	errs := []error{}

	for _, arg := range args {
		if _, ignore := toIgnore[arg]; ignore {
			continue
		}

		if len(arg) >= 2 && strings.HasPrefix(arg, "-") {
			errs = append(errs, fmt.Errorf("unknown flag/option '%s'", arg))
		}
	}

	if len(errs) > 0 {
		return Options{}, errors.Join(errs...)
	}

	if !opts.ShowHelp && !opts.ShowVersion {
		opts.SourceFiles = args
	}

	return opts, nil
}
