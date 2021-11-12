package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUI(t *testing.T) {
	ui := createUI()
	assert.NotNil(t, ui)
}
