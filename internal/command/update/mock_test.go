package update

import (
	"context"
	"io"

	"github.com/gardenbed/go-github"
)

type (
	LatestMock struct {
		InContext   context.Context
		OutRelease  *github.Release
		OutResponse *github.Response
		OutError    error
	}

	DownloadAssetMock struct {
		InContext    context.Context
		InReleaseTag string
		InAssetName  string
		InWriter     io.Writer
		OutResponse  *github.Response
		OutError     error
	}

	MockReleaseService struct {
		LatestIndex int
		LatestMocks []LatestMock

		DownloadAssetIndex int
		DownloadAssetMocks []DownloadAssetMock
	}
)

func (m *MockReleaseService) Latest(ctx context.Context) (*github.Release, *github.Response, error) {
	i := m.LatestIndex
	m.LatestIndex++
	m.LatestMocks[i].InContext = ctx
	return m.LatestMocks[i].OutRelease, m.LatestMocks[i].OutResponse, m.LatestMocks[i].OutError
}

func (m *MockReleaseService) DownloadAsset(ctx context.Context, releaseTag, assetName string, writer io.Writer) (*github.Response, error) {
	i := m.DownloadAssetIndex
	m.DownloadAssetIndex++
	m.DownloadAssetMocks[i].InContext = ctx
	m.DownloadAssetMocks[i].InReleaseTag = releaseTag
	m.DownloadAssetMocks[i].InAssetName = assetName
	m.DownloadAssetMocks[i].InWriter = writer
	return m.DownloadAssetMocks[i].OutResponse, m.DownloadAssetMocks[i].OutError
}
