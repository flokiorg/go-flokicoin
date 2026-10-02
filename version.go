// Copyright (c) 2013-2014 The btcsuite developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// semanticAlphabet
const semanticAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"

// appVersion is the application version, set at release-build time with
// '-ldflags "-X main.appVersion=0.26.2"' from the release tag. It is the only
// place the version is recorded, so what this binary reports cannot drift from
// what was actually published. A plain `go build` leaves it empty and
// devVersion is reported instead.
var appVersion string

// devVersion is reported by builds that had no version injected. It is
// deliberately not a real version number, so a development build is never
// mistaken for a release.
const devVersion = "0.0.0-dev"

// appBuild is defined as a variable so it can be overridden during the build
// process with '-ldflags "-X main.appBuild foo' if needed.  It MUST only
// contain characters from semanticAlphabet per the semantic versioning spec.
var appBuild string

// The numeric components and pre-release label of the effective version, used
// by the P2P user agent and the version RPCs. These are derived from
// appVersion rather than maintained by hand. They are variable initializers
// rather than assignments in init() so that package-level variables which
// depend on them -- userAgentVersion in server.go -- are guaranteed to be
// initialized after them.
var appMajor, appMinor, appPatch, appPreRelease = parseVersion(effectiveVersion())

// effectiveVersion returns the injected version, or the development
// placeholder when nothing was injected.
func effectiveVersion() string {
	if appVersion != "" {
		return appVersion
	}
	return devVersion
}

// parseVersion splits a semantic version string into its numeric components
// and pre-release label, discarding any build metadata. A string that does not
// parse yields zeroed components, which is preferable to reporting numbers
// that were never released.
func parseVersion(v string) (major, minor, patch uint, preRelease string) {
	if i := strings.IndexByte(v, '+'); i >= 0 {
		v = v[:i]
	}
	if i := strings.IndexByte(v, '-'); i >= 0 {
		preRelease, v = v[i+1:], v[:i]
	}

	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return 0, 0, 0, preRelease
	}

	var nums [3]uint
	for i, part := range parts {
		n, err := strconv.ParseUint(part, 10, 32)
		if err != nil {
			return 0, 0, 0, preRelease
		}
		nums[i] = uint(n)
	}

	return nums[0], nums[1], nums[2], preRelease
}

// version returns the application version as a properly formed string per the
// semantic versioning 2.0.0 spec (http://semver.org/).
func version() string {
	version := effectiveVersion()

	// Append build metadata if there is any.  The plus called for
	// by the semantic versioning spec is automatically appended and should
	// not be contained in the build metadata string.  The build metadata
	// string is not appended if it contains invalid characters.
	build := normalizeVerString(appBuild)
	if build != "" {
		version = fmt.Sprintf("%s+%s", version, build)
	}

	return version
}

// normalizeVerString returns the passed string stripped of all characters which
// are not valid according to the semantic versioning guidelines for pre-release
// version and build metadata strings.  In particular they MUST only contain
// characters in semanticAlphabet.
func normalizeVerString(str string) string {
	var result bytes.Buffer
	for _, r := range str {
		if strings.ContainsRune(semanticAlphabet, r) {
			result.WriteRune(r)
		}
	}
	return result.String()
}
