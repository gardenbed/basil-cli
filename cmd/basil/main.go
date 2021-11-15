package main

import (
	"os"

	"github.com/gardenbed/basil-cli/internal/command"
	"github.com/gardenbed/basil-cli/internal/config"
	"github.com/gardenbed/basil-cli/internal/spec"
	"github.com/gardenbed/basil-cli/internal/ui"
)

func main() {
	u := ui.NewInteractive(ui.Info)

	// Read the config from file if any
	config, err := config.Read()
	if err != nil {
		u.Errorf(ui.Red, "Cannot read the config file: %s", err)
		os.Exit(command.ConfigError)
	}

	// Read the spec from file if any
	spec, err := spec.Read()
	if err != nil {
		u.Errorf(ui.Red, "Cannot read the spec file: %s", err)
		os.Exit(command.SpecError)
	}
	spec = spec.WithDefaults()

	// Create the CLI app
	app := createCLI(u, config, spec)

	code, err := app.Run()
	if err != nil {
		u.Errorf(ui.Red, "%s", err)
	}

	os.Exit(code)
}
