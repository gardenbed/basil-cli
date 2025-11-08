//go:build !expr

package main

import (
	"os"

	"github.com/mitchellh/cli"

	"github.com/gardenbed/basil-cli/internal/config"
	"github.com/gardenbed/basil-cli/internal/spec"
	"github.com/gardenbed/basil-cli/internal/ui"
	"github.com/gardenbed/basil-cli/metadata"

	configcmd "github.com/gardenbed/basil-cli/internal/command/config"
	createmonorepocmd "github.com/gardenbed/basil-cli/internal/command/monorepo/create"
	buildcmd "github.com/gardenbed/basil-cli/internal/command/project/build"
	createprojectcmd "github.com/gardenbed/basil-cli/internal/command/project/create"
	releasecmd "github.com/gardenbed/basil-cli/internal/command/project/release"
	semvercmd "github.com/gardenbed/basil-cli/internal/command/project/semver"
	updatecmd "github.com/gardenbed/basil-cli/internal/command/update"
)

func createCLI(ui ui.UI, config config.Config, spec spec.Spec) *cli.CLI {
	c := cli.NewCLI("basil", metadata.String())
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"update":          updatecmd.NewFactory(ui, config),
		"config":          configcmd.NewFactory(ui, config),
		"monorepo create": createmonorepocmd.NewFactory(ui, config),
		"project create":  createprojectcmd.NewFactory(ui, config),
		"project semver":  semvercmd.NewFactory(ui),
		"project build":   buildcmd.NewFactory(ui, spec),
		"project release": releasecmd.NewFactory(ui, config, spec),
	}

	return c
}
