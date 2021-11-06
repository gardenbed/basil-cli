package release

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"
	"time"

	changelog "github.com/gardenbed/changelog/generate"
	changelogspec "github.com/gardenbed/changelog/spec"
	"github.com/gardenbed/charm/flagit"
	"github.com/gardenbed/charm/shell"
	"github.com/gardenbed/go-github"
	"github.com/mitchellh/cli"
	"golang.org/x/sync/errgroup"

	"github.com/gardenbed/basil-cli/internal/command"
	buildcmd "github.com/gardenbed/basil-cli/internal/command/project/build"
	semvercmd "github.com/gardenbed/basil-cli/internal/command/project/semver"
	"github.com/gardenbed/basil-cli/internal/config"
	"github.com/gardenbed/basil-cli/internal/git"
	"github.com/gardenbed/basil-cli/internal/semver"
	"github.com/gardenbed/basil-cli/internal/spec"
)

const (
	timeout  = 10 * time.Minute
	synopsis = `Create a release`
	help     = `
  Use this command for creating a new release.
  Currently, only GitHub repositories are supported.

  It assumes the remote repository name is origin.
  The initial semantic version is always 0.1.0.

  DIRECT Release:
  A new release commit will be created, tagged, and directly psuhed to the default branch.
  A new GitHub release will also be created and published.

  INDIRECT Release:
  A new release commit will be created and a pull request will be opened for it to be reviewed and merged.
  A new draft GitHub release will also be created.
  After the pull request is merged, you need to tag the release commit and publish the draft release.
  The last step can be done manually or through a GitGub action.

  Usage:  basil project release [flags]

  Flags:
    -patch      create a patch release (default: true)
    -minor      create a minor release (default: false)
    -major      create a major release (default: false)
    -comment    add a description for the release
    -mode       the release mode, either direct or indirect (default: {{.Project.Release.Mode}})

  Examples:
    basil project release
    basil project release -patch
    basil project release -minor
    basil project release -major
    basil project release -comment="Fixing Bugs!"
    basil project release -minor -comment "New Features!"
    basil project release -major -comment "Breaking Changes!"
  `
)

const (
	remoteName = "origin"
)

var (
	h2Regex = regexp.MustCompile(`##[^\n]*\n`)
)

type (
	gitService interface {
		Remote(string) (string, string, error)
	}

	repoService interface {
		Get(context.Context) (*github.Repository, *github.Response, error)
		Permission(context.Context, string) (github.Permission, *github.Response, error)
		BranchProtection(context.Context, string, bool) (*github.Response, error)
	}

	releaseService interface {
		List(context.Context, int, int) ([]github.Release, *github.Response, error)
		Create(context.Context, github.ReleaseParams) (*github.Release, *github.Response, error)
		Update(context.Context, int, github.ReleaseParams) (*github.Release, *github.Response, error)
		UploadAsset(context.Context, int, string, string) (*github.ReleaseAsset, *github.Response, error)
	}

	pullService interface {
		Get(context.Context, int) (*github.Pull, *github.Response, error)
		Create(context.Context, github.CreatePullParams) (*github.Pull, *github.Response, error)
		Update(context.Context, int, github.UpdatePullParams) (*github.Pull, *github.Response, error)
	}

	userService interface {
		User(context.Context) (*github.User, *github.Response, error)
	}

	searchService interface {
		SearchIssues(context.Context, int, int, github.SearchResultSort, github.SearchResultOrder, github.SearchQuery) (*github.SearchIssuesResult, *github.Response, error)
	}

	changelogService interface {
		Generate(context.Context, changelogspec.Spec) (string, error)
	}

	semverCommand interface {
		Run([]string) int
		SemVer() semver.SemVer
	}

	buildCommand interface {
		Run([]string) int
		Artifacts() []buildcmd.Artifact
	}
)

