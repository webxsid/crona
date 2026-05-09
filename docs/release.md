# Release Process

Crona uses `main` as the only long-lived code branch.

Release builds may come from `main`, short-lived `release/*` branches, or version tags. These branches run the same validation, and the release workflow publishes prereleases by default. Stable promotion is manual.

## Validation

Before tagging a release, run:

```bash
make ci
make test-e2e
make release VERSION=v1.0.0
```

`make ci` runs release metadata checks, unit tests, vet, lint, and coverage generation. `make test-e2e` runs the local engine IPC e2e suite and requires an environment that permits local Unix sockets or Windows named pipes.

## Version Metadata

The release version must stay consistent across:

- `Makefile`
- `shared/version/version.go`
- `README.md`
- `docs/install.md`
- `docs/changelog.md`

`make release-check` validates these references and keeps the protocol version pinned to `1.0` until an external GUI compatibility requirement forces a protocol bump.

## Publishing

1. Update version metadata and changelog.
2. Commit the release prep.
3. Tag the commit with a version tag such as `v1.0.0`.
4. Push the tag.

The release workflow builds cross-platform bundles, installer scripts, bundled assets, checksums, and size reports, then attaches public release notes from `docs/release-notes/<tag>.md` when present. The changelog remains the internal development log and a fallback source during the transition. Prereleases are the default publication mode; stable releases are promoted manually.

## Branch Cleanup

Keep `main` as the only long-lived branch. Delete merged or stale `release/*`, feature, and dependabot branches after they are no longer needed.
