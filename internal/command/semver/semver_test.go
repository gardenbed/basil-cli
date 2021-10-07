package semver

import (
	"errors"
	"testing"
	"time"

	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"

	"github.com/gardenbed/basil-cli/internal/command"
	"github.com/gardenbed/basil-cli/internal/git"
	"github.com/gardenbed/basil-cli/internal/semver"
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
		c.Run([]string{})

		assert.NotNil(t, c.services.git)
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
			c := &Command{ui: cli.NewMockUi()}
			exitCode := c.parseFlags(tc.args)

			assert.Equal(t, tc.expectedExitCode, exitCode)
		})
	}
}

func TestCommand_exec(t *testing.T) {
	tests := []struct {
		name             string
		git              *MockGitService
		expectedExitCode int
		expectedSemver   string
	}{
		{
			name: "IsCleanFails",
			git: &MockGitService{
				IsCleanMocks: []IsCleanMock{
					{OutError: errors.New("git error")},
				},
			},
			expectedExitCode: command.GitError,
		},
		{
			name: "HEADFails",
			git: &MockGitService{
				IsCleanMocks: []IsCleanMock{
					{OutBool: false},
				},
				HEADMocks: []HEADMock{
					{OutError: errors.New("git error")},
				},
			},
			expectedExitCode: command.GitError,
		},
		{
			name: "TagsFails",
			git: &MockGitService{
				IsCleanMocks: []IsCleanMock{
					{OutBool: false},
				},
				HEADMocks: []HEADMock{
					{OutHash: "8d2f15295f28f28355178250ede5cf43a40f0d14", OutBranch: "main"},
				},
				TagsMocks: []TagsMock{
					{OutError: errors.New("git error")},
				},
			},
			expectedExitCode: command.GitError,
		},
		{
			name: "CommitsInFails",
			git: &MockGitService{
				IsCleanMocks: []IsCleanMock{
					{OutBool: false},
				},
				HEADMocks: []HEADMock{
					{OutHash: "8d2f15295f28f28355178250ede5cf43a40f0d14", OutBranch: "main"},
				},
				TagsMocks: []TagsMock{
					{
						OutTags: git.Tags{},
					},
				},
				CommitsInMocks: []CommitsInMock{
					{OutError: errors.New("git error")},
				},
			},
			expectedExitCode: command.GitError,
		},
		{
			name: "WithoutTags_WithoutCommits_WorkingTreeNotClean",
			git: &MockGitService{
				IsCleanMocks: []IsCleanMock{
					{OutBool: false},
				},
				HEADMocks: []HEADMock{
					{OutHash: "8d2f15295f28f28355178250ede5cf43a40f0d14", OutBranch: "main"},
				},
				TagsMocks: []TagsMock{
					{
						OutTags: git.Tags{},
					},
				},
				CommitsInMocks: []CommitsInMock{
					{
						OutCommits: git.Commits{},
					},
				},
			},
			expectedExitCode: command.Success,
			expectedSemver:   "0.1.0-0.dev",
		},
		{
			name: "WithoutTags_WithCommits_WorkingTreeClean",
			git: &MockGitService{
				IsCleanMocks: []IsCleanMock{
					{OutBool: true},
				},
				HEADMocks: []HEADMock{
					{OutHash: "8d2f15295f28f28355178250ede5cf43a40f0d14", OutBranch: "main"},
				},
				TagsMocks: []TagsMock{
					{
						OutTags: git.Tags{},
					},
				},
				CommitsInMocks: []CommitsInMock{
					{
						OutCommits: git.Commits{
							{Hash: "8d2f15295f28f28355178250ede5cf43a40f0d14"},
							{Hash: "3a1960ec0cec18d2dca14d270d11c5bc4138abf6"},
						},
					},
				},
			},
			expectedExitCode: command.Success,
			expectedSemver:   "0.1.0-2.8d2f152",
		},
		{
			name: "WithoutTags_WithCommits_WorkingTreeNotClean",
			git: &MockGitService{
				IsCleanMocks: []IsCleanMock{
					{OutBool: false},
				},
				HEADMocks: []HEADMock{
					{OutHash: "8d2f15295f28f28355178250ede5cf43a40f0d14", OutBranch: "main"},
				},
				TagsMocks: []TagsMock{
					{
						OutTags: git.Tags{},
					},
				},
				CommitsInMocks: []CommitsInMock{
					{
						OutCommits: git.Commits{
							{Hash: "8d2f15295f28f28355178250ede5cf43a40f0d14"},
							{Hash: "3a1960ec0cec18d2dca14d270d11c5bc4138abf6"},
						},
					},
				},
			},
			expectedExitCode: command.Success,
			expectedSemver:   "0.1.0-2.dev",
		},
		{
			name: "WithTags_WithoutNewCommits_WorkingTreeClean",
			git: &MockGitService{
				IsCleanMocks: []IsCleanMock{
					{OutBool: true},
				},
				HEADMocks: []HEADMock{
					{OutHash: "8d2f15295f28f28355178250ede5cf43a40f0d14", OutBranch: "main"},
				},
				TagsMocks: []TagsMock{
					{
						OutTags: git.Tags{
							{
								Name:   "v0.1.0",
								Commit: git.Commit{Hash: "8d2f15295f28f28355178250ede5cf43a40f0d14"},
							},
						},
					},
				},
				CommitsInMocks: []CommitsInMock{
					{
						OutCommits: git.Commits{
							{Hash: "8d2f15295f28f28355178250ede5cf43a40f0d14"},
							{Hash: "3a1960ec0cec18d2dca14d270d11c5bc4138abf6"},
						},
					},
				},
			},
			expectedExitCode: command.Success,
			expectedSemver:   "0.1.0",
		},
		{
			name: "WithTags_WithoutNewCommits_WorkingTreeNotClean",
			git: &MockGitService{
				IsCleanMocks: []IsCleanMock{
					{OutBool: false},
				},
				HEADMocks: []HEADMock{
					{OutHash: "8d2f15295f28f28355178250ede5cf43a40f0d14", OutBranch: "main"},
				},
				TagsMocks: []TagsMock{
					{
						OutTags: git.Tags{
							{
								Name:   "v0.1.0",
								Commit: git.Commit{Hash: "8d2f15295f28f28355178250ede5cf43a40f0d14"},
							},
						},
					},
				},
				CommitsInMocks: []CommitsInMock{
					{
						OutCommits: git.Commits{
							{Hash: "8d2f15295f28f28355178250ede5cf43a40f0d14"},
							{Hash: "3a1960ec0cec18d2dca14d270d11c5bc4138abf6"},
						},
					},
				},
			},
			expectedExitCode: command.Success,
			expectedSemver:   "0.1.1-0.dev",
		},
		{
			name: "WithTags_WithNewCommits_WorkingTreeClean",
			git: &MockGitService{
				IsCleanMocks: []IsCleanMock{
					{OutBool: true},
				},
				HEADMocks: []HEADMock{
					{OutHash: "605a46c79d2500fef8d34145e4831624a7244bd1", OutBranch: "main"},
				},
				TagsMocks: []TagsMock{
					{
						OutTags: git.Tags{
							{
								Name:   "v0.1.0",
								Commit: git.Commit{Hash: "8d2f15295f28f28355178250ede5cf43a40f0d14"},
							},
						},
					},
				},
				CommitsInMocks: []CommitsInMock{
					{
						OutCommits: git.Commits{
							{Hash: "605a46c79d2500fef8d34145e4831624a7244bd1"},
							{Hash: "7fa23333fbc158af08d5b8073fa4828addde9c6b"},
							{Hash: "8d2f15295f28f28355178250ede5cf43a40f0d14"},
							{Hash: "3a1960ec0cec18d2dca14d270d11c5bc4138abf6"},
						},
					},
				},
			},
			expectedExitCode: command.Success,
			expectedSemver:   "0.1.1-2.605a46c",
		},
		{
			name: "WithTags_WithNewCommits_WorkingTreeNotClean",
			git: &MockGitService{
				IsCleanMocks: []IsCleanMock{
					{OutBool: false},
				},
				HEADMocks: []HEADMock{
					{OutHash: "605a46c79d2500fef8d34145e4831624a7244bd1", OutBranch: "main"},
				},
				TagsMocks: []TagsMock{
					{
						OutTags: git.Tags{
							{
								Name:   "v0.1.0",
								Commit: git.Commit{Hash: "8d2f15295f28f28355178250ede5cf43a40f0d14"},
							},
						},
					},
				},
				CommitsInMocks: []CommitsInMock{
					{
						OutCommits: git.Commits{
							{Hash: "605a46c79d2500fef8d34145e4831624a7244bd1"},
							{Hash: "7fa23333fbc158af08d5b8073fa4828addde9c6b"},
							{Hash: "8d2f15295f28f28355178250ede5cf43a40f0d14"},
							{Hash: "3a1960ec0cec18d2dca14d270d11c5bc4138abf6"},
						},
					},
				},
			},
			expectedExitCode: command.Success,
			expectedSemver:   "0.1.1-2.dev",
		},
		{
			name: "WithTags_WithNewCommits_WorkingTreeClean_WithMiscTags",
			git: &MockGitService{
				IsCleanMocks: []IsCleanMock{
					{OutBool: true},
				},
				HEADMocks: []HEADMock{
					{OutHash: "605a46c79d2500fef8d34145e4831624a7244bd1", OutBranch: "main"},
				},
				TagsMocks: []TagsMock{
					{
						OutTags: git.Tags{
							{
								Name: "non-semver",
							},
							{
								Name:   "v0.1.0",
								Commit: git.Commit{Hash: "8d2f15295f28f28355178250ede5cf43a40f0d14"},
							},
						},
					},
				},
				CommitsInMocks: []CommitsInMock{
					{
						OutCommits: git.Commits{
							{Hash: "605a46c79d2500fef8d34145e4831624a7244bd1"},
							{Hash: "7fa23333fbc158af08d5b8073fa4828addde9c6b"},
							{Hash: "8d2f15295f28f28355178250ede5cf43a40f0d14"},
							{Hash: "3a1960ec0cec18d2dca14d270d11c5bc4138abf6"},
						},
					},
				},
			},
			expectedExitCode: command.Success,
			expectedSemver:   "0.1.1-2.605a46c",
		},
		{
			name: "WithTags_WithNewCommits_WorkingTreeClean_WithTagsAfterHEAD",
			git: &MockGitService{
				IsCleanMocks: []IsCleanMock{
					{OutBool: true},
				},
				HEADMocks: []HEADMock{
					{OutHash: "605a46c79d2500fef8d34145e4831624a7244bd1", OutBranch: "main"},
				},
				TagsMocks: []TagsMock{
					{
						OutTags: git.Tags{
							{
								Name: "v0.2.0",
								Commit: git.Commit{
									Hash: "9df3723fd334bbff67db8149e6e0893769d5a9d3",
									Committer: git.Signature{
										Time: time.Date(2020, time.November, 25, 12, 0, 0, 0, time.UTC),
									},
								},
							},
							{
								Name: "v0.1.0",
								Commit: git.Commit{
									Hash: "8d2f15295f28f28355178250ede5cf43a40f0d14",
									Committer: git.Signature{
										Time: time.Date(2020, time.November, 10, 12, 0, 0, 0, time.UTC),
									},
								},
							},
						},
					},
				},
				CommitsInMocks: []CommitsInMock{
					{
						OutCommits: git.Commits{
							{
								Hash: "605a46c79d2500fef8d34145e4831624a7244bd1",
								Committer: git.Signature{
									Time: time.Date(2020, time.November, 20, 12, 0, 0, 0, time.UTC),
								},
							},
							{
								Hash: "7fa23333fbc158af08d5b8073fa4828addde9c6b",
								Committer: git.Signature{
									Time: time.Date(2020, time.November, 15, 12, 0, 0, 0, time.UTC),
								},
							},
							{
								Hash: "8d2f15295f28f28355178250ede5cf43a40f0d14",
								Committer: git.Signature{
									Time: time.Date(2020, time.November, 10, 12, 0, 0, 0, time.UTC),
								},
							},
							{
								Hash: "3a1960ec0cec18d2dca14d270d11c5bc4138abf6",
								Committer: git.Signature{
									Time: time.Date(2020, time.November, 5, 12, 0, 0, 0, time.UTC),
								},
							},
						},
					},
				},
			},
			expectedExitCode: command.Success,
			expectedSemver:   "0.1.1-2.605a46c",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Command{ui: cli.NewMockUi()}
			c.services.git = tc.git

			exitCode := c.exec()

			assert.Equal(t, tc.expectedExitCode, exitCode)

			if tc.expectedExitCode == command.Success {
				assert.Equal(t, tc.expectedSemver, c.outputs.semver.String())
			} else {
				assert.Empty(t, c.outputs.semver)
			}
		})
	}
}

func TestCommand_SemVer(t *testing.T) {
	smv := semver.SemVer{
		Major: 0, Minor: 1, Patch: 0,
		Prerelease: []string{"7", "aaaaaaa"},
	}

	c := new(Command)
	c.outputs.semver = smv

	assert.Equal(t, smv, c.SemVer())
}
