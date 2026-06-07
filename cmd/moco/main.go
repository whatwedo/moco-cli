// Command moco is a command-line client for the MOCO API v1.
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/whatwedo/moco-cli/internal/cli"
	"github.com/whatwedo/moco-cli/internal/commands"
)

// version is overridden at build time via -ldflags.
var version = "dev"

func main() {
	root, app := cli.NewRootCmd(version)
	commands.AddCommands(root, app)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(exitCode(err))
	}
}

// exitCode maps an error to a process exit code: 2 for usage/configuration
// errors, 1 otherwise.
func exitCode(err error) int {
	var usage *cli.UsageError
	if errors.As(err, &usage) {
		return 2
	}
	return 1
}