// Command is the cli.Command implementation for release command.
type Command struct {
	ui     cli.Ui
	config config.Config
	spec   spec.Spec
	flags  struct {
		patch, minor, major bool
		comment             string
	}
	data struct {
		owner, repo   string
		changelogSpec changelogspec.Spec
	}
	funcs struct {
		goList        shell.RunnerFunc
		gitStatus     shell.RunnerFunc
		gitRevBranch  shell.RunnerFunc
		gitBranch     shell.RunnerFunc
		gitCheckout   shell.RunnerFunc
		gitAdd        shell.RunnerFunc
		gitCommit     shell.RunnerFunc
		gitTag        shell.RunnerFunc
		gitPull       shell.RunnerFunc
		gitPush       shell.RunnerFunc
		gitPushTag    shell.RunnerFunc
		gitPushBranch shell.RunnerFunc
	}
	services struct {
		git       gitService
		repo      repoService
		releases  releaseService
		pulls     pullService
		users     userService
		search    searchService
		changelog changelogService
	}
	commands struct {
		semver semverCommand
		build  buildCommand
	}
	outputs struct {
		version semver.SemVer
	}
}

// New creates a new command.
func New(ui cli.Ui, config config.Config, spec spec.Spec) *Command {
	return &Command{
		ui:     ui,
		config: config,
		spec:   spec,
	}
}

// NewFactory returns a cli.CommandFactory for creating a new command.
func NewFactory(ui cli.Ui, config config.Config, spec spec.Spec) cli.CommandFactory {
	return func() (cli.Command, error) {
		return New(ui, config, spec), nil
	}
}

// Synopsis returns a short one-line synopsis for the command.
func (c *Command) Synopsis() string {
	return synopsis
}

// Help returns a long help text including usage, description, and list of flags for the command.
func (c *Command) Help() string {
	var buf bytes.Buffer
	t := template.Must(template.New("help").Parse(help))
	_ = t.Execute(&buf, c.spec)
	return buf.String()
}

// Run runs the actual command with the given command-line arguments.
// This method is used as a proxy for creating dependencies and the actual command execution is delegated to the run method for testing purposes.
func (c *Command) Run(args []string) int {
	if code := c.parseFlags(args); code != command.Success {
		return code
	}

	git, err := git.Open(".")
	if err != nil {
		c.ui.Error(err.Error())
		return command.GitError
	}

	domain, path, err := git.Remote(remoteName)
	if err != nil {
		c.ui.Error(err.Error())
		return command.GitError
	}

	if domain != "github.com" {
		c.ui.Error(fmt.Sprintf("unsupported git platform: %s", domain))
		return command.GitHubError
	}

	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		c.ui.Error("Unexpected GitHub repository: cannot parse owner and repo.")
		return command.GitHubError
	}
	ownerName, repoName := parts[0], parts[1]

	if c.config.GitHub.AccessToken == "" {
		c.ui.Error("A GitHub access token is required.")
		return command.GitHubError
	}

	client := github.NewClient(c.config.GitHub.AccessToken)
	repo := client.Repo(ownerName, repoName)

	changelogSpec, err := changelogspec.Default().FromFile()
	if err != nil {
		c.ui.Error(err.Error())
		return command.ChangelogError
	}

	changelogSpec = changelogSpec.WithRepo(domain, path)
	changelogSpec.Repo.AccessToken = c.config.GitHub.AccessToken
	changelog, err := changelog.New(changelogSpec, newLogger(c.ui))
	if err != nil {
		c.ui.Error(err.Error())
		return command.ChangelogError
	}

	c.data.owner = ownerName
	c.data.repo = repoName
	c.data.changelogSpec = changelogSpec

	c.funcs.goList = shell.Runner("go", "list", "./...")
	c.funcs.gitStatus = shell.Runner("git", "status", "--porcelain")
	c.funcs.gitRevBranch = shell.Runner("git", "rev-parse", "--abbrev-ref", "HEAD")
	c.funcs.gitBranch = shell.Runner("git", "branch")
	c.funcs.gitCheckout = shell.Runner("git", "checkout")
	c.funcs.gitAdd = shell.Runner("git", "add", c.data.changelogSpec.General.File)
	c.funcs.gitCommit = shell.Runner("git", "commit", "-m")
	c.funcs.gitTag = shell.Runner("git", "tag")
	c.funcs.gitPull = shell.Runner("git", "pull")
	c.funcs.gitPush = shell.Runner("git", "push")
	c.funcs.gitPushTag = shell.Runner("git", "push", remoteName)
	c.funcs.gitPushBranch = shell.Runner("git", "push", "-u", remoteName)
	c.services.git = git
	c.services.repo = repo
	c.services.releases = repo.Releases
	c.services.pulls = repo.Pulls
	c.services.users = client.Users
	c.services.search = client.Search
	c.services.changelog = changelog
	c.commands.semver = semvercmd.New(cli.NewMockUi())
	c.commands.build = buildcmd.New(c.ui, c.spec)

	return c.exec()
}

