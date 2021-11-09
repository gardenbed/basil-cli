package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	tests := []struct {
		name          string
		specFiles     []string
		expectedSpec  Spec
		expectedError string
	}{
		{
			name:         "NoSpecFile",
			specFiles:    []string{"test/null"},
			expectedSpec: Spec{},
		},
		{
			name:          "UnknownFile",
			specFiles:     []string{"test/unknown"},
			expectedError: "unknown spec file",
		},
		{
			name:          "EmptyJSON",
			specFiles:     []string{"test/empty.json"},
			expectedError: "EOF",
		},
		{
			name:          "EmptyYAML",
			specFiles:     []string{"test/empty.yaml"},
			expectedError: "EOF",
		},
		{
			name:          "InvalidJSON",
			specFiles:     []string{"test/invalid.json"},
			expectedError: "invalid character",
		},
		{
			name:          "InvalidYAML",
			specFiles:     []string{"test/invalid.yaml"},
			expectedError: "cannot unmarshal",
		},
		{
			name:      "ValidJSON",
			specFiles: []string{"test/valid.json"},
			expectedSpec: Spec{
				Version: "1.0",
				Project: Project{
					Owner:    "my-team",
					Language: ProjectLanguageGo,
					Profile:  ProjectProfileGeneric,
					Build: Build{
						CrossCompile: true,
						Platforms:    []string{"linux-386", "linux-amd64", "linux-arm", "linux-arm64", "darwin-amd64", "windows-386", "windows-amd64"},
					},
					Release: Release{
						Mode: ReleaseModeDirect,
					},
				},
			},
		},
		{
			name:      "ValidYAML",
			specFiles: []string{"test/valid.yaml"},
			expectedSpec: Spec{
				Version: "1.0",
				Project: Project{
					Owner:    "my-team",
					Language: ProjectLanguageGo,
					Profile:  ProjectProfileGeneric,
					Build: Build{
						CrossCompile: true,
						Platforms:    []string{"linux-386", "linux-amd64", "linux-arm", "linux-arm64", "darwin-amd64", "windows-386", "windows-amd64"},
					},
					Release: Release{
						Mode: ReleaseModeDirect,
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			specFiles = tc.specFiles
			spec, err := Read()

			if tc.expectedError != "" {
				assert.Empty(t, spec)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedSpec, spec)
			}
		})
	}
}

func TestSpec_WithDefaults(t *testing.T) {
	tests := []struct {
		name         string
		spec         Spec
		expectedSpec Spec
	}{
		{
			"DefaultsRequired",
			Spec{},
			Spec{
				Version: "1.0",
				Project: Project{
					Language: ProjectLanguageGo,
					Profile:  ProjectProfileGeneric,
					Build: Build{
						Platforms: defaultPlatforms,
					},
					Release: Release{
						Mode: ReleaseModeIndirect,
					},
				},
			},
		},
		{
			"DefaultsNotRequired",
			Spec{
				Version: "2.0",
				Project: Project{
					Owner:    "my-team",
					Language: ProjectLanguageGo,
					Profile:  ProjectProfileGeneric,
					Build: Build{
						CrossCompile: true,
						Platforms:    []string{"linux-amd64", "darwin-amd64", "windows-amd64"},
					},
					Release: Release{
						Mode: ReleaseModeDirect,
					},
				},
			},
			Spec{
				Version: "2.0",
				Project: Project{
					Owner:    "my-team",
					Language: ProjectLanguageGo,
					Profile:  ProjectProfileGeneric,
					Build: Build{
						CrossCompile: true,
						Platforms:    []string{"linux-amd64", "darwin-amd64", "windows-amd64"},
					},
					Release: Release{
						Mode: ReleaseModeDirect,
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedSpec, tc.spec.WithDefaults())
		})
	}
}

func TestProject_WithDefaults(t *testing.T) {
	tests := []struct {
		name            string
		project         Project
		expectedProject Project
	}{
		{
			"DefaultsRequired",
			Project{},
			Project{
				Language: ProjectLanguageGo,
				Profile:  ProjectProfileGeneric,
				Build: Build{
					Platforms: defaultPlatforms,
				},
				Release: Release{
					Mode: ReleaseModeIndirect,
				},
			},
		},
		{
			"DefaultsNotRequired",
			Project{
				Owner:    "my-team",
				Language: ProjectLanguageGo,
				Profile:  ProjectProfileGeneric,
				Build: Build{
					CrossCompile: true,
					Platforms:    []string{"linux-amd64", "darwin-amd64", "windows-amd64"},
				},
				Release: Release{
					Mode: ReleaseModeDirect,
				},
			},
			Project{
				Owner:    "my-team",
				Language: ProjectLanguageGo,
				Profile:  ProjectProfileGeneric,
				Build: Build{
					CrossCompile: true,
					Platforms:    []string{"linux-amd64", "darwin-amd64", "windows-amd64"},
				},
				Release: Release{
					Mode: ReleaseModeDirect,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedProject, tc.project.WithDefaults())
		})
	}
}

func TestBuild_WithDefaults(t *testing.T) {
	tests := []struct {
		name          string
		build         Build
		expectedBuild Build
	}{
		{
			"DefaultsRequired",
			Build{},
			Build{
				Platforms: defaultPlatforms,
			},
		},
		{
			"DefaultsNotRequired",
			Build{
				CrossCompile: true,
				Platforms:    []string{"linux-amd64", "darwin-amd64", "windows-amd64"},
			},
			Build{
				CrossCompile: true,
				Platforms:    []string{"linux-amd64", "darwin-amd64", "windows-amd64"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedBuild, tc.build.WithDefaults())
		})
	}
}

func TestReleaseMode_String(t *testing.T) {
	assert.Equal(t, "INDIRECT", ReleaseModeIndirect.String())
	assert.Equal(t, "DIRECT", ReleaseModeDirect.String())
}

func TestRelease_WithDefaults(t *testing.T) {
	tests := []struct {
		name            string
		release         Release
		expectedRelease Release
	}{
		{
			"DefaultsRequired",
			Release{},
			Release{
				Mode: ReleaseModeIndirect,
			},
		},
		{
			"DefaultsNotRequired",
			Release{
				Mode: ReleaseModeDirect,
			},
			Release{
				Mode: ReleaseModeDirect,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedRelease, tc.release.WithDefaults())
		})
	}
}
