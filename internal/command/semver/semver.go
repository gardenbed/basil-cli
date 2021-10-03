package semver

import (
	"flag"
	"strconv"

	"github.com/mitchellh/cli"

	"github.com/gardenbed/basil-cli/internal/command"
	"github.com/gardenbed/basil-cli/internal/git"
	"github.com/gardenbed/basil-cli/internal/semver"
)

const (
	synopsis = `Print the current semantic version`
	help     = `
  Use this command for getting the current semantic version.

  Usage:  basil project semver

  Examples:
    basil project semver
  `
)

type gitService interface {
	IsClean() (bool, error)
	HEAD() (string, string, error)
	Tags() (git.Tags, error)
	CommitsIn(string) (git.Commits, error)
}

// Command is the cli.Command implementation for semver command.
type Command struct {
	ui       cli.Ui
	services struct {
		git gitService
	}
	outputs struct {
		semver semver.SemVer
	}
}

// New creates a semver command.
func New(ui cli.Ui) *Command {
	return &Command{
		ui: ui,
	}
}

// NewFactory returns a cli.CommandFactory for creating a semver command.
func NewFactory(ui cli.Ui) cli.CommandFactory {
	return func() (cli.Command, error) {
		return New(ui), nil
	}
}

// Synopsis returns a short one-line synopsis for the command.
func (c *Command) Synopsis() string {
	return synopsis
}

// Help returns a long help text including usage, description, and list of flags for the command.
func (c *Command) Help() string {
	return help
}

// Run runs the actual command with the given command-line arguments.
// This method is used as a proxy for creating dependencies and the actual command execution is delegated to the run method for testing purposes.
func (c *Command) Run(args []string) int {
	git, err := git.Open(".")
	if err != nil {
		c.ui.Error(err.Error())
		return command.GitError
	}

	c.services.git = git

	return c.run(args)
}

// run in an auxiliary method, so we can test the business logic with mock dependencies.
func (c *Command) run(args []string) int {
	fs := flag.NewFlagSet("semver", flag.ContinueOnError)
	fs.Usage = func() {
		c.ui.Output(c.Help())
	}

	if err := fs.Parse(args); err != nil {
		return command.FlagError
	}

	// ==============================> GET GIT INFORMATION <==============================

	isClean, err := c.services.git.IsClean()
	if err != nil {
		c.ui.Error(err.Error())
		return command.GitError
	}

	gitSHA, _, err := c.services.git.HEAD()
	if err != nil {
		c.ui.Error(err.Error())
		return command.GitError
	}

	// Get all git tags
	tags, err := c.services.git.Tags()
	if err != nil {
		c.ui.Error(err.Error())
		return command.GitError
	}

	// Get all git commits in the current branch
	commits, err := c.services.git.CommitsIn("HEAD")
	if err != nil {
		c.ui.Error(err.Error())
		return command.GitError
	}

	// Get the most recent tag that is a semantic version
	tag, _ := tags.First(func(t git.Tag) bool {
		// Make sure the tag falls in the commits range
		if t.Commit.After(commits[0]) {
			return false
		}

		// Make sure the tag is a semantic version
		if _, ok := semver.Parse(t.Name); !ok {
			return false
		}

		return true
	})

	// ==============================> RESOLVE THE CURRENT SEMANTIC VERSION <==============================

	var sv semver.SemVer

	var signature string
	if isClean {
		signature = gitSHA[:7]
	} else {
		signature = "dev"
	}

	if tag.IsZero() {
		// No git tag and no previous semantic version -> using the default initial semantic version
		sv = semver.SemVer{Major: 0, Minor: 1, Patch: 0}
		count := strconv.Itoa(len(commits))
		sv.Prerelease = append(sv.Prerelease, count, signature)
	} else {
		// The most recent tag either points to the HEAD commit or is reachable from the HEAD commit
		// The tag is guaranteed to be a valid semantic version thanks to the predicte for selecting it
		sv, _ = semver.Parse(tag.Name)

		// Count how many commits HEAD is ahead of the most recent tag
		var count int
		for i, c := range commits {
			if c.Equal(tag.Commit) {
				count = i
				break
			}
		}

		// If there are any changes since the most recent tag, we are on next semantic version
		// If the the most recent tag points to the HEAD commit and the working tree is clean, we are just at current semantic version
		if count > 0 || !isClean {
			sv = sv.Next()
			sv.Prerelease = append(sv.Prerelease, strconv.Itoa(count), signature)
		}
	}

	c.outputs.semver = sv

	c.ui.Info(sv.String())

	// ==============================> DONE <==============================

	return command.Success
}

// SemVer returns the semantic version output.
func (c *Command) SemVer() semver.SemVer {
	return c.outputs.semver
}
