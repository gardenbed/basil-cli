package semver

import "github.com/gardenbed/basil-cli/internal/git"

type (
	TagsMock struct {
		OutTags  git.Tags
		OutError error
	}

	CommitsInMock struct {
		InRev      string
		OutCommits git.Commits
		OutError   error
	}

	MockGitService struct {
		TagsIndex int
		TagsMocks []TagsMock

		CommitsInIndex int
		CommitsInMocks []CommitsInMock
	}
)

func (m *MockGitService) Tags() (git.Tags, error) {
	i := m.TagsIndex
	m.TagsIndex++
	return m.TagsMocks[i].OutTags, m.TagsMocks[i].OutError
}

func (m *MockGitService) CommitsIn(rev string) (git.Commits, error) {
	i := m.CommitsInIndex
	m.CommitsInIndex++
	m.CommitsInMocks[i].InRev = rev
	return m.CommitsInMocks[i].OutCommits, m.CommitsInMocks[i].OutError
}
