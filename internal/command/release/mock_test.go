package release

import (
	"context"

	changelogspec "github.com/gardenbed/changelog/spec"
	"github.com/gardenbed/go-github"

	buildcmd "github.com/gardenbed/basil-cli/internal/command/build"
	"github.com/gardenbed/basil-cli/internal/semver"
)

type (
	RemoteMock struct {
		InName    string
		OutDomain string
		OutPath   string
		OutError  error
	}

	HEADMock struct {
		OutHash   string
		OutBranch string
		OutError  error
	}

	IsCleanMock struct {
		OutBool  bool
		OutError error
	}

	ResetMock struct {
		InRev    string
		InHard   bool
		OutError error
	}

	CreateBranchMock struct {
		InName   string
		OutError error
	}

	DeleteBranchMock struct {
		InName   string
		OutError error
	}

	PullMock struct {
		InContext context.Context
		OutError  error
	}

	PushMock struct {
		InContext    context.Context
		InRemoteName string
		OutError     error
	}

	PushTagMock struct {
		InContext    context.Context
		InRemoteName string
		InTagName    string
		OutError     error
	}

	PushBranchMock struct {
		InContext    context.Context
		InRemoteName string
		InBranchName string
		OutError     error
	}

	MockGitService struct {
		RemoteIndex int
		RemoteMocks []RemoteMock

		HEADIndex int
		HEADMocks []HEADMock

		IsCleanIndex int
		IsCleanMocks []IsCleanMock

		ResetIndex int
		ResetMocks []ResetMock

		CreateBranchIndex int
		CreateBranchMocks []CreateBranchMock

		DeleteBranchIndex int
		DeleteBranchMocks []DeleteBranchMock

		PullIndex int
		PullMocks []PullMock

		PushIndex int
		PushMocks []PushMock

		PushTagIndex int
		PushTagMocks []PushTagMock

		PushBranchIndex int
		PushBranchMocks []PushBranchMock
	}
)

func (m *MockGitService) Remote(name string) (string, string, error) {
	i := m.RemoteIndex
	m.RemoteIndex++
	m.RemoteMocks[i].InName = name
	return m.RemoteMocks[i].OutDomain, m.RemoteMocks[i].OutPath, m.RemoteMocks[i].OutError
}

func (m *MockGitService) HEAD() (string, string, error) {
	i := m.HEADIndex
	m.HEADIndex++
	return m.HEADMocks[i].OutHash, m.HEADMocks[i].OutBranch, m.HEADMocks[i].OutError
}

func (m *MockGitService) IsClean() (bool, error) {
	i := m.IsCleanIndex
	m.IsCleanIndex++
	return m.IsCleanMocks[i].OutBool, m.IsCleanMocks[i].OutError
}

func (m *MockGitService) Reset(rev string, hard bool) error {
	i := m.ResetIndex
	m.ResetIndex++
	m.ResetMocks[i].InRev = rev
	m.ResetMocks[i].InHard = hard
	return m.ResetMocks[i].OutError
}

func (m *MockGitService) CreateBranch(name string) error {
	i := m.CreateBranchIndex
	m.CreateBranchIndex++
	m.CreateBranchMocks[i].InName = name
	return m.CreateBranchMocks[i].OutError
}

func (m *MockGitService) DeleteBranch(name string) error {
	i := m.DeleteBranchIndex
	m.DeleteBranchIndex++
	m.DeleteBranchMocks[i].InName = name
	return m.DeleteBranchMocks[i].OutError
}

func (m *MockGitService) Pull(ctx context.Context) error {
	i := m.PullIndex
	m.PullIndex++
	m.PullMocks[i].InContext = ctx
	return m.PullMocks[i].OutError
}

func (m *MockGitService) Push(ctx context.Context, remoteName string) error {
	i := m.PushIndex
	m.PushIndex++
	m.PushMocks[i].InContext = ctx
	m.PushMocks[i].InRemoteName = remoteName
	return m.PushMocks[i].OutError
}

func (m *MockGitService) PushTag(ctx context.Context, remoteName, tagName string) error {
	i := m.PushTagIndex
	m.PushTagIndex++
	m.PushTagMocks[i].InContext = ctx
	m.PushTagMocks[i].InRemoteName = remoteName
	m.PushTagMocks[i].InTagName = tagName
	return m.PushTagMocks[i].OutError
}

