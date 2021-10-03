package build

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gardenbed/flagit"
	"github.com/mitchellh/cli"
	"golang.org/x/sync/errgroup"

	"github.com/gardenbed/basil-cli/internal/command"
	semvercmd "github.com/gardenbed/basil-cli/internal/command/semver"
	"github.com/gardenbed/basil-cli/internal/git"
	"github.com/gardenbed/basil-cli/internal/semver"
	"github.com/gardenbed/basil-cli/internal/shell"
	"github.com/gardenbed/basil-cli/internal/spec"
	"github.com/gardenbed/basil-cli/metadata"
)

const (
	timeout  = 5 * time.Minute
	synopsis = `Build artifacts`
	help     = `
  Use this command for building artifacts.
  Currently, the build command only builds binaries for Go applications.

  By convention, It assumes the current directory is a main package if it contains a main.go file.
  It also assumes every directory inside cmd is a main package for a binary with the same name as the directory name.

  Usage:  basil project build [flags]

  Flags:
    -cross-compile    build the binary for all platforms (default: {{.Build.CrossCompile}})

  Examples:
    basil project build
    basil project build -cross-compile
  `
)

const (
	cmdDir       = "cmd"
	binPath      = "./bin/"
	metadataPath = "./metadata"
	timeFormat   = "2006-01-02 15:04:05 MST"
)

var (
	goVersionRE = regexp.MustCompile(`\d+\.\d+(\.\d+)?`)
)

type (
	gitService interface {
		HEAD() (string, string, error)
	}

	semverCommand interface {
		Run([]string) int
		SemVer() semver.SemVer
	}
)

// Artifact is a build artifact.
type Artifact struct {
	Path  string
	Label string
}

// Command is the cli.Command implementation for build command.
type Command struct {
	sync.Mutex
	ui    cli.Ui
	spec  spec.Spec
	funcs struct {
		goList  shell.RunnerFunc
		goBuild shell.RunnerWithFunc
	}
	services struct {
		git gitService
	}
	commands struct {
		semver semverCommand
	}
	outputs struct {
		artifacts []Artifact
	}
}

// New creates a build command.
func New(ui cli.Ui, spec spec.Spec) *Command {
	return &Command{
		ui:   ui,
		spec: spec,
	}
}

