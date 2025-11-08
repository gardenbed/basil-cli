//go:build expr

package main

import (
	"os"

	"github.com/mitchellh/cli"

	"github.com/gardenbed/basil-cli/internal/config"
	"github.com/gardenbed/basil-cli/internal/spec"
	"github.com/gardenbed/basil-cli/internal/ui"

	buildcmd "github.com/gardenbed/basil-cli/internal/command/code/build"
	mockcmd "github.com/gardenbed/basil-cli/internal/command/code/mock"
	createmonorepocmd "github.com/gardenbed/basil-cli/internal/command/monorepo/create"
	buildcmd "github.com/gardenbed/basil-cli/internal/command/project/build"
	createprojectcmd "github.com/gardenbed/basil-cli/internal/command/project/create"
	semvercmd "github.com/gardenbed/basil-cli/internal/command/project/semver"
)

func createCLI(ui ui.UI, config config.Config, spec spec.Spec) *cli.CLI {
	c := cli.NewCLI("basil", "experimental")
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"monorepo create": createmonorepocmd.NewFactory(ui, config),
		"project create":  createprojectcmd.NewFactory(ui, config),
		"project semver":  semvercmd.NewFactory(ui),
		"project build":   buildcmd.NewFactory(ui, spec),
		"code mock":       mockcmd.NewFactory(ui),
		"code build":      buildcmd.NewFactory(ui),
	}

	return c
}
