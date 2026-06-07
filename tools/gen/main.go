// Command gen generates the Cobra commands under internal/commands from the
// MOCO OpenAPI spec. Run via `go generate ./...`.
package main

import (
	"flag"
	"fmt"
	"os"
)

// defaultSpecURL is MOCO's official OpenAPI spec. It is fetched at generation
// time; MOCO's spec is not vendored into this (AGPL-licensed) repository.
const defaultSpecURL = "https://docs.mocoapp.com/api/docs/v1.yaml"

func main() {
	specPath := flag.String("spec", defaultSpecURL, "OpenAPI spec source: an https URL or a local file path")
	outDir := flag.String("out", "internal/commands", "output directory for generated commands")
	dump := flag.Bool("dump", false, "print the command table instead of generating code")
	flag.Parse()

	if err := run(*specPath, *outDir, *dump); err != nil {
		fmt.Fprintln(os.Stderr, "gen:", err)
		os.Exit(1)
	}
}

func run(specPath, outDir string, dump bool) error {
	spec, err := LoadSpec(specPath)
	if err != nil {
		return err
	}
	cmds := BuildCommands(spec, LoadTranslations())

	if dump {
		return dumpCommands(cmds)
	}
	return emit(cmds, outDir)
}

func dumpCommands(cmds []Command) error {
	fallback := 0
	for _, c := range cmds {
		mark := ""
		if c.Short == c.Summary {
			mark = "  <FALLBACK>"
			fallback++
		}
		fmt.Printf("%-32s %-26s %-6s\t%s%s\n", c.Group, c.Name, c.Method, c.Short, mark)
	}
	fmt.Fprintf(os.Stderr, "\n%d commands, %d English fallbacks\n", len(cmds), fallback)
	return nil
}
