package semver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name           string
		semver         string
		expectedSemver SemVer
		expectedOK     bool
	}{
		{
			name:           "Empty",
			semver:         "",
			expectedSemver: SemVer{},
			expectedOK:     false,
		},
		{
			name:           "NoMinor",
			semver:         "1",
			expectedSemver: SemVer{},
			expectedOK:     false,
		},
		{
			name:           "NoPatch",
			semver:         "0.1",
			expectedSemver: SemVer{},
			expectedOK:     false,
		},
		{
			name:           "InvalidMajor",
			semver:         "X.1.0",
			expectedSemver: SemVer{},
			expectedOK:     false,
		},
		{
			name:           "InvalidMinor",
			semver:         "0.Y.0",
			expectedSemver: SemVer{},
			expectedOK:     false,
		},
		{
			name:           "InvalidPatch",
			semver:         "0.1.Z",
			expectedSemver: SemVer{},
			expectedOK:     false,
		},
		{
			name:           "InvalidPrerelease",
			semver:         "0.1.0-",
			expectedSemver: SemVer{},
			expectedOK:     false,
		},
		{
			name:           "InvalidPrerelease",
			semver:         "0.1.0-beta.",
			expectedSemver: SemVer{},
			expectedOK:     false,
		},
		{
			name:           "InvalidMetadata",
			semver:         "0.1.0-beta.1+",
			expectedSemver: SemVer{},
			expectedOK:     false,
		},
		{
			name:           "InvalidMetadata",
			semver:         "0.1.0-beta.1+20200818.",
			expectedSemver: SemVer{},
			expectedOK:     false,
		},
		{
			name:   "Release",
			semver: "0.1.0",
			expectedSemver: SemVer{
				Major: 0,
				Minor: 1,
				Patch: 0,
			},
			expectedOK: true,
		},
		{
			name:   "Release",
			semver: "v0.1.0",
			expectedSemver: SemVer{
				Major: 0,
				Minor: 1,
				Patch: 0,
			},
			expectedOK: true,
		},
		{
			name:   "WithPrerelease",
			semver: "0.1.0-beta",
			expectedSemver: SemVer{
				Major:      0,
				Minor:      1,
				Patch:      0,
				Prerelease: []string{"beta"},
			},
			expectedOK: true,
		},
		{
			name:   "WithPrerelease",
			semver: "v0.1.0-rc.1",
			expectedSemver: SemVer{
				Major:      0,
				Minor:      1,
				Patch:      0,
				Prerelease: []string{"rc", "1"},
			},
			expectedOK: true,
		},
		{
			name:   "WithMetadata",
			semver: "0.1.0+20200920",
			expectedSemver: SemVer{
				Major:    0,
				Minor:    1,
				Patch:    0,
				Metadata: []string{"20200920"},
			},
			expectedOK: true,
		},
		{
			name:   "WithMetadata",
			semver: "v0.1.0+sha.abcdeff",
			expectedSemver: SemVer{
				Major:    0,
				Minor:    1,
				Patch:    0,
				Metadata: []string{"sha", "abcdeff"},
			},
			expectedOK: true,
		},
		{
			name:   "WithPrereleaseAndMetadata",
			semver: "0.1.0-beta+20200920",
			expectedSemver: SemVer{
				Major:      0,
				Minor:      1,
				Patch:      0,
				Prerelease: []string{"beta"},
				Metadata:   []string{"20200920"},
			},
			expectedOK: true,
		},
		{
			name:   "WithPrereleaseAndMetadata",
			semver: "v0.1.0-rc.1+sha.abcdeff.20200920",
			expectedSemver: SemVer{
				Major:      0,
				Minor:      1,
				Patch:      0,
				Prerelease: []string{"rc", "1"},
				Metadata:   []string{"sha", "abcdeff", "20200920"},
			},
			expectedOK: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			semver, ok := Parse(tc.semver)

			assert.Equal(t, tc.expectedSemver, semver)
			assert.Equal(t, tc.expectedOK, ok)
		})
	}
}

func TestSemVer_Next(t *testing.T) {
	tests := []struct {
		semver       SemVer
		expectedNext SemVer
	}{
		{
			SemVer{},
			SemVer{Major: 0, Minor: 0, Patch: 1},
		},
		{
			SemVer{Major: 0, Minor: 1, Patch: 0},
			SemVer{Major: 0, Minor: 1, Patch: 1},
		},
		{
			SemVer{Major: 1, Minor: 0, Patch: 0},
			SemVer{Major: 1, Minor: 0, Patch: 1},
		},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.expectedNext, tc.semver.Next())
	}
}

