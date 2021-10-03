package main

import (
	"testing"

	"github.com/gardenbed/basil-cli/internal/spec"
	"github.com/mitchellh/cli"

	"github.com/stretchr/testify/assert"
)

func TestCreateUI(t *testing.T) {
	ui := createUI()
	assert.NotNil(t, ui)
}

func TestCreateCLI(t *testing.T) {
	ui := cli.NewMockUi()
	spec := spec.Spec{}
	cli := createCLI(ui, spec)
	assert.NotNil(t, cli)
}
