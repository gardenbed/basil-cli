package main

import (
	"os"

	"github.com/mitchellh/cli"

	"github.com/gardenbed/basil-cli/internal/command/semver"
	"github.com/gardenbed/basil-cli/internal/command/update"
	"github.com/gardenbed/basil-cli/metadata"
)

func main() {
	ui := createUI()
	c := createCLI(ui)

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
			OutputColor: cli.UiColorCyan,
			InfoColor:   cli.UiColorGreen,
			WarnColor:   cli.UiColorYellow,
			ErrorColor:  cli.UiColorRed,
		},
	}
}

func createCLI(ui cli.Ui) *cli.CLI {
	c := cli.NewCLI("basil", metadata.String())
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"semver": semver.New(ui),
		"update": update.New(ui),
	}

	return c
}
