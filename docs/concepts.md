# Concepts

Crona is a local-first work tracker for developers. A background local engine owns state, and the TUI and CLI act as clients over local IPC.

## Terminology

The codebase and socket API still use the term `kernel` for the internal engine process and its command namespace. In user-facing docs, this is usually called the local engine or background engine because it is the small local service that owns storage, timers, reminders, update checks, and IPC.

## Core Ideas

- Local-first state, with the background engine as the source of truth.
- Terminal-native interaction through the TUI and CLI.
- Structured work objects instead of loose notes.
- Deterministic exports and local automation hooks instead of cloud coupling.
- UIs are clients, not controllers.

## Runtime Model

Crona has three main runtime pieces:

- `crona-kernel`: the background local engine that owns storage, timers, updates, and IPC.
- `crona-tui`: the interactive terminal UI.
- `crona`: the scriptable CLI and default launcher.

All clients talk to the local engine over the shared IPC surface documented in [api/socket.md](api/socket.md).

The TUI owns the terminal tab/window title while it is running. Idle titles show Crona plus the active repo/stream and current view when available; active focus sessions show Crona plus the issue/session context and elapsed timer state. The title is reset on exit on a best-effort basis.

## Core Entities

### Repository

A top-level bucket for work.

Examples:
- Office
- Personal
- Research

### Stream

A long-lived subdivision inside a repository.

Examples:
- main
- backend
- experiments

### Issue

The smallest intentional unit of work. An issue can carry a title, estimate, notes, and lifecycle state.

### Session

A focused work interval tied to an issue.

Sessions:
- are started and stopped via the timer
- contain one or more segments
- end with a commit-style summary message

### Session Segments

A session is composed of:
- `work`
- `short_break`
- `long_break`
- `rest`

### Timer

The timer is derived state, not stored state.

It:
- starts and stops sessions
- transitions segments
- enforces structured boundaries
- emits events for subscribed clients

### Stash

A stash suspends the current context and can preserve timer state.

If a user starts a focus session on an issue that already has a stash, the local engine blocks the fresh start and returns a structured conflict. Clients should show the matching stash or stashes and let the user either resume a stash or explicitly continue with a fresh session. Continuing fresh keeps the existing stash for later.

### Active Context

The shared `{ repo -> stream -> issue }` selection across local clients.

### Scratchpads

Scratchpads are filesystem-backed notes rather than scoped database metadata.

Example:

```text
notes/[[date]]-daily.md
```

## Wellbeing Metrics

The Wellbeing view combines a selected-day check-in with a rolling metrics window. Mood, energy, sleep, screen time, burnout, focus, and habit rollups are still shown for the recent 7-day window ending on the selected wellbeing date.

Momentum is separate from that 7-day window. It uses all stored local history up to the selected wellbeing date, so focus, check-in, and habit streaks can exceed the visible metrics window. The range-based streak API remains available for reports and callers that intentionally want date-window streak calculations.

Custom habit momentum follows the habit streak definitions configured in Settings:

- daily definitions count matching completed days
- weekly definitions count weeks that meet their configured completion threshold
- monthly definitions count months that meet their configured completion threshold

For weekly and monthly custom momentum, the current in-progress bucket does not break an existing streak just because it has not reached its threshold yet. It only extends the streak once the threshold is met.

The Momentum pane visualizes current streak length with a cadence-specific ladder. Filled blocks are milestones reached by the current streak; empty blocks are future milestones.

```text
Daily/check-in/focus: 1d, 3d, 7d, 14d, 30d, 60d, 100d
Weekly customs:       1w, 2w, 4w, 8w, 13w, 26w, 52w
Monthly customs:      1mo, 2mo, 3mo, 6mo, 12mo, 24mo
```

The exact current and best values are shown next to the ladder, for example `14d current · 30d best`.

## Terminal UI Surfaces

On wider terminals, the Wellbeing dashboard splits its lower region into a 7-day Metrics Window pane and a separate Momentum pane. The Momentum pane is focusable and scrollable independently so custom habit momentum can grow without clipping the metrics content.

Calendar surfaces use terminal background styling for selected dates, date ranges, today, and the current week rather than bracket markers. This keeps date cells fixed-width while relying on color and background state to distinguish selection and today.

## Notifications And Automation

### Notifications

Crona can trigger local OS notifications and bundled alert sounds from the local engine itself. The TUI configures and tests alerts, but notification timing, scheduled reminder evaluation, and delivery decisions remain local-engine-owned. Today this uses platform-specific local helpers rather than a separate native companion layer.

Focus inactivity alerts are also local-engine-owned. If a focus session keeps running without recent TUI activity for the configured threshold, Crona can notify the user to review, pause, stash, or end the session.

### Calendar Export

Crona can generate deterministic local `.ics` files for external automations.

Typical workflow:
- Crona writes `.ics` files into the configured export directory.
- Local automations watch that directory.
- External tools import or react to those files.

Crona does not require direct Google Calendar or iCloud API integration for this flow.

## Design Principles

- local-first
- authoritative data over derived state
- replayable operations
- no hidden background jobs
- deterministic local artifacts
- a git-like mental model for work state

## Project Status

The core workflow is usable for general users, while validation builds remain available for faster feedback on upcoming changes.

Current mainline focus:
- stable-channel maintenance
- installer/updater/support polish
- documentation and contributor-facing references
- tester feedback for upcoming releases
