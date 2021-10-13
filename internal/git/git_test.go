package git

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
)

const (
	testPath  = "./test"
	gitmodule = `[submodule "make"]
	path = make
	url = git@github.com:octocat/module.git
	branch = main
`
)

func setupGitRepo() (*git.Repository, func(), error) {
	repo, err := git.PlainInit(testPath, false)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		os.RemoveAll(testPath)
	}

	c, err := repo.Config()
	if err != nil {
		cleanup()
		return nil, nil, err
	}

	// Required configs
	c.Author.Name = "Jane Doe"
	c.Author.Email = "jane.doe@example.com"

	if err := repo.SetConfig(c); err != nil {
		cleanup()
		return nil, nil, err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		cleanup()
		return nil, nil, err
	}

	// CREATE FIRST COMMIT

	if err := ioutil.WriteFile(testPath+"/README.md", []byte(""), 0644); err != nil {
		cleanup()
		return nil, nil, err
	}

	if _, err := worktree.Add("."); err != nil {
		cleanup()
		return nil, nil, err
	}

	h1, err := worktree.Commit("First commit", &git.CommitOptions{})
	if err != nil {
		cleanup()
		return nil, nil, err
	}

	// CREATE SECOND COMMIT

	if err := ioutil.WriteFile(testPath+"/LICENSE", []byte(""), 0644); err != nil {
		cleanup()
		return nil, nil, err
	}

	if _, err := worktree.Add("."); err != nil {
		cleanup()
		return nil, nil, err
	}

	h2, err := worktree.Commit("Second commit", &git.CommitOptions{})
	if err != nil {
		cleanup()
		return nil, nil, err
	}

	// THIRD COMMIT: ADD SUBMODULE

	if err := ioutil.WriteFile(testPath+"/.gitmodules", []byte(gitmodule), 0644); err != nil {
		cleanup()
		return nil, nil, err
	}

	if _, err := worktree.Add("."); err != nil {
		cleanup()
		return nil, nil, err
	}

	h3, err := worktree.Commit("Third commit", &git.CommitOptions{})
	if err != nil {
		cleanup()
		return nil, nil, err
	}

	// CREATE A BRANCH

	br := plumbing.NewHashReference("refs/heads/feature-branch", h3)
	if err := repo.Storer.SetReference(br); err != nil {
		cleanup()
		return nil, nil, err
	}

	// CREATE TAGS

	if _, err := repo.CreateTag("v0.1.0", h1, nil); err != nil {
		cleanup()
		return nil, nil, err
	}

	opts := &git.CreateTagOptions{
		Message: "second tag",
	}

	if _, err := repo.CreateTag("v0.2.0", h2, opts); err != nil {
		cleanup()
		return nil, nil, err
	}

	// ADD REMOTE

	rc := &config.RemoteConfig{
		Name: "origin",
		URLs: []string{
			"https://github.com/octocat/Hello-World",
		},
	}

	if _, err := repo.CreateRemote(rc); err != nil {
		cleanup()
		return nil, nil, err
	}

	return repo, cleanup, nil
}

