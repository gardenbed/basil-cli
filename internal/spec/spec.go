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
	Language Language `json:"language" yaml:"language"`
	Profile  Profile  `json:"profile" yaml:"profile"`
	Build    Build    `json:"build" yaml:"build"`
}

// Language is the type for the project language.
type Language string

const (
	// LanguageGo represents the Go programming language.
	LanguageGo Language = "go"
)

// Profile is the type for the project profile.
type Profile string

const (
	// ProfileGeneric represents a generic application/library.
	ProfileGeneric Profile = "generic"
)

// WithDefaults returns a new object with default values.
func (p Project) WithDefaults() Project {
	if p.Language == "" {
		p.Language = LanguageGo
	}

	if p.Profile == "" {
		p.Profile = ProfileGeneric
	}

	p.Build = p.Build.WithDefaults()

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
