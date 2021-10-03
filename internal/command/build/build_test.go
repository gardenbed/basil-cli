package build

import (
	"context"
	"errors"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"

	"github.com/gardenbed/basil-cli/internal/command"
	"github.com/gardenbed/basil-cli/internal/semver"
	"github.com/gardenbed/basil-cli/internal/shell"
	"github.com/gardenbed/basil-cli/internal/spec"
)

func TestNew(t *testing.T) {
	ui := cli.NewMockUi()
	spec := spec.Spec{}
	c := New(ui, spec)

	assert.NotNil(t, c)
}

func TestNewFactory(t *testing.T) {
	ui := cli.NewMockUi()
	spec := spec.Spec{}
	c, err := NewFactory(ui, spec)()

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

	assert.NotNil(t, c.funcs.goList)
	assert.NotNil(t, c.funcs.goBuild)
	assert.NotNil(t, c.services.git)
	assert.NotNil(t, c.commands.semver)
}

func TestCommand_run(t *testing.T) {
	tests := []struct {
		name             string
		spec             spec.Spec
		goList           shell.RunnerFunc
		goBuild          shell.RunnerWithFunc
		git              *MockGitService
		semver           *MockSemverCommand
		args             []string
		expectedExitCode int
	}{
		{
			name: "UndefinedFlag",
			spec: spec.Spec{
				Project: spec.Project{
					Build: spec.Build{},
				},
			},
			args:             []string{"--undefined"},
			expectedExitCode: command.FlagError,
		},
		{
			name: "GoListAndGitHEADFail",
			spec: spec.Spec{
				Project: spec.Project{
					Build: spec.Build{},
				},
			},
			goList: func(ctx context.Context, args ...string) (int, string, error) {
				return 1, "", errors.New("go error")
			},
			git: &MockGitService{
				HEADMocks: []HEADMock{
					{OutError: errors.New("git error")},
				},
			},
			args:             []string{},
			expectedExitCode: command.GitError,
		},
		{
			name: "SemverRunFails",
			spec: spec.Spec{
				Project: spec.Project{
					Build: spec.Build{},
				},
			},
			goList: func(ctx context.Context, args ...string) (int, string, error) {
				return 0, "github.com/foo/bar/metadata", nil
			},
			git: &MockGitService{
				HEADMocks: []HEADMock{
					{OutHash: "7813389d2b09cdf851665b7848daa212b27e4e82", OutBranch: "main"},
				},
			},
			semver: &MockSemverCommand{
				RunMocks: []RunMock{
					{OutCode: command.GitError},
				},
			},
			args:             []string{},
			expectedExitCode: command.GitError,
		},
		{
			name: "Success_NoArtifact",
			spec: spec.Spec{
				Project: spec.Project{
					Build: spec.Build{},
				},
			},
			goList: func(ctx context.Context, args ...string) (int, string, error) {
				return 0, "github.com/foo/bar/metadata", nil
			},
			git: &MockGitService{
				HEADMocks: []HEADMock{
					{OutHash: "7813389d2b09cdf851665b7848daa212b27e4e82", OutBranch: "main"},
				},
			},
			semver: &MockSemverCommand{
				RunMocks: []RunMock{
					{OutCode: command.Success},
				},
				SemVerMocks: []SemVerMock{
					{
						OutSemVer: semver.SemVer{
							Major: 1,
							Minor: 0,
							Patch: 0,
						},
					},
				},
			},
			args:             []string{},
			expectedExitCode: command.Success,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Command{
				ui:   cli.NewMockUi(),
				spec: tc.spec,
			}

			c.funcs.goList = tc.goList
			c.funcs.goBuild = tc.goBuild
			c.services.git = tc.git
			c.commands.semver = tc.semver

			exitCode := c.run(tc.args)

			assert.Equal(t, tc.expectedExitCode, exitCode)
		})
	}
}

func TestCommand_buildAll(t *testing.T) {
	tests := []struct {
		name          string
		buildSpec     spec.Build
		goBuild       shell.RunnerWithFunc
		ctx           context.Context
		ldFlags       string
		mainPkg       string
		output        string
		expectedError string
	}{
		{
			name: "WithoutCrossCompile_BuildFails",
			buildSpec: spec.Build{
				CrossCompile: false,
			},
			goBuild: func(ctx context.Context, opts shell.RunOptions, args ...string) (int, string, error) {
				return 1, "", errors.New("go build error")
			},
			ctx:           context.Background(),
			ldFlags:       `-X "github.com/foo/bar/metadata.Version=1.0.0"`,
			mainPkg:       "./cmd/app",
			output:        "./bin/app",
			expectedError: "go build error",
		},
		{
			name: "WithoutCrossCompile_BuildSucceeds",
			buildSpec: spec.Build{
				CrossCompile: false,
			},
			goBuild: func(ctx context.Context, opts shell.RunOptions, args ...string) (int, string, error) {
				return 0, "github.com/foo/bar/metadata", nil
			},
			ctx:           context.Background(),
			ldFlags:       `-X "github.com/foo/bar/metadata.Version=1.0.0"`,
			mainPkg:       "./cmd/app",
			output:        "./bin/app",
			expectedError: "",
		},
		{
			name: "WithCrossCompile_BuildFails",
			buildSpec: spec.Build{
				CrossCompile: true,
				Platforms:    []string{"linux-amd64", "darwin-amd64"},
			},
			goBuild: func(ctx context.Context, opts shell.RunOptions, args ...string) (int, string, error) {
				return 1, "", errors.New("go build error")
			},
			ctx:           context.Background(),
			ldFlags:       `-X "github.com/foo/bar/metadata.Version=1.0.0"`,
			mainPkg:       "./cmd/app",
			output:        "./bin/app",
			expectedError: "go build error",
		},
		{
			name: "WithCrossCompile_BuildSucceeds",
			buildSpec: spec.Build{
				CrossCompile: true,
				Platforms:    []string{"linux-amd64", "darwin-amd64"},
			},
			goBuild: func(ctx context.Context, opts shell.RunOptions, args ...string) (int, string, error) {
				return 0, "github.com/foo/bar/metadata", nil
			},
			ctx:           context.Background(),
			ldFlags:       `-X "github.com/foo/bar/metadata.Version=1.0.0"`,
			mainPkg:       "./cmd/app",
			output:        "./bin/app",
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Command{
				ui: cli.NewMockUi(),
				spec: spec.Spec{
					Project: spec.Project{
						Build: tc.buildSpec,
					},
				},
			}

			c.funcs.goBuild = tc.goBuild

			err := c.buildAll(tc.ctx, tc.ldFlags, tc.mainPkg, tc.output)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestCommand_Artifacts(t *testing.T) {
	artifacts := []Artifact{
		{"bin/app", "linux"},
	}

	c := new(Command)
	c.outputs.artifacts = artifacts

	assert.Equal(t, artifacts, c.Artifacts())
}
