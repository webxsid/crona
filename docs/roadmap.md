# Roadmap

Current mainline focus is the `1.0.2` stable line. Native OS integration is out of scope for the current mainline and remains deferred until the product, packaging, and distribution model is settled.

## Phase 1 — TUI Core
Foundation for all future phases. TUI must be stable and usable before anything is layered on top.

- [x] Go monorepo workspace (`kernel`, `tui`, `cli`, `shared`)
- [x] Go kernel established as the local runtime
- [x] SQLite store, repositories, commands, and Unix socket IPC in Go
- [x] Go-native e2e coverage for the local engine IPC boundary
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

## Phase 5 — Dashboards

- [x] Wellbeing dashboard visual refresh with compact cards and an activity heatmap
- [x] Narrative report template refresh with preset styles and stronger PDF presentation
- [x] Dedicated rollup dashboard with a default weekly window and adjustable start/end dates
- [x] Shared kernel summary APIs for dashboard rollups so TUI dashboards and reports read from the same data surface
- [x] Rollup summaries for execution, focus, time distribution, and estimate-vs-actual progress
- [x] Rollup day detail drill-down from the daily-status pane
- [x] Estimate-bias summary metric for average over/under-estimation across worked issues
- [x] Small-screen dashboard layouts keep daily and wellbeing views glanceable without hiding critical content
- [x] Dashboard and report copy stays short, numeric, and visually scannable by default

## v1.0 Hardening And Release Prep
- [x] Kernel, TUI, and CLI code sanitation is complete
- [x] Documentation, install docs, and API docs are current and navigable
- [x] Installer, updater, and support flows are stable enough for stable users
- [x] Feedback and issue intake paths are clear for users and beta testers
- [x] Core workflows are stable enough for `v1.0.0`
- [x] `v1.0.2` is the active stable release target
- [x] CI, coverage, release-check, and tag-driven release publishing are defined for the stable release path
- [x] Kernel IPC e2e tests are isolated behind an explicit target for reliable local and CI validation

**Mainline exit criteria**
- [x] Mainline is focused on stabilization rather than new feature families
- [x] Stable and beta channel documentation and upgrade guidance are clear
- [x] The codebase is ready for the first `1.0.0` stable release

## Post-1.0 Product Work

These are candidate tracks after the `1.0` stabilization path, not blockers for the stable release.

### Bulk Issue Import
Introduce a watched inbox for bulk issue creation so users can drop structured files into Crona without going through the TUI one item at a time.

- [ ] Watched import inbox path for bulk issue intake
- [ ] Predefined YAML-based issue import format
- [ ] Create-only import flow in v1
- [ ] Processed / archived / failed file handling for imports
- [ ] Import status visibility and last-run feedback in Crona

### Configurable Dashboards
Introduce a constrained YAML-driven dashboard composition layer on top of the stable summary APIs and terminal-native section renderers. This is post-stable product work, not a blocker for the stable release.

- [ ] YAML-based dashboard configuration
- [ ] Configurable section ordering and enable/disable per dashboard view
- [ ] Configurable scope and date-window per section
- [ ] Configurable streak sections
- [ ] Configurable heatmap sections
- [ ] Configurable line-graph sections
- [ ] Configurable bar-graph sections
- [ ] Pre-built starter YAML layouts for daily-focused, wellbeing-focused, and accountability-focused setups
- [ ] Keep customization terminal-native: stacked sections, simple splits, and fixed section types instead of freeform layout

### Multi-Device Sync
See [`feature-design.md`](feature-design.md) for design proposal.

- [ ] Op log export/import
- [ ] File-based sync (iCloud Drive / Dropbox / Google Drive)
- [ ] Self-hosted sync relay (Docker, optional)
- [ ] Conflict resolution strategy
- [ ] Per-device context isolation

## Deferred Exploration

These are intentionally not on the active mainline path right now.

### Native OS Integration
- [ ] Finalize the native client and distribution strategy before implementation resumes
- [ ] Decide whether native OS work belongs in a separate macOS client, a companion, or another distribution model
- [ ] Revisit native notifications, sounds, tray surfaces, and launch-at-login only after that decision is settled
- [ ] Revisit EventKit / local calendar integration only after the native client shape is clear

### Local Companion
Introduce an explicit local desktop companion for users running Crona remotely so notifications and open actions can still happen on their local machine.

- [ ] Opt-in local companion process for desktop-local actions
- [ ] Bridge local notifications for remote/server-hosted Crona
- [ ] Bridge local open actions for URL, file, and editor targets
- [ ] Clear unavailable-state UX when the companion is disabled or unreachable
- [ ] Keep the companion constrained to typed desktop actions instead of arbitrary command execution

## Deferred
- [ ] Full public CLI expansion beyond the current scriptable shell surface
- [ ] Full CRUD command trees for `repo`, `stream`, `issue`, and `habit`
- [ ] Non-interactive flag-driven create and update flows for all core entities
- [ ] Interactive add/edit flows in the CLI for repos, streams, issues, and habits
- [ ] Interactive CLI context picker
- [ ] Proper per-command help docs and examples for all CRUD surfaces
- [ ] Command palette / `:` command mode
- [ ] Fuzzy command search
- [ ] Context-aware command suggestions
- [ ] Vim-style command-line editing
