// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"testing"
)

// TestParseVersion covers the single place the reported version now comes
// from, including the build-metadata and pre-release forms the release
// tooling produces.
func TestParseVersion(t *testing.T) {
	tests := []struct {
		name       string
		in         string
		major      uint
		minor      uint
		patch      uint
		preRelease string
	}{
		{name: "release", in: "0.26.2", major: 0, minor: 26, patch: 2},
		{name: "pre-release", in: "0.26.1-alpha", major: 0, minor: 26, patch: 1, preRelease: "alpha"},
		{name: "dotted pre-release", in: "1.2.3-rc.1", major: 1, minor: 2, patch: 3, preRelease: "rc.1"},
		{name: "build metadata", in: "0.26.2+a1b2c3d", major: 0, minor: 26, patch: 2},
		{name: "pre-release and build metadata", in: "0.26.2-beta+a1b2c3d", major: 0, minor: 26, patch: 2, preRelease: "beta"},
		{name: "development placeholder", in: devVersion, major: 0, minor: 0, patch: 0, preRelease: "dev"},
		{name: "too few components", in: "1.2"},
		{name: "not numeric", in: "notaversion"},
		{name: "empty", in: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			major, minor, patch, preRelease := parseVersion(test.in)
			if major != test.major || minor != test.minor || patch != test.patch {
				t.Errorf("parseVersion(%q) components = %d.%d.%d, want %d.%d.%d",
					test.in, major, minor, patch, test.major, test.minor, test.patch)
			}
			if preRelease != test.preRelease {
				t.Errorf("parseVersion(%q) preRelease = %q, want %q",
					test.in, preRelease, test.preRelease)
			}
		})
	}
}

// TestVersionReportsInjectedValue asserts that an injected version is what
// gets reported, which is the whole point of injecting it.
func TestVersionReportsInjectedValue(t *testing.T) {
	defer func(v, b string) { appVersion, appBuild = v, b }(appVersion, appBuild)

	appVersion, appBuild = "0.26.2", "a1b2c3d"
	if got, want := version(), "0.26.2+a1b2c3d"; got != want {
		t.Errorf("version() = %q, want %q", got, want)
	}

	appVersion, appBuild = "", ""
	if got, want := version(), devVersion; got != want {
		t.Errorf("version() with nothing injected = %q, want %q", got, want)
	}
}

// TestRPCSubVersionString guards the trailing-hyphen case: the string used to
// be formatted as if a pre-release label were always present, so a release
// without one advertised "/lokid:0.26.2-/" to peers.
func TestRPCSubVersionString(t *testing.T) {
	defer func(ma, mi, pa uint, pr, b string) {
		appMajor, appMinor, appPatch, appPreRelease, appBuild = ma, mi, pa, pr, b
	}(appMajor, appMinor, appPatch, appPreRelease, appBuild)

	tests := []struct {
		name string
		pre  string
		buil string
		want string
	}{
		{name: "no pre-release", want: "/lokid:0.26.2/"},
		{name: "pre-release", pre: "alpha", want: "/lokid:0.26.2-alpha/"},
		{name: "build only", buil: "a1b2c3d", want: "/lokid:0.26.2+a1b2c3d/"},
		{name: "pre-release and build", pre: "alpha", buil: "a1b2c3d", want: "/lokid:0.26.2-alpha+a1b2c3d/"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			appMajor, appMinor, appPatch = 0, 26, 2
			appPreRelease, appBuild = test.pre, test.buil
			if got := rpcSubVersionString(); got != test.want {
				t.Errorf("rpcSubVersionString() = %q, want %q", got, test.want)
			}
		})
	}
}

// TestUserAgentVersionMatchesComponents guards initialization order.
// userAgentVersion in server.go is a package-level variable computed from
// appMajor/appMinor/appPatch, which are themselves variable initializers
// derived from appVersion. If those ever moved into init(), userAgentVersion
// would be built from zero values instead. Run with an injected version to
// exercise the case that can actually distinguish the two:
//
//	go test -ldflags "-X main.appVersion=1.2.3" -run TestUserAgentVersionMatchesComponents
func TestUserAgentVersionMatchesComponents(t *testing.T) {
	want := fmt.Sprintf("%d.%d.%d", appMajor, appMinor, appPatch)
	if userAgentVersion != want {
		t.Errorf("userAgentVersion = %q, want %q (initialization order regression)",
			userAgentVersion, want)
	}
}
