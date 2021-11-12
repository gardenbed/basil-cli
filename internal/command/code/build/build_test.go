package build

import (
	"errors"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"

	"github.com/gardenbed/basil-cli/internal/command"
)

func TestNew(t *testing.T) {
	ui := cli.NewMockUi()
	c := New(ui)

	assert.NotNil(t, c)
}

func TestNewFactory(t *testing.T) {
	ui := cli.NewMockUi()
	c, err := NewFactory(ui)()

	assert.NoError(t, err)
	assert.NotNil(t, c)
}

func TestCommand_Synopsis(t *testing.T) {
	c := new(Command)
	synopsis := c.Synopsis()

	assert.NotEmpty(t, synopsis)
}

func TestCommand_Help(t *testing.T) {
	c := new(Command)
	help := c.Help()

	assert.NotEmpty(t, help)
}

func TestCommand_Run(t *testing.T) {
	t.Run("InvalidFlag", func(t *testing.T) {
		c := &Command{ui: cli.NewMockUi()}
		exitCode := c.Run([]string{"-undefined"})

		assert.Equal(t, command.FlagError, exitCode)
	})

	t.Run("OK", func(t *testing.T) {
		c := &Command{ui: cli.NewMockUi()}
		c.Run([]string{"/dev/null"})

		assert.NotNil(t, c.services.compiler)
	})
}

func TestCommand_parseFlags(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		expectedExitCode int
	}{
		{
			name:             "InvalidFlag",
			args:             []string{"-undefined"},
			expectedExitCode: command.FlagError,
		},
		{
			name:             "NoFlag",
			args:             []string{},
			expectedExitCode: command.Success,
		},
		{
			name: "OK",
			args: []string{
				"-exported",
				"-names", "Request,Response",
				"-regexp", `[A-Z][0-9a-z]Entity$`,
				"./...",
			},
			expectedExitCode: command.Success,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Command{ui: cli.NewMockUi()}
			exitCode := c.parseFlags(tc.args)

			assert.Equal(t, tc.expectedExitCode, exitCode)
		})
	}
}

func TestCommand_exec(t *testing.T) {
	tests := []struct {
		name             string
		compiler         *MockCompilerService
		expectedExitCode int
	}{
		{
			name: "CompileFails",
			compiler: &MockCompilerService{
				CompileMocks: []CompileMock{
					{OutError: errors.New("compile error")},
				},
			},
			expectedExitCode: command.CompileError,
		},
		{
			name: "Success",
			compiler: &MockCompilerService{
				CompileMocks: []CompileMock{
					{OutError: nil},
				},
			},
			expectedExitCode: command.Success,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Command{
				ui: cli.NewMockUi(),
			}

			c.services.compiler = tc.compiler

			exitCode := c.exec()

			assert.Equal(t, tc.expectedExitCode, exitCode)
		})
	}
}
