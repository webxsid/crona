# Repository Guidelines

## Project Structure & Module Organization

Crona is a Go workspace split into four main modules. `shared/` contains shared DTOs, protocol constants, version metadata, and utilities. `kernel/` is the local daemon, including SQLite storage, IPC handlers, timers, alerts, exports, and e2e tests. `tui/` contains the Bubble Tea terminal UI, and `cli/` contains scriptable command entry points. Bundled templates and alert sounds live under `assets/`. User and contributor docs live in `docs/`; release notes are in `docs/release-notes/`.

## Build, Test, and Development Commands

- `make build`: builds shared, daemon, TUI, and CLI modules.
- `make run-daemon`: runs the local daemon from `kernel/cmd/crona-kernel`.
- `make run-tui`: runs the terminal UI with local binaries on `PATH`.
- `make test-unit`: runs all non-e2e Go tests across modules.
- `make test-e2e`: runs daemon IPC e2e tests with the `e2e` build tag.
- `make ci`: runs release metadata checks, unit tests, vet, lint, and coverage.
- `make release-check`: validates version metadata and release notes consistency.

Use `GOCACHE=/tmp/crona-go-cache` when running Go commands directly, matching the Makefile default.

## Coding Style & Naming Conventions

Use idiomatic Go formatting. Run `make fmt` before larger changes; it applies `gofmt` and `golines`. Keep package names short and lowercase. Prefer existing local patterns over new abstractions, especially around IPC request handling, Bubble Tea model updates, and export/report types. Public protocol names live in `shared/protocol` and should stay stable unless a compatibility change is intentional.

## Testing Guidelines

Go tests use the standard `testing` package. Place tests next to the package they cover and name them `Test...`. Use focused package tests while developing, for example `go test ./kernel/internal/notify`, then run `make test-unit` or `make ci` before release-facing changes. Use `make test-e2e` when touching daemon startup, shutdown, IPC transport, or runtime files.

## Commit & Pull Request Guidelines

Recent history uses short imperative commit subjects, sometimes with a scope, for example `Notify on export completion` or `refactor(tui): simplify session action labels`. Keep commits focused and avoid mixing unrelated docs, UI, and daemon changes. PRs should describe the behavior change, list validation commands run, link relevant issues, and include screenshots or terminal captures for visible TUI changes.

## Security & Configuration Tips

Crona is local-first. Do not introduce network dependencies into core workflows without an explicit product reason. Runtime state lives under Crona's resolved runtime directory; avoid hardcoding user paths. Keep destructive operations explicit and guarded, especially deletes, runtime wipes, and release packaging scripts.
