// Package create implements the command for creating (scaffolding) a new monorepo.
package create

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"time"

	"github.com/gardenbed/go-github"
	"github.com/mitchellh/cli"

	"github.com/gardenbed/basil-cli/internal/archive"
	"github.com/gardenbed/basil-cli/internal/command"
	"github.com/gardenbed/basil-cli/internal/config"
	"github.com/gardenbed/basil-cli/internal/template"
	"github.com/gardenbed/basil-cli/internal/ui"
)

const (
	timeout  = time.Minute
	synopsis = `Create a new monorepo`
	help     = `
  Use this command for creating a new monorepo.

  Usage:  basil monorepo create [flags]

  Flags:
    -name    the name of the new monorepo

  Examples:
    basil monorepo create
    basil monorepo create -name=go-monorepo
  `
)

const (
	templateOwner = "gardenbed"
	templateRepo  = "basil-templates"
)

var (
	nameRegexp        = regexp.MustCompile(`^[a-z][0-9a-z-]+$`)
	archivePathRegexp = regexp.MustCompile(fmt.Sprintf("^%s-%s-[0-9a-f]{7,40}/go/monorepo/", templateOwner, templateRepo))
)

type (
	repoService interface {
		DownloadTarArchive(context.Context, string, io.Writer) (*github.Response, error)
	}

	archiveService interface {
		Extract(string, io.Reader, archive.Selector) error
	}

	templateService interface {
		Load(string) error
		Params() template.Params
		Template(interface{}) (*template.Template, error)
	}
)

// Command is the cli.Command implementation for create command.
type Command struct {
	ui     ui.UI
	config config.Config
	flags  struct {
		revision string
		name     string
	}
	services struct {
		repo     repoService
		archive  archiveService
		template templateService
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

	// GitHub access token is optional
	token := c.config.GitHub.AccessToken
	c.services.repo = github.NewClient(token).Repo(templateOwner, templateRepo)
	c.services.archive = archive.NewTarArchive(c.ui)
	c.services.template = template.NewService(c.ui)

	return c.exec()
}

func (c *Command) parseFlags(args []string) int {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	fs.StringVar(&c.flags.revision, "revision", "main", "")
	fs.StringVar(&c.flags.name, "name", "", "")

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

	// ==============================> RUN PREFLIGHT CHECKS <==============================

	checklist := command.PreflightChecklist{}

	info, err := command.RunPreflightChecks(ctx, checklist)
	if err != nil {
		c.ui.Errorf(ui.Red, "%s", err)
		return command.PreflightError
	}

	// ==============================> GET REQUIRED INPUTS <==============================

	if c.flags.name == "" {
		c.flags.name, err = c.ui.Ask("Monorepo name", "", validateInputName)
		if err != nil {
			c.ui.Errorf(ui.Red, "%s", err)
			return command.InputError
		}
	}

	// ==============================> DOWNLOAD & EXTRACT TEMPLATE <==============================

	c.ui.Infof(ui.Green, "Downloading monorepo template revision %q ...", c.flags.revision)

	buf := new(bytes.Buffer)
	if _, err := c.services.repo.DownloadTarArchive(ctx, c.flags.revision, buf); err != nil {
		c.ui.Errorf(ui.Red, "Failed to download templates: %s", err)
		return command.GitHubError
	}

	c.ui.Printf("Extracting monorepo template revision %q ...", c.flags.revision)

	if err = c.services.archive.Extract(info.WorkingDirectory, buf, c.selectTemplatePath); err != nil {
		c.ui.Errorf(ui.Red, "Failed to extract template: %s", err)
		return command.ArchiveError
	}

	// ==============================> LOAD TEMPLATE <==============================

	projectPath := filepath.Join(info.WorkingDirectory, c.flags.name)

	if err := c.services.template.Load(projectPath); err != nil {
		c.ui.Errorf(ui.Red, "%s", err)
		return command.TemplateError
	}

	// ==============================> APPLY TEMPLATE CHANGES <==============================

	c.ui.Infof(ui.Green, "Editing %s ...", projectPath)

	inputs := struct {
		Name string
	}{
		Name: c.flags.name,
	}

	template, err := c.services.template.Template(inputs)
	if err != nil {
		c.ui.Errorf(ui.Red, "Template error: %s", err)
		return command.TemplateError
	}

	if err := template.Execute(c.ui, projectPath); err != nil {
		c.ui.Errorf(ui.Red, "%s", err)
		return command.TemplateError
	}

	// ==============================> DONE <==============================

	return command.Success
}

func (c *Command) selectTemplatePath(path string) (string, bool) {
	if archivePathRegexp.MatchString(path) {
		return archivePathRegexp.ReplaceAllString(path, c.flags.name+"/"), true
	}
	return "", false
}

func validateInputName(val string) error {
	if !nameRegexp.MatchString(val) {
		return fmt.Errorf("invalid name: %s", val)
	}
	return nil
}
