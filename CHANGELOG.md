# Changelog

All notable changes to **Crona** are documented here.

## [Unreleased] - 2026-03-09

### Added
- Go TUI workspace with Default, Meta, Scratchpads, Ops, Session, and Daily Dashboard views.
- Session-focused workflow from issue panes with auto-context checkout, session lock, stash/end prompts, and scratchpad access during active sessions.
- Daily Dashboard with date navigation, planned-task list, worked-vs-estimate stats, and resolved-task progress.
- UI-local filtering across repos, streams, issues, scratchpads, and ops.
- Searchable repo and stream selectors in the Default issue-create dialog.
- Optional due date on issue creation, with a calendar picker in the Go TUI dialogs.
- Issue due-date picker action from issue tables/lists, backed by a date-aware todo API.
- Kernel shutdown hotkey from the Go TUI.

### Changed
- Scratchpad reading now stays confined to its pane instead of taking over the full screen.
- Scratchpad editing now opens the real file under the kernel scratch directory, with `.md` fallback when metadata paths omit the extension.
- Scratchpad previews render markdown again after fixing the reload path.
- Default issues are rendered as a table with status, estimate, repo, and stream metadata.
- Meta issues now show lifecycle status inline.
- Ops moved from a plain list to a table and now load newest-first.
- Ops fetch size is user-adjustable instead of fixed.
- View navigation moved from a top bar to a grouped left sidebar.
- Header was simplified back to a stable context row plus an active-session row.
- Issue lifecycle actions now follow the core transition rules, with one cycle key and explicit abandon behavior.
- Session progress uses cumulative worked time for the active issue based on kernel session history.
- Status colors are applied consistently across issue lists and dashboard indicators.

### Fixed
- Session timer acceleration caused by overlapping local tick loops.
- Meta issue switching now updates issue context correctly.
- Scratchpad editor saves now reload properly in the Go TUI.
- Go client repo creation now uses the correct kernel route.
- Go client ops loading now uses the kernel's latest-ops endpoint.
- Todo-for-date clearing now actually removes the stored date.
- Issue completion and abandonment timestamps are persisted for dashboard reporting.

### API / Core
- `POST /issue` accepts `todoForDate` during issue creation.
- `PUT /issue/:id/todo` accepts an explicit date payload instead of only computing today.
- Added daily summary by arbitrary date in the kernel/core issue summary flow.
- Added kernel shutdown HTTP route for TUI-triggered shutdown.

### Verification
- `go build ./...` passes for `packages/tui-go`.
- `pnpm --filter @crona/core build` passes.
- `pnpm exec tsc -p packages/kernel/tsconfig.json --noEmit` passes.
