package config

import (
	"context"
	"flag"
	"time"

	"github.com/mitchellh/cli"

	"github.com/gardenbed/basil-cli/internal/command"
)

const (
	timeout  = time.Minute
	synopsis = `Configure Basil`
	help     = `
  Use this command for setting basil global configurations.

  Usage:  basil config

  Examples:
    basil config
  `
)

// Command is the cli.Command implementation for config command.
type Command struct {
	ui cli.Ui
}

// New creates a config command.
func New(ui cli.Ui) *Command {
	return &Command{
		ui: ui,
	}
}

// NewFactory returns a cli.CommandFactory for creating a config command.
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
	fs := flag.NewFlagSet("config", flag.ContinueOnError)

	fs.Usage = func() {
		c.ui.Output(c.Help())
	}

	if err := fs.Parse(args); err != nil {
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
