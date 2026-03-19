# Roadmap

## Phase 1 — TUI Core
Foundation for all future phases. TUI must be stable and usable before anything is layered on top.

- [x] Go monorepo workspace (`kernel`, `tui`, `cli`, `shared`)
- [x] Go kernel replacing the Node core/kernel runtime
- [x] SQLite store, repositories, commands, and Unix socket IPC in Go
- [x] Go-native e2e coverage for the kernel IPC boundary
- [x] Pane-based navigation (1/2/3/4 + j/k)
- [x] Kernel auto-launch & discovery
- [x] Real-time Unix socket event sync
- [x] Scratchpad CRUD
- [x] Daily summary & todo-for-date
- [x] Status bar (active context + timer state at a glance)
- [x] List filtering & search
- [x] Terminal resize handling
- [x] Settings view
- [x] Session history view
- [x] Active-session history access with issue-scoped session history
- [x] Active-session sidebar reduction to session-only views
- [x] Temporary stash dialog with pop/apply
- [x] Transient toast messages for errors and status updates
- [x] Modal key-help overlay for small terminals
- [x] Minimum TUI size guard with compact small-height layouts for Daily, Default, and Wellbeing views
- [x] Active-session safeguards for issue status changes
- [x] SQLite single-connection kernel mode for stable local IPC writes
- [x] Layout finalization (panel locking)
- [x] Timer boundary recovery after kernel restart
- [x] Session amend (rewrite notes safely)
- [x] Stash + timer interaction rules

**Phase 1 exit criteria**
- [x] Session history view is implemented
- [x] Stash management view is implemented
- [x] Core TUI flow is stable enough to move on to metrics and dashboards

## Phase 2 — Metrics, Check-ins & Habits
Capture the non-work signals that give work data context, add lightweight personal habit tracking, and build the summary primitives needed for richer dashboards.

- [x] Daily check-in (mood, energy level — lightweight prompt in TUI)
- [x] Optional inputs: sleep hours, sleep score, screen time
- [x] Burnout indicator model (derived from session density, break compliance, mood trend, work/rest ratio over rolling window)
- [x] Daily check-in storage schema (new table, not polluting issue/session data)
- [x] Kernel API endpoints for check-in CRUD
- [x] Retrospective entry (backfill past days)
- [x] Reusable kernel summary primitives for streaks, rollups, and date-range analytics
- [x] Habit definitions with daily / weekdays / weekly schedules
- [x] Habit completion tracking, history, and due-for-date queries
- [x] Daily dashboard habit lane with completion/failure and time logging
- [x] Wellbeing dashboard view with check-in summary, streaks, rollups, and burnout status
- [x] Session streak summary (current streak and longest streak)
- [x] Burnout indicator view (rolling composite score from session data + wellbeing inputs)
- [x] Daily log export (Markdown)
- [x] Editable export templates in runtime assets with bundled defaults and variable docs
- [x] Config view for export templates, reports directory, and renderer status
- [x] Export browser view with generated report listing
- [x] Daily PDF export with dedicated template and runtime renderer detection
- [x] Timeline-like export report list in TUI
- [x] Dev helper entrypoint for seed / clear workflows
- [ ] Windows support

**Phase 2 exit criteria**
- [x] Daily check-ins are editable from the TUI for any date
- [x] Rolling wellbeing summaries are available from kernel metrics APIs
- [x] Habits are part of the daily workflow in both kernel and TUI

## Phase 3 — Exports & Reports
Make work history reviewable and portable.

- [x] Weekly summary export
- [x] Session → Issue rollups
- [x] Repo-level reports
- [x] Stream-level reports
- [x] CSV export for external analysis
- [x] Editable report templates and variable docs for weekly, repo, stream, and issue-rollup exports
- [x] Editable CSV export spec plus docs
- [x] Config view exposure for report templates, docs, and CSV spec assets

**Phase 3 exit criteria**
- [x] Narrative reports are generated from editable runtime templates
- [x] CSV export is configurable through an editable runtime spec
- [x] Report templates/specs are editable from the TUI Config view

## Phase 4 — TUI Dashboard System
Make dashboards a first-class terminal feature. Focus on strong summaries, multiple dashboard views, and practical customization that fits a terminal UI.

### Built-in Dashboards
- [ ] Daily Dashboard expansion (weekly rollups, carry-over, missed-vs-done summary)
- [ ] Activity heatmap (terminal-friendly, date-range configurable)
- [ ] Configurable streak scope
- [ ] Time distribution by repo, stream, issue, or segment type
- [ ] Daily/weekly focus score (work vs break ratio vs target)
- [ ] Goal progress (estimated vs actual time per issue/stream/repo)

### TUI Customisation
- [ ] Multiple dashboard views under the `DASHBOARD` group
- [ ] Widget-style dashboard sections instead of a fixed page
- [ ] Add/remove/reorder widgets in TUI
- [ ] Widget configuration for scope, metric, and date range
- [ ] Saved dashboard presets in the kernel
- [ ] Pre-built layout presets (default, focus-heavy, wellbeing-focused)

### Constraints
- [ ] Keep customization terminal-native: stacked widgets, simple grids, no freeform layout
- [ ] Avoid one-off dashboard endpoints by building reusable summary APIs

## Phase 5 — CLI
Non-TUI interface for scripting, shell aliases, and integration with other tools, after the dashboard and metrics model is stable.

- [ ] `crona` binary with subcommands
- [ ] JSON output mode (`--json`)
- [ ] Kernel attach/detach commands
- [ ] Context management from shell (`crona context set`, `crona issue start`)
- [ ] Session lifecycle from shell (`crona timer start|pause|end`)
- [ ] Shell completions (zsh, bash, fish)

## Phase 6 — Public Beta Release
- [ ] Cross-platform packaging and install docs are ready for external users
- [ ] Public beta release notes and upgrade path are documented
- [ ] Feedback / issue intake path is defined for beta users
- [ ] Core TUI and kernel flows are stable enough for public beta usage

## Phase 7 — Multi-Device Sync
See `FEATURE.md` for design proposal.

- [ ] Op log export/import
- [ ] File-based sync (iCloud Drive / Dropbox / Google Drive)
- [ ] Self-hosted sync relay (Docker, optional)
- [ ] Conflict resolution strategy
- [ ] Per-device context isolation

## Deferred
- [ ] Command palette / `:` command mode
- [ ] Fuzzy command search
- [ ] Context-aware command suggestions
- [ ] Vim-style command-line editing
