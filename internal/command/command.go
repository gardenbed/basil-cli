package command

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"golang.org/x/sync/errgroup"

	"github.com/gardenbed/charm/shell"
)

const (
	// Success is the exit code when a command execution is successful.
	Success int = iota
	// GenericError is the generic exit code when something fails.
	GenericError
	// ConfigError is the exit code when reading or writing the config file fails.
	ConfigError
	// SpecError is the exit code when reading the spec file fails.
	SpecError
	// FlagError is the exit code when an undefined or invalid flag is provided to a command.
	FlagError
	// PreflightError is the exit code when a preflight check fails.
	PreflightError
	// OSError is the exit code when an OS operation fails.
	OSError
	// GoError is the exit code when a go command fails.
	GoError
	// GitError is the exit code when a git command fails.
	GitError
	// GPGError is the exit code when a gpg command fails.
	GPGError
	// GitHubError is the exit code when a GitHub request fails.
	GitHubError
	// ChangelogError is the exit code when generating the changelog fails.
	ChangelogError
)

var (
	gpgVersionRegexp = regexp.MustCompile(`gpg \(GnuPG\) ([0-9]+\.[0-9]+\.[0-9]+)`)
	gitVersionRegexp = regexp.MustCompile(`git version ([0-9]+\.[0-9]+\.[0-9]+)`)
	goVersionRegexp  = regexp.MustCompile(`go([0-9]+\.[0-9]+\.[0-9]+)`)
)

type (
	// PreflightChecklist is a list of common preflight checks for commands.
	PreflightChecklist struct {
		GPG bool
		Git bool
		Go  bool
	}

	// PreflightInfo is a list of common preflight information for commands.
	PreflightInfo struct {
		WorkingDirectory string

		GPG struct {
			Version string
		}

		Git struct {
			Version string
		}

		Go struct {
			Version string
		}
	}
)

// RunPreflightChecks runs a list of preflight checks to ensure they are fulfilled.
// It returns a list of preflight information.
func RunPreflightChecks(ctx context.Context, checklist PreflightChecklist) (PreflightInfo, error) {
	var info PreflightInfo

	group, ctx := errgroup.WithContext(ctx)

	// Get the current working directory
	group.Go(func() (err error) {
		if info.WorkingDirectory, err = os.Getwd(); err != nil {
			return fmt.Errorf("error on getting the current working directory: %s", err)
		}
		return nil
	})

	// Check the gpg
	if checklist.GPG {
		group.Go(func() (err error) {
			var out string
			if _, out, err = shell.Run(ctx, "gpg", "--version"); err != nil {
				return fmt.Errorf("error on checking gpg: %s", err)
			}

			matches := gpgVersionRegexp.FindStringSubmatch(out)
			if len(matches) != 2 {
				return fmt.Errorf("invalid gpg version: %s", out)
			}

			info.GPG.Version = matches[1]
			return nil
		})
	}

	// Check the git
	if checklist.Git {
		group.Go(func() (err error) {
			var out string
			if _, out, err = shell.Run(ctx, "git", "--version"); err != nil {
				return fmt.Errorf("error on checking go: %s", err)
			}

			matches := gitVersionRegexp.FindStringSubmatch(out)
			if len(matches) != 2 {
				return fmt.Errorf("invalid git version: %s", out)
			}

			info.Git.Version = matches[1]
			return nil
		})
	}

	// Check the go
	if checklist.Go {
		group.Go(func() (err error) {
			var out string
			if _, out, err = shell.Run(ctx, "go", "version"); err != nil {
				return fmt.Errorf("error on checking go: %s", err)
			}

			matches := goVersionRegexp.FindStringSubmatch(out)
			if len(matches) != 2 {
				return fmt.Errorf("invalid go version: %s", out)
			}

			info.Go.Version = matches[1]
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return info, err
	}

	return info, nil
}
