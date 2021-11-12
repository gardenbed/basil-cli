package main

import (
	"fmt"
	"os"

	"github.com/mitchellh/cli"

	"github.com/gardenbed/basil-cli/internal/command"
	"github.com/gardenbed/basil-cli/internal/config"
	"github.com/gardenbed/basil-cli/internal/spec"
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

	// Create the CLI app
	app := createCLI(ui, config, spec)

	code, err := app.Run()
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
