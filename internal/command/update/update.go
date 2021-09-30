package update

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/gardenbed/go-github"
	"github.com/mitchellh/cli"

	"github.com/gardenbed/basil-cli/internal/command"
)

const (
	timeout  = time.Minute
	synopsis = `Update Basil`
	help     = `
  Use this command for updating basil to the latest release.

  Usage:  basil update

  Examples:
    basil update
  `
)

const (
	owner = "gardenbed"
	repo  = "basil-cli"
)

type repoService interface {
	LatestRelease(context.Context) (*github.Release, *github.Response, error)
	DownloadReleaseAsset(context.Context, string, string, io.Writer) (*github.Response, error)
}

// Command is the cli.Command implementation for update command.
type Command struct {
	ui       cli.Ui
	services struct {
		repo repoService
	}
}

// New returns a cli.CommandFactory for creating an update command.
func New(ui cli.Ui) cli.CommandFactory {
	return func() (cli.Command, error) {
		return &Command{
			ui: ui,
		}, nil
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
	// If no access token is provided, we try without it!
	token := os.Getenv("BASIL_GITHUB_TOKEN")

	c.services.repo = github.NewClient(token).Repo(owner, repo)

	return c.run(args)
}

// run in an auxiliary method, so we can test the business logic with mock dependencies.
func (c *Command) run(args []string) int {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	fs.Usage = func() {
		c.ui.Output(c.Help())
	}

	if err := fs.Parse(args); err != nil {
		return command.FlagError
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// ==============================> GET THE LATEST RELEASE <==============================

	c.ui.Output("Finding the latest release of basil ...")

	release, _, err := c.services.repo.LatestRelease(ctx)
	if err != nil {
		c.ui.Error(err.Error())
		return command.GitHubError
	}

	// ==============================> DOWNLOAD THE LATEST BINARY <==============================

	c.ui.Output(fmt.Sprintf("Downloading Basil %s ...", release.TagName))

	assetName := fmt.Sprintf("basil-%s-%s", runtime.GOOS, runtime.GOARCH)

	binPath, err := exec.LookPath(os.Args[0])
	if err != nil {
		c.ui.Error(fmt.Sprintf("Cannot find the path for Basil binary: %s", err))
		return command.OSError
	}

	f, err := os.OpenFile(binPath, os.O_WRONLY, 0755)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Cannot open file for writing: %s", err))
		return command.OSError
	}
	defer f.Close()

	_, err = c.services.repo.DownloadReleaseAsset(ctx, release.TagName, assetName, f)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Failed to download and update Basil binary: %s", err))
		return command.GitHubError
	}

	c.ui.Info(fmt.Sprintf("🌿 Basil %s written to %s", release.Name, binPath))

	// ==============================> DONE <==============================

	return command.Success
}
