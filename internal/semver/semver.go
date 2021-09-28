// Package semver provides functionalities for working with semantic versions.
// For more information about semantic versioning, visit https://semver.org.
package semver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// SemVer represents a semantic version.
type SemVer struct {
	Major      uint
	Minor      uint
	Patch      uint
	Prerelease []string
	Metadata   []string
}

// Parse gets a semantic version string and returns a SemVer.
// If the second return value is false, it implies that the input semver was incorrect.
func Parse(semver string) (SemVer, bool) {
	var major, minor, patch uint64
	var prerelease, metadata []string

	// Make sure the string is a valid semantic version
	if re := regexp.MustCompile(`^v?[0-9]+\.[0-9]+\.[0-9]+(\-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$`); !re.MatchString(semver) {
		return SemVer{}, false
	}

	re := regexp.MustCompile(`^v?([0-9]+)\.([0-9]+)\.([0-9]+)(\-[\.[0-9A-Za-z-]+)?(\+[\.[0-9A-Za-z-]*)?$`)
	subs := re.FindStringSubmatch(semver)

	major, _ = strconv.ParseUint(subs[1], 10, 64)
	minor, _ = strconv.ParseUint(subs[2], 10, 64)
	patch, _ = strconv.ParseUint(subs[3], 10, 64)

	if subs[4] != "" {
		prerelease = strings.Split(subs[4][1:], ".")
	}

	if subs[5] != "" {
		metadata = strings.Split(subs[5][1:], ".")
	}

	return SemVer{
		Major:      uint(major),
		Minor:      uint(minor),
		Patch:      uint(patch),
		Prerelease: prerelease,
		Metadata:   metadata,
	}, true
}

// Next creates a new semantic version by increasing the patch version by one.
func (v SemVer) Next() SemVer {
	return SemVer{
		Major: v.Major,
		Minor: v.Minor,
		Patch: v.Patch + 1,
	}
}

// ReleasePatch creates a new semantic version for a patch release.
func (v SemVer) ReleasePatch() SemVer {
	return SemVer{
		Major: v.Major,
		Minor: v.Minor,
		Patch: v.Patch,
	}
}

// ReleaseMinor creates a new semantic version for a minor release.
func (v SemVer) ReleaseMinor() SemVer {
	return SemVer{
		Major: v.Major,
		Minor: v.Minor + 1,
		Patch: 0,
	}
}

// ReleaseMajor creates a new semantic version for a major release.
func (v SemVer) ReleaseMajor() SemVer {
	return SemVer{
		Major: v.Major + 1,
		Minor: 0,
		Patch: 0,
	}
}

// String returns the string representation of the current semantic version.
func (v SemVer) String() string {
	var tail string

	if len(v.Prerelease) > 0 {
		tail += "-" + strings.Join(v.Prerelease, ".")
	}

	if len(v.Metadata) > 0 {
		tail += "+" + strings.Join(v.Metadata, ".")
	}

	return fmt.Sprintf("%d.%d.%d%s", v.Major, v.Minor, v.Patch, tail)
}
