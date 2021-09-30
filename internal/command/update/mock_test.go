package update

import (
	"context"
	"io"

	"github.com/gardenbed/go-github"
)

type (
	LatestReleaseMock struct {
		InContext   context.Context
		OutRelease  *github.Release
		OutResponse *github.Response
		OutError    error
	}

	DownloadReleaseAssetMock struct {
		InContext    context.Context
		InReleaseTag string
		InAssetName  string
		InWriter     io.Writer
		OutResponse  *github.Response
		OutError     error
	}

	MockRepoService struct {
		LatestReleaseIndex int
		LatestReleaseMocks []LatestReleaseMock

		DownloadReleaseAssetIndex int
		DownloadReleaseAssetMocks []DownloadReleaseAssetMock
	}
)

func (m *MockRepoService) LatestRelease(ctx context.Context) (*github.Release, *github.Response, error) {
	i := m.LatestReleaseIndex
	m.LatestReleaseIndex++
	m.LatestReleaseMocks[i].InContext = ctx
	return m.LatestReleaseMocks[i].OutRelease, m.LatestReleaseMocks[i].OutResponse, m.LatestReleaseMocks[i].OutError
}

func (m *MockRepoService) DownloadReleaseAsset(ctx context.Context, releaseTag, assetName string, writer io.Writer) (*github.Response, error) {
	i := m.DownloadReleaseAssetIndex
	m.DownloadReleaseAssetIndex++
	m.DownloadReleaseAssetMocks[i].InContext = ctx
	m.DownloadReleaseAssetMocks[i].InReleaseTag = releaseTag
	m.DownloadReleaseAssetMocks[i].InAssetName = assetName
	m.DownloadReleaseAssetMocks[i].InWriter = writer
	return m.DownloadReleaseAssetMocks[i].OutResponse, m.DownloadReleaseAssetMocks[i].OutError
}
