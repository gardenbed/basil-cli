// Package update implements the command for updating Basil CLI.
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
	"github.com/gardenbed/basil-cli/internal/config"
	"github.com/gardenbed/basil-cli/internal/ui"
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

type (
	releaseService interface {
		Latest(context.Context) (*github.Release, *github.Response, error)
		DownloadAsset(context.Context, string, string, io.Writer) (*github.Response, error)
	}
)

// Command is the cli.Command implementation for update command.
type Command struct {
	ui       ui.UI
	config   config.Config
	services struct {
		releases releaseService
	}
}

// New creates a new command.
func New(ui ui.UI, config config.Config) *Command {
	return &Command{
		ui:     ui,
		config: config,
	}
}

// NewFactory returns a cli.CommandFactory for creating a new command.
func NewFactory(ui ui.UI, config config.Config) cli.CommandFactory {
	return func() (cli.Command, error) {
		return New(ui, config), nil
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
	if code := c.parseFlags(args); code != command.Success {
		return code
	}

	// GitHub access token for update command is optional
	token := c.config.GitHub.AccessToken
	c.services.releases = github.NewClient(token).Repo(owner, repo).Releases

	return c.exec()
}

func (c *Command) parseFlags(args []string) int {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)

	fs.Usage = func() {
		c.ui.Printf(c.Help())
	}

	if err := fs.Parse(args); err != nil {
		// In case of error, the error and help will be printed by the Parse method
		return command.FlagError
	}

	return command.Success
}

// exec in an auxiliary method, so we can test the business logic with mock dependencies.
func (c *Command) exec() int {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	binPath, err := exec.LookPath(os.Args[0])
	if err != nil {
		c.ui.Errorf(ui.Red, "Cannot find the path for Basil binary: %s", err)
		return command.OSError
	}

	// ==============================> GET THE LATEST RELEASE <==============================

	c.ui.Printf("Finding the latest release of basil ...")

	release, _, err := c.services.releases.Latest(ctx)
	if err != nil {
		c.ui.Errorf(ui.Red, "%s", err)
		return command.GitHubError
	}

	// ==============================> DOWNLOAD THE LATEST BINARY <==============================

	c.ui.Printf("Downloading Basil %s ...", release.TagName)

	assetName := fmt.Sprintf("basil-%s-%s", runtime.GOOS, runtime.GOARCH)

	f, err := os.OpenFile(binPath, os.O_WRONLY, 0755)
	if err != nil {
		c.ui.Errorf(ui.Red, "Cannot open file for writing: %s", err)
		return command.OSError
	}

	_, err = c.services.releases.DownloadAsset(ctx, release.TagName, assetName, f)
	if err != nil {
		c.ui.Errorf(ui.Red, "Failed to download and update Basil binary: %s", err)
		_ = f.Close() // Ignore error here since we are already failing.
		return command.GitHubError
	}

	if err := f.Close(); err != nil {
		c.ui.Errorf(ui.Red, "Failed to close the file: %s", err)
		return command.OSError
	}

	c.ui.Infof(ui.Green, "🌿 Basil %s written to %s", release.Name, binPath)

	// ==============================> DONE <==============================

	return command.Success
}
