package create

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gardenbed/go-github"
	"github.com/mitchellh/cli"

	"github.com/gardenbed/basil-cli/internal/archive"
	"github.com/gardenbed/basil-cli/internal/command"
	"github.com/gardenbed/basil-cli/internal/config"
	"github.com/gardenbed/basil-cli/internal/debug"
	"github.com/gardenbed/basil-cli/internal/spec"
	"github.com/gardenbed/basil-cli/internal/template"
)

const (
	timeout  = time.Minute
	synopsis = `Create a new project`
	help     = `
  Use this command for creating a new project.

  Usage:  basil project create [flags]

  Flags:
    -name        the name of the new project
    -owner       the owner id for the new project (team name, id, email, etc.)
    -profile     the profile (template) name for creating the new project based off it
    -dockerid    the Docker ID for building container images for the new project

  Examples:
    basil project create
    basil project create -name=my-service -owner=my-team -profile=grpc-service -dockerid=orca
  `
)

const (
	templateOwner = "gardenbed"
	templateRepo  = "basil-templates"
	templateLang  = "go"
)

var (
	nameRegexp     = regexp.MustCompile(`^[a-z][0-9a-z-]+$`)
	ownerRegexp    = regexp.MustCompile(`^[a-z][0-9a-z-]+$`)
	dockeridRegexp = regexp.MustCompile(`^[a-z][0-9a-z-]+$`)

	allProfiles = []string{
		string(spec.ProjectProfileGeneric),
		string(spec.ProjectProfileCLI),
		string(spec.ProjectProfileGRPCService), string(spec.ProjectProfileGRPCServiceHorizontal),
		string(spec.ProjectProfileHTTPService), string(spec.ProjectProfileHTTPServiceHorizontal),
	}
)

type (
	repoService interface {
		DownloadTarArchive(context.Context, string, io.Writer) (*github.Response, error)
	}

	archiveService interface {
		Extract(string, io.Reader, archive.Selector) error
	}

	templateService interface {
		Execute(template.Template) error
	}
)

// Command is the cli.Command implementation for create command.
type Command struct {
	ui     cli.Ui
	config config.Config
	flags  struct {
		name     string
		owner    string
		profile  string
		dockerid string
		revision string
	}
	services struct {
		repo     repoService
		archive  archiveService
		template templateService
	}
}

// New creates a new command.
func New(ui cli.Ui, config config.Config) *Command {
	return &Command{
		ui:     ui,
		config: config,
	}
}

// NewFactory returns a cli.CommandFactory for creating a new command.
func NewFactory(ui cli.Ui, config config.Config) cli.CommandFactory {
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

	// GitHub access token is optional
	token := c.config.GitHub.AccessToken
	c.services.repo = github.NewClient(token).Repo(templateOwner, templateRepo)
	c.services.archive = archive.NewTarArchive(debug.None)
	c.services.template = template.NewService(debug.None)

	return c.exec()
}

func (c *Command) parseFlags(args []string) int {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	fs.StringVar(&c.flags.name, "name", "", "")
	fs.StringVar(&c.flags.owner, "owner", "", "")
	fs.StringVar(&c.flags.profile, "profile", "", "")
	fs.StringVar(&c.flags.dockerid, "dockerid", "", "")
	fs.StringVar(&c.flags.revision, "revision", "main", "")

	fs.Usage = func() {
		c.ui.Output(c.Help())
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

	checklist := command.PreflightChecklist{}

	info, err := command.RunPreflightChecks(ctx, checklist)
	if err != nil {
		c.ui.Error(err.Error())
		return command.PreflightError
	}

	// ==============================> GET INPUTS <==============================

	if c.flags.name == "" {
		c.flags.name, err = c.ui.Ask("Project name:")
		if err != nil {
			c.ui.Error(err.Error())
			return command.InputError
		}

		if !nameRegexp.MatchString(c.flags.name) {
			c.ui.Error(fmt.Sprintf("Invalid name: %s", c.flags.name))
			return command.InputError
		}
	}

	if c.flags.owner == "" {
		c.flags.owner, err = c.ui.Ask("Project owner (team name, id, email, etc.):")
		if err != nil {
			c.ui.Error(err.Error())
			return command.InputError
		}

		if !ownerRegexp.MatchString(c.flags.owner) {
			c.ui.Error(fmt.Sprintf("Invalid owner: %s", c.flags.owner))
			return command.InputError
		}
	}

	if c.flags.profile == "" {
		c.flags.profile, err = c.ui.Ask(fmt.Sprintf("Project profile (%s):", strings.Join(allProfiles, ", ")))
		if err != nil {
			c.ui.Error(err.Error())
			return command.InputError
		}

		if !c.isProfileValid(c.flags.profile) {
			c.ui.Error(fmt.Sprintf("Invalid profile: %s", c.flags.profile))
			return command.InputError
		}
	}

	if c.flags.dockerid == "" {
		c.flags.dockerid, err = c.ui.Ask("Docker ID:")
		if err != nil {
			c.ui.Error(err.Error())
			return command.InputError
		}

		if !dockeridRegexp.MatchString(c.flags.dockerid) {
			c.ui.Error(fmt.Sprintf("Invalid Docker ID: %s", c.flags.dockerid))
			return command.InputError
		}
	}

	// ==============================> DOWNLOAD & EXTRACT TEMPLATE <==============================

	c.ui.Info(fmt.Sprintf("Downloading template %q revision %q ...", c.flags.profile, c.flags.revision))

	buf := new(bytes.Buffer)
	if _, err := c.services.repo.DownloadTarArchive(ctx, c.flags.revision, buf); err != nil {
		c.ui.Error(fmt.Sprintf("Failed to download templates: %s", err))
		return command.GitHubError
	}

	c.ui.Output(fmt.Sprintf("Extracting template %q revision %q ...", c.flags.profile, c.flags.revision))

	if err := c.services.archive.Extract(info.WorkingDirectory, buf, c.selectTemplatePath()); err != nil {
		c.ui.Error(fmt.Sprintf("Failed to extract template: %s", err))
		return command.ArchiveError
	}

	// ==============================> APPLY TEMPLATE CHANGES <==============================

	path := filepath.Join(info.WorkingDirectory, c.flags.name)

	c.ui.Info(fmt.Sprintf("Finalizing %s ...", path))

	params := struct {
		Name     string
		Owner    string
		DockerID string
	}{
		Name:     c.flags.name,
		Owner:    c.flags.owner,
		DockerID: c.flags.dockerid,
	}

	t, err := template.Read(path, params)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Template error: %s", err))
		return command.TemplateError
	}

	if err := c.services.template.Execute(t); err != nil {
		c.ui.Error(fmt.Sprintf("Template error: %s", err))
		return command.TemplateError
	}

	// ==============================> DONE <==============================

	return command.Success
}

func (c *Command) isProfileValid(profile string) bool {
	for _, p := range allProfiles {
		if profile == p {
			return true
		}
	}
	return false
}

func (c *Command) selectTemplatePath() func(path string) (string, bool) {
	// c.flags.profile is already validated
	archivePathRegexp := regexp.MustCompile(fmt.Sprintf("^%s-%s-[0-9a-f]{7,40}/%s/%s/", templateOwner, templateRepo, templateLang, c.flags.profile))

	return func(path string) (string, bool) {
		if archivePathRegexp.MatchString(path) {
			return archivePathRegexp.ReplaceAllString(path, c.flags.name+"/"), true
		}
		return "", false
	}
}
