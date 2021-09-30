package semver

import "github.com/gardenbed/basil-cli/internal/git"

type (
	IsCleanMock struct {
		OutBool  bool
		OutError error
	}

	HEADMock struct {
		OutHash   string
		OutBranch string
		OutError  error
	}

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
		IsCleanIndex int
		IsCleanMocks []IsCleanMock

		HEADIndex int
		HEADMocks []HEADMock

		TagsIndex int
		TagsMocks []TagsMock

		CommitsInIndex int
		CommitsInMocks []CommitsInMock
	}
)

func (m *MockGitService) IsClean() (bool, error) {
	i := m.IsCleanIndex
	m.IsCleanIndex++
	return m.IsCleanMocks[i].OutBool, m.IsCleanMocks[i].OutError
}

func (m *MockGitService) HEAD() (string, string, error) {
	i := m.HEADIndex
	m.HEADIndex++
	return m.HEADMocks[i].OutHash, m.HEADMocks[i].OutBranch, m.HEADMocks[i].OutError
}

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
