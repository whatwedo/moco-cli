// Command moco is a command-line client for the MOCO API v1.
package main

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/whatwedo/moco-cli/internal/cli"
	"github.com/whatwedo/moco-cli/internal/commands"
)

// version is overridden at build time via -ldflags (e.g. by GoReleaser).
var version = "dev"

func main() {
	root, app := cli.NewRootCmd(resolveVersion())
	commands.AddCommands(root, app)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(exitCode(err))
	}
}

// resolveVersion returns the ldflags-injected version, or falls back to the
// module version recorded in the build info (set for `go install module@vX`).
func resolveVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
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