// NewFactory returns a cli.CommandFactory for creating a build command.
func NewFactory(ui cli.Ui, spec spec.Spec) cli.CommandFactory {
	return func() (cli.Command, error) {
		return New(ui, spec), nil
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
	git, err := git.Open(".")
	if err != nil {
		c.ui.Error(err.Error())
		return command.GitError
	}

	c.funcs.goList = shell.Runner("go", "list", metadataPath)
	c.funcs.goBuild = shell.RunnerWith("go", "build")
	c.services.git = git
	c.commands.semver = semvercmd.New(cli.NewMockUi())

	return c.run(args)
}

// run in an auxiliary method, so we can test the business logic with mock dependencies.
func (c *Command) run(args []string) int {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	fs.Usage = func() {
		c.ui.Output(c.Help())
	}

	if err := flagit.Register(fs, &c.spec.Project.Build, false); err != nil {
		return command.GenericError
	}

	if err := fs.Parse(args); err != nil {
		return command.FlagError
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// ==============================> RUN PREFLIGHT CHECKS <==============================

	checklist := command.PreflightChecklist{
		Go: true,
	}

	info, err := command.RunPreflightChecks(ctx, checklist)
	if err != nil {
		c.ui.Error(err.Error())
		return command.PreflightError
	}

	// ==============================> GATHER METADATA <==============================

	_, metadataPkg, err := c.funcs.goList(ctx)
	if err != nil {
		c.ui.Warn(err.Error())
		// If metadata package not found, we simply skip it
	}

	gitSHA, gitBranch, err := c.services.git.HEAD()
	if err != nil {
		c.ui.Error(err.Error())
		return command.GitError
	}

	// Run semver command
	if code := c.commands.semver.Run(nil); code != command.Success {
		return code
	}

	semver := c.commands.semver.SemVer()

	// ==============================> CONSTRUCT LD FLAGS <==============================

	var ldFlags string

	// Construct the LD flags only if the version package exist
	if metadataPkg != "" {
		goVersion := goVersionRE.FindString(info.GoVersion)
		buildTime := time.Now().UTC().Format(timeFormat)
		buildTool := "Basil"

		if metadata.Version != "" {
			buildTool += " " + metadata.Version
		}

		ldFlags = strings.Join([]string{
			fmt.Sprintf(`-X "%s.Version=%s"`, metadataPkg, semver),
			fmt.Sprintf(`-X "%s.Commit=%s"`, metadataPkg, gitSHA[:7]),
			fmt.Sprintf(`-X "%s.Branch=%s"`, metadataPkg, gitBranch),
			fmt.Sprintf(`-X "%s.GoVersion=%s"`, metadataPkg, goVersion),
			fmt.Sprintf(`-X "%s.BuildTool=%s"`, metadataPkg, buildTool),
			fmt.Sprintf(`-X "%s.BuildTime=%s"`, metadataPkg, buildTime),
		}, " ")
	}

	// ==============================> BUILD BINARIES <==============================

	cmdPath := fmt.Sprintf("./%s/", cmdDir)

	// By convention, we assume every directory inside cmd is a main package for a binary with the same name as the directory name.
	if _, err := os.Stat(cmdPath); err == nil {
		files, err := ioutil.ReadDir(cmdPath)
		if err != nil {
			c.ui.Error(err.Error())
			return command.OSError
		}

		for _, file := range files {
			if file.IsDir() {
				mainPkg := cmdPath + file.Name()
				output := binPath + file.Name()

				if err := c.buildAll(ctx, ldFlags, mainPkg, output); err != nil {
					c.ui.Error(err.Error())
					return command.GoError
				}
			}
		}
	}

	// We also assume the current directory is a main package if it contains a main.go file.
	if _, err := os.Stat("./main.go"); err == nil {
		mainPkg := "."
		output := binPath + filepath.Base(info.WorkingDirectory)

		if err := c.buildAll(ctx, ldFlags, mainPkg, output); err != nil {
			c.ui.Error(err.Error())
			return command.GoError
		}
	}

	if len(c.outputs.artifacts) == 0 {
		c.ui.Warn("No main package found.")
		c.ui.Warn("Run basil project build -help for more information.")
	}

	// ==============================> DONE <==============================

	return command.Success
}

func (c *Command) buildAll(ctx context.Context, ldFlags, mainPkg, output string) error {
	if !c.spec.Project.Build.CrossCompile {
		return c.build(ctx, "", "", ldFlags, mainPkg, output)
	}

	// Cross-compiling
	group, groupCtx := errgroup.WithContext(ctx)
	for _, platform := range c.spec.Project.Build.Platforms {
		output := output + "-" + platform
		vals := strings.Split(platform, "-")

		group.Go(func() error {
			return c.build(groupCtx, vals[0], vals[1], ldFlags, mainPkg, output)
		})
	}

	return group.Wait()
}

func (c *Command) build(ctx context.Context, os, arch, ldFlags, mainPkg, output string) error {
	opts := shell.RunOptions{
		Environment: map[string]string{
			"GOOS":   os,
			"GOARCH": arch,
		},
	}

	args := []string{}
	if ldFlags != "" {
		args = append(args, "-ldflags", ldFlags)
	}
	if output != "" {
		args = append(args, "-o", output)
	}
	args = append(args, mainPkg)

	_, _, err := c.funcs.goBuild(ctx, opts, args...)
	if err != nil {
		return err
	}

	c.Mutex.Lock()
	c.outputs.artifacts = append(c.outputs.artifacts, Artifact{
		Path: output,
	})
	c.Mutex.Unlock()

	c.ui.Output("🍨 " + output)

	return nil
}

// Artifacts returns the build artifacts after the command is run.
func (c *Command) Artifacts() []Artifact {
	return c.outputs.artifacts
}