func (m *MockGitService) PushBranch(ctx context.Context, remoteName, branchName string) error {
	i := m.PushBranchIndex
	m.PushBranchIndex++
	m.PushBranchMocks[i].InContext = ctx
	m.PushBranchMocks[i].InRemoteName = remoteName
	m.PushBranchMocks[i].InBranchName = branchName
	return m.PushBranchMocks[i].OutError
}

type (
	UserMock struct {
		InContext   context.Context
		OutUser     *github.User
		OutResponse *github.Response
		OutError    error
	}

	MockUsersService struct {
		UserIndex int
		UserMocks []UserMock
	}
)

func (m *MockUsersService) User(ctx context.Context) (*github.User, *github.Response, error) {
	i := m.UserIndex
	m.UserIndex++
	m.UserMocks[i].InContext = ctx
	return m.UserMocks[i].OutUser, m.UserMocks[i].OutResponse, m.UserMocks[i].OutError
}

type (
	GetMock struct {
		InContext     context.Context
		OutRepository *github.Repository
		OutResponse   *github.Response
		OutError      error
	}

	PermissionMock struct {
		InContext     context.Context
		InUsername    string
		OutPermission github.Permission
		OutResponse   *github.Response
		OutError      error
	}

	BranchProtectionMock struct {
		InContext   context.Context
		InBranch    string
		InEnabled   bool
		OutResponse *github.Response
		OutError    error
	}

	CreateReleaseMock struct {
		InContext   context.Context
		InParams    github.ReleaseParams
		OutRelease  *github.Release
		OutResponse *github.Response
		OutError    error
	}

	UpdateReleaseMock struct {
		InContext   context.Context
		InReleaseID int
		InParams    github.ReleaseParams
		OutRelease  *github.Release
		OutResponse *github.Response
		OutError    error
	}

	UploadReleaseAssetMock struct {
		InContext       context.Context
		InReleaseID     int
		InAssetFile     string
		InAssetLabel    string
		OutReleaseAsset *github.ReleaseAsset
		OutResponse     *github.Response
		OutError        error
	}

	MockRepoService struct {
		GetIndex int
		GetMocks []GetMock

		PermissionIndex int
		PermissionMocks []PermissionMock

		BranchProtectionIndex int
		BranchProtectionMocks []BranchProtectionMock

		CreateReleaseIndex int
		CreateReleaseMocks []CreateReleaseMock

		UpdateReleaseIndex int
		UpdateReleaseMocks []UpdateReleaseMock

		UploadReleaseAssetIndex int
		UploadReleaseAssetMocks []UploadReleaseAssetMock
	}
)

func (m *MockRepoService) Get(ctx context.Context) (*github.Repository, *github.Response, error) {
	i := m.GetIndex
	m.GetIndex++
	m.GetMocks[i].InContext = ctx
	return m.GetMocks[i].OutRepository, m.GetMocks[i].OutResponse, m.GetMocks[i].OutError
}

func (m *MockRepoService) Permission(ctx context.Context, username string) (github.Permission, *github.Response, error) {
	i := m.PermissionIndex
	m.PermissionIndex++
	m.PermissionMocks[i].InContext = ctx
	m.PermissionMocks[i].InUsername = username
	return m.PermissionMocks[i].OutPermission, m.PermissionMocks[i].OutResponse, m.PermissionMocks[i].OutError
}

func (m *MockRepoService) BranchProtection(ctx context.Context, branch string, enabled bool) (*github.Response, error) {
	i := m.BranchProtectionIndex
	m.BranchProtectionIndex++
	m.BranchProtectionMocks[i].InContext = ctx
	m.BranchProtectionMocks[i].InBranch = branch
	m.BranchProtectionMocks[i].InEnabled = enabled
	return m.BranchProtectionMocks[i].OutResponse, m.BranchProtectionMocks[i].OutError
}

func (m *MockRepoService) CreateRelease(ctx context.Context, params github.ReleaseParams) (*github.Release, *github.Response, error) {
	i := m.CreateReleaseIndex
	m.CreateReleaseIndex++
	m.CreateReleaseMocks[i].InContext = ctx
	m.CreateReleaseMocks[i].InParams = params
	return m.CreateReleaseMocks[i].OutRelease, m.CreateReleaseMocks[i].OutResponse, m.CreateReleaseMocks[i].OutError
}

