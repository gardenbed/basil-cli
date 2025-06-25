// Package build implements the command for generating struct builders.
package build

import (
	"context"
	"flag"
	"regexp"
	"time"

	"github.com/gardenbed/charm/flagit"
	"github.com/gardenbed/go-parser"
	"github.com/mitchellh/cli"

	"github.com/gardenbed/basil-cli/internal/builder"
	"github.com/gardenbed/basil-cli/internal/command"
	"github.com/gardenbed/basil-cli/internal/ui"
)

const (
	timeout  = time.Minute
	synopsis = `Build Go structs`
	help     = `
  Use this command for generating builders for structs in Go.

  Usage:  basil code build [flags] [args]

  Flags:
    -exported    build exported structs and ignore unexported ones (default: false)
    -names       build structs matching these names (default: all)
    -regexp      build structs matching this regular expression (default: all)

  Examples:
    basil code build
  `
)

type (
	compilerService interface {
		Compile(string, parser.ParseOptions) error
	}
)

// Command is the cli.Command implementation for build command.
type Command struct {
	ui    ui.UI
	flags struct {
		Exported bool           `flag:"exported"`
		Names    []string       `flag:"names"`
		Regexp   *regexp.Regexp `flag:"regexp"`
	}
	args struct {
		packages string
	}
	services struct {
		compiler compilerService
	}
}

// New creates a new command.
func New(ui ui.UI) *Command {
	return &Command{
		ui: ui,
	}
}

// NewFactory returns a cli.CommandFactory for creating a new command.
func NewFactory(ui ui.UI) cli.CommandFactory {
	return func() (cli.Command, error) {
		return New(ui), nil
	}
}

// Synopsis returns a short one-line synopsis for the command.
func (c *Command) Synopsis() string {
	return synopsis
}

// Help returns a long help text including usage, description, and list of flags for the command.
func (c *Command) Help() string {
	return help
}

// Run runs the actual command with the given command-line arguments.
// This method is used as a proxy for creating dependencies and the actual command execution is delegated to the run method for testing purposes.
func (c *Command) Run(args []string) int {
	if code := c.parseFlags(args); code != command.Success {
		return code
	}

	c.services.compiler = builder.New(c.ui)

	return c.exec()
}

func (c *Command) parseFlags(args []string) int {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)

	fs.Usage = func() {
		c.ui.Printf(c.Help())
	}

	if err := flagit.Register(fs, &c.flags, false); err != nil {
		return command.GenericError
	}

	if err := fs.Parse(args); err != nil {
		// In case of error, the error and help will be printed by the Parse method
		return command.FlagError
	}

	c.args.packages = fs.Arg(0)

	return command.Success
}

// exec in an auxiliary method, so we can test the business logic with mock dependencies.
func (c *Command) exec() int {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// ==============================> RUN PREFLIGHT CHECKS <==============================

	checklist := command.PreflightChecklist{}

	info, err := command.RunPreflightChecks(ctx, checklist)
	if err != nil {
		c.ui.Errorf(ui.Red, "%s", err)
		return command.PreflightError
	}

	// ==============================> GENERATE CODE <==============================

	if c.args.packages == "" {
		c.args.packages = info.WorkingDirectory
	}

	opts := parser.ParseOptions{
		SkipTestFiles: true,
		TypeFilter: parser.TypeFilter{
			Exported: c.flags.Exported,
			Names:    c.flags.Names,
			Regexp:   c.flags.Regexp,
		},
	}

	if err := c.services.compiler.Compile(c.args.packages, opts); err != nil {
		c.ui.Errorf(ui.Red, "%s", err)
		return command.CompileError
	}

	// ==============================> DONE <==============================

	return command.Success
}
