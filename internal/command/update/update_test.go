package update

import (
	"errors"
	"os"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"

	"github.com/gardenbed/basil-cli/internal/command"
	"github.com/gardenbed/basil-cli/internal/config"
	"github.com/gardenbed/go-github"
)

func TestNew(t *testing.T) {
	ui := cli.NewMockUi()
	config := config.Config{}
	c := New(ui, config)

	assert.NotNil(t, c)
}

func TestNewFactory(t *testing.T) {
	ui := cli.NewMockUi()
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
	c := &Command{ui: cli.NewMockUi()}
	c.Run([]string{})

	assert.NotNil(t, c.services.repo)
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
			c := &Command{ui: cli.NewMockUi()}
			exitCode := c.parseFlags(tc.args)

			assert.Equal(t, tc.expectedExitCode, exitCode)
		})
	}
}

func TestCommand_exec(t *testing.T) {
	tests := []struct {
		name             string
		repo             *MockRepoService
		expectedExitCode int
	}{
		{
			name: "LatestReleaseFails",
			repo: &MockRepoService{
				LatestReleaseMocks: []LatestReleaseMock{
					{OutError: errors.New("error on getting the latest GitHub release")},
				},
			},
			expectedExitCode: command.GitHubError,
		},
		{
			name: "DownloadReleaseAssetFails",
			repo: &MockRepoService{
				LatestReleaseMocks: []LatestReleaseMock{
					{
						OutRelease: &github.Release{
							Name:    "1.0.0",
							TagName: "v1.0.0",
						},
						OutResponse: &github.Response{},
					},
				},
				DownloadReleaseAssetMocks: []DownloadReleaseAssetMock{
					{OutError: errors.New("error on downloading the release asset")},
				},
			},
			expectedExitCode: command.GitHubError,
		},
		{
			name: "Success",
			repo: &MockRepoService{
				LatestReleaseMocks: []LatestReleaseMock{
					{
						OutRelease: &github.Release{
							Name:    "1.0.0",
							TagName: "v1.0.0",
						},
						OutResponse: &github.Response{},
					},
				},
				DownloadReleaseAssetMocks: []DownloadReleaseAssetMock{
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
	defer os.Remove(f.Name())

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
			c := &Command{ui: cli.NewMockUi()}
			c.services.repo = tc.repo

			exitCode := c.exec()

			assert.Equal(t, tc.expectedExitCode, exitCode)
		})
	}
}
