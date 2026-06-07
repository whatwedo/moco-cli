// Package cli holds the hand-written CLI infrastructure: the root command, the
// application context and helpers used by the generated command code.
package cli

import (
	"context"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/whatwedo/moco-cli/internal/client"
	"github.com/whatwedo/moco-cli/internal/config"
)

// App bundles the global flags and output streams. Generated commands receive
// a pointer to it and call Execute.
type App struct {
	Version  string
	Endpoint string
	Token    string
	Output   string

	Out io.Writer
	Err io.Writer

	// Client, if set, is used instead of building one from the resolved
	// configuration. It exists as a seam for tests and embedding.
	Client *client.Client
}

// NewRootCmd creates the root command together with its global flags and
// returns the associated App context.
func NewRootCmd(version string) (*cobra.Command, *App) {
	app := &App{Version: version, Out: os.Stdout, Err: os.Stderr}

	root := &cobra.Command{
		Use:           "moco",
		Short:         "Command-line client for the MOCO API",
		Long:          "moco is a command-line client for the MOCO API v1 (https://www.mocoapp.com).",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return &UsageError{Err: err}
	})

	pf := root.PersistentFlags()
	pf.StringVar(&app.Endpoint, "endpoint", "", "MOCO host, e.g. whatwedo.mocoapp.com (or "+config.EnvEndpoint+")")
	pf.StringVar(&app.Token, "token", "", "API token (or "+config.EnvToken+")")
	pf.StringVarP(&app.Output, "output", "o", "json", "output format: json|raw")

	return root, app
}

// Execute resolves the configuration, sends the request and prints the
// response in the selected format.
func (a *App) Execute(ctx context.Context, req client.Request) error {
	c := a.Client
	if c == nil {
		cfg, err := config.Resolve(a.Endpoint, a.Token)
		if err != nil {
			return &UsageError{Err: err}
		}
		c = client.New(cfg.BaseURL(), cfg.Token, "moco-cli/"+a.Version)
	}

	resp, err := c.Do(ctx, req)
	if err != nil {
		return err
	}

	return a.writeOutput(resp)
}
