# Changelog

All notable changes to **Crona** are documented here.

## [Unreleased]

### Added
- Wellbeing tracking flow with daily check-ins for mood, energy, sleep hours, sleep score, screen time, and notes.
- Bubble Tea `Wellbeing` view with per-day check-in details, rolling metrics, streak summaries, and burnout status.
- Habit management across kernel and TUI, including create/edit/delete flows, due-by-date queries, and completion history.
- Daily dashboard habit lane with completion, failure, and optional duration logging.
- Kernel metrics APIs for date-range rollups, burnout indicators, and focus/check-in streak summaries.
- Kernel e2e coverage for daily check-ins, metrics, and persisted sort settings.

### Changed
- Daily dashboard now combines planned issues with due habits for the selected date.
- Repo, stream, and issue ordering is now user-configurable through persisted sort settings in core settings.
- Default issue scoping and create/checkout dialogs now prefill from the active repo/stream context when available.
- Roadmap documentation now reflects the implemented Phase 2 check-ins, metrics, and habit work present in the current branch.

### API / Core
- Added kernel RPC methods for habit CRUD, habit completion/uncompletion/history, daily check-in CRUD/range, and metrics range/rollup/streak queries.
- Added daily check-in persistence plus habit and habit-completion repositories to the SQLite kernel store.
- Added shared domain types and DTOs for habits, check-ins, metrics rollups, streaks, and burnout indicators.
- Added persisted `repoSort`, `streamSort`, and `issueSort` core settings that drive kernel list ordering.

### Verification
- `make build` passes for the current workspace.
- `make test` passes for `kernel`.

## [0.1.0-beta.2] - 2026-03-14

### Added
- Go monorepo workspace with `kernel`, `tui`, `cli`, and `shared`.
- Go TUI workspace with Default, Meta, Session History, Active Session, Scratchpads, Ops, Settings, and Daily Dashboard views.
- Session-focused workflow from issue panes with auto-context checkout, session lock, stash/end prompts, and scratchpad access during active sessions.
- Session detail overlay in Session History, with richer kernel-backed session context and amend entrypoint.
- Daily Dashboard with date navigation, planned-task list, worked-vs-estimate stats, and resolved-task progress.
- UI-local filtering across repos, streams, issues, scratchpads, and ops.
- Searchable repo and stream selectors in the Default issue-create dialog.
- Optional due date on issue creation, with a calendar picker in the Go TUI dialogs.
- Issue due-date picker action from issue tables/lists, backed by a date-aware todo API.
- Kernel shutdown hotkey from the Go TUI.
- Idle-only stash dialog in the TUI with stash pop/apply.
- Root `.env`-driven runtime mode plus dev-only seed / clear workflows.
- Root `Makefile` and helper scripts for workspace tasks and dev data management.
- Release builder and TUI installer flow for shipping standalone `crona-tui` and `crona-kernel` binaries.
- Go end-to-end tests under `kernel/e2e`.

### Changed
- Repo, stream, and issue public IDs now use numeric IDs.
- The entire local runtime path moved from Node/HTTP to Go/Unix socket IPC.
- Kernel auto-launch now prefers an adjacent Go kernel binary and falls back to repo-local `go run` when developing from source.
- Scratchpad reading now stays confined to its pane instead of taking over the full screen.
- Scratchpad editing now opens the real file under the kernel scratch directory, with `.md` fallback when metadata paths omit the extension.
- Scratchpad previews render markdown again after fixing the reload path.
- Pane sizing now uses fixed sidebar/content budgeting and deterministic vertical splits instead of letting overlays and narrow terminals distort row math.
- Default issues are now prioritized by due/open work, split into active vs completed panes, and support direct `1/2`/`tab` section switching.
- Meta issues now show lifecycle status inline.
- Ops moved from a plain list to a table and now load newest-first.
- Ops fetch size is user-adjustable instead of fixed.
- View navigation moved from a top bar to a grouped left sidebar.
- Header was simplified back to a stable context row plus an active-session row.
- Issue lifecycle actions now follow the core transition rules, with one cycle key and explicit abandon behavior.
- Session progress uses cumulative worked time for the active issue based on kernel session history.
- Focus-session start/end now drive issue status transitions through the kernel timer flow.
- Direct issue-status changes are now blocked while the same issue has an active focus session; the end-session transition flow now stops the timer before applying terminal statuses.
- Session amend is now exposed in the TUI as a commit-message rewrite flow for ended sessions.
- Status colors are applied consistently across issue lists and dashboard indicators.
- Release packaging now treats TUI and kernel as independent deliverables instead of bundling them together.

### Fixed
- Footer/status errors now render as transient toast overlays instead of permanently consuming layout space.
- `?` key help now opens as an overlay modal instead of expanding the footer and breaking small-screen layouts.
- Daily and Settings panes no longer overflow unpredictably on small terminals because row-height and list-window calculations now match the rendered layout.
- Session detail and help overlays now match the rest of the pane styling, and session-detail actions stay visible in a fixed footer.
- Dev seed data now follows the current issue lifecycle rules.
- Stash restore no longer intermittently fails with `SQLITE_BUSY` under overlapping local kernel activity.
- Stash apply now fails cleanly while another focus session is active, without mutating context or consuming the stash.
- Focus-session auto-transition to `in_progress` now bypasses the active-session status guard used for manual changes.
- Structured timer boundaries are now recovered when the kernel restarts with an active session still persisted.
- Session timer acceleration caused by overlapping local tick loops.
- Meta issue switching now updates issue context correctly.
- Scratchpad editor saves now reload properly in the Go TUI.
- Go client repo creation now uses the correct kernel route.
- Go client ops loading now uses the kernel's latest-ops endpoint.
- Todo-for-date clearing now actually removes the stored date.
- Issue completion and abandonment timestamps are persisted for dashboard reporting.
- Commit-message dialogs no longer treat typed confirmation characters as submit/cancel.
- Focus-session start no longer races separate issue-status and timer writes in the TUI.

### API / Core
- Added shared Go contracts for domain types, DTOs, and Unix socket IPC envelopes.
- Added daily summary by arbitrary date in the kernel issue summary flow.
- Added kernel shutdown IPC support for TUI-triggered shutdown.
- Added session history and stash IPC consumption in the Go TUI.
- Added kernel session-detail IPC for the Session History overlay.
- Added `kernel.dev.seed` and `kernel.dev.clear` dev-only IPC methods guarded by `CRONA_ENV=Dev`.
- Migrated kernel storage, commands, timer, stash, scratchpad, and settings flows from TypeScript to Go.
- Switched the TUI from HTTP/SSE to Unix socket IPC.

### Verification
- `go build ./...` passes for `shared`, `kernel`, `tui`, and `cli`.
- `go test ./...` passes for `kernel`.
