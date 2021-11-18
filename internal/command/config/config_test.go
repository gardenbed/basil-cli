package config

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gardenbed/basil-cli/internal/command"
	"github.com/gardenbed/basil-cli/internal/config"
	"github.com/gardenbed/basil-cli/internal/ui"
)

func TestNew(t *testing.T) {
	ui := ui.NewNop()
	config := config.Config{}
	c := New(ui, config)

	assert.NotNil(t, c)
}

func TestNewFactory(t *testing.T) {
	ui := ui.NewNop()
	config := config.Config{}
	c, err := NewFactory(ui, config)()

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
		c := &Command{ui: ui.NewNop()}
		exitCode := c.Run([]string{"-undefined"})

		assert.Equal(t, command.FlagError, exitCode)
	})

	t.Run("OK", func(t *testing.T) {
		c := &Command{
			ui: &MockUI{
				UI: ui.NewNop(),
				AskSecretMocks: []AskSecretMock{
					{OutError: errors.New("io error")},
				},
			},
		}

		c.Run([]string{})

		assert.NotNil(t, c.funcs.writeConfig)
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Command{ui: ui.NewNop()}
			exitCode := c.parseFlags(tc.args)

			assert.Equal(t, tc.expectedExitCode, exitCode)
		})
	}
}

func TestCommand_exec(t *testing.T) {
	tests := []struct {
		name             string
		ui               *MockUI
		config           config.Config
		writeConfig      writeConfigFunc
		expectedExitCode int
	}{
		{
			name: "ConfirmFails",
			ui: &MockUI{
				UI: ui.NewNop(),
				ConfirmMocks: []ConfirmMock{
					{OutError: errors.New("io error")},
				},
			},
			config: config.Config{
				GitHub: config.GitHub{
					AccessToken: "access token",
				},
			},
			expectedExitCode: command.InputError,
		},
		{
			name: "AskSecretFails",
			ui: &MockUI{
				UI: ui.NewNop(),
				ConfirmMocks: []ConfirmMock{
					{OutConfirmed: true},
				},
				AskSecretMocks: []AskSecretMock{
					{OutError: errors.New("io error")},
				},
			},
			config: config.Config{
				GitHub: config.GitHub{
					AccessToken: "access token",
				},
			},
			expectedExitCode: command.InputError,
		},
		{
			name: "WriteConfigFails",
			ui: &MockUI{
				UI: ui.NewNop(),
				ConfirmMocks: []ConfirmMock{
					{OutConfirmed: true},
				},
				AskSecretMocks: []AskSecretMock{
					{OutValue: "ghp_personalaccesstoken"},
				},
			},
			config: config.Config{
				GitHub: config.GitHub{
					AccessToken: "access token",
				},
			},
			writeConfig: func(config.Config) (string, error) {
				return "", errors.New("io error")
			},
			expectedExitCode: command.ConfigError,
		},
		{
			name: "Success",
			ui: &MockUI{
				UI: ui.NewNop(),
				ConfirmMocks: []ConfirmMock{
					{OutConfirmed: true},
				},
				AskSecretMocks: []AskSecretMock{
					{OutValue: "ghp_personalaccesstoken"},
				},
			},
			config: config.Config{
				GitHub: config.GitHub{
					AccessToken: "access token",
				},
			},
			writeConfig: func(config.Config) (string, error) {
				return "", nil
			},
			expectedExitCode: command.Success,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Command{
				ui:     tc.ui,
				config: tc.config,
			}

			c.funcs.writeConfig = tc.writeConfig

			exitCode := c.exec()

			assert.Equal(t, tc.expectedExitCode, exitCode)
		})
	}
}

func TestValidateInputToken(t *testing.T) {
	tests := []struct {
		name          string
		val           string
		expectedError string
	}{
		{
			name:          "InvalidToken",
			val:           "access token",
			expectedError: "invalid GitHub personal access token",
		},
		{
			name:          "ValidToken",
			val:           "ghp_0123456789abcdefghijklmnopqrstuvwxyz",
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateInputToken(tc.val)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
