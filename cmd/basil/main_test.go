package main

import (
	"testing"

	"github.com/mitchellh/cli"

	"github.com/stretchr/testify/assert"
)

func TestCreateUI(t *testing.T) {
	ui := createUI()
	assert.NotNil(t, ui)
}

func TestCreateCLI(t *testing.T) {
	ui := cli.NewMockUi()
	cli := createCLI(ui)
	assert.NotNil(t, cli)
}
