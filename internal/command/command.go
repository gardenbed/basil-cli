package command

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"golang.org/x/sync/errgroup"

	"github.com/gardenbed/basil-cli/internal/shell"
)

const (
	// Success is the exit code when a command execution is successful.
	Success int = iota
	// GenericError is the generic exit code when something fails.
	GenericError
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
	// GitHubError is the exit code when a GitHub request fails.
	GitHubError
	// ChangelogError is the exit code when generating the changelog fails.
	ChangelogError
)

var (
	goVersionRegexp = regexp.MustCompile(`go([0-9]+\.[0-9]+\.[0-9]+)`)
)

type (
	// PreflightChecklist is a list of common preflight checks for commands.
	PreflightChecklist struct {
		Go bool
	}

	// PreflightInfo is a list of common preflight information for commands.
	PreflightInfo struct {
		WorkingDirectory string
		GoVersion        string
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

	// Get the Go compiler version
	if checklist.Go {
		group.Go(func() (err error) {
			var out string
			if _, out, err = shell.Run(ctx, "go", "version"); err != nil {
				return fmt.Errorf("error on getting Go compiler version: %s", err)
			}

			matches := goVersionRegexp.FindStringSubmatch(out)
			if len(matches) != 2 {
				return fmt.Errorf("invalid Go compiler version: %s", out)
			}

			info.GoVersion = matches[1]
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return info, err
	}

	return info, nil
}
