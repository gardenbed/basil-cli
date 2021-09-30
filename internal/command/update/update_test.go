package update

import (
	"errors"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"

	"github.com/gardenbed/basil-cli/internal/command"
	"github.com/gardenbed/go-github"
)

func TestNew(t *testing.T) {
	ui := cli.NewMockUi()
	c, err := New(ui)()

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
	c.Run([]string{"--undefined"})

	assert.NotNil(t, c.services.repo)
}

func TestCommand_run(t *testing.T) {
	tests := []struct {
		name             string
		repo             *MockRepoService
		args             []string
		expectedExitCode int
	}{
		{
			name:             "UndefinedFlag",
			repo:             &MockRepoService{},
			args:             []string{"--undefined"},
			expectedExitCode: command.FlagError,
		},
		{
			name: "LatestReleaseFails",
			repo: &MockRepoService{
				LatestReleaseMocks: []LatestReleaseMock{
					{OutError: errors.New("error on getting the latest GitHub release")},
				},
			},
			args:             []string{},
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
			args:             []string{},
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
			args:             []string{},
			expectedExitCode: command.Success,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Command{ui: cli.NewMockUi()}
			c.services.repo = tc.repo

			exitCode := c.run(tc.args)

			assert.Equal(t, tc.expectedExitCode, exitCode)
		})
	}
}
