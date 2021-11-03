package build

import (
	"context"
	"errors"
	"testing"

	"github.com/gardenbed/charm/shell"
	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"

	"github.com/gardenbed/basil-cli/internal/command"
	"github.com/gardenbed/basil-cli/internal/semver"
	"github.com/gardenbed/basil-cli/internal/spec"
	"github.com/gardenbed/basil-cli/metadata"
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
	t.Run("InvalidFlag", func(t *testing.T) {
		c := &Command{ui: cli.NewMockUi()}
		exitCode := c.Run([]string{"-undefined"})

		assert.Equal(t, command.FlagError, exitCode)
	})

	t.Run("OK", func(t *testing.T) {
		c := &Command{ui: cli.NewMockUi()}
		c.Run([]string{})

		assert.NotNil(t, c.funcs.gitRevSHA)
		assert.NotNil(t, c.funcs.gitRevBranch)
		assert.NotNil(t, c.funcs.goList)
		assert.NotNil(t, c.funcs.goBuild)
		assert.NotNil(t, c.commands.semver)
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
			name: "ValidFlags",
			args: []string{
				"-cross-compile",
				"-platforms", "linux-amd64,darwin-amd64,windows-amd64",
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
		spec             spec.Spec
		gitRevSHA        shell.RunnerFunc
		gitRevBranch     shell.RunnerFunc
		goList           shell.RunnerFunc
		goBuild          shell.RunnerWithFunc
		semver           *MockSemverCommand
		expectedExitCode int
	}{
		{
			name: "GitRevSHAFails",
			spec: spec.Spec{
				Project: spec.Project{
					Build: spec.Build{},
				},
			},
			gitRevSHA: func(context.Context, ...string) (int, string, error) {
				return 1, "", errors.New("go error")
			},
			expectedExitCode: command.GitError,
		},
		{
			name: "GitRevBranchFails",
			spec: spec.Spec{
				Project: spec.Project{
					Build: spec.Build{},
				},
			},
			gitRevSHA: func(context.Context, ...string) (int, string, error) {
				return 0, "7813389d2b09cdf851665b7848daa212b27e4e82", nil
			},
			gitRevBranch: func(context.Context, ...string) (int, string, error) {
				return 1, "", errors.New("go error")
			},
			expectedExitCode: command.GitError,
		},
		{
			name: "SemverRunFails",
			spec: spec.Spec{
				Project: spec.Project{
					Build: spec.Build{},
				},
			},
			gitRevSHA: func(context.Context, ...string) (int, string, error) {
				return 0, "7813389d2b09cdf851665b7848daa212b27e4e82", nil
			},
			gitRevBranch: func(context.Context, ...string) (int, string, error) {
				return 0, "main", nil
			},
			goList: func(context.Context, ...string) (int, string, error) {
				return 0, "github.com/foo/bar/metadata", nil
			},
			semver: &MockSemverCommand{
				RunMocks: []RunMock{
					{OutCode: command.GitError},
				},
			},
			expectedExitCode: command.GitError,
		},
		{
			name: "Success_NoArtifact",
			spec: spec.Spec{
				Project: spec.Project{
					Build: spec.Build{},
				},
			},
			gitRevSHA: func(context.Context, ...string) (int, string, error) {
				return 0, "7813389d2b09cdf851665b7848daa212b27e4e82", nil
			},
			gitRevBranch: func(context.Context, ...string) (int, string, error) {
				return 0, "main", nil
			},
			goList: func(context.Context, ...string) (int, string, error) {
				return 0, "github.com/foo/bar/metadata", nil
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
			expectedExitCode: command.Success,
		},
	}

	metadata.Version = "0.1.0-test"
	defer func() {
		metadata.Version = ""
	}()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Command{
				ui:   cli.NewMockUi(),
				spec: tc.spec,
			}

			c.funcs.gitRevSHA = tc.gitRevSHA
			c.funcs.gitRevBranch = tc.gitRevBranch
			c.funcs.goList = tc.goList
			c.funcs.goBuild = tc.goBuild
			c.commands.semver = tc.semver

			exitCode := c.exec()

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
			goBuild: func(context.Context, shell.RunOptions, ...string) (int, string, error) {
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
			goBuild: func(context.Context, shell.RunOptions, ...string) (int, string, error) {
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
			goBuild: func(context.Context, shell.RunOptions, ...string) (int, string, error) {
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
			goBuild: func(context.Context, shell.RunOptions, ...string) (int, string, error) {
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
