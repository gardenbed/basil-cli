package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/cli"

	"github.com/gardenbed/basil-cli/internal/config"
	"github.com/gardenbed/basil-cli/internal/spec"
	"github.com/gardenbed/basil-cli/metadata"

	"github.com/gardenbed/basil-cli/internal/command"
	mockcmd "github.com/gardenbed/basil-cli/internal/command/code/mock"
	configcmd "github.com/gardenbed/basil-cli/internal/command/config"
	createmonorepocmd "github.com/gardenbed/basil-cli/internal/command/monorepo/create"
	buildcmd "github.com/gardenbed/basil-cli/internal/command/project/build"
	createprojectcmd "github.com/gardenbed/basil-cli/internal/command/project/create"
	releasecmd "github.com/gardenbed/basil-cli/internal/command/project/release"
	semvercmd "github.com/gardenbed/basil-cli/internal/command/project/semver"
	updatecmd "github.com/gardenbed/basil-cli/internal/command/update"
)

const expEnvVar = "BASIL_EXPERIMENT"

func main() {
	ui := createUI()

	// Read the config from file if any
	config, err := config.Read()
	if err != nil {
		ui.Error(fmt.Sprintf("Cannot read the config file: %s", err))
		os.Exit(command.ConfigError)
	}

	// Read the spec from file if any
	spec, err := spec.Read()
	if err != nil {
		ui.Error(fmt.Sprintf("Cannot read the spec file: %s", err))
		os.Exit(command.SpecError)
	}
	spec = spec.WithDefaults()

	code, err := createCLI(ui, config, spec).Run()
	if err != nil {
		ui.Error(err.Error())
	}

	os.Exit(code)
}

func createUI() cli.Ui {
	return &cli.ConcurrentUi{
		Ui: &cli.ColoredUi{
			Ui: &cli.BasicUi{
				Reader:      os.Stdin,
				Writer:      os.Stdout,
				ErrorWriter: os.Stderr,
			},
			OutputColor: cli.UiColorNone,
			InfoColor:   cli.UiColorGreen,
			WarnColor:   cli.UiColorYellow,
			ErrorColor:  cli.UiColorRed,
		},
	}
}

func createCLI(ui cli.Ui, config config.Config, spec spec.Spec) *cli.CLI {
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

	// Enable experimental features
	if val := os.Getenv(expEnvVar); strings.ToLower(val) == "true" {
		c.Commands["code mock"] = mockcmd.NewFactory(ui)
	}

	return c
}
