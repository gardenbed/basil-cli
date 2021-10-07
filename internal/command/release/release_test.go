package release

import (
	"context"
	"errors"
	"testing"

	buildcmd "github.com/gardenbed/basil-cli/internal/command/build"
	"github.com/gardenbed/basil-cli/internal/semver"
	"github.com/gardenbed/basil-cli/internal/shell"
	changelogspec "github.com/gardenbed/changelog/spec"
	"github.com/gardenbed/go-github"

	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"

	"github.com/gardenbed/basil-cli/internal/command"
	"github.com/gardenbed/basil-cli/internal/config"
	"github.com/gardenbed/basil-cli/internal/spec"
)

var (
	user = github.User{
		Login: "octocat",
	}

	repo = github.Repository{
		Name:          "Hello-World",
		FullName:      "octocat/Hello-World",
		DefaultBranch: "main",
	}

	version = semver.SemVer{
		Major: 0, Minor: 1, Patch: 0,
		Prerelease: []string{"10", "aaaaaaa"},
	}

	draftRelease = &github.Release{
		Name:       "0.1.0",
		TagName:    "v0.1.0",
		Target:     "main",
		Draft:      true,
		Prerelease: false,
	}

	artifacts = []buildcmd.Artifact{
		{
			Path:  "bin/app",
			Label: "linux",
		},
	}

	asset = github.ReleaseAsset{
		Name:  "bin/app",
		Label: "linux",
	}

	release = &github.Release{
		Name:       "0.1.0",
		TagName:    "v0.1.0",
		Target:     "main",
		Draft:      false,
		Prerelease: false,
	}
)

func TestNew(t *testing.T) {
	ui := cli.NewMockUi()
	config := config.Config{}
	spec := spec.Spec{}
	c := New(ui, config, spec)

	assert.NotNil(t, c)
}

