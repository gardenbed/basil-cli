//go:build expr

package main

import (
	"os"

	"github.com/mitchellh/cli"

	"github.com/gardenbed/basil-cli/internal/config"
	"github.com/gardenbed/basil-cli/internal/spec"

	mockcmd "github.com/gardenbed/basil-cli/internal/command/code/mock"
)

func createCLI(ui cli.Ui, config config.Config, spec spec.Spec) *cli.CLI {
	c := cli.NewCLI("basil", "experimental")
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"code mock": mockcmd.NewFactory(ui),
	}

	return c
}
