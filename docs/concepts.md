# Concepts

Crona is a local-first work kernel for developers. The kernel owns state, and the TUI and CLI act as clients over local IPC.

## Core Ideas

- Local-first state, with the kernel as the source of truth.
- Terminal-native interaction through the TUI and CLI.
- Structured work objects instead of loose notes.
- Deterministic exports and local automation hooks instead of cloud coupling.
- UIs are clients, not controllers.

## Runtime Model

Crona has three main runtime pieces:

- `crona-kernel`: the local daemon that owns storage, timers, updates, and IPC.
- `crona-tui`: the interactive terminal UI.
- `crona`: the scriptable CLI and default launcher.

All clients talk to the local kernel over the shared IPC surface documented in [api/socket.md](api/socket.md).

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

If a user starts a focus session on an issue that already has a stash, the kernel blocks the fresh start and returns a structured conflict. Clients should show the matching stash or stashes and let the user either resume a stash or explicitly continue with a fresh session. Continuing fresh keeps the existing stash for later.

### Active Context

The shared `{ repo -> stream -> issue }` selection across kernel clients.

### Scratchpads

Scratchpads are filesystem-backed notes rather than scoped database metadata.

Example:

```text
notes/[[date]]-daily.md
```

## Notifications And Automation

### Notifications

Crona can trigger local OS notifications and bundled alert sounds from the kernel itself. The TUI configures and tests alerts, but notification timing, scheduled reminder evaluation, and delivery decisions remain kernel-owned. Today this uses platform-specific local helpers rather than a separate native companion layer.

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

Crona is in beta. The core workflow is usable, but the product is still settling and some APIs, storage details, and update flows may continue to change before `1.0.0`.

Current mainline focus:
- hardening the kernel, TUI, and CLI
- installer/updater/support polish
- documentation and contributor-facing references
- beta tester feedback before the first stable release
