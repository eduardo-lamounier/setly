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
	}
}

func printError(err error) {
	for errMsg := range strings.SplitSeq(err.Error(), "\n") {
		fmt.Printf("ERROR: %s.\n", errMsg)
	}
}