func (m *MockRepoService) UpdateRelease(ctx context.Context, releaseID int, params github.ReleaseParams) (*github.Release, *github.Response, error) {
	i := m.UpdateReleaseIndex
	m.UpdateReleaseIndex++
	m.UpdateReleaseMocks[i].InContext = ctx
	m.UpdateReleaseMocks[i].InReleaseID = releaseID
	m.UpdateReleaseMocks[i].InParams = params
	return m.UpdateReleaseMocks[i].OutRelease, m.UpdateReleaseMocks[i].OutResponse, m.UpdateReleaseMocks[i].OutError
}

func (m *MockRepoService) UploadReleaseAsset(ctx context.Context, releaseID int, assetFile, assetLabel string) (*github.ReleaseAsset, *github.Response, error) {
	i := m.UploadReleaseAssetIndex
	m.UploadReleaseAssetIndex++
	m.UploadReleaseAssetMocks[i].InContext = ctx
	m.UploadReleaseAssetMocks[i].InReleaseID = releaseID
	m.UploadReleaseAssetMocks[i].InAssetFile = assetFile
	m.UploadReleaseAssetMocks[i].InAssetLabel = assetLabel
	return m.UploadReleaseAssetMocks[i].OutReleaseAsset, m.UploadReleaseAssetMocks[i].OutResponse, m.UploadReleaseAssetMocks[i].OutError
}

type (
	PullsCreateMock struct {
		InContext   context.Context
		InParams    github.CreatePullParams
		OutPull     *github.Pull
		OutResponse *github.Response
		OutError    error
	}

	MockPullsService struct {
		CreateIndex int
		CreateMocks []PullsCreateMock
	}
)

func (m *MockPullsService) Create(ctx context.Context, params github.CreatePullParams) (*github.Pull, *github.Response, error) {
	i := m.CreateIndex
	m.CreateIndex++
	m.CreateMocks[i].InContext = ctx
	m.CreateMocks[i].InParams = params
	return m.CreateMocks[i].OutPull, m.CreateMocks[i].OutResponse, m.CreateMocks[i].OutError
}

type (
	GenerateMock struct {
		InContext  context.Context
		InSpec     changelogspec.Spec
		OutContent string
		OutError   error
	}

	MockChangelogService struct {
		GenerateIndex int
		GenerateMocks []GenerateMock
	}
)

func (m *MockChangelogService) Generate(ctx context.Context, spec changelogspec.Spec) (string, error) {
	i := m.GenerateIndex
	m.GenerateIndex++
	m.GenerateMocks[i].InContext = ctx
	m.GenerateMocks[i].InSpec = spec
	return m.GenerateMocks[i].OutContent, m.GenerateMocks[i].OutError
}

type (
	SemverRunMock struct {
		InArgs  []string
		OutCode int
	}

	SemVerMock struct {
		OutSemVer semver.SemVer
	}

	MockSemverCommand struct {
		RunIndex int
		RunMocks []SemverRunMock

		SemVerIndex int
		SemVerMocks []SemVerMock
	}
)

func (m *MockSemverCommand) Run(args []string) int {
	i := m.RunIndex
	m.RunIndex++
	m.RunMocks[i].InArgs = args
	return m.RunMocks[i].OutCode
}

func (m *MockSemverCommand) SemVer() semver.SemVer {
	i := m.SemVerIndex
	m.SemVerIndex++
	return m.SemVerMocks[i].OutSemVer
}

type (
	BuildRunMock struct {
		InArgs  []string
		OutCode int
	}

	ArtifactsMock struct {
		OutArtifacts []buildcmd.Artifact
	}

	MockBuildCommand struct {
		RunIndex int
		RunMocks []BuildRunMock

		ArtifactsIndex int
		ArtifactsMocks []ArtifactsMock
	}
)

func (m *MockBuildCommand) Run(args []string) int {
	i := m.RunIndex
	m.RunIndex++
	m.RunMocks[i].InArgs = args
	return m.RunMocks[i].OutCode
}

func (m *MockBuildCommand) Artifacts() []buildcmd.Artifact {
	i := m.ArtifactsIndex
	m.ArtifactsIndex++
	return m.ArtifactsMocks[i].OutArtifacts
}
