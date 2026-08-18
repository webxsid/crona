---
hosted: false
title: "Release Process"
description: "Release validation, version metadata, publishing, and branch cleanup."
order: 6.3
---

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

`make ci` runs release metadata checks, unit tests, vet, lint, and coverage generation. `make test-e2e` runs the local daemon IPC e2e suite and requires an environment that permits local Unix sockets or Windows named pipes.

## Version Metadata

The release version must stay consistent across:

- `Makefile`
- `shared/version/version.go`
- `docs/release-notes/<tag>.md`

`make release-check` validates these references, confirms the matching release notes file exists, and verifies the current local IPC protocol metadata. The current protocol version is `1.5`.

## Publishing

1. Update version metadata and release notes.
2. Commit the release prep.
3. Tag the commit with a version tag such as `v1.0.0`.
4. Push the tag.

The release workflow runs tests, invokes GoReleaser, and lets GoReleaser publish the GitHub release assets, including the legacy installer scripts and shared assets tarball, so the checksum file covers the full release asset set. It then overwrites the GitHub release notes from `docs/release-notes/<tag>.md`, normalizes the release state based on tag shape, and pushes the Homebrew tap update to `webxsid/homebrew-tap` plus the stable winget manifest update to the configured winget-pkgs fork. Stable tags refresh both `Formula/crona.rb` and `Formula/crona-beta.rb` so beta users can move onto the stable build without a manual uninstall/reinstall cycle; `-beta` tags publish only `Formula/crona-beta.rb`. The canonical binary source remains GitHub Releases, and the TUI and CLI keep using the release body and source-aware update command for guidance only.

The isolated Homebrew validation workflow runs in CI on both macOS and Linux so formula and archive issues are caught before tagging.


## Branch Cleanup

Keep `main` as the only long-lived branch. Delete merged or stale `release/*`, feature, and dependabot branches after they are no longer needed.
