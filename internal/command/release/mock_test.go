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

	MockGitService struct {
		RemoteIndex int
		RemoteMocks []RemoteMock
	}
)

func (m *MockGitService) Remote(name string) (string, string, error) {
	i := m.RemoteIndex
	m.RemoteIndex++
	m.RemoteMocks[i].InName = name
	return m.RemoteMocks[i].OutDomain, m.RemoteMocks[i].OutPath, m.RemoteMocks[i].OutError
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

	MockRepoService struct {
		GetIndex int
		GetMocks []GetMock

		PermissionIndex int
		PermissionMocks []PermissionMock

		BranchProtectionIndex int
		BranchProtectionMocks []BranchProtectionMock
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

type (
	ReleaseListMock struct {
		InContext   context.Context
		InPageSize  int
		InPageNo    int
		OutReleases []github.Release
		OutResponse *github.Response
		OutError    error
	}

	ReleaseCreateMock struct {
		InContext   context.Context
		InParams    github.ReleaseParams
		OutRelease  *github.Release
		OutResponse *github.Response
		OutError    error
	}

	ReleaseUpdateMock struct {
		InContext   context.Context
		InReleaseID int
		InParams    github.ReleaseParams
		OutRelease  *github.Release
		OutResponse *github.Response
		OutError    error
	}

	ReleaseUploadAssetMock struct {
		InContext       context.Context
		InReleaseID     int
		InAssetFile     string
		InAssetLabel    string
		OutReleaseAsset *github.ReleaseAsset
		OutResponse     *github.Response
		OutError        error
	}

	MockReleaseService struct {
		ListIndex int
		ListMocks []ReleaseListMock

		CreateIndex int
		CreateMocks []ReleaseCreateMock

		UpdateIndex int
		UpdateMocks []ReleaseUpdateMock

		UploadAssetIndex int
		UploadAssetMocks []ReleaseUploadAssetMock
	}
)

func (m *MockReleaseService) List(ctx context.Context, pageSize, pageNo int) ([]github.Release, *github.Response, error) {
	i := m.ListIndex
	m.ListIndex++
	m.ListMocks[i].InContext = ctx
	m.ListMocks[i].InPageSize = pageSize
	m.ListMocks[i].InPageNo = pageNo
	return m.ListMocks[i].OutReleases, m.ListMocks[i].OutResponse, m.ListMocks[i].OutError
}

func (m *MockReleaseService) Create(ctx context.Context, params github.ReleaseParams) (*github.Release, *github.Response, error) {
	i := m.CreateIndex
	m.CreateIndex++
	m.CreateMocks[i].InContext = ctx
	m.CreateMocks[i].InParams = params
	return m.CreateMocks[i].OutRelease, m.CreateMocks[i].OutResponse, m.CreateMocks[i].OutError
}

func (m *MockReleaseService) Update(ctx context.Context, releaseID int, params github.ReleaseParams) (*github.Release, *github.Response, error) {
	i := m.UpdateIndex
	m.UpdateIndex++
	m.UpdateMocks[i].InContext = ctx
	m.UpdateMocks[i].InReleaseID = releaseID
	m.UpdateMocks[i].InParams = params
	return m.UpdateMocks[i].OutRelease, m.UpdateMocks[i].OutResponse, m.UpdateMocks[i].OutError
}

func (m *MockReleaseService) UploadAsset(ctx context.Context, releaseID int, assetFile, assetLabel string) (*github.ReleaseAsset, *github.Response, error) {
	i := m.UploadAssetIndex
	m.UploadAssetIndex++
	m.UploadAssetMocks[i].InContext = ctx
	m.UploadAssetMocks[i].InReleaseID = releaseID
	m.UploadAssetMocks[i].InAssetFile = assetFile
	m.UploadAssetMocks[i].InAssetLabel = assetLabel
	return m.UploadAssetMocks[i].OutReleaseAsset, m.UploadAssetMocks[i].OutResponse, m.UploadAssetMocks[i].OutError
}

type (
	PullGetMock struct {
		InContext   context.Context
		InNumber    int
		OutPull     *github.Pull
		OutResponse *github.Response
		OutError    error
	}

	PullCreateMock struct {
		InContext   context.Context
		InParams    github.CreatePullParams
		OutPull     *github.Pull
		OutResponse *github.Response
		OutError    error
	}

	PullUpdateMock struct {
		InContext   context.Context
		InNumber    int
		InParams    github.UpdatePullParams
		OutPull     *github.Pull
		OutResponse *github.Response
		OutError    error
	}

	MockPullService struct {
		GetIndex int
		GetMocks []PullGetMock

		CreateIndex int
		CreateMocks []PullCreateMock

		UpdateIndex int
		UpdateMocks []PullUpdateMock
	}
)

func (m *MockPullService) Get(ctx context.Context, number int) (*github.Pull, *github.Response, error) {
	i := m.GetIndex
	m.GetIndex++
	m.GetMocks[i].InContext = ctx
	m.GetMocks[i].InNumber = number
	return m.GetMocks[i].OutPull, m.GetMocks[i].OutResponse, m.GetMocks[i].OutError
}

func (m *MockPullService) Create(ctx context.Context, params github.CreatePullParams) (*github.Pull, *github.Response, error) {
	i := m.CreateIndex
	m.CreateIndex++
	m.CreateMocks[i].InContext = ctx
	m.CreateMocks[i].InParams = params
	return m.CreateMocks[i].OutPull, m.CreateMocks[i].OutResponse, m.CreateMocks[i].OutError
}

func (m *MockPullService) Update(ctx context.Context, number int, params github.UpdatePullParams) (*github.Pull, *github.Response, error) {
	i := m.UpdateIndex
	m.UpdateIndex++
	m.UpdateMocks[i].InContext = ctx
	m.UpdateMocks[i].InNumber = number
	m.UpdateMocks[i].InParams = params
	return m.UpdateMocks[i].OutPull, m.UpdateMocks[i].OutResponse, m.UpdateMocks[i].OutError
}

type (
	UserMock struct {
		InContext   context.Context
		OutUser     *github.User
		OutResponse *github.Response
		OutError    error
	}

	MockUserService struct {
		UserIndex int
		UserMocks []UserMock
	}
)

func (m *MockUserService) User(ctx context.Context) (*github.User, *github.Response, error) {
	i := m.UserIndex
	m.UserIndex++
	m.UserMocks[i].InContext = ctx
	return m.UserMocks[i].OutUser, m.UserMocks[i].OutResponse, m.UserMocks[i].OutError
}

type (
	SearchIssuesMock struct {
		InContext   context.Context
		InPageSize  int
		InPageNo    int
		InSort      github.SearchResultSort
		InOrder     github.SearchResultOrder
		InQuery     github.SearchQuery
		OutResult   *github.SearchIssuesResult
		OutResponse *github.Response
		OutError    error
	}

	MockSearchService struct {
		SearchIssuesIndex int
		SearchIssuesMocks []SearchIssuesMock
	}
)

func (m *MockSearchService) SearchIssues(ctx context.Context, pageSize, pageNo int, sort github.SearchResultSort, order github.SearchResultOrder, query github.SearchQuery) (*github.SearchIssuesResult, *github.Response, error) {
	i := m.SearchIssuesIndex
	m.SearchIssuesIndex++
	m.SearchIssuesMocks[i].InContext = ctx
	m.SearchIssuesMocks[i].InPageSize = pageSize
	m.SearchIssuesMocks[i].InPageNo = pageNo
	m.SearchIssuesMocks[i].InSort = sort
	m.SearchIssuesMocks[i].InOrder = order
	m.SearchIssuesMocks[i].InQuery = query
	return m.SearchIssuesMocks[i].OutResult, m.SearchIssuesMocks[i].OutResponse, m.SearchIssuesMocks[i].OutError
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
