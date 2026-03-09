# Roadmap

## Phase 1 — TUI Core (Current)
Foundation for all future phases. TUI must be stable and usable before anything is layered on top.

- [x] Pane-based navigation (1/2/3/4 + j/k)
- [x] Kernel auto-launch & discovery
- [x] Real-time SSE event sync
- [x] Scratchpad CRUD
- [x] Daily summary & todo-for-date
- [x] Status bar (active context + timer state at a glance)
- [x] List filtering & search
- [x] Terminal resize handling
- [ ] Layout finalization (panel locking)
- [ ] Timer boundary recovery after kernel restart
- [ ] Session amend (rewrite notes safely)
- [ ] Stash + timer interaction rules

## Phase 2 — Command Palette
Make the TUI fully keyboard-driven without requiring mouse or menu navigation.

- [ ] `:` command vocabulary (create, update, delete, switch context)
- [ ] Fuzzy match on command input
- [ ] Context-aware command suggestions
- [ ] Vim-style line editing in command mode
- [x] Scratchpad editor integration (open in `$EDITOR`)

## Phase 3 — CLI
Non-TUI interface for scripting, shell aliases, and integration with other tools.

- [ ] `crona` binary with subcommands
- [ ] JSON output mode (`--json`)
- [ ] Kernel attach/detach commands
- [ ] Context management from shell (`crona context set`, `crona issue start`)
- [ ] Session lifecycle from shell (`crona timer start|pause|end`)
- [ ] Shell completions (zsh, bash, fish)

## Phase 4 — Personal Metrics & Wellbeing Data
Capture the non-work signals that give work data context. Without these inputs, productivity metrics are incomplete.

- [ ] Daily check-in (mood, energy level — lightweight prompt in TUI or CLI)
- [ ] Optional inputs: sleep hours, sleep score, screen time
- [ ] Burnout indicator model (derived from session density, break compliance, mood trend, work/rest ratio over rolling window)
- [ ] Daily check-in storage schema (new table, not polluting issue/session data)
- [ ] Kernel API endpoints for check-in CRUD
- [ ] Retrospective entry (backfill past days)

## Phase 5 — Insights Dashboard
Local web dashboard served by the kernel (`http://127.0.0.1:<port>/dashboard`). No external server. The kernel already runs HTTP — this is a static asset route + new analytics API endpoints.

**Why web over TUI**: Real SVG charts, heatmaps with color gradients, and interactive customization aren't feasible in a terminal. TUI gets a condensed quick-stats pane; the browser gets the full dashboard. Both are localhost-only.

### Built-in Components
- [ ] Activity heatmap (GitHub-style, configurable time range + entity)
- [ ] Session streaks (current streak, longest streak, configurable per repo/stream/issue)
- [ ] Time distribution (pie/bar by repo, stream, or segment type)
- [ ] Daily/weekly focus score (work vs break ratio vs target)
- [ ] Burnout indicator (rolling composite score from session data + wellbeing inputs)
- [ ] Mood & energy trend overlay (correlate wellbeing inputs with session data)
- [ ] Goal progress (estimated vs actual time per issue/stream)

### Customisation System
- [ ] Dashboard is a grid of user-arranged widget slots
- [ ] Each widget is a component (heatmap, streak, graph) bound to a user-defined data query
- [ ] Data query config: entity scope (all repos / specific repo / stream), metric (time spent, sessions, issues closed), time range, aggregation (daily/weekly)
- [ ] Streak config: what counts as a "day" (any session, minimum N minutes, specific repo only)
- [ ] Dashboard layouts saved to kernel (stored in DB, portable via sync)
- [ ] Pre-built layout presets (default, focus-heavy, wellbeing-focused)

### TUI Quick Stats Pane
- [ ] Condensed today-at-a-glance in TUI (streak count, today's focus time, burnout signal)
- [ ] Accessible without opening a browser

## Phase 6 — Exports & Reports
Make work history reviewable and portable.

- [ ] Daily log export (Markdown)
- [ ] Weekly summary export
- [ ] Session → Issue rollups
- [ ] Repo-level time reports
- [ ] Timeline view in TUI
- [ ] CSV export for external analysis

## Phase 7 — Multi-Device Sync
See `FEATURE.md` for design proposal.

- [ ] Op log export/import
- [ ] File-based sync (iCloud Drive / Dropbox / Google Drive)
- [ ] Self-hosted sync relay (Docker, optional)
- [ ] Conflict resolution strategy
- [ ] Per-device context isolation
