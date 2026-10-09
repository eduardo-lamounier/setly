package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/eduardo-lamounier/setly/internal/cli"
	"github.com/eduardo-lamounier/setly/internal/setly"
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
		os.Exit(RunPrompt())
	} else {
		os.Exit(RunFile(*opts.SourceFilePath))
	}
}

func RunPrompt() int {
	exitCode := 0
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("You are now in interactive mode. Press Ctrl+C to finish.")
	for {
		fmt.Print("> ")

		if !scanner.Scan() {
			break
		}

		prompt := scanner.Text()

		err := setly.Run(prompt)
		if err != nil {
			printError(err)
			exitCode = 1
		}
	}

	if err := scanner.Err(); err != nil {
		printError(fmt.Errorf("problem when reading from interactive input: %w", err))
		exitCode = 1
	}

	return exitCode
}

func RunFile(sourceFilePath string) int {
	source, err := os.ReadFile(sourceFilePath)
	if err != nil {
		printError(fmt.Errorf("error when opening the source file: %w", err))
		return 1
	}

	err = setly.Run(string(source))
	if err != nil {
		printError(err)
		return 1
	}

	return 0
}

func printError(err error) {
	for errMsg := range strings.SplitSeq(err.Error(), "\n") {
		fmt.Printf("ERROR: %s.\n", errMsg)
	}
}
