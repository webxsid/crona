# Roadmap

## Phase 1 — TUI Core
Foundation for all future phases. TUI must be stable and usable before anything is layered on top.

- [x] Go monorepo workspace (`kernel`, `tui`, `cli`, `shared`)
- [x] Go kernel established as the local runtime
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

## Phase 4 — Automation, Notifications & Calendar Hooks
Prioritise machine-friendly flows and local integrations before deeper TUI dashboard expansion.

- [x] `crona` binary with scriptable subcommands
- [x] JSON output mode (`--json`)
- [x] Context management from shell (`crona context get|set|clear`, `crona issue start`)
- [x] Session lifecycle from shell (`crona timer start|pause|resume|end|status`)
- [x] Calendar export via local `.ics` generation
- [x] Separate configurable ICS export directory for local automation workflows
- [x] Stable calendar-export file workflow for Shortcuts, Folder Actions, and local import automations
- [x] Apple Shortcuts-friendly non-interactive CLI surface
- [x] Structured timer boundary notifications
- [x] Audible timer-boundary cues where the local OS supports them
- [x] Kernel attach/detach commands
- [x] Shell completions (zsh, bash, fish)
- [x] Notification settings docs and platform-specific fallback guidance
- [x] Repo-scoped ICS bundle export (`issues.ics` + `sessions.ics`)

### Phase 4.1 — Windows Support V1
- [x] Windows named-pipe IPC for kernel RPC and event streaming
- [x] Windows-aware kernel runtime metadata (`transport` + `endpoint`)
- [x] Windows PowerShell support for `crona`, `crona-tui`, and `crona-kernel`
- [x] Windows-aware binary discovery and launch (`.exe` binaries)
- [x] Windows runtime path support
- [x] Windows coverage for kernel attach/info/status flows
- [x] Self-update explicitly unavailable on Windows with clear manual guidance

### Phase 4.2 — Windows Support V2
- [x] Windows release artifacts and checksums
- [x] Native Windows installer flow
- [x] Windows in-app self-update from the `Updates` view
- [x] Windows relaunch flow for TUI + kernel after install
- [x] Windows install/update docs

**Phase 4 exit criteria**
- [x] Structured timer boundaries can notify outside the TUI
- [x] Calendar exports are generated as local `.ics` files
- [x] ICS exports can be written to a dedicated configurable directory suitable for local automations
- [x] Core focus/context/export flows are scriptable through `crona`
- [x] Phase 4.1 Windows runtime support works from PowerShell without requiring in-app install
- [x] Phase 4.2 native Windows packaging and self-update are available

## Phase 5 — Dashboards, Customisation, and CLI
Expand the product surface after the kernel, TUI shell, and packaging work have settled.

### Phase 5.1 — Dashboards And Export Template Rewrite
- [ ] Daily Dashboard expansion (weekly rollups, carry-over, missed-vs-done summary)
- [ ] Activity heatmap (terminal-friendly, date-range configurable)
- [ ] Configurable streak scope
- [ ] Time distribution by repo, stream, issue, or segment type
- [ ] Daily/weekly focus score (work vs break ratio vs target)
- [ ] Goal progress (estimated vs actual time per issue/stream/repo)
- [ ] Rewrite bundled export templates to match the new dashboard and reporting model
- [ ] Remove one-off summary formatting from the TUI and drive dashboard/export views from reusable summary APIs

### Phase 5.2 — TUI Customisations
- [ ] Multiple dashboard views under the `DASHBOARD` group
- [ ] Widget-style dashboard sections instead of a fixed page
- [ ] Add/remove/reorder widgets in TUI
- [ ] Widget configuration for scope, metric, and date range
- [ ] Saved dashboard presets in the kernel
- [ ] Pre-built layout presets (default, focus-heavy, wellbeing-focused)
- [ ] Keep customization terminal-native: stacked widgets, simple grids, no freeform layout

### Phase 5.3 — CLI
- [ ] Ship the `crona` CLI in public release artifacts and installers
- [ ] Full CRUD command trees for `repo`, `stream`, `issue`, and `habit`
- [ ] Non-interactive flag-driven create and update flows for all core entities
- [ ] Interactive add/edit flows in the CLI for repos, streams, issues, and habits
- [ ] Interactive CLI context picker
- [ ] Proper per-command help docs and examples for all CRUD surfaces

## Phase 6 — OS Integrations
- [ ] Native OS bridge for notifications, audible timer alerts, and permission-aware delivery
- [ ] Stronger bundled timer-alert sounds with OS-native playback instead of shell bell fallbacks
- [ ] Test notification and test sound actions exposed from settings
- [ ] Native macOS menu bar companion
- [ ] Native Windows tray companion
- [ ] Native Linux tray/status notifier companion where desktop support allows
- [ ] Live timer and checked-out context status from local kernel IPC
- [ ] Global hotkeys for core timer actions
- [ ] Launch-at-login support
- [ ] Local Calendar.app integration via EventKit
- [ ] Calendar sync into a dedicated local Crona calendar
- [ ] Optional Shortcuts/Automation entrypoints for Crona sync and timer actions
- [ ] Best-effort high-visibility timer notifications per OS without promising Focus/DND bypass

**Phase 6 exit criteria**
- [ ] Native notification and sound delivery is handled through OS-aware bridges instead of shell-only fallbacks
- [ ] Users can verify notification and sound behavior from Crona settings without waiting for a real timer boundary
- [ ] macOS users can monitor and control active context and timer state without opening the TUI
- [ ] Local Calendar.app sync works without direct cloud API dependencies
- [ ] OS companions remain thin clients over the Crona kernel

## Phase 7 — Public Beta Release
- [ ] Cross-platform packaging and install docs are ready for external users
- [ ] Public beta release notes and upgrade path are documented
- [ ] Feedback / issue intake path is defined for beta users
- [ ] Core TUI and kernel flows are stable enough for public beta usage

## Phase 8 — Multi-Device Sync
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
