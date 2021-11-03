package release

import (
	"context"
	"errors"
	"testing"

	buildcmd "github.com/gardenbed/basil-cli/internal/command/project/build"
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
	}

	artifacts = []buildcmd.Artifact{
		{
			Path:  "bin/app",
			Label: "linux",
		},
	}

	draftRelease = github.Release{
		Name:       "0.1.0",
		TagName:    "v0.1.0",
		Target:     "main",
		Draft:      true,
		Prerelease: false,
	}

	release = github.Release{
		Name:       "0.1.0",
		TagName:    "v0.1.0",
		Target:     "main",
		Draft:      false,
		Prerelease: false,
	}

	asset = github.ReleaseAsset{
		Name:  "bin/app",
		Label: "linux",
	}

	emptySearchResult = &github.SearchIssuesResult{
		TotalCount:        0,
		IncompleteResults: false,
		Items:             []github.Issue{},
	}

	mergedSearchResult = &github.SearchIssuesResult{
		TotalCount:        0,
		IncompleteResults: false,
		Items: []github.Issue{
			{
				ID:     1,
				Number: 1001,
				State:  "merged",
				Title:  "Release 0.1.0",
			},
		},
	}

	openSearchResult = &github.SearchIssuesResult{
		TotalCount:        0,
		IncompleteResults: false,
		Items: []github.Issue{
			{
				ID:     1,
				Number: 1001,
				State:  "open",
				Title:  "Release 0.1.0",
			},
		},
	}

	mergedPull = &github.Pull{
		ID:             1,
		Number:         1001,
		Merged:         true,
		MergeCommitSHA: "e9e71afc9382f03807042fd2e1bda25bf4f099fb",
		HTMLURL:        "https://github.com/octocat/Hello-World/pull/1001",
	}

	openPull = &github.Pull{
		ID:      1,
		Number:  1001,
		State:   "open",
		HTMLURL: "https://github.com/octocat/Hello-World/pull/1001",
	}

	successRunnerFunc = func(context.Context, ...string) (int, string, error) {
		return 0, "", nil
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
	t.Run("InvalidFlag", func(t *testing.T) {
		c := &Command{ui: cli.NewMockUi()}
		exitCode := c.Run([]string{"-undefined"})

		assert.Equal(t, command.FlagError, exitCode)
	})

	t.Run("OK", func(t *testing.T) {
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
		assert.NotNil(t, c.funcs.goList)
		assert.NotNil(t, c.funcs.gitStatus)
		assert.NotNil(t, c.funcs.gitRevBranch)
		assert.NotNil(t, c.funcs.gitBranch)
		assert.NotNil(t, c.funcs.gitCheckout)
		assert.NotNil(t, c.funcs.gitAdd)
		assert.NotNil(t, c.funcs.gitCommit)
		assert.NotNil(t, c.funcs.gitTag)
		assert.NotNil(t, c.funcs.gitPull)
		assert.NotNil(t, c.funcs.gitPush)
		assert.NotNil(t, c.funcs.gitPushTag)
		assert.NotNil(t, c.funcs.gitPushBranch)
		assert.NotNil(t, c.services.git)
		assert.NotNil(t, c.services.repo)
		assert.NotNil(t, c.services.releases)
		assert.NotNil(t, c.services.pulls)
		assert.NotNil(t, c.services.users)
		assert.NotNil(t, c.services.search)
		assert.NotNil(t, c.services.changelog)
		assert.NotNil(t, c.commands.semver)
		assert.NotNil(t, c.commands.build)
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
		spec             spec.Spec
		patchFlag        bool
		minorFlag        bool
		majorFlag        bool
		gitRevBranch     shell.RunnerFunc
		gitStatus        shell.RunnerFunc
		gitPull          shell.RunnerFunc
		repo             *MockRepoService
		users            *MockUserService
		search           *MockSearchService
		semver           *MockSemverCommand
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
			name: "GitRevBranchFails",
			gitRevBranch: func(context.Context, ...string) (int, string, error) {
				return 1, "", errors.New("git error")
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
			gitRevBranch: func(context.Context, ...string) (int, string, error) {
				return 0, "feature-branch", nil
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
			gitRevBranch: func(context.Context, ...string) (int, string, error) {
				return 0, "main", nil
			},
			gitStatus: func(context.Context, ...string) (int, string, error) {
				return 1, "", errors.New("git error")
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
			gitRevBranch: func(context.Context, ...string) (int, string, error) {
				return 0, "main", nil
			},
			gitStatus: func(context.Context, ...string) (int, string, error) {
				return 0, "M foo/bar", nil
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
			gitRevBranch: func(context.Context, ...string) (int, string, error) {
				return 0, "main", nil
			},
			gitStatus: successRunnerFunc,
			gitPull: func(context.Context, ...string) (int, string, error) {
				return 1, "", errors.New("git error")
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
			gitRevBranch: func(context.Context, ...string) (int, string, error) {
				return 0, "main", nil
			},
			gitStatus: successRunnerFunc,
			gitPull:   successRunnerFunc,
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
			name: "DirectReleaseMode",
			spec: spec.Spec{
				Project: spec.Project{
					Release: spec.Release{
						Mode: spec.ReleaseModeDirect,
					},
				},
			},
			patchFlag: true,
			gitRevBranch: func(context.Context, ...string) (int, string, error) {
				return 0, "main", nil
			},
			gitStatus: successRunnerFunc,
			gitPull:   successRunnerFunc,
			repo: &MockRepoService{
				GetMocks: []GetMock{
					{OutRepository: &repo, OutResponse: &github.Response{}},
				},
			},
			users: &MockUserService{
				UserMocks: []UserMock{
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
			name: "IndirectReleaseMode",
			spec: spec.Spec{
				Project: spec.Project{
					Release: spec.Release{
						Mode: spec.ReleaseModeIndirect,
					},
				},
			},
			minorFlag: true,
			gitRevBranch: func(context.Context, ...string) (int, string, error) {
				return 0, "main", nil
			},
			gitStatus: successRunnerFunc,
			gitPull:   successRunnerFunc,
			repo: &MockRepoService{
				GetMocks: []GetMock{
					{OutRepository: &repo, OutResponse: &github.Response{}},
				},
			},
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
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
			name: "InvalidReleaseMode",
			spec: spec.Spec{
				Project: spec.Project{
					Release: spec.Release{
						Mode: spec.ReleaseMode(""),
					},
				},
			},
			majorFlag: true,
			gitRevBranch: func(context.Context, ...string) (int, string, error) {
				return 0, "main", nil
			},
			gitStatus: successRunnerFunc,
			gitPull:   successRunnerFunc,
			repo: &MockRepoService{
				GetMocks: []GetMock{
					{OutRepository: &repo, OutResponse: &github.Response{}},
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Command{
				ui:   cli.NewMockUi(),
				spec: tc.spec,
			}

			c.flags.patch = tc.patchFlag
			c.flags.minor = tc.minorFlag
			c.flags.major = tc.majorFlag

			c.data.owner = "octocat"
			c.data.repo = "Hello-World"

			c.funcs.gitRevBranch = tc.gitRevBranch
			c.funcs.gitStatus = tc.gitStatus
			c.funcs.gitPull = tc.gitPull
			c.services.repo = tc.repo
			c.services.users = tc.users
			c.services.search = tc.search
			c.commands.semver = tc.semver

			exitCode := c.exec()

			assert.Equal(t, tc.expectedExitCode, exitCode)
		})
	}
}

func TestCommand_directRelease(t *testing.T) {
	tests := []struct {
		name             string
		commentFlag      string
		gitAdd           shell.RunnerFunc
		gitCommit        shell.RunnerFunc
		gitTag           shell.RunnerFunc
		goList           shell.RunnerFunc
		gitPush          shell.RunnerFunc
		gitPushTag       shell.RunnerFunc
		users            *MockUserService
		repo             *MockRepoService
		releases         *MockReleaseService
		changelog        *MockChangelogService
		build            *MockBuildCommand
		version          semver.SemVer
		ctx              context.Context
		defaultBranch    string
		expectedExitCode int
	}{
		{
			name: "UsersUserFails",
			users: &MockUserService{
				UserMocks: []UserMock{
					{OutError: errors.New("github error")},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitHubError,
		},
		{
			name: "RepoPermissionFails",
			users: &MockUserService{
				UserMocks: []UserMock{
					{OutUser: &user, OutResponse: &github.Response{}},
				},
			},
			repo: &MockRepoService{
				PermissionMocks: []PermissionMock{
					{OutError: errors.New("github error")},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitHubError,
		},
		{
			name: "InvalidUserPermission",
			users: &MockUserService{
				UserMocks: []UserMock{
					{OutUser: &user, OutResponse: &github.Response{}},
				},
			},
			repo: &MockRepoService{
				PermissionMocks: []PermissionMock{
					{OutPermission: github.PermissionWrite, OutResponse: &github.Response{}},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitHubError,
		},
		{
			name: "CreateReleaseFails",
			users: &MockUserService{
				UserMocks: []UserMock{
					{OutUser: &user, OutResponse: &github.Response{}},
				},
			},
			repo: &MockRepoService{
				PermissionMocks: []PermissionMock{
					{OutPermission: github.PermissionAdmin, OutResponse: &github.Response{}},
				},
			},
			releases: &MockReleaseService{
				CreateMocks: []ReleaseCreateMock{
					{OutError: errors.New("github error")},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitHubError,
		},
		{
			name: "ChangelogGenerateFails",
			users: &MockUserService{
				UserMocks: []UserMock{
					{OutUser: &user, OutResponse: &github.Response{}},
				},
			},
			repo: &MockRepoService{
				PermissionMocks: []PermissionMock{
					{OutPermission: github.PermissionAdmin, OutResponse: &github.Response{}},
				},
			},
			releases: &MockReleaseService{
				CreateMocks: []ReleaseCreateMock{
					{OutRelease: &draftRelease, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutError: errors.New("changelog error")},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.ChangelogError,
		},
		{
			name: "GitAddFails",
			gitAdd: func(context.Context, ...string) (int, string, error) {
				return 1, "", errors.New("git error")
			},
			users: &MockUserService{
				UserMocks: []UserMock{
					{OutUser: &user, OutResponse: &github.Response{}},
				},
			},
			repo: &MockRepoService{
				PermissionMocks: []PermissionMock{
					{OutPermission: github.PermissionAdmin, OutResponse: &github.Response{}},
				},
			},
			releases: &MockReleaseService{
				CreateMocks: []ReleaseCreateMock{
					{OutRelease: &draftRelease, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitError,
		},
		{
			name:   "GitCommitFails",
			gitAdd: successRunnerFunc,
			gitCommit: func(context.Context, ...string) (int, string, error) {
				return 1, "", errors.New("git error")
			},
			users: &MockUserService{
				UserMocks: []UserMock{
					{OutUser: &user, OutResponse: &github.Response{}},
				},
			},
			repo: &MockRepoService{
				PermissionMocks: []PermissionMock{
					{OutPermission: github.PermissionAdmin, OutResponse: &github.Response{}},
				},
			},
			releases: &MockReleaseService{
				CreateMocks: []ReleaseCreateMock{
					{OutRelease: &draftRelease, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitError,
		},
		{
			name:      "GitTagFails",
			gitAdd:    successRunnerFunc,
			gitCommit: successRunnerFunc,
			gitTag: func(context.Context, ...string) (int, string, error) {
				return 1, "", errors.New("git error")
			},
			users: &MockUserService{
				UserMocks: []UserMock{
					{OutUser: &user, OutResponse: &github.Response{}},
				},
			},
			repo: &MockRepoService{
				PermissionMocks: []PermissionMock{
					{OutPermission: github.PermissionAdmin, OutResponse: &github.Response{}},
				},
			},
			releases: &MockReleaseService{
				CreateMocks: []ReleaseCreateMock{
					{OutRelease: &draftRelease, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitError,
		},
		{
			name:      "BuildRunFails",
			gitAdd:    successRunnerFunc,
			gitCommit: successRunnerFunc,
			gitTag:    successRunnerFunc,
			goList:    successRunnerFunc,
			users: &MockUserService{
				UserMocks: []UserMock{
					{OutUser: &user, OutResponse: &github.Response{}},
				},
			},
			repo: &MockRepoService{
				PermissionMocks: []PermissionMock{
					{OutPermission: github.PermissionAdmin, OutResponse: &github.Response{}},
				},
			},
			releases: &MockReleaseService{
				CreateMocks: []ReleaseCreateMock{
					{OutRelease: &draftRelease, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
				},
			},
			build: &MockBuildCommand{
				RunMocks: []BuildRunMock{
					{OutCode: command.GoError},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GoError,
		},
		{
			name:      "UploadReleaseAssetFails",
			gitAdd:    successRunnerFunc,
			gitCommit: successRunnerFunc,
			gitTag:    successRunnerFunc,
			goList:    successRunnerFunc,
			users: &MockUserService{
				UserMocks: []UserMock{
					{OutUser: &user, OutResponse: &github.Response{}},
				},
			},
			repo: &MockRepoService{
				PermissionMocks: []PermissionMock{
					{OutPermission: github.PermissionAdmin, OutResponse: &github.Response{}},
				},
			},
			releases: &MockReleaseService{
				CreateMocks: []ReleaseCreateMock{
					{OutRelease: &draftRelease, OutResponse: &github.Response{}},
				},
				UploadAssetMocks: []ReleaseUploadAssetMock{
					{OutError: errors.New("github error")},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
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
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitHubError,
		},
		{
			name:      "BranchProtectionFails",
			gitAdd:    successRunnerFunc,
			gitCommit: successRunnerFunc,
			gitTag:    successRunnerFunc,
			goList:    successRunnerFunc,
			users: &MockUserService{
				UserMocks: []UserMock{
					{OutUser: &user, OutResponse: &github.Response{}},
				},
			},
			repo: &MockRepoService{
				PermissionMocks: []PermissionMock{
					{OutPermission: github.PermissionAdmin, OutResponse: &github.Response{}},
				},
				BranchProtectionMocks: []BranchProtectionMock{
					{OutError: errors.New("github error")},
				},
			},
			releases: &MockReleaseService{
				CreateMocks: []ReleaseCreateMock{
					{OutRelease: &draftRelease, OutResponse: &github.Response{}},
				},
				UploadAssetMocks: []ReleaseUploadAssetMock{
					{OutReleaseAsset: &asset, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
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
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitHubError,
		},
		{
			name:      "GitPushFails",
			gitAdd:    successRunnerFunc,
			gitCommit: successRunnerFunc,
			gitTag:    successRunnerFunc,
			goList:    successRunnerFunc,
			gitPush: func(context.Context, ...string) (int, string, error) {
				return 1, "", errors.New("git error")
			},
			users: &MockUserService{
				UserMocks: []UserMock{
					{OutUser: &user, OutResponse: &github.Response{}},
				},
			},
			repo: &MockRepoService{
				PermissionMocks: []PermissionMock{
					{OutPermission: github.PermissionAdmin, OutResponse: &github.Response{}},
				},
				BranchProtectionMocks: []BranchProtectionMock{
					{OutResponse: &github.Response{}},
					{OutResponse: &github.Response{}},
				},
			},
			releases: &MockReleaseService{
				CreateMocks: []ReleaseCreateMock{
					{OutRelease: &draftRelease, OutResponse: &github.Response{}},
				},
				UploadAssetMocks: []ReleaseUploadAssetMock{
					{OutReleaseAsset: &asset, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
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
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitError,
		},
		{
			name:      "GitPushTagFails",
			gitAdd:    successRunnerFunc,
			gitCommit: successRunnerFunc,
			gitTag:    successRunnerFunc,
			goList:    successRunnerFunc,
			gitPush:   successRunnerFunc,
			gitPushTag: func(context.Context, ...string) (int, string, error) {
				return 1, "", errors.New("git error")
			},
			users: &MockUserService{
				UserMocks: []UserMock{
					{OutUser: &user, OutResponse: &github.Response{}},
				},
			},
			repo: &MockRepoService{
				PermissionMocks: []PermissionMock{
					{OutPermission: github.PermissionAdmin, OutResponse: &github.Response{}},
				},
				BranchProtectionMocks: []BranchProtectionMock{
					{OutResponse: &github.Response{}},
					{OutResponse: &github.Response{}},
				},
			},
			releases: &MockReleaseService{
				CreateMocks: []ReleaseCreateMock{
					{OutRelease: &draftRelease, OutResponse: &github.Response{}},
				},
				UploadAssetMocks: []ReleaseUploadAssetMock{
					{OutReleaseAsset: &asset, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
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
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitError,
		},
		{
			name:       "UpdateReleaseFails",
			gitAdd:     successRunnerFunc,
			gitCommit:  successRunnerFunc,
			gitTag:     successRunnerFunc,
			goList:     successRunnerFunc,
			gitPush:    successRunnerFunc,
			gitPushTag: successRunnerFunc,
			users: &MockUserService{
				UserMocks: []UserMock{
					{OutUser: &user, OutResponse: &github.Response{}},
				},
			},
			repo: &MockRepoService{
				PermissionMocks: []PermissionMock{
					{OutPermission: github.PermissionAdmin, OutResponse: &github.Response{}},
				},
				BranchProtectionMocks: []BranchProtectionMock{
					{OutResponse: &github.Response{}},
					{OutResponse: &github.Response{}},
				},
			},
			releases: &MockReleaseService{
				CreateMocks: []ReleaseCreateMock{
					{OutRelease: &draftRelease, OutResponse: &github.Response{}},
				},
				UploadAssetMocks: []ReleaseUploadAssetMock{
					{OutReleaseAsset: &asset, OutResponse: &github.Response{}},
				},
				UpdateMocks: []ReleaseUpdateMock{
					{OutError: errors.New("github error")},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
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
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitHubError,
		},
		{
			name:        "Success",
			commentFlag: "description",
			gitAdd:      successRunnerFunc,
			gitCommit:   successRunnerFunc,
			gitTag:      successRunnerFunc,
			goList:      successRunnerFunc,
			gitPush:     successRunnerFunc,
			gitPushTag:  successRunnerFunc,
			users: &MockUserService{
				UserMocks: []UserMock{
					{OutUser: &user, OutResponse: &github.Response{}},
				},
			},
			repo: &MockRepoService{
				PermissionMocks: []PermissionMock{
					{OutPermission: github.PermissionAdmin, OutResponse: &github.Response{}},
				},
				BranchProtectionMocks: []BranchProtectionMock{
					{OutResponse: &github.Response{}},
					{OutResponse: &github.Response{}},
				},
			},
			releases: &MockReleaseService{
				CreateMocks: []ReleaseCreateMock{
					{OutRelease: &draftRelease, OutResponse: &github.Response{}},
				},
				UploadAssetMocks: []ReleaseUploadAssetMock{
					{OutReleaseAsset: &asset, OutResponse: &github.Response{}},
				},
				UpdateMocks: []ReleaseUpdateMock{
					{OutRelease: &release, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
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
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.Success,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Command{
				ui: cli.NewMockUi(),
			}

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
			c.funcs.goList = tc.goList
			c.funcs.gitPush = tc.gitPush
			c.funcs.gitPushTag = tc.gitPushTag
			c.services.users = tc.users
			c.services.repo = tc.repo
			c.services.releases = tc.releases
			c.services.changelog = tc.changelog
			c.commands.build = tc.build

			c.outputs.version = tc.version

			exitCode := c.directRelease(tc.ctx, tc.defaultBranch)

			assert.Equal(t, tc.expectedExitCode, exitCode)
		})
	}
}

func TestCommand_indirectRelease(t *testing.T) {
	tests := []struct {
		name             string
		commentFlag      string
		gitPull          shell.RunnerFunc
		gitTag           shell.RunnerFunc
		gitPushTag       shell.RunnerFunc
		goList           shell.RunnerFunc
		gitCheckout      shell.RunnerFunc
		gitAdd           shell.RunnerFunc
		gitCommit        shell.RunnerFunc
		gitPushBranch    shell.RunnerFunc
		gitBranch        shell.RunnerFunc
		search           *MockSearchService
		pulls            *MockPullService
		releases         *MockReleaseService
		changelog        *MockChangelogService
		build            *MockBuildCommand
		version          semver.SemVer
		ctx              context.Context
		defaultBranch    string
		expectedExitCode int
	}{
		{
			name: "SearchForMergedPullRequestFails",
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutError: errors.New("github error")},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitHubError,
		},
		{
			name: "FinishRelease_GetPullFails",
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: mergedSearchResult, OutResponse: &github.Response{}},
				},
			},
			pulls: &MockPullService{
				GetMocks: []PullGetMock{
					{OutError: errors.New("github error")},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitHubError,
		},
		{
			name: "FinishRelease_GetReleaseFails",
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: mergedSearchResult, OutResponse: &github.Response{}},
				},
			},
			pulls: &MockPullService{
				GetMocks: []PullGetMock{
					{OutPull: mergedPull, OutResponse: &github.Response{}},
				},
			},
			releases: &MockReleaseService{
				ListMocks: []ReleaseListMock{
					{OutError: errors.New("github error")},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitHubError,
		},
		{
			name: "FinishRelease_GitPullFails",
			gitPull: func(context.Context, ...string) (int, string, error) {
				return 1, "", errors.New("git error")
			},
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: mergedSearchResult, OutResponse: &github.Response{}},
				},
			},
			pulls: &MockPullService{
				GetMocks: []PullGetMock{
					{OutPull: mergedPull, OutResponse: &github.Response{}},
				},
			},
			releases: &MockReleaseService{
				ListMocks: []ReleaseListMock{
					{OutReleases: []github.Release{draftRelease}, OutResponse: &github.Response{}},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitError,
		},
		{
			name:    "FinishRelease_GitTagFails",
			gitPull: successRunnerFunc,
			gitTag: func(context.Context, ...string) (int, string, error) {
				return 1, "", errors.New("git error")
			},
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: mergedSearchResult, OutResponse: &github.Response{}},
				},
			},
			pulls: &MockPullService{
				GetMocks: []PullGetMock{
					{OutPull: mergedPull, OutResponse: &github.Response{}},
				},
			},
			releases: &MockReleaseService{
				ListMocks: []ReleaseListMock{
					{OutReleases: []github.Release{draftRelease}, OutResponse: &github.Response{}},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitError,
		},
		{
			name:    "FinishRelease_GitPushTagFails",
			gitPull: successRunnerFunc,
			gitTag:  successRunnerFunc,
			gitPushTag: func(context.Context, ...string) (int, string, error) {
				return 1, "", errors.New("git error")
			},
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: mergedSearchResult, OutResponse: &github.Response{}},
				},
			},
			pulls: &MockPullService{
				GetMocks: []PullGetMock{
					{OutPull: mergedPull, OutResponse: &github.Response{}},
				},
			},
			releases: &MockReleaseService{
				ListMocks: []ReleaseListMock{
					{OutReleases: []github.Release{draftRelease}, OutResponse: &github.Response{}},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitError,
		},
		{
			name:       "FinishRelease_BuildRunFails",
			gitPull:    successRunnerFunc,
			gitTag:     successRunnerFunc,
			gitPushTag: successRunnerFunc,
			goList:     successRunnerFunc,
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: mergedSearchResult, OutResponse: &github.Response{}},
				},
			},
			pulls: &MockPullService{
				GetMocks: []PullGetMock{
					{OutPull: mergedPull, OutResponse: &github.Response{}},
				},
			},
			releases: &MockReleaseService{
				ListMocks: []ReleaseListMock{
					{OutReleases: []github.Release{draftRelease}, OutResponse: &github.Response{}},
				},
			},
			build: &MockBuildCommand{
				RunMocks: []BuildRunMock{
					{OutCode: command.GoError},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GoError,
		},
		{
			name:       "FinishRelease_UploadReleaseAssetFails",
			gitPull:    successRunnerFunc,
			gitTag:     successRunnerFunc,
			gitPushTag: successRunnerFunc,
			goList:     successRunnerFunc,
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: mergedSearchResult, OutResponse: &github.Response{}},
				},
			},
			pulls: &MockPullService{
				GetMocks: []PullGetMock{
					{OutPull: mergedPull, OutResponse: &github.Response{}},
				},
			},
			releases: &MockReleaseService{
				ListMocks: []ReleaseListMock{
					{OutReleases: []github.Release{draftRelease}, OutResponse: &github.Response{}},
				},
				UploadAssetMocks: []ReleaseUploadAssetMock{
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
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitHubError,
		},
		{
			name:       "FinishRelease_UpdateReleaseFails",
			gitPull:    successRunnerFunc,
			gitTag:     successRunnerFunc,
			gitPushTag: successRunnerFunc,
			goList:     successRunnerFunc,
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: mergedSearchResult, OutResponse: &github.Response{}},
				},
			},
			pulls: &MockPullService{
				GetMocks: []PullGetMock{
					{OutPull: mergedPull, OutResponse: &github.Response{}},
				},
			},
			releases: &MockReleaseService{
				ListMocks: []ReleaseListMock{
					{OutReleases: []github.Release{draftRelease}, OutResponse: &github.Response{}},
				},
				UploadAssetMocks: []ReleaseUploadAssetMock{
					{OutReleaseAsset: &asset, OutResponse: &github.Response{}},
				},
				UpdateMocks: []ReleaseUpdateMock{
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
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitHubError,
		},
		{
			name:       "FinishRelease_Success",
			gitPull:    successRunnerFunc,
			gitTag:     successRunnerFunc,
			gitPushTag: successRunnerFunc,
			goList:     successRunnerFunc,
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: mergedSearchResult, OutResponse: &github.Response{}},
				},
			},
			pulls: &MockPullService{
				GetMocks: []PullGetMock{
					{OutPull: mergedPull, OutResponse: &github.Response{}},
				},
			},
			releases: &MockReleaseService{
				ListMocks: []ReleaseListMock{
					{OutReleases: []github.Release{draftRelease}, OutResponse: &github.Response{}},
				},
				UploadAssetMocks: []ReleaseUploadAssetMock{
					{OutReleaseAsset: &asset, OutResponse: &github.Response{}},
				},
				UpdateMocks: []ReleaseUpdateMock{
					{OutRelease: &release, OutResponse: &github.Response{}},
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
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.Success,
		},
		{
			name: "SearchForOpenPullRequestFails",
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
					{OutError: errors.New("github error")},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitHubError,
		},
		{
			name: "CreatePullRequest_ChangelogGenerateFails",
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutError: errors.New("changelog error")},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.ChangelogError,
		},
		{
			name: "CreatePullRequest_GitCheckoutReleaseBranchFails",
			gitCheckout: func(context.Context, ...string) (int, string, error) {
				return 1, "", errors.New("git error")
			},
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitError,
		},
		{
			name:        "CreatePullRequest_GitAddFails",
			gitCheckout: successRunnerFunc,
			gitAdd: func(context.Context, ...string) (int, string, error) {
				return 1, "", errors.New("git error")
			},
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitError,
		},
		{
			name:        "CreatePullRequest_GitCommitFails",
			gitCheckout: successRunnerFunc,
			gitAdd:      successRunnerFunc,
			gitCommit: func(context.Context, ...string) (int, string, error) {
				return 1, "", errors.New("git error")
			},
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitError,
		},
		{
			name:        "CreatePullRequest_GitPushBranchFails",
			gitCheckout: successRunnerFunc,
			gitAdd:      successRunnerFunc,
			gitCommit:   successRunnerFunc,
			gitPushBranch: func(context.Context, ...string) (int, string, error) {
				return 1, "", errors.New("git error")
			},
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitError,
		},
		{
			name: "CreatePullRequest_GitCheckoutDefaultBranchFails",
			gitCheckout: func(ctx context.Context, args ...string) (int, string, error) {
				if args[0] == "-b" {
					return 0, "", nil
				}
				return 1, "", errors.New("git error")
			},
			gitAdd:        successRunnerFunc,
			gitCommit:     successRunnerFunc,
			gitPushBranch: successRunnerFunc,
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitError,
		},
		{
			name:          "CreatePullRequest_GitDeleteBranchFails",
			gitCheckout:   successRunnerFunc,
			gitAdd:        successRunnerFunc,
			gitCommit:     successRunnerFunc,
			gitPushBranch: successRunnerFunc,
			gitBranch: func(context.Context, ...string) (int, string, error) {
				return 1, "", errors.New("git error")
			},
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitError,
		},
		{
			name:          "CreatePullRequest_CreatePullFails",
			gitCheckout:   successRunnerFunc,
			gitAdd:        successRunnerFunc,
			gitCommit:     successRunnerFunc,
			gitPushBranch: successRunnerFunc,
			gitBranch:     successRunnerFunc,
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
				},
			},
			pulls: &MockPullService{
				CreateMocks: []PullCreateMock{
					{OutError: errors.New("github error")},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitHubError,
		},
		{
			name:          "CreatePullRequest_CreateDraftReleaseFails",
			gitCheckout:   successRunnerFunc,
			gitAdd:        successRunnerFunc,
			gitCommit:     successRunnerFunc,
			gitPushBranch: successRunnerFunc,
			gitBranch:     successRunnerFunc,
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
				},
			},
			pulls: &MockPullService{
				CreateMocks: []PullCreateMock{
					{OutPull: openPull, OutResponse: &github.Response{}},
				},
			},
			releases: &MockReleaseService{
				CreateMocks: []ReleaseCreateMock{
					{OutError: errors.New("github error")},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitHubError,
		},
		{
			name:          "CreatePullRequest_Success",
			commentFlag:   "description",
			gitCheckout:   successRunnerFunc,
			gitAdd:        successRunnerFunc,
			gitCommit:     successRunnerFunc,
			gitPushBranch: successRunnerFunc,
			gitBranch:     successRunnerFunc,
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
				},
			},
			pulls: &MockPullService{
				CreateMocks: []PullCreateMock{
					{OutPull: openPull, OutResponse: &github.Response{}},
				},
			},
			releases: &MockReleaseService{
				CreateMocks: []ReleaseCreateMock{
					{OutRelease: &draftRelease, OutResponse: &github.Response{}},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.Success,
		},
		{
			name: "UpdatePullRequest_ChangelogGenerateFails",
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
					{OutResult: openSearchResult, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutError: errors.New("changelog error")},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.ChangelogError,
		},
		{
			name: "UpdatePullRequest_GitCheckoutReleaseBranchFails",
			gitCheckout: func(context.Context, ...string) (int, string, error) {
				return 1, "", errors.New("git error")
			},
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
					{OutResult: openSearchResult, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitError,
		},
		{
			name:        "UpdatePullRequest_GitAddFails",
			gitCheckout: successRunnerFunc,
			gitAdd: func(context.Context, ...string) (int, string, error) {
				return 1, "", errors.New("git error")
			},
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
					{OutResult: openSearchResult, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitError,
		},
		{
			name:        "UpdatePullRequest_GitCommitFails",
			gitCheckout: successRunnerFunc,
			gitAdd:      successRunnerFunc,
			gitCommit: func(context.Context, ...string) (int, string, error) {
				return 1, "", errors.New("git error")
			},
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
					{OutResult: openSearchResult, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitError,
		},
		{
			name:        "UpdatePullRequest_GitPushBranchFails",
			gitCheckout: successRunnerFunc,
			gitAdd:      successRunnerFunc,
			gitCommit:   successRunnerFunc,
			gitPushBranch: func(context.Context, ...string) (int, string, error) {
				return 1, "", errors.New("git error")
			},
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
					{OutResult: openSearchResult, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitError,
		},
		{
			name: "UpdatePullRequest_GitCheckoutDefaultBranchFails",
			gitCheckout: func(ctx context.Context, args ...string) (int, string, error) {
				if args[0] == "-b" {
					return 0, "", nil
				}
				return 1, "", errors.New("git error")
			},
			gitAdd:        successRunnerFunc,
			gitCommit:     successRunnerFunc,
			gitPushBranch: successRunnerFunc,
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
					{OutResult: openSearchResult, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitError,
		},
		{
			name:          "UpdatePullRequest_GitDeleteBranchFails",
			gitCheckout:   successRunnerFunc,
			gitAdd:        successRunnerFunc,
			gitCommit:     successRunnerFunc,
			gitPushBranch: successRunnerFunc,
			gitBranch: func(context.Context, ...string) (int, string, error) {
				return 1, "", errors.New("git error")
			},
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
					{OutResult: openSearchResult, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitError,
		},
		{
			name:          "UpdatePullRequest_UpdatePullFails",
			gitCheckout:   successRunnerFunc,
			gitAdd:        successRunnerFunc,
			gitCommit:     successRunnerFunc,
			gitPushBranch: successRunnerFunc,
			gitBranch:     successRunnerFunc,
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
					{OutResult: openSearchResult, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
				},
			},
			pulls: &MockPullService{
				UpdateMocks: []PullUpdateMock{
					{OutError: errors.New("github error")},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitHubError,
		},
		{
			name:          "UpdatePullRequest_GetReleaseFails",
			gitCheckout:   successRunnerFunc,
			gitAdd:        successRunnerFunc,
			gitCommit:     successRunnerFunc,
			gitPushBranch: successRunnerFunc,
			gitBranch:     successRunnerFunc,
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
					{OutResult: openSearchResult, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
				},
			},
			pulls: &MockPullService{
				UpdateMocks: []PullUpdateMock{
					{OutPull: openPull, OutResponse: &github.Response{}},
				},
			},
			releases: &MockReleaseService{
				ListMocks: []ReleaseListMock{
					{OutError: errors.New("github error")},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitHubError,
		},
		{
			name:          "UpdatePullRequest_UpdateReleaseFails",
			gitCheckout:   successRunnerFunc,
			gitAdd:        successRunnerFunc,
			gitCommit:     successRunnerFunc,
			gitPushBranch: successRunnerFunc,
			gitBranch:     successRunnerFunc,
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
					{OutResult: openSearchResult, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
				},
			},
			pulls: &MockPullService{
				UpdateMocks: []PullUpdateMock{
					{OutPull: openPull, OutResponse: &github.Response{}},
				},
			},
			releases: &MockReleaseService{
				ListMocks: []ReleaseListMock{
					{OutReleases: []github.Release{draftRelease}, OutResponse: &github.Response{}},
				},
				UpdateMocks: []ReleaseUpdateMock{
					{OutError: errors.New("github error")},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.GitHubError,
		},
		{
			name:          "UpdatePullRequest_Success",
			commentFlag:   "description",
			gitCheckout:   successRunnerFunc,
			gitAdd:        successRunnerFunc,
			gitCommit:     successRunnerFunc,
			gitPushBranch: successRunnerFunc,
			gitBranch:     successRunnerFunc,
			search: &MockSearchService{
				SearchIssuesMocks: []SearchIssuesMock{
					{OutResult: emptySearchResult, OutResponse: &github.Response{}},
					{OutResult: openSearchResult, OutResponse: &github.Response{}},
				},
			},
			changelog: &MockChangelogService{
				GenerateMocks: []GenerateMock{
					{OutContent: "changelog content"},
				},
			},
			pulls: &MockPullService{
				UpdateMocks: []PullUpdateMock{
					{OutPull: openPull, OutResponse: &github.Response{}},
				},
			},
			releases: &MockReleaseService{
				ListMocks: []ReleaseListMock{
					{OutReleases: []github.Release{draftRelease}, OutResponse: &github.Response{}},
				},
				UpdateMocks: []ReleaseUpdateMock{
					{OutRelease: &draftRelease, OutResponse: &github.Response{}},
				},
			},
			version:          version,
			ctx:              context.Background(),
			defaultBranch:    "main",
			expectedExitCode: command.Success,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Command{
				ui: cli.NewMockUi(),
			}

			c.flags.comment = tc.commentFlag

			c.data.owner = "octocat"
			c.data.repo = "Hello-World"
			c.data.changelogSpec = changelogspec.Spec{
				General: changelogspec.General{
					File: "CHANGELOG.md",
				},
			}

			c.funcs.gitPull = tc.gitPull
			c.funcs.gitTag = tc.gitTag
			c.funcs.gitPushTag = tc.gitPushTag
			c.funcs.goList = tc.goList
			c.funcs.gitCheckout = tc.gitCheckout
			c.funcs.gitAdd = tc.gitAdd
			c.funcs.gitCommit = tc.gitCommit
			c.funcs.gitPushBranch = tc.gitPushBranch
			c.funcs.gitBranch = tc.gitBranch
			c.services.search = tc.search
			c.services.pulls = tc.pulls
			c.services.releases = tc.releases
			c.services.changelog = tc.changelog
			c.commands.build = tc.build

			c.outputs.version = tc.version

			exitCode := c.indirectRelease(tc.ctx, tc.defaultBranch)

			assert.Equal(t, tc.expectedExitCode, exitCode)
		})
	}
}

func TestCommand_findDraftRelease(t *testing.T) {
	tests := []struct {
		name             string
		releases         *MockReleaseService
		ctx              context.Context
		tag              string
		expectedRelease  *github.Release
		expectedExitCode int
	}{
		{
			name: "FirstListReleasesFails",
			releases: &MockReleaseService{
				ListMocks: []ReleaseListMock{
					{OutError: errors.New("github error")},
				},
			},
			ctx:              context.Background(),
			tag:              "v0.1.0",
			expectedExitCode: command.GitHubError,
		},
		{
			name: "ReleaseFoundInFirstPage",
			releases: &MockReleaseService{
				ListMocks: []ReleaseListMock{
					{
						OutReleases: []github.Release{draftRelease},
						OutResponse: &github.Response{},
					},
				},
			},
			ctx:              context.Background(),
			tag:              "v0.1.0",
			expectedRelease:  &draftRelease,
			expectedExitCode: command.Success,
		},
		{
			name: "SecondListReleasesFails",
			releases: &MockReleaseService{
				ListMocks: []ReleaseListMock{
					{
						OutReleases: []github.Release{},
						OutResponse: &github.Response{
							Pages: github.Pages{
								Next: 2,
								Last: 2,
							},
						},
					},
					{OutError: errors.New("github error")},
				},
			},
			ctx:              context.Background(),
			tag:              "v0.1.0",
			expectedExitCode: command.GitHubError,
		},
		{
			name: "ReleaseFoundInSecondPage",
			releases: &MockReleaseService{
				ListMocks: []ReleaseListMock{
					{
						OutReleases: []github.Release{},
						OutResponse: &github.Response{
							Pages: github.Pages{
								Next: 2,
								Last: 2,
							},
						},
					},
					{
						OutReleases: []github.Release{draftRelease},
						OutResponse: &github.Response{},
					},
				},
			},
			ctx:              context.Background(),
			tag:              "v0.1.0",
			expectedRelease:  &draftRelease,
			expectedExitCode: command.Success,
		},
		{
			name: "ReleaseNotFound",
			releases: &MockReleaseService{
				ListMocks: []ReleaseListMock{
					{
						OutReleases: []github.Release{},
						OutResponse: &github.Response{
							Pages: github.Pages{
								Next: 2,
								Last: 2,
							},
						},
					},
					{
						OutReleases: []github.Release{},
						OutResponse: &github.Response{},
					},
				},
			},
			ctx:              context.Background(),
			tag:              "v0.1.0",
			expectedExitCode: command.GitHubError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Command{
				ui: cli.NewMockUi(),
			}

			c.services.releases = tc.releases

			release, exitCode := c.findDraftRelease(tc.ctx, tc.tag)

			assert.Equal(t, tc.expectedRelease, release)
			assert.Equal(t, tc.expectedExitCode, exitCode)
		})
	}
}