func (c *Command) parseFlags(args []string) int {
	fs := flag.NewFlagSet("release", flag.ContinueOnError)
	fs.BoolVar(&c.flags.patch, "patch", true, "")
	fs.BoolVar(&c.flags.minor, "minor", false, "")
	fs.BoolVar(&c.flags.major, "major", false, "")
	fs.StringVar(&c.flags.comment, "comment", "", "")

	fs.Usage = func() {
		c.ui.Output(c.Help())
	}

	if err := flagit.Register(fs, &c.spec.Project.Release, false); err != nil {
		return command.GenericError
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

	// ==============================> RUN PREFLIGHT CHECKS <==============================

	c.ui.Output("Running preflight checks ...")

	checklist := command.PreflightChecklist{
		Git: true,
		Go:  true,
	}

	if _, err := command.RunPreflightChecks(ctx, checklist); err != nil {
		c.ui.Error(err.Error())
		return command.PreflightError
	}

	// ==============================> VALIDATE REPO STATE <==============================

	repo, _, err := c.services.repo.Get(ctx)
	if err != nil {
		c.ui.Error(err.Error())
		return command.GitHubError
	}

	_, gitBranch, err := c.funcs.gitRevBranch(ctx)
	if err != nil {
		c.ui.Error(err.Error())
		return command.GitError
	}

	if gitBranch != repo.DefaultBranch {
		c.ui.Error("The repository can only be released from the default branch.")
		return command.GitError
	}

	_, gitStatus, err := c.funcs.gitStatus(ctx)
	if err != nil {
		c.ui.Error(err.Error())
		return command.GitError
	}

	if gitStatus != "" {
		c.ui.Error("Working directory is not clean and has uncommitted changes.")
		return command.GitError
	}

	c.ui.Info(fmt.Sprintf("Pulling the latest changes on the %s branch ...", gitBranch))

	if _, _, err := c.funcs.gitPull(ctx); err != nil {
		c.ui.Error(err.Error())
		return command.GitError
	}

	// ==============================> RESOLVE SEMANTIC VERSION <==============================

	c.ui.Output(fmt.Sprintf("Releasing %s/%s in %s mode ...", c.data.owner, c.data.repo, c.spec.Project.Release.Mode))

	// Run semver command
	if code := c.commands.semver.Run(nil); code != command.Success {
		return code
	}

	switch {
	case c.flags.major:
		c.outputs.version = c.commands.semver.SemVer().ReleaseMajor()
	case c.flags.minor:
		c.outputs.version = c.commands.semver.SemVer().ReleaseMinor()
	case c.flags.patch:
		fallthrough
	default:
		c.outputs.version = c.commands.semver.SemVer().ReleasePatch()
	}

	// ==============================> BRANCH BASED ON MODE <==============================

	switch c.spec.Project.Release.Mode {
	case spec.ReleaseModeDirect:
		return c.directRelease(ctx, gitBranch)
	case spec.ReleaseModeIndirect:
		return c.indirectRelease(ctx, gitBranch)
	default:
		c.ui.Error(fmt.Sprintf("Invalid release mode: %s", c.spec.Project.Release.Mode))
		return command.SpecError
	}
}

// For direct mode
func (c *Command) directRelease(ctx context.Context, defaultBranch string) int {
	// ==============================> CHECK GITHUB PERMISSION <==============================

	c.ui.Output("Checking GitHub permission for direct mode ...")

	user, _, err := c.services.users.User(ctx)
	if err != nil {
		c.ui.Error(err.Error())
		return command.GitHubError
	}

	perm, _, err := c.services.repo.Permission(ctx, user.Login)
	if err != nil {
		c.ui.Error(err.Error())
		return command.GitHubError
	}

	if perm != github.PermissionAdmin {
		c.ui.Error("The provided GitHub access token does not have admin permission for direct mode.")
		return command.GitHubError
	}

	// ==============================> CREATE A DRAFT RELEASE <==============================

	c.ui.Info(fmt.Sprintf("Creating the draft release %s ...", c.outputs.version))

	release, _, err := c.services.releases.Create(ctx, github.ReleaseParams{
		Name:       c.outputs.version.String(),
		TagName:    c.outputs.version.TagName(),
		Target:     defaultBranch,
		Draft:      true,
		Prerelease: false,
	})

	if err != nil {
		c.ui.Error(err.Error())
		return command.GitHubError
	}

	// ==============================> GENERATE CHANGELOG <==============================

	c.ui.Info("Creating/Updating the changelog ...")

	c.data.changelogSpec.Tags.Future = c.outputs.version.TagName()

	changelog, err := c.services.changelog.Generate(ctx, c.data.changelogSpec)
	if err != nil {
		c.ui.Error(err.Error())
		return command.ChangelogError
	}

	// Remove the H2 title
	changelog = h2Regex.ReplaceAllString(changelog, "")
	changelog = strings.TrimLeft(changelog, "\n")

	// ==============================> CREATE RELEASE COMMIT & TAG <==============================

	c.ui.Info(fmt.Sprintf("Creating the release commit %s ...", c.outputs.version))

	if _, _, err := c.funcs.gitAdd(ctx); err != nil {
		c.ui.Error(err.Error())
		return command.GitError
	}

	// We need to create the commit using the git command.
	// So, all user configurations (author, committer, signing key, etc.) will be picked up correctly and automatically.
	message := fmt.Sprintf("Release %s", c.outputs.version)
	if _, _, err := c.funcs.gitCommit(ctx, message); err != nil {
		c.ui.Error(err.Error())
		return command.GitError
	}

	c.ui.Info(fmt.Sprintf("Creating the release tag %s ...", c.outputs.version))

	// We need to create the tag using the git command.
	// So, all user configurations (author, committer, signing key, etc.) will be picked up correctly and automatically.
	if _, _, err := c.funcs.gitTag(ctx, "-a", c.outputs.version.TagName(), "-m", message); err != nil {
		c.ui.Error(err.Error())
		return command.GitError
	}

	// ==============================> BUILD AND UPLOAD ARTIFACTS <==============================

	// Check if we can build any artifacts
	if _, _, err := c.funcs.goList(ctx); err == nil {
		c.ui.Output("Building artifacts ...")

		// Run build command
		if code := c.commands.build.Run(nil); code != command.Success {
			return code
		}

		if artifacts := c.commands.build.Artifacts(); len(artifacts) > 0 {
			c.ui.Info(fmt.Sprintf("Uploading artifacts to release %s ...", release.Name))

			group, groupCtx := errgroup.WithContext(ctx)

			for _, artifact := range artifacts {
				artifact := artifact // https://golang.org/doc/faq#closures_and_goroutines
				group.Go(func() error {
					_, _, err := c.services.releases.UploadAsset(groupCtx, release.ID, artifact.Path, artifact.Label)
					return err
				})
			}

			if err := group.Wait(); err != nil {
				c.ui.Error(err.Error())
				return command.GitHubError
			}
		}
	}

	// ==============================> TEMPORARILY DISABLE DEFAULT BRANCH PROTECTION <==============================

	c.ui.Warn(fmt.Sprintf("Temporarily enabling push to %s branch ...", defaultBranch))

	if _, err := c.services.repo.BranchProtection(ctx, defaultBranch, false); err != nil {
		c.ui.Error(err.Error())
		return command.GitHubError
	}

	// Make sure we re-enable the branch protection
	defer func() {
		c.ui.Warn(fmt.Sprintf("🔒 Re-disabling push to %s branch ...", defaultBranch))
		if _, err := c.services.repo.BranchProtection(ctx, defaultBranch, true); err != nil {
			c.ui.Error(err.Error())
			os.Exit(command.GitHubError)
		}
	}()

	// ==============================> PUSH RELEASE COMMIT & TAG <==============================

	c.ui.Info(fmt.Sprintf("Pushing the release commit %s ...", c.outputs.version))

	if _, _, err := c.funcs.gitPush(ctx); err != nil {
		c.ui.Error(err.Error())
		return command.GitError
	}

	c.ui.Info(fmt.Sprintf("Pushing the release tag %s ...", c.outputs.version.TagName()))

	if _, _, err := c.funcs.gitPushTag(ctx, c.outputs.version.TagName()); err != nil {
		c.ui.Error(err.Error())
		return command.GitError
	}

	// ==============================> PUBLISH THE RELEASE <==============================

	c.ui.Info(fmt.Sprintf("Publishing the release %s ...", release.Name))

	description := changelog
	if c.flags.comment != "" {
		description = fmt.Sprintf("%s\n\n%s", c.flags.comment, description)
	}

	release, _, err = c.services.releases.Update(ctx, release.ID, github.ReleaseParams{
		Name:       release.Name,
		TagName:    release.TagName,
		Target:     release.Target,
		Draft:      false,
		Prerelease: false,
		Body:       description,
	})

	if err != nil {
		c.ui.Error(err.Error())
		return command.GitHubError
	}

	// ==============================> DONE <==============================

	return command.Success
}

// For indirect mode
func (c *Command) indirectRelease(ctx context.Context, defaultBranch string) int {
	title := fmt.Sprintf("RELEASE %s", c.outputs.version)
	releaseBranch := fmt.Sprintf("release-%s", c.outputs.version)

	// ==============================> CHECK FOR A MERGED PULL REQUEST <==============================

	c.ui.Output(fmt.Sprintf("Checking for a merged pull request for release %s ...", c.outputs.version))

	// Search for a merged pull request matching the release version
	mergedQuery := github.SearchQuery{}
	mergedQuery.IncludeKeywords(title)
	mergedQuery.IncludeQualifiers(
		github.QualifierIsPR,
		github.QualifierIsMerged,
		github.QualifierInTitle,
		github.QualifierRepo(c.data.owner, c.data.repo),
	)

	mergedResult, _, err := c.services.search.SearchIssues(ctx, 1, 1, github.SortByDefault, github.DefaultOrder, mergedQuery)
	if err != nil {
		c.ui.Error(err.Error())
		return command.GitHubError
	}

	// ==============================> FINISH THE RELEASE <==============================

	if len(mergedResult.Items) > 0 {
		return c.tagAndPublishRelease(ctx, mergedResult.Items[0].Number, defaultBranch)
	}

	// ==============================> CHECK FOR AN OPEN PULL REQUEST <==============================

	c.ui.Output(fmt.Sprintf("Checking for an open pull request for release %s ...", c.outputs.version))

	// Search for an open pull request matching the release version
	openQuery := github.SearchQuery{}
	openQuery.IncludeKeywords(title)
	openQuery.IncludeQualifiers(
		github.QualifierIsPR,
		github.QualifierIsOpen,
		github.QualifierInTitle,
		github.QualifierRepo(c.data.owner, c.data.repo),
	)

	openResult, _, err := c.services.search.SearchIssues(ctx, 1, 1, github.SortByDefault, github.DefaultOrder, openQuery)
	if err != nil {
		c.ui.Error(err.Error())
		return command.GitHubError
	}

	// ==============================> CREATE/UPDATE PULL REQUEST <==============================

	changelog, code := c.pushReleaseBranch(ctx, defaultBranch, releaseBranch)
	if code != command.Success {
		return code
	}

	description := changelog
	if c.flags.comment != "" {
		description = fmt.Sprintf("%s\n\n%s", c.flags.comment, description)
	}

	if len(openResult.Items) == 0 {
		code := c.createPullAndRelease(ctx, defaultBranch, releaseBranch, title, description)
		if code != command.Success {
			return code
		}
	} else {
		code := c.updatePullAndRelease(ctx, openResult.Items[0].Number, defaultBranch, title, description)
		if code != command.Success {
			return code
		}
	}

	// ==============================> DONE <==============================

	return command.Success
}

func (c *Command) tagAndPublishRelease(ctx context.Context, number int, defaultBranch string) int {
	// ==============================> GET MERGED PULL REQUEST & DRAFT RELEASE <==============================

	pull, _, err := c.services.pulls.Get(ctx, number)
	if err != nil {
		c.ui.Error(err.Error())
		return command.GitHubError
	}

	release, code := c.findDraftRelease(ctx, c.outputs.version.TagName())
	if code != command.Success {
		return code
	}

	// ==============================> CREATE & PUSH TAG <==============================

	c.ui.Info(fmt.Sprintf("Pulling the latest changes on the %s branch ...", defaultBranch))

	if _, _, err := c.funcs.gitPull(ctx); err != nil {
		c.ui.Error(err.Error())
		return command.GitError
	}

	c.ui.Info(fmt.Sprintf("Creating the release tag %s ...", c.outputs.version))

	// We need to create the tag using the git command.
	// So, all user configurations (author, committer, signing key, etc.) will be picked up correctly and automatically.
	message := fmt.Sprintf("Release %s", c.outputs.version)
	if _, _, err := c.funcs.gitTag(ctx, "-a", c.outputs.version.TagName(), pull.MergeCommitSHA, "-m", message); err != nil {
		c.ui.Error(err.Error())
		return command.GitError
	}

	c.ui.Info(fmt.Sprintf("Pushing the release tag %s ...", c.outputs.version.TagName()))

	if _, _, err := c.funcs.gitPushTag(ctx, c.outputs.version.TagName()); err != nil {
		c.ui.Error(err.Error())
		return command.GitError
	}

	// ==============================> BUILD AND UPLOAD ARTIFACTS <==============================

	// Check if we can build any artifacts
	if _, _, err := c.funcs.goList(ctx); err == nil {
		c.ui.Output("Building artifacts ...")

		// Run build command
		if code := c.commands.build.Run(nil); code != command.Success {
			return code
		}

		if artifacts := c.commands.build.Artifacts(); len(artifacts) > 0 {
			c.ui.Info(fmt.Sprintf("Uploading artifacts to release %s ...", c.outputs.version))

			group, groupCtx := errgroup.WithContext(ctx)

			for _, artifact := range artifacts {
				artifact := artifact // https://golang.org/doc/faq#closures_and_goroutines
				group.Go(func() error {
					_, _, err := c.services.releases.UploadAsset(groupCtx, release.ID, artifact.Path, artifact.Label)
					return err
				})
			}

			if err := group.Wait(); err != nil {
				c.ui.Error(err.Error())
				return command.GitHubError
			}
		}
	}

	// ==============================> PUBLISH THE RELEASE <==============================

	c.ui.Info(fmt.Sprintf("Publishing the release %s ...", release.Name))

	release, _, err = c.services.releases.Update(ctx, release.ID, github.ReleaseParams{
		Name:       release.Name,
		TagName:    release.TagName,
		Target:     release.Target,
		Draft:      false,
		Prerelease: false,
		Body:       release.Body,
	})

	if err != nil {
		c.ui.Error(err.Error())
		return command.GitHubError
	}

	// ==============================> DONE <==============================

	return command.Success
}

func (c *Command) pushReleaseBranch(ctx context.Context, defaultBranch, releaseBranch string) (string, int) {
	// ==============================> GENERATE CHANGELOG <==============================

	c.ui.Info("Creating/Updating the changelog ...")

	c.data.changelogSpec.Tags.Future = c.outputs.version.TagName()

	changelog, err := c.services.changelog.Generate(ctx, c.data.changelogSpec)
	if err != nil {
		c.ui.Error(err.Error())
		return changelog, command.ChangelogError
	}

	// Remove the H2 title
	changelog = h2Regex.ReplaceAllString(changelog, "")
	changelog = strings.TrimLeft(changelog, "\n")

	// ==============================> CREATE RELEASE BRANCH & COMMIT <==============================

	c.ui.Info(fmt.Sprintf("Creating the release branch %s ...", c.outputs.version))

	// Create a new branch for the release
	if _, _, err := c.funcs.gitCheckout(ctx, "-b", releaseBranch); err != nil {
		c.ui.Error(err.Error())
		return changelog, command.GitError
	}

	c.ui.Info(fmt.Sprintf("Creating the release commit %s ...", c.outputs.version))

	if _, _, err := c.funcs.gitAdd(ctx); err != nil {
		c.ui.Error(err.Error())
		return changelog, command.GitError
	}

	// We need to create the commit using the git command.
	// So, all user configurations (author, committer, signing key, etc.) will be picked up correctly and automatically.
	message := fmt.Sprintf("Release %s", c.outputs.version)
	if _, _, err := c.funcs.gitCommit(ctx, message); err != nil {
		c.ui.Error(err.Error())
		return changelog, command.GitError
	}

	// ==============================> PUSH RELEASE BRANCH <==============================

	c.ui.Info(fmt.Sprintf("Pushing the release branch %s ...", c.outputs.version))

	// Push the release branch
	if _, _, err := c.funcs.gitPushBranch(ctx, "-f", releaseBranch); err != nil {
		c.ui.Error(err.Error())
		return changelog, command.GitError
	}

	// ==============================> DELETE RELEASE BRANCH <==============================

	// Check out to default branch
	if _, _, err := c.funcs.gitCheckout(ctx, defaultBranch); err != nil {
		c.ui.Error(err.Error())
		return changelog, command.GitError
	}

	// Delete the release branch
	if _, _, err := c.funcs.gitBranch(ctx, "-D", releaseBranch); err != nil {
		c.ui.Error(err.Error())
		return changelog, command.GitError
	}

	return changelog, command.Success
}

func (c *Command) createPullAndRelease(ctx context.Context, defaultBranch, releaseBranch, title, description string) int {
	// ==============================> CREATE PULL REQUEST <==============================

	c.ui.Info(fmt.Sprintf("Creating pull request for release %s ...", c.outputs.version))

	pull, _, err := c.services.pulls.Create(ctx, github.CreatePullParams{
		Title: title,
		Body:  description,
		Head:  releaseBranch,
		Base:  defaultBranch,
	})

	if err != nil {
		c.ui.Error(err.Error())
		return command.GitHubError
	}

	// ==============================> CREATE DRAFT RELEASE <==============================

	c.ui.Info(fmt.Sprintf("Creating the draft release %s ...", c.outputs.version))

	release, _, err := c.services.releases.Create(ctx, github.ReleaseParams{
		Name:       c.outputs.version.String(),
		TagName:    c.outputs.version.TagName(),
		Target:     defaultBranch,
		Draft:      true,
		Prerelease: false,
		Body:       description,
	})

	if err != nil {
		c.ui.Error(err.Error())
		return command.GitHubError
	}

	// ==============================> DONE <==============================

	c.ui.Output(fmt.Sprintf("Pull request created:  %s", pull.HTMLURL))
	c.ui.Output(fmt.Sprintf("Draft release created: %s", release.HTMLURL))
	c.ui.Warn("🔖 Re-run this command to update the pull request.")
	c.ui.Warn("🔖 After merging the pull request, re-run this command to create the release.")

	return command.Success
}

func (c *Command) updatePullAndRelease(ctx context.Context, number int, defaultBranch, title, description string) int {
	// ==============================> UPDATE PULL REQUEST <==============================

	c.ui.Info(fmt.Sprintf("Updating pull request for release %s ...", c.outputs.version))

	pull, _, err := c.services.pulls.Update(ctx, number, github.UpdatePullParams{
		Title: title,
		Body:  description,
		Base:  defaultBranch,
	})

	if err != nil {
		c.ui.Error(err.Error())
		return command.GitHubError
	}

	// ==============================> UPDATE DRAFT RELEASE <==============================

	c.ui.Info(fmt.Sprintf("Updating the draft release %s ...", c.outputs.version))

	release, code := c.findDraftRelease(ctx, c.outputs.version.TagName())
	if code != command.Success {
		return code
	}

	release, _, err = c.services.releases.Update(ctx, release.ID, github.ReleaseParams{
		Name:       c.outputs.version.String(),
		TagName:    c.outputs.version.TagName(),
		Target:     defaultBranch,
		Draft:      true,
		Prerelease: false,
		Body:       description,
	})

	if err != nil {
		c.ui.Error(err.Error())
		return command.GitHubError
	}

	// ==============================> DONE <==============================

	c.ui.Output(fmt.Sprintf("Pull request updated:  %s", pull.HTMLURL))
	c.ui.Output(fmt.Sprintf("Draft release updated: %s", release.HTMLURL))
	c.ui.Warn("🔖 Re-run this command to update the pull request.")
	c.ui.Warn("🔖 After merging the pull request, re-run this command to create the release.")

	return command.Success
}

func (c *Command) findDraftRelease(ctx context.Context, tag string) (*github.Release, int) {
	releases, resp, err := c.services.releases.List(ctx, 100, 1)
	if err != nil {
		c.ui.Error(err.Error())
		return nil, command.GitHubError
	}

	for _, release := range releases {
		if release.Draft && release.TagName == tag {
			return &release, command.Success
		}
	}

	var release *github.Release

	group, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for page := 2; page <= resp.Pages.Last; page++ {
		group.Go(func() error {
			releases, _, err := c.services.releases.List(ctx, 100, 1)
			if err != nil {
				return err
			}

			for _, r := range releases {
				if r.Draft && r.TagName == tag {
					release = &r
					cancel()
					break
				}
			}

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		c.ui.Error(err.Error())
		return nil, command.GitHubError
	}

	if release == nil {
		c.ui.Error(fmt.Sprintf("Draft release not found for tag %s", tag))
		return nil, command.GitHubError
	}

	return release, command.Success
}