func TestSemVer_ReleasePatch(t *testing.T) {
	tests := []struct {
		semver          SemVer
		expectedRelease SemVer
	}{
		{
			SemVer{},
			SemVer{Major: 0, Minor: 0, Patch: 0},
		},
		{
			SemVer{Major: 0, Minor: 1, Patch: 0},
			SemVer{Major: 0, Minor: 1, Patch: 0},
		},
		{
			SemVer{Major: 1, Minor: 2, Patch: 0},
			SemVer{Major: 1, Minor: 2, Patch: 0},
		},
	}

	for _, tc := range tests {
		release := tc.semver.ReleasePatch()

		assert.Equal(t, tc.expectedRelease, release)
	}
}

func TestSemVer_ReleaseMinor(t *testing.T) {
	tests := []struct {
		semver          SemVer
		expectedRelease SemVer
	}{
		{
			SemVer{},
			SemVer{Major: 0, Minor: 1, Patch: 0},
		},
		{
			SemVer{Major: 0, Minor: 1, Patch: 0},
			SemVer{Major: 0, Minor: 2, Patch: 0},
		},
		{
			SemVer{Major: 1, Minor: 2, Patch: 0},
			SemVer{Major: 1, Minor: 3, Patch: 0},
		},
	}

	for _, tc := range tests {
		release := tc.semver.ReleaseMinor()

		assert.Equal(t, tc.expectedRelease, release)
	}
}

func TestSemVer_ReleaseMajor(t *testing.T) {
	tests := []struct {
		semver          SemVer
		expectedRelease SemVer
	}{
		{
			SemVer{},
			SemVer{Major: 1, Minor: 0, Patch: 0},
		},
		{
			SemVer{Major: 0, Minor: 1, Patch: 0},
			SemVer{Major: 1, Minor: 0, Patch: 0},
		},
		{
			SemVer{Major: 1, Minor: 2, Patch: 0},
			SemVer{Major: 2, Minor: 0, Patch: 0},
		},
	}

	for _, tc := range tests {
		release := tc.semver.ReleaseMajor()

		assert.Equal(t, tc.expectedRelease, release)
	}
}

func TestSemVer_String(t *testing.T) {
	tests := []struct {
		name            string
		semver          SemVer
		expectedVersion string
	}{
		{
			name: "OK",
			semver: SemVer{
				Major: 0,
				Minor: 2,
				Patch: 7,
			},
			expectedVersion: "0.2.7",
		},
		{
			name: "WithPrerelease",
			semver: SemVer{
				Major:      0,
				Minor:      2,
				Patch:      7,
				Prerelease: []string{"rc", "1"},
			},
			expectedVersion: "0.2.7-rc.1",
		},
		{
			name: "WithMetadata",
			semver: SemVer{
				Major:    0,
				Minor:    2,
				Patch:    7,
				Metadata: []string{"20200820"},
			},
			expectedVersion: "0.2.7+20200820",
		},
		{
			name: "WithPrereleaseAndMetadata",
			semver: SemVer{
				Major:      0,
				Minor:      2,
				Patch:      7,
				Prerelease: []string{"rc", "1"},
				Metadata:   []string{"20200820"},
			},
			expectedVersion: "0.2.7-rc.1+20200820",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedVersion, tc.semver.String())
		})
	}
}

func TestSemVer_TagName(t *testing.T) {
	tests := []struct {
		name            string
		semver          SemVer
		expectedTagName string
	}{
		{
			name: "OK",
			semver: SemVer{
				Major: 0,
				Minor: 2,
				Patch: 7,
			},
			expectedTagName: "v0.2.7",
		},
		{
			name: "WithPrerelease",
			semver: SemVer{
				Major:      0,
				Minor:      2,
				Patch:      7,
				Prerelease: []string{"rc", "1"},
			},
			expectedTagName: "v0.2.7-rc.1",
		},
		{
			name: "WithMetadata",
			semver: SemVer{
				Major:    0,
				Minor:    2,
				Patch:    7,
				Metadata: []string{"20200820"},
			},
			expectedTagName: "v0.2.7+20200820",
		},
		{
			name: "WithPrereleaseAndMetadata",
			semver: SemVer{
				Major:      0,
				Minor:      2,
				Patch:      7,
				Prerelease: []string{"rc", "1"},
				Metadata:   []string{"20200820"},
			},
			expectedTagName: "v0.2.7-rc.1+20200820",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedTagName, tc.semver.TagName())
		})
	}
}
