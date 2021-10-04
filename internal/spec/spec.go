package spec

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

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

// FromFile reads specifications from a file.
// If no spec file is found, an empty spec will be returned.
func FromFile() (Spec, error) {
	var spec Spec

	for _, specFile := range specFiles {
		file, err := os.Open(specFile)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return Spec{}, err
		}
		defer file.Close()

		if ext := filepath.Ext(specFile); ext == ".yml" || ext == ".yaml" {
			err = yaml.NewDecoder(file).Decode(&spec)
		} else if ext == ".json" {
			err = json.NewDecoder(file).Decode(&spec)
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
	if b.CrossCompile && len(b.Platforms) == 0 {
		b.Platforms = defaultPlatforms
	}

	return b
}

// Release has the specifications for the release command.
type Release struct {
	Model ReleaseModel `json:"model" yaml:"model" flag:"model"`
}

// ReleaseModel is the type for the release model.
type ReleaseModel string

const (
	// ReleaseModelIndirect creates a release commit through a pull request.
	ReleaseModelIndirect ReleaseModel = "indirect"
	// ReleaseModelDirect creates a release commit and pushes it to the default branch.
	ReleaseModelDirect ReleaseModel = "direct"
)

// WithDefaults returns a new object with default values.
func (r Release) WithDefaults() Release {
	if r.Model == "" {
		r.Model = ReleaseModelIndirect
	}

	return r
}
