package update

import (
	"errors"
	"os"
	"testing"

	"github.com/gardenbed/go-github"
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
		c := &Command{ui: ui.NewNop()}
		c.Run([]string{})

		assert.NotNil(t, c.services.releases)
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
	t.Run("LookPathFails", func(t *testing.T) {
		arg := os.Args[0]
		os.Args[0] = "/dev/null"
		defer func() {
			os.Args[0] = arg
		}()

		c := &Command{ui: ui.NewNop()}
		assert.Equal(t, command.OSError, c.exec())
	})

	tests := []struct {
		name             string
		releases         *MockReleaseService
		expectedExitCode int
	}{
		{
			name: "LatestFails",
			releases: &MockReleaseService{
				LatestMocks: []LatestMock{
					{OutError: errors.New("error on getting the latest GitHub release")},
				},
			},
			expectedExitCode: command.GitHubError,
		},
		{
			name: "DownloadAssetFails",
			releases: &MockReleaseService{
				LatestMocks: []LatestMock{
					{
						OutRelease: &github.Release{
							Name:    "1.0.0",
							TagName: "v1.0.0",
						},
						OutResponse: &github.Response{},
					},
				},
				DownloadAssetMocks: []DownloadAssetMock{
					{OutError: errors.New("error on downloading the release asset")},
				},
			},
			expectedExitCode: command.GitHubError,
		},
		{
			name: "Success",
			releases: &MockReleaseService{
				LatestMocks: []LatestMock{
					{
						OutRelease: &github.Release{
							Name:    "1.0.0",
							TagName: "v1.0.0",
						},
						OutResponse: &github.Response{},
					},
				},
				DownloadAssetMocks: []DownloadAssetMock{
					{
						OutResponse: &github.Response{},
					},
				},
			},
			expectedExitCode: command.Success,
		},
	}

	// LookPath requires the test file to be an executable.
	// We also need ensure that the test file is accessible.

	// Creating a temporary file
	f, err := os.CreateTemp("", "basil-*")
	assert.NoError(t, err)

	defer func() {
		assert.NoError(t, os.Remove(f.Name()))
	}()

	// Set execute permission
	err = os.Chmod(f.Name(), 0755)
	assert.NoError(t, err)

	// Temporarily, replace the executable name for testing
	arg := os.Args[0]
	os.Args[0] = f.Name()
	defer func() {
		os.Args[0] = arg
	}()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Command{
				ui: ui.NewNop(),
			}

			c.services.releases = tc.releases

			exitCode := c.exec()

			assert.Equal(t, tc.expectedExitCode, exitCode)
		})
	}
}