func TestDetectGit(t *testing.T) {
	_, cleanup, err := setupGitRepo()
	assert.NoError(t, err)
	defer cleanup()

	tests := []struct {
		name          string
		path          string
		expectedError string
	}{
		{
			name:          "PathNotExist",
			path:          "/invalid",
			expectedError: "stat /invalid: no such file or directory",
		},
		{
			name:          "NoGit",
			path:          "/opt",
			expectedError: "git path not found",
		},
		{
			name:          "Success",
			path:          testPath,
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gitPath, err := DetectGit(tc.path)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.NotEmpty(t, gitPath)
			} else {
				assert.Empty(t, gitPath)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestOpen(t *testing.T) {
	_, cleanup, err := setupGitRepo()
	assert.NoError(t, err)
	defer cleanup()

	tests := []struct {
		name          string
		path          string
		expectedError string
	}{
		{
			name:          "PathNotExist",
			path:          "/foo",
			expectedError: "repository does not exist",
		},
		{
			name:          "Success",
			path:          testPath,
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g, err := Open(tc.path)

			if tc.expectedError != "" {
				assert.Nil(t, g)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, g)
				assert.NotNil(t, g.repo)
			}
		})
	}
}

func TestGit_Path(t *testing.T) {
	repo, cleanup, err := setupGitRepo()
	assert.NoError(t, err)
	defer cleanup()

	g := &Git{repo: repo}

	path, err := g.Path()

	assert.NoError(t, err)
	assert.Equal(t, testPath, path)
}

func TestGit_Remote(t *testing.T) {
	repo, cleanup, err := setupGitRepo()
	assert.NoError(t, err)
	defer cleanup()

	tests := []struct {
		name           string
		remoteName     string
		expectedDomain string
		expectedPath   string
		expectedError  string
	}{
		{
			name:          "RemoteNotExist",
			remoteName:    "foo",
			expectedError: "remote not found",
		},
		{
			name:           "Success",
			remoteName:     "origin",
			expectedDomain: "github.com",
			expectedPath:   "octocat/Hello-World",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := &Git{repo: repo}

			domain, path, err := g.Remote(tc.remoteName)

			if tc.expectedError != "" {
				assert.Empty(t, domain)
				assert.Empty(t, path)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedDomain, domain)
				assert.Equal(t, tc.expectedPath, path)
			}
		})
	}
}

func TestParseRemoteURL(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		expectedDomain string
		expectedPath   string
		expectedError  string
	}{
		{
			name:          "Empty",
			url:           "",
			expectedError: "invalid git remote url: ",
		},
		{
			name:          "Invalid",
			url:           "octocat/Hello-World",
			expectedError: "invalid git remote url: octocat/Hello-World",
		},
		{
			name:           "SSH",
			url:            "git@github.com:octocat/Hello-World",
			expectedDomain: "github.com",
			expectedPath:   "octocat/Hello-World",
		},
		{
			name:           "SSH_git",
			url:            "git@github.com:octocat/Hello-World.git",
			expectedDomain: "github.com",
			expectedPath:   "octocat/Hello-World",
		},
		{
			name:           "HTTPS",
			url:            "https://github.com/octocat/Hello-World",
			expectedDomain: "github.com",
			expectedPath:   "octocat/Hello-World",
		},
		{
			name:           "HTTPS_git",
			url:            "https://github.com/octocat/Hello-World.git",
			expectedDomain: "github.com",
			expectedPath:   "octocat/Hello-World",
		},
		{
			name:          "SSHVariant",
			url:           "ssh://git@github.com/octocat/Hello-World",
			expectedError: "invalid git remote url: ssh://git@github.com/octocat/Hello-World",
		},
		{
			name:          "SSHVariant_git",
			url:           "ssh://git@github.com/octocat/Hello-World.git",
			expectedError: "invalid git remote url: ssh://git@github.com/octocat/Hello-World.git",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			domain, path, err := parseRemoteURL(tc.url)

			if tc.expectedError != "" {
				assert.Empty(t, domain)
				assert.Empty(t, path)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedDomain, domain)
				assert.Equal(t, tc.expectedPath, path)
			}
		})
	}
}

func TestGit_Tags(t *testing.T) {
	repo, cleanup, err := setupGitRepo()
	assert.NoError(t, err)
	defer cleanup()

	g := &Git{repo: repo}

	tags, err := g.Tags()
	assert.NoError(t, err)
	assert.Len(t, tags, 2)
}

func TestGit_CommitsIn(t *testing.T) {
	repo, cleanup, err := setupGitRepo()
	assert.NoError(t, err)
	defer cleanup()

	g := &Git{repo: repo}

	commits, err := g.CommitsIn("HEAD")
	assert.NoError(t, err)
	assert.Len(t, commits, 3)
}
