package main

import (
	"testing"

	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"

	"github.com/gardenbed/basil-cli/internal/config"
	"github.com/gardenbed/basil-cli/internal/spec"
)

func TestCreateCLI(t *testing.T) {
	ui := cli.NewMockUi()
	config := config.Config{}
	spec := spec.Spec{}
	cli := createCLI(ui, config, spec)

	assert.NotNil(t, cli)
}
