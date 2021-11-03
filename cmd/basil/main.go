package main

import (
	"fmt"
	"os"

	"github.com/mitchellh/cli"

	"github.com/gardenbed/basil-cli/internal/config"
	"github.com/gardenbed/basil-cli/internal/spec"
	"github.com/gardenbed/basil-cli/metadata"

	"github.com/gardenbed/basil-cli/internal/command"
	configcmd "github.com/gardenbed/basil-cli/internal/command/config"
	buildcmd "github.com/gardenbed/basil-cli/internal/command/project/build"
	releasecmd "github.com/gardenbed/basil-cli/internal/command/project/release"
	semvercmd "github.com/gardenbed/basil-cli/internal/command/project/semver"
	updatecmd "github.com/gardenbed/basil-cli/internal/command/update"
)

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

	c := createCLI(ui, config, spec)
	code, err := c.Run()
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
		"project semver":  semvercmd.NewFactory(ui),
		"project build":   buildcmd.NewFactory(ui, spec),
		"project release": releasecmd.NewFactory(ui, config, spec),
	}

	return c
}
