package main

import (
	"os"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"

	"github.com/gardenbed/basil-cli/internal/config"
	"github.com/gardenbed/basil-cli/internal/spec"
)

func TestCreateUI(t *testing.T) {
	ui := createUI()

	assert.NotNil(t, ui)
}

func TestCreateCLI(t *testing.T) {
	os.Setenv(expEnvVar, "true")
	defer os.Unsetenv(expEnvVar)

	ui := cli.NewMockUi()
	config := config.Config{}
	spec := spec.Spec{}
	cli := createCLI(ui, config, spec)

	assert.NotNil(t, cli)
}