func TestNewFactory(t *testing.T) {
	ui := cli.NewMockUi()
	config := config.Config{}
	spec := spec.Spec{}
	c, err := NewFactory(ui, config, spec)()

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
	c := &Command{
		ui: cli.NewMockUi(),
		config: config.Config{
			GitHub: config.GitHub{
				AccessToken: "access-token",
			},
		},
	}

	c.Run([]string{})

	assert.Equal(t, "gardenbed", c.data.owner)
	assert.Equal(t, "basil-cli", c.data.repo)
	assert.NotEmpty(t, c.data.changelogSpec)
	assert.NotNil(t, c.funcs.gitAdd)
	assert.NotNil(t, c.funcs.gitCommit)
	assert.NotNil(t, c.funcs.gitTag)
	assert.NotNil(t, c.funcs.goList)
	assert.NotNil(t, c.services.git)
	assert.NotNil(t, c.services.users)
	assert.NotNil(t, c.services.repo)
	assert.NotNil(t, c.services.pulls)
	assert.NotNil(t, c.services.changelog)
	assert.NotNil(t, c.commands.semver)
	assert.NotNil(t, c.commands.build)
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
				"-patch",
				"-minor",
				"-major",
				"-comment", "description",
				"-mode", "direct",
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
		config           config.Config
		spec             spec.Spec
		patchFlag        bool
		minorFlag        bool
		majorFlag        bool
		commentFlag      string
		gitAdd           shell.RunnerFunc
		gitCommit        shell.RunnerFunc
		gitTag           shell.RunnerFunc
		git              *MockGitService
		users            *MockUsersService
		repo             *MockRepoService
		pulls            *MockPullsService
		changelog        *MockChangelogService
		semver           *MockSemverCommand
		build            *MockBuildCommand
		expectedExitCode int
	}{
		{
			name: "RepoGetFails",
			repo: &MockRepoService{
				GetMocks: []GetMock{
					{OutError: errors.New("github error")},
				},
			},
			expectedExitCode: command.GitHubError,
		},
		{
			name: "GitHEADFails",
			git: &MockGitService{
				HEADMocks: []HEADMock{
					{OutError: errors.New("git error")},
				},
			},
			repo: &MockRepoService{
				GetMocks: []GetMock{
					{OutRepository: &repo, OutResponse: &github.Response{}},
				},
			},
			expectedExitCode: command.GitError,
		},
		{
			name: "NotOnDefaultBranch",
			git: &MockGitService{
				HEADMocks: []HEADMock{
					{OutBranch: "feature-branch"},
				},
			},
			repo: &MockRepoService{
				GetMocks: []GetMock{
					{OutRepository: &repo, OutResponse: &github.Response{}},
				},
			},
			expectedExitCode: command.GitError,
		},
		{
			name: "GitIsCleanFails",
			git: &MockGitService{
				HEADMocks: []HEADMock{
					{OutBranch: "main"},
				},
				IsCleanMocks: []IsCleanMock{
					{OutError: errors.New("git error")},
				},
			},
			repo: &MockRepoService{
				GetMocks: []GetMock{
					{OutRepository: &repo, OutResponse: &github.Response{}},
				},
			},
			expectedExitCode: command.GitError,
		},
		{
			name: "RepoNotClean",
			git: &MockGitService{
				HEADMocks: []HEADMock{
					{OutBranch: "main"},
				},
				IsCleanMocks: []IsCleanMock{
					{OutBool: false},
				},
			},
			repo: &MockRepoService{
				GetMocks: []GetMock{
					{OutRepository: &repo, OutResponse: &github.Response{}},
				},
			},
			expectedExitCode: command.GitError,
		},
		{
			name: "GitPullFails",
			git: &MockGitService{
				HEADMocks: []HEADMock{
					{OutBranch: "main"},
				},
				IsCleanMocks: []IsCleanMock{
					{OutBool: true},
				},
				PullMocks: []PullMock{
					{OutError: errors.New("git error")},
				},
			},
			repo: &MockRepoService{
				GetMocks: []GetMock{
					{OutRepository: &repo, OutResponse: &github.Response{}},
				},
			},
			expectedExitCode: command.GitError,
		},
		{
			name: "SemverRunFails",
			git: &MockGitService{
				HEADMocks: []HEADMock{
					{OutBranch: "main"},
				},
				IsCleanMocks: []IsCleanMock{
					{OutBool: true},
				},
				PullMocks: []PullMock{
					{OutError: nil},
				},
			},
			repo: &MockRepoService{
				GetMocks: []GetMock{
					{OutRepository: &repo, OutResponse: &github.Response{}},
				},
			},
			semver: &MockSemverCommand{
				RunMocks: []SemverRunMock{
					{OutCode: command.GitError},
				},
			},
			expectedExitCode: command.GitError,
		},
		{
			name: "CreateReleaseFails",
			git: &MockGitService{
				HEADMocks: []HEADMock{
					{OutBranch: "main"},
				},
				IsCleanMocks: []IsCleanMock{
					{OutBool: true},
				},
				PullMocks: []PullMock{
					{OutError: nil},
				},
			},
			repo: &MockRepoService{
				GetMocks: []GetMock{
					{OutRepository: &repo, OutResponse: &github.Response{}},
				},
				CreateReleaseMocks: []CreateReleaseMock{
					{OutError: errors.New("github error")},
				},
			},
			semver: &MockSemverCommand{
				RunMocks: []SemverRunMock{
					{OutCode: command.Success},
				},
				SemVerMocks: []SemVerMock{
					{OutSemVer: version},
				},
			},
			expectedExitCode: command.GitHubError,
		},
		{
			name:      "ChangelogGenerateFails",
			patchFlag: true,
			git: &MockGitService{
				HEADMocks: []HEADMock{
					{OutBranch: "main"},
				},
				IsCleanMocks: []IsCleanMock{
					{OutBool: true},
				},
				PullMocks: []PullMock{
					{OutError: nil},
				},
			},
			repo: &MockRepoService{
				GetMocks: []GetMock{
					{OutRepository: &repo, OutResponse: &github.Response{}},
				},
				CreateReleaseMocks: []CreateReleaseMock{
					{OutRelease: draftRelease, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutError: errors.New("changelog error")},
				},
			},
			semver: &MockSemverCommand{
				RunMocks: []SemverRunMock{
					{OutCode: command.Success},
				},
				SemVerMocks: []SemVerMock{
					{OutSemVer: version},
				},
			},
			expectedExitCode: command.ChangelogError,
		},
		{
			name:      "GitAddFails",
			minorFlag: true,
			gitAdd: func(context.Context, ...string) (int, string, error) {
				return 1, "", errors.New("git error")
			},
			git: &MockGitService{
				HEADMocks: []HEADMock{
					{OutBranch: "main"},
				},
				IsCleanMocks: []IsCleanMock{
					{OutBool: true},
				},
				PullMocks: []PullMock{
					{OutError: nil},
				},
			},
			repo: &MockRepoService{
				GetMocks: []GetMock{
					{OutRepository: &repo, OutResponse: &github.Response{}},
				},
				CreateReleaseMocks: []CreateReleaseMock{
					{OutRelease: draftRelease, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
				},
			},
			semver: &MockSemverCommand{
				RunMocks: []SemverRunMock{
					{OutCode: command.Success},
				},
				SemVerMocks: []SemVerMock{
					{OutSemVer: version},
				},
			},
			expectedExitCode: command.GitError,
		},
		{
			name:      "GitCommitFails",
			majorFlag: true,
			gitAdd: func(context.Context, ...string) (int, string, error) {
				return 0, "", nil
			},
			gitCommit: func(context.Context, ...string) (int, string, error) {
				return 1, "", errors.New("git error")
			},
			git: &MockGitService{
				HEADMocks: []HEADMock{
					{OutBranch: "main"},
				},
				IsCleanMocks: []IsCleanMock{
					{OutBool: true},
				},
				PullMocks: []PullMock{
					{OutError: nil},
				},
			},
			repo: &MockRepoService{
				GetMocks: []GetMock{
					{OutRepository: &repo, OutResponse: &github.Response{}},
				},
				CreateReleaseMocks: []CreateReleaseMock{
					{OutRelease: draftRelease, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
				},
			},
			semver: &MockSemverCommand{
				RunMocks: []SemverRunMock{
					{OutCode: command.Success},
				},
				SemVerMocks: []SemVerMock{
					{OutSemVer: version},
				},
			},
			expectedExitCode: command.GitError,
		},
		{
			name: "InvalidReleaseMode",
			spec: spec.Spec{
				Project: spec.Project{
					Release: spec.Release{
						Mode: spec.ReleaseMode(""),
					},
				},
			},
			gitAdd: func(context.Context, ...string) (int, string, error) {
				return 0, "", nil
			},
			gitCommit: func(context.Context, ...string) (int, string, error) {
				return 0, "", nil
			},
			git: &MockGitService{
				HEADMocks: []HEADMock{
					{OutBranch: "main"},
				},
				IsCleanMocks: []IsCleanMock{
					{OutBool: true},
				},
				PullMocks: []PullMock{
					{OutError: nil},
				},
			},
			users: &MockUsersService{
				UserMocks: []UserMock{
					{OutError: errors.New("github error")},
				},
			},
			repo: &MockRepoService{
				GetMocks: []GetMock{
					{OutRepository: &repo, OutResponse: &github.Response{}},
				},
				CreateReleaseMocks: []CreateReleaseMock{
					{OutRelease: draftRelease, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
				},
			},
			semver: &MockSemverCommand{
				RunMocks: []SemverRunMock{
					{OutCode: command.Success},
				},
				SemVerMocks: []SemVerMock{
					{OutSemVer: version},
				},
			},
			expectedExitCode: command.SpecError,
		},
		{
			name: "DirectReleaseFails",
			spec: spec.Spec{
				Project: spec.Project{
					Release: spec.Release{
						Mode: spec.ReleaseModeDirect,
					},
				},
			},
			gitAdd: func(context.Context, ...string) (int, string, error) {
				return 0, "", nil
			},
			gitCommit: func(context.Context, ...string) (int, string, error) {
				return 0, "", nil
			},
			git: &MockGitService{
				HEADMocks: []HEADMock{
					{OutBranch: "main"},
				},
				IsCleanMocks: []IsCleanMock{
					{OutBool: true},
				},
				PullMocks: []PullMock{
					{OutError: nil},
				},
			},
			users: &MockUsersService{
				UserMocks: []UserMock{
					{OutError: errors.New("github error")},
				},
			},
			repo: &MockRepoService{
				GetMocks: []GetMock{
					{OutRepository: &repo, OutResponse: &github.Response{}},
				},
				CreateReleaseMocks: []CreateReleaseMock{
					{OutRelease: draftRelease, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
				},
			},
			semver: &MockSemverCommand{
				RunMocks: []SemverRunMock{
					{OutCode: command.Success},
				},
				SemVerMocks: []SemVerMock{
					{OutSemVer: version},
				},
			},
			expectedExitCode: command.GitHubError,
		},
		{
			name: "IndirectReleaseFails",
			spec: spec.Spec{
				Project: spec.Project{
					Release: spec.Release{
						Mode: spec.ReleaseModeIndirect,
					},
				},
			},
			gitAdd: func(context.Context, ...string) (int, string, error) {
				return 0, "", nil
			},
			gitCommit: func(context.Context, ...string) (int, string, error) {
				return 0, "", nil
			},
			git: &MockGitService{
				HEADMocks: []HEADMock{
					{OutBranch: "main"},
				},
				IsCleanMocks: []IsCleanMock{
					{OutBool: true},
				},
				PullMocks: []PullMock{
					{OutError: nil},
				},
			},
			repo: &MockRepoService{
				GetMocks: []GetMock{
					{OutRepository: &repo, OutResponse: &github.Response{}},
				},
				CreateReleaseMocks: []CreateReleaseMock{
					{OutRelease: draftRelease, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
				},
			},
			semver: &MockSemverCommand{
				RunMocks: []SemverRunMock{
					{OutCode: command.Success},
				},
				SemVerMocks: []SemVerMock{
					{OutSemVer: version},
				},
			},
			expectedExitCode: command.GenericError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Command{
				ui:     cli.NewMockUi(),
				config: tc.config,
				spec:   tc.spec,
			}

			c.flags.patch = tc.patchFlag
			c.flags.minor = tc.minorFlag
			c.flags.major = tc.majorFlag
			c.flags.comment = tc.commentFlag

			c.data.owner = "octocat"
			c.data.repo = "Hello-World"
			c.data.changelogSpec = changelogspec.Spec{
				General: changelogspec.General{
					File: "CHANGELOG.md",
				},
			}

			c.funcs.gitAdd = tc.gitAdd
			c.funcs.gitCommit = tc.gitCommit
			c.funcs.gitTag = tc.gitTag
			c.services.git = tc.git
			c.services.users = tc.users
			c.services.repo = tc.repo
			c.services.pulls = tc.pulls
			c.services.changelog = tc.changelog
			c.commands.semver = tc.semver
			c.commands.build = tc.build

			exitCode := c.exec()

			assert.Equal(t, tc.expectedExitCode, exitCode)
		})
	}
}

func TestCommand_releaseDirectly(t *testing.T) {
	tests := []struct {
		name             string
		commentFlag      string
		goList           shell.RunnerFunc
		gitTag           shell.RunnerFunc
		git              *MockGitService
		users            *MockUsersService
		repo             *MockRepoService
		build            *MockBuildCommand
		commit           string
		version          semver.SemVer
		ctx              context.Context
		release          *github.Release
		defaultBranch    string
		changelog        string
		expectedExitCode int
	}{
		{
			name: "UsersUserFails",
			users: &MockUsersService{
				UserMocks: []UserMock{
					{OutError: errors.New("github error")},
				},
			},
			expectedExitCode: command.GitHubError,
		},
		{
			name: "RepoPermissionFails",
			users: &MockUsersService{
				UserMocks: []UserMock{
					{OutUser: &user, OutResponse: &github.Response{}},
				},
			},
			repo: &MockRepoService{
				PermissionMocks: []PermissionMock{
					{OutError: errors.New("github error")},
				},
			},
			expectedExitCode: command.GitHubError,
		},
		{
			name: "InvalidUserPermission",
			users: &MockUsersService{
				UserMocks: []UserMock{
					{OutUser: &user, OutResponse: &github.Response{}},
				},
			},
			repo: &MockRepoService{
				PermissionMocks: []PermissionMock{
					{OutPermission: github.PermissionWrite, OutResponse: &github.Response{}},
				},
			},
			expectedExitCode: command.GitHubError,
		},
		{
			name: "BuildRunFails",
			goList: func(context.Context, ...string) (int, string, error) {
				return 0, "", nil
			},
			users: &MockUsersService{
				UserMocks: []UserMock{
					{OutUser: &user, OutResponse: &github.Response{}},
				},
			},
			repo: &MockRepoService{
				PermissionMocks: []PermissionMock{
					{OutPermission: github.PermissionAdmin, OutResponse: &github.Response{}},
				},
			},
			build: &MockBuildCommand{
				RunMocks: []BuildRunMock{
					{OutCode: command.GoError},
				},
			},
			commit:           "6e8c7d217faab1d88905d4c75b4e7995a42c81d5",
			version:          version,
			ctx:              context.Background(),
			release:          draftRelease,
			defaultBranch:    "main",
			changelog:        "changelog content",
			expectedExitCode: command.GoError,
		},
		{
			name: "UploadReleaseAssetFails",
			goList: func(context.Context, ...string) (int, string, error) {
				return 0, "", nil
			},
			users: &MockUsersService{
				UserMocks: []UserMock{
					{OutUser: &user, OutResponse: &github.Response{}},
				},
			},
			repo: &MockRepoService{
				PermissionMocks: []PermissionMock{
					{OutPermission: github.PermissionAdmin, OutResponse: &github.Response{}},
				},
				UploadReleaseAssetMocks: []UploadReleaseAssetMock{
					{OutError: errors.New("github error")},
				},
			},
			build: &MockBuildCommand{
				RunMocks: []BuildRunMock{
					{OutCode: command.Success},
				},
				ArtifactsMocks: []ArtifactsMock{
					{OutArtifacts: artifacts},
				},
			},
			commit:           "6e8c7d217faab1d88905d4c75b4e7995a42c81d5",
			version:          version,
			ctx:              context.Background(),
			release:          draftRelease,
			defaultBranch:    "main",
			changelog:        "changelog content",
			expectedExitCode: command.GitHubError,
		},
		{
			name: "BranchProtectionFails",
			goList: func(context.Context, ...string) (int, string, error) {
				return 0, "", nil
			},
			users: &MockUsersService{
				UserMocks: []UserMock{
					{OutUser: &user, OutResponse: &github.Response{}},
				},
			},
			repo: &MockRepoService{
				PermissionMocks: []PermissionMock{
					{OutPermission: github.PermissionAdmin, OutResponse: &github.Response{}},
				},
				UploadReleaseAssetMocks: []UploadReleaseAssetMock{
					{OutReleaseAsset: &asset, OutResponse: &github.Response{}},
				},
				BranchProtectionMocks: []BranchProtectionMock{
					{OutError: errors.New("github error")},
				},
			},
			build: &MockBuildCommand{
				RunMocks: []BuildRunMock{
					{OutCode: command.Success},
				},
				ArtifactsMocks: []ArtifactsMock{
					{OutArtifacts: artifacts},
				},
			},
			commit:           "6e8c7d217faab1d88905d4c75b4e7995a42c81d5",
			version:          version,
			ctx:              context.Background(),
			release:          draftRelease,
			defaultBranch:    "main",
			changelog:        "changelog content",
			expectedExitCode: command.GitHubError,
		},
		{
			name: "CreateTagFails",
			goList: func(context.Context, ...string) (int, string, error) {
				return 0, "", nil
			},
			gitTag: func(context.Context, ...string) (int, string, error) {
				return 1, "", errors.New("git error")
			},
			users: &MockUsersService{
				UserMocks: []UserMock{
					{OutUser: &user, OutResponse: &github.Response{}},
				},
			},
			repo: &MockRepoService{
				PermissionMocks: []PermissionMock{
					{OutPermission: github.PermissionAdmin, OutResponse: &github.Response{}},
				},
				UploadReleaseAssetMocks: []UploadReleaseAssetMock{
					{OutReleaseAsset: &asset, OutResponse: &github.Response{}},
				},
				BranchProtectionMocks: []BranchProtectionMock{
					{OutResponse: &github.Response{}},
					{OutResponse: &github.Response{}},
				},
			},
			build: &MockBuildCommand{
				RunMocks: []BuildRunMock{
					{OutCode: command.Success},
				},
				ArtifactsMocks: []ArtifactsMock{
					{OutArtifacts: artifacts},
				},
			},
			commit:           "6e8c7d217faab1d88905d4c75b4e7995a42c81d5",
			version:          version,
			ctx:              context.Background(),
			release:          draftRelease,
			defaultBranch:    "main",
			changelog:        "changelog content",
			expectedExitCode: command.GitError,
		},
		{
			name: "GitPushFails",
			goList: func(context.Context, ...string) (int, string, error) {
				return 0, "", nil
			},
			gitTag: func(context.Context, ...string) (int, string, error) {
				return 0, "", nil
			},
			git: &MockGitService{
				PushMocks: []PushMock{
					{OutError: errors.New("git error")},
				},
			},
			users: &MockUsersService{
				UserMocks: []UserMock{
					{OutUser: &user, OutResponse: &github.Response{}},
				},
			},
			repo: &MockRepoService{
				PermissionMocks: []PermissionMock{
					{OutPermission: github.PermissionAdmin, OutResponse: &github.Response{}},
				},
				UploadReleaseAssetMocks: []UploadReleaseAssetMock{
					{OutReleaseAsset: &asset, OutResponse: &github.Response{}},
				},
				BranchProtectionMocks: []BranchProtectionMock{
					{OutResponse: &github.Response{}},
					{OutResponse: &github.Response{}},
				},
			},
			build: &MockBuildCommand{
				RunMocks: []BuildRunMock{
					{OutCode: command.Success},
				},
				ArtifactsMocks: []ArtifactsMock{
					{OutArtifacts: artifacts},
				},
			},
			commit:           "6e8c7d217faab1d88905d4c75b4e7995a42c81d5",
			version:          version,
			ctx:              context.Background(),
			release:          draftRelease,
			defaultBranch:    "main",
			changelog:        "changelog content",
			expectedExitCode: command.GitError,
		},
		{
			name: "GitPushTagFails",
			goList: func(context.Context, ...string) (int, string, error) {
				return 0, "", nil
			},
			gitTag: func(context.Context, ...string) (int, string, error) {
				return 0, "", nil
			},
			git: &MockGitService{
				PushMocks: []PushMock{
					{OutError: nil},
				},
				PushTagMocks: []PushTagMock{
					{OutError: errors.New("git error")},
				},
			},
			users: &MockUsersService{
				UserMocks: []UserMock{
					{OutUser: &user, OutResponse: &github.Response{}},
				},
			},
			repo: &MockRepoService{
				PermissionMocks: []PermissionMock{
					{OutPermission: github.PermissionAdmin, OutResponse: &github.Response{}},
				},
				UploadReleaseAssetMocks: []UploadReleaseAssetMock{
					{OutReleaseAsset: &asset, OutResponse: &github.Response{}},
				},
				BranchProtectionMocks: []BranchProtectionMock{
					{OutResponse: &github.Response{}},
					{OutResponse: &github.Response{}},
				},
			},
			build: &MockBuildCommand{
				RunMocks: []BuildRunMock{
					{OutCode: command.Success},
				},
				ArtifactsMocks: []ArtifactsMock{
					{OutArtifacts: artifacts},
				},
			},
			commit:           "6e8c7d217faab1d88905d4c75b4e7995a42c81d5",
			version:          version,
			ctx:              context.Background(),
			release:          draftRelease,
			defaultBranch:    "main",
			changelog:        "changelog content",
			expectedExitCode: command.GitError,
		},
		{
			name: "UpdateReleaseFails",
			goList: func(context.Context, ...string) (int, string, error) {
				return 0, "", nil
			},
			gitTag: func(context.Context, ...string) (int, string, error) {
				return 0, "", nil
			},
			git: &MockGitService{
				PushMocks: []PushMock{
					{OutError: nil},
				},
				PushTagMocks: []PushTagMock{
					{OutError: nil},
				},
			},
			users: &MockUsersService{
				UserMocks: []UserMock{
					{OutUser: &user, OutResponse: &github.Response{}},
				},
			},
			repo: &MockRepoService{
				PermissionMocks: []PermissionMock{
					{OutPermission: github.PermissionAdmin, OutResponse: &github.Response{}},
				},
				UploadReleaseAssetMocks: []UploadReleaseAssetMock{
					{OutReleaseAsset: &asset, OutResponse: &github.Response{}},
				},
				BranchProtectionMocks: []BranchProtectionMock{
					{OutResponse: &github.Response{}},
					{OutResponse: &github.Response{}},
				},
				UpdateReleaseMocks: []UpdateReleaseMock{
					{OutError: errors.New("github error")},
				},
			},
			build: &MockBuildCommand{
				RunMocks: []BuildRunMock{
					{OutCode: command.Success},
				},
				ArtifactsMocks: []ArtifactsMock{
					{OutArtifacts: artifacts},
				},
			},
			commit:           "6e8c7d217faab1d88905d4c75b4e7995a42c81d5",
			version:          version,
			ctx:              context.Background(),
			release:          draftRelease,
			defaultBranch:    "main",
			changelog:        "changelog content",
			expectedExitCode: command.GitHubError,
		},
		{
			name:        "Success",
			commentFlag: "release description",
			goList: func(context.Context, ...string) (int, string, error) {
				return 0, "", nil
			},
			gitTag: func(context.Context, ...string) (int, string, error) {
				return 0, "", nil
			},
			git: &MockGitService{
				PushMocks: []PushMock{
					{OutError: nil},
				},
				PushTagMocks: []PushTagMock{
					{OutError: nil},
				},
			},
			users: &MockUsersService{
				UserMocks: []UserMock{
					{OutUser: &user, OutResponse: &github.Response{}},
				},
			},
			repo: &MockRepoService{
				PermissionMocks: []PermissionMock{
					{OutPermission: github.PermissionAdmin, OutResponse: &github.Response{}},
				},
				UploadReleaseAssetMocks: []UploadReleaseAssetMock{
					{OutReleaseAsset: &asset, OutResponse: &github.Response{}},
				},
				BranchProtectionMocks: []BranchProtectionMock{
					{OutResponse: &github.Response{}},
					{OutResponse: &github.Response{}},
				},
				UpdateReleaseMocks: []UpdateReleaseMock{
					{OutRelease: release, OutResponse: &github.Response{}},
				},
			},
			build: &MockBuildCommand{
				RunMocks: []BuildRunMock{
					{OutCode: command.Success},
				},
				ArtifactsMocks: []ArtifactsMock{
					{OutArtifacts: artifacts},
				},
			},
			commit:           "6e8c7d217faab1d88905d4c75b4e7995a42c81d5",
			version:          version,
			ctx:              context.Background(),
			release:          draftRelease,
			defaultBranch:    "main",
			changelog:        "changelog content",
			expectedExitCode: command.Success,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Command{
				ui: cli.NewMockUi(),
			}

			c.flags.comment = tc.commentFlag
			c.funcs.gitTag = tc.gitTag
			c.funcs.goList = tc.goList
			c.services.git = tc.git
			c.services.users = tc.users
			c.services.repo = tc.repo
			c.commands.build = tc.build
			c.outputs.commit = tc.commit
			c.outputs.version = tc.version

			exitCode := c.releaseDirectly(tc.ctx, tc.release, tc.defaultBranch, tc.changelog)

			assert.Equal(t, tc.expectedExitCode, exitCode)
		})
	}
}

func TestCommand_releaseIndirectly(t *testing.T) {
	tests := []struct {
		name             string
		ctx              context.Context
		release          *github.Release
		expectedExitCode int
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Command{
				ui: cli.NewMockUi(),
			}

			exitCode := c.releaseIndirectly(tc.ctx, tc.release)

			assert.Equal(t, tc.expectedExitCode, exitCode)
		})
	}
}
