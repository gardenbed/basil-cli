package git

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

var (
	idPattern       = `[A-Za-z][0-9A-Za-z-]+[0-9A-Za-z]`
	domainPattern   = fmt.Sprintf(`%s\.[A-Za-z]{2,63}`, idPattern)
	repoPathPattern = fmt.Sprintf(`(%s/){1,20}(%s)`, idPattern, idPattern)
	httpsPattern    = fmt.Sprintf(`^https://(%s)/(%s)(.git)?$`, domainPattern, repoPathPattern)
	sshPattern      = fmt.Sprintf(`^git@(%s):(%s)(.git)?$`, domainPattern, repoPathPattern)
	httpsRE         = regexp.MustCompile(httpsPattern)
	sshRE           = regexp.MustCompile(sshPattern)
)

// DetectGit determines if a given path or any of its parents has a .git directory.
func DetectGit(path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(path); err != nil {
		return "", err
	}

	if path == "/" {
		return "", errors.New("git path not found")
	}

	gitPath := filepath.Join(path, ".git")
	if _, err = os.Stat(gitPath); err == nil {
		return path, nil
	}

	if os.IsNotExist(err) {
		topPath := filepath.Dir(path)
		return DetectGit(topPath)
	}

	return "", err
}

// Git provides Git functionalities.
type Git struct {
	repo *git.Repository
}

// Open creates a new Git service for an existing git repository.
func Open(path string) (*Git, error) {
	repo, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{
		DetectDotGit: true,
	})

	if err != nil {
		return nil, err
	}

	return &Git{
		repo: repo,
	}, nil
}

// Path returns the root path of the Git repository.
func (g *Git) Path() (string, error) {
	worktree, err := g.repo.Worktree()
	if err != nil {
		return "", err
	}

	return worktree.Filesystem.Root(), nil
}

// Remote returns the domain part and path part of a Git remote repository URL.
// It assumes the remote repository is named origin.
func (g *Git) Remote(name string) (string, string, error) {
	remote, err := g.repo.Remote(name)
	if err != nil {
		return "", "", err
	}

	// TODO: Should we handle all URLs and not just the first one?
	var remoteURL string
	if config := remote.Config(); len(config.URLs) > 0 {
		remoteURL = config.URLs[0]
	}

	return parseRemoteURL(remoteURL)
}

func parseRemoteURL(url string) (string, string, error) {
	// Parse the origin remote URL into a domain part a path part
	if m := httpsRE.FindStringSubmatch(url); len(m) == 6 { // HTTPS Git Remote URL
		//  Example:
		//    https://github.com/gardenbed/basil-cli.git
		//    m = []string{"https://github.com/gardenbed/basil-cli.git", "github.com", "gardenbed/basil-cli", "gardenbed/", "basil-cli", ".git"}
		return m[1], m[2], nil
	} else if m := sshRE.FindStringSubmatch(url); len(m) == 6 { // SSH Git Remote URL
		//  Example:
		//    git@github.com:gardenbed/basil-cli.git
		//    m = []string{"git@github.com:gardenbed/basil-cli.git", "github.com", "gardenbed/basil-cli, "gardenbed/", "basil-cli", ".git"}
		return m[1], m[2], nil
	}

	return "", "", fmt.Errorf("invalid git remote url: %s", url)
}

// Tags returns the list of all tags.
func (g *Git) Tags() (Tags, error) {
	refs, err := g.repo.Tags()
	if err != nil {
		return nil, err
	}

	tags := []Tag{}

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		t, err := g.repo.TagObject(ref.Hash())
		switch err {
		// Annotated tag
		case nil:
			c, err := g.repo.CommitObject(t.Target)
			if err != nil {
				return err
			}
			tags = append(tags, toAnnotatedTag(t, c))

		// Lightweight tag
		case plumbing.ErrObjectNotFound:
			c, err := g.repo.CommitObject(ref.Hash())
			if err != nil {
				return err
			}
			tags = append(tags, toLightweightTag(ref, c))

		default:
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort tags
	sort.Slice(tags, func(i, j int) bool {
		// The order of the tags should be from the most recent to the least recent
		return tags[i].After(tags[j])
	})

	return tags, nil
}

// CommitsIn returns all commits reachable from a revision.
func (g *Git) CommitsIn(rev string) (Commits, error) {
	hash, err := g.repo.ResolveRevision(plumbing.Revision(rev))
	if err != nil {
		return nil, err
	}

	commitsMap := make(map[plumbing.Hash]*object.Commit)
	err = g.parentCommits(commitsMap, *hash)
	if err != nil {
		return nil, err
	}

	commits := make([]Commit, 0)
	for _, c := range commitsMap {
		commits = append(commits, toCommit(c))
	}

	// Sort commits
	sort.Slice(commits, func(i, j int) bool {
		// The order of the commits should be from the most recent to the least recent
		return commits[i].Committer.After(commits[j].Committer)
	})

	return commits, nil
}

func (g *Git) parentCommits(commitsMap map[plumbing.Hash]*object.Commit, h plumbing.Hash) error {
	if _, ok := commitsMap[h]; ok {
		return nil
	}

	c, err := g.repo.CommitObject(h)
	if err != nil {
		return err
	}

	commitsMap[c.Hash] = c

	for _, h := range c.ParentHashes {
		if err := g.parentCommits(commitsMap, h); err != nil {
			return err
		}
	}

	return nil
}
