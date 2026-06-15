# Distribution

This page covers local validation of Crona's GoReleaser release flow and the isolated Homebrew workflow on macOS and Linux.

## A. GoReleaser Build-Only Test

```bash
goreleaser build --snapshot --clean
```

Expected:

- binaries appear under `dist/`
- no GitHub release is created
- no Homebrew formula is published

## B. GoReleaser Snapshot Release Test

```bash
goreleaser release --snapshot --clean --skip=publish
```

Expected:

- release archives are generated
- `checksums.txt` is generated
- nothing is uploaded

## C. Isolated Homebrew Validation

Use the repo-managed workflow so the developer's real Crona installation is not touched:

```bash
make brew-test
```

What it does:

- runs the GoReleaser snapshot release flow
- generates a local tap under `/tmp/crona-homebrew-test`
- uses a configurable temporary tap name, defaulting to `crona/local-test`
- installs from local `dist/` artifacts via `file://` URLs
- validates `crona`, `crona-kernel`, and `crona-tui`
- can generate either `crona.rb` or `crona-beta.rb` for release-tag vs beta-tag publishes
- checks that update metadata reports:
  - `install-source: brew`
  - `update-command: brew upgrade crona`
- uninstalls and untaps the disposable tap
- removes `/tmp/crona-homebrew-test` on exit

Generate-only mode is also available:

```bash
make brew-generate
```

This leaves the temporary tap and generated formula in place for debugging.

## D. Local Formula Test

Point a local Homebrew formula at generated release artifacts with `file://` URLs.

Example flow:

```bash
shasum -a 256 dist/<archive>.zip
brew install --formula ./Formula/crona.rb
crona --version
brew uninstall crona
```

## E. Local Tap Test

```bash
mkdir -p ~/dev/homebrew-tap/Formula
cp Formula/crona.rb ~/dev/homebrew-tap/Formula/crona.rb
brew tap crona/local-test ~/dev/homebrew-tap
brew install crona/local-test/crona
crona --version
brew upgrade crona/local-test/crona
brew uninstall crona
brew untap crona/local-test
```

## F. Simulated Upgrade Test

Use the repo target for the upgrade simulation:

```bash
make brew-upgrade-test
```

The workflow builds a fake `v0.0.1-test`, installs it, then regenerates the formula for `v0.0.2-test` and runs:

```bash
brew upgrade crona
```

Confirm:

- `crona --version` changes
- the update view identifies the install source as Homebrew
- the update command shown is `brew upgrade crona`
