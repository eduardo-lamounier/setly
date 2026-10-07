package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/eduardo-lamounier/setly/internal/cli"
)

const (
	appName    = "setly"
	appVersion = "1.0.0"
)

func main() {
	opts, err := cli.GetOptions()
	if err != nil {
		printError(err)
		os.Exit(1)
	}

	if opts.ShowHelp {
		fmt.Println(cli.HelpMessage)
		return
	} else if opts.ShowVersion {
		fmt.Println(appName, appVersion)
		return
	}

	if opts.SourceFilePath == nil {
		err = RunPrompt()
	} else {
		err = RunFile(*opts.SourceFilePath)
	}

	if err != nil {
		printError(err)
		os.Exit(1)
	}
}

func RunPrompt() error {
	return fmt.Errorf("REPL mode not yet implemented")
}

func RunFile(sourceFilePath string) error {
	return fmt.Errorf("file running mode not yet implemented")
}

func printError(err error) {
	for errMsg := range strings.SplitSeq(err.Error(), "\n") {
		fmt.Printf("ERROR: %s.\n", errMsg)
	}
}
