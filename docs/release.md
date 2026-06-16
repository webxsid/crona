# Release Process

Crona uses `main` as the only long-lived code branch.

Release builds come from version tags.

## Validation

Before tagging a release, run:

```bash
make ci
make test-e2e
goreleaser build --snapshot --clean
goreleaser release --snapshot --clean --skip=publish
make brew-test
```

`make ci` runs release metadata checks, unit tests, vet, lint, and coverage generation. `make test-e2e` runs the local engine IPC e2e suite and requires an environment that permits local Unix sockets or Windows named pipes.

## Version Metadata

The release version must stay consistent across:

- `Makefile`
- `shared/version/version.go`
- `docs/release-notes/<tag>.md`

`make release-check` validates these references, confirms the matching release notes file exists, and keeps the protocol version pinned to `1.0` until an external GUI compatibility requirement forces a protocol bump.

## Publishing

1. Update version metadata and release notes.
2. Commit the release prep.
3. Tag the commit with a version tag such as `v1.0.0`.
4. Push the tag.

The release workflow runs tests, invokes GoReleaser, publishes GitHub Releases, uploads the legacy installer scripts plus shared assets tarball, then normalizes the GitHub release state based on tag shape before pushing the Homebrew tap update to `webxsid/homebrew-tap` and the stable winget manifest update to the configured winget-pkgs fork. Stable tags become latest releases and publish `Formula/crona.rb`; `-beta` tags become prereleases and publish `Formula/crona-beta.rb`. The canonical binary source remains GitHub Releases, and the TUI and CLI keep using the release body and source-aware update command.

The isolated Homebrew validation workflow runs in CI on both macOS and Linux so formula and archive issues are caught before tagging.

For local release validation and tap testing, see [distribution.md](distribution.md).
For end-user migration between install methods, see [migration.md](migration.md).

## Branch Cleanup

Keep `main` as the only long-lived branch. Delete merged or stale `release/*`, feature, and dependabot branches after they are no longer needed.
