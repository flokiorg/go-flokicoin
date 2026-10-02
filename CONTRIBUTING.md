# Contributing to go-flokicoin

## Building and testing

```sh
go build ./...
go vet ./...
go test ./...
```

These are the same three commands CI runs, so run them before opening a pull
request.

## Pull requests

- Keep each change focused; split unrelated work into separate pull requests.
- Add a `CHANGELOG.md` entry for anything that changes behaviour, under the
  topmost `## [X.Y.Z]` heading in the matching `### Added` / `### Changed` /
  `### Fixed` subsection. If the last release just shipped and no heading is
  open yet, add one with the version the change warrants.
- Once your pull request has a number, append `(#N)` to the changelog bullets it
  introduces. The release notes are generated from that text, so a bullet
  without its reference loses the link back to the discussion.

## Versioning

There is no `VERSION` file. `CHANGELOG.md` is the only place the version is
recorded, and the version is injected into both binaries at release-build time
from the git tag, via `main.appVersion`.

Nothing else should hold a copy of the version. `lokid`'s numeric components,
its P2P user agent and the `/lokid:.../` RPC sub-version string are all derived
from that one value — previously each was composed from hand-edited constants,
which is how the daemon ended up reporting 0.25.12-alpha and the CLI
0.23.2-beta long after those releases had passed.

A plain `go build` injects nothing, so a development build reports `0.0.0-dev`.
That is deliberate: a development build should never be mistaken for a release.

## How releases are cut

Releases are manual. `.github/workflows/release.yml` is `workflow_dispatch`-only
and does the tagging itself:

```sh
gh workflow run release.yml --repo flokiorg/go-flokicoin
```

It resolves the version from the topmost `## [X.Y.Z]` heading in `CHANGELOG.md`
(or from the optional `version` input, given as a bare number with no `v`),
re-runs the build/vet/test gate, extracts that changelog section as the release
notes, creates and pushes the annotated `vX.Y.Z` tag, then publishes the
`lokid` and `lokid-cli` binaries with GoReleaser.

Do not create the tag by hand — the workflow creates it, and a manual tag would
collide with the one it makes.
