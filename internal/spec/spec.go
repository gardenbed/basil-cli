package spec

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

var (
	specFiles        = []string{"basil.yml", "basil.yaml", "basil.json"}
	defaultPlatforms = []string{"linux-386", "linux-amd64", "linux-arm", "linux-arm64", "darwin-amd64", "windows-386", "windows-amd64"}
)

// Spec is the model for all specifications.
type Spec struct {
	Version string  `json:"version" yaml:"version"`
	Project Project `json:"project" yaml:"project"`
}

// Read reads specifications from a file.
// If no spec file is found, an empty spec will be returned.
func Read() (Spec, error) {
	var spec Spec

	for _, specFile := range specFiles {
		f, err := os.Open(specFile)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return Spec{}, err
		}
		defer f.Close()

		if ext := filepath.Ext(specFile); ext == ".yml" || ext == ".yaml" {
			err = yaml.NewDecoder(f).Decode(&spec)
		} else if ext == ".json" {
			err = json.NewDecoder(f).Decode(&spec)
		} else {
			err = errors.New("unknown spec file")
		}

		if err != nil {
			return Spec{}, err
		}

		break
	}

	return spec, nil
}

// WithDefaults returns a new object with default values.
func (s Spec) WithDefaults() Spec {
	if s.Version == "" {
		s.Version = "1.0"
	}

	s.Project = s.Project.WithDefaults()

	return s
}

// Project has the specifications for a Basil project.
type Project struct {
	Owner    string          `json:"owner" yaml:"owner"`
	Language ProjectLanguage `json:"language" yaml:"language"`
	Profile  ProjectProfile  `json:"profile" yaml:"profile"`
	Build    Build           `json:"build" yaml:"build"`
	Release  Release         `json:"release" yaml:"release"`
}

// ProjectLanguage is the type for the project language.
type ProjectLanguage string

const (
	// ProjectLanguageGo represents the Go programming language.
	ProjectLanguageGo ProjectLanguage = "go"
)

// ProjectProfile is the type for the project profile.
type ProjectProfile string

const (
	// ProjectProfileGeneric represents a generic application/library.
	ProjectProfileGeneric ProjectProfile = "generic"
	// ProjectProfileLibrary represents a library/package.
	ProjectProfileLibrary ProjectProfile = "library"
	// ProjectProfileCLI represents a command-line application.
	ProjectProfileCLI ProjectProfile = "command-line-app"
	// ProjectProfileGRPCService represents a gRPC service.
	ProjectProfileGRPCService ProjectProfile = "grpc-service"
	// ProjectProfileGRPCServiceHorizontal represents a gRPC service.
	ProjectProfileGRPCServiceHorizontal ProjectProfile = "grpc-service-horizontal"
	// ProjectProfileHTTPService represents an HTTP service.
	ProjectProfileHTTPService ProjectProfile = "http-service"
	// ProjectProfileHTTPServiceHorizontal represents an HTTP service.
	ProjectProfileHTTPServiceHorizontal ProjectProfile = "http-service-horizontal"
)

// WithDefaults returns a new object with default values.
func (p Project) WithDefaults() Project {
	if p.Language == "" {
		p.Language = ProjectLanguageGo
	}

	if p.Profile == "" {
		p.Profile = ProjectProfileGeneric
	}

	p.Build = p.Build.WithDefaults()
	p.Release = p.Release.WithDefaults()

	return p
}

// Build has the specifications for the build command.
type Build struct {
	CrossCompile bool     `json:"crossCompile" yaml:"cross_compile" flag:"cross-compile"`
	Platforms    []string `json:"platforms" yaml:"platforms" flag:"platforms"`
}

// WithDefaults returns a new object with default values.
func (b Build) WithDefaults() Build {
	if len(b.Platforms) == 0 {
		b.Platforms = defaultPlatforms
	}

	return b
}

// Release has the specifications for the release command.
type Release struct {
	Mode ReleaseMode `json:"mode" yaml:"mode" flag:"mode"`
}

// ReleaseModelis the type for the release mode.
type ReleaseMode string

const (
	// ReleaseModeIndirect creates a release commit through a pull request.
	ReleaseModeIndirect ReleaseMode = "indirect"
	// ReleaseModeDirect creates a release commit and pushes it to the default branch.
	ReleaseModeDirect ReleaseMode = "direct"
)

// String returns the release mode in upper case.
func (m ReleaseMode) String() string {
	return strings.ToUpper(string(m))
}

// WithDefaults returns a new object with default values.
func (r Release) WithDefaults() Release {
	if r.Mode == "" {
		r.Mode = ReleaseModeIndirect
	}

	return r
}
