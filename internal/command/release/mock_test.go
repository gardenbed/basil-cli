package release

import (
	"context"

	"github.com/ProtonMail/go-crypto/openpgp"
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

	CreateCommitMock struct {
		InMessage string
		InSignKey *openpgp.Entity
		InPaths   []string
		OutHash   string
		OutError  error
	}

	CreateTagMock struct {
		InCommit  string
		InName    string
		InMessage string
		InSignKey *openpgp.Entity
		OutHash   string
		OutError  error
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

	MockGitService struct {
		RemoteIndex int
		RemoteMocks []RemoteMock

		HEADIndex int
		HEADMocks []HEADMock

		IsCleanIndex int
		IsCleanMocks []IsCleanMock

		CreateCommitIndex int
		CreateCommitMocks []CreateCommitMock

		CreateTagIndex int
		CreateTagMocks []CreateTagMock

		PullIndex int
		PullMocks []PullMock

		PushIndex int
		PushMocks []PushMock

		PushTagIndex int
		PushTagMocks []PushTagMock
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

func (m *MockGitService) CreateCommit(message string, signKey *openpgp.Entity, paths ...string) (string, error) {
	i := m.CreateCommitIndex
	m.CreateCommitIndex++
	m.CreateCommitMocks[i].InMessage = message
	m.CreateCommitMocks[i].InSignKey = signKey
	m.CreateCommitMocks[i].InPaths = paths
	return m.CreateCommitMocks[i].OutHash, m.CreateCommitMocks[i].OutError
}

func (m *MockGitService) CreateTag(commit, name, message string, signKey *openpgp.Entity) (string, error) {
	i := m.CreateTagIndex
	m.CreateTagIndex++
	m.CreateTagMocks[i].InCommit = commit
	m.CreateTagMocks[i].InName = name
	m.CreateTagMocks[i].InMessage = message
	m.CreateTagMocks[i].InSignKey = signKey
	return m.CreateTagMocks[i].OutHash, m.CreateTagMocks[i].OutError
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
	m.PushMocks[i].InRemoteName = remoteName
	return m.PushMocks[i].OutError
}

func (m *MockGitService) PushTag(ctx context.Context, remoteName, tagName string) error {
	i := m.PushTagIndex
	m.PushTagIndex++
	m.PushTagMocks[i].InRemoteName = remoteName
	m.PushTagMocks[i].InTagName = tagName
	return m.PushTagMocks[i].OutError
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

const mockGPGKey = `
-----BEGIN PGP PRIVATE KEY BLOCK-----

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
hJk7yKzcUwWyMTmo9F7CwEkTXnj1KEk2oQ==
=LZXO
-----END PGP PRIVATE KEY BLOCK-----
`
