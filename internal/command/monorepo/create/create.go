package create

import (
	"context"
	"flag"
	"time"

	"github.com/mitchellh/cli"

	"github.com/gardenbed/basil-cli/internal/command"
)

const (
	timeout  = time.Minute
	synopsis = `Create a new monorepo`
	help     = `
  Use this command for creating a new monorepo.

  Usage:  basil monorepo create [flags]

  Flags:
    -name    the name of the new monorepo

  Examples:
    basil monorepo create
    basil monorepo create -name=go-monorepo
  `
)

// Command is the cli.Command implementation for create command.
type Command struct {
	ui cli.Ui
}

// New creates a new command.
func New(ui cli.Ui) *Command {
	return &Command{
		ui: ui,
	}
}

// NewFactory returns a cli.CommandFactory for creating a new command.
func NewFactory(ui cli.Ui) cli.CommandFactory {
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

	return c.exec()
}

func (c *Command) parseFlags(args []string) int {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)

	fs.Usage = func() {
		c.ui.Output(c.Help())
	}

	if err := fs.Parse(args); err != nil {
		// In case of error, the error and help will be printed by the Parse method
		return command.FlagError
	}

	return command.Success
}

// exec in an auxiliary method, so we can test the business logic with mock dependencies.
func (c *Command) exec() int {
	_, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// ==============================> DONE <==============================

	return command.Success
}
