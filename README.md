# Crona

**Crona** is a local-first, developer-centric work tracking system that blends ideas from Git, Pomodoro timers, and task managers into a single, opinionated workflow.

This repository is a Go monorepo containing the Crona kernel, shared contracts, terminal UI, and future CLI.

## Public Beta Install

The current beta ships as two binaries:
- `crona-tui`
- `crona-kernel`

End users do not need Go installed. The installer downloads the correct TUI and kernel pair for the current machine.

```bash
curl -fsSL https://github.com/webxsid/crona-node/releases/download/v0.1.0-beta.1/install-crona-tui.sh | sh
```

By default this installs into `~/.local/bin`.

## Repository Structure

```text
crona-node/
├─ Makefile  # Project metadata and common tasks
├─ kernel/   # Local daemon: storage, commands, timer, IPC
├─ tui/      # Terminal UI (Bubble Tea)
├─ cli/      # Future CLI
├─ shared/   # DTOs, types, protocol envelopes
├─ go.work
└─ README.md
```

## Core Concepts

### Repository

A high-level bucket for work.

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

The smallest unit of intentional work.

Each issue can have:
- title
- estimate
- notes
- status (`todo`, `active`, `done`, `abandoned`)

### Session

A focused work interval tied to an issue.

Sessions:
- are started and stopped via the timer
- contain one or more segments
- end with a commit-style message

### Session Segments

A session is composed of segments:
- `work`
- `short_break`
- `long_break`
- `rest`

### Timer

The timer is derived state, not stored state.

It:
- starts and stops sessions
- transitions segments
- enforces Pomodoro-style boundaries
- emits events for UIs to subscribe to

### Stash

A stash suspends your current context and timer state.

It captures:
- active context
- optional timer state snapshot

### Active Context

The current `{ repo -> stream -> issue }` selection, shared across kernel clients.

### Scratchpads

Scratchpads are filesystem-backed notes, not scoped metadata.

- Arbitrary files
- Multiple buffers
- Variable path templates

Example:

```text
notes/[[date]]-daily.md
```

## Components

### `kernel`

A local kernel process that:
- owns the SQLite database
- exposes Unix socket IPC
- emits real-time events
- auto-starts when needed

### `tui`

A Bubble Tea terminal UI with:
- kernel auto-launch and discovery
- real-time updates via Unix socket events
- contextual views for planning, tracking, scratchpads, ops, and settings

### `shared`

Shared types, request DTOs, and IPC envelopes used by the kernel, TUI, and future CLI.

## Development

### Prerequisites

- Go 1.26+
- `make` for the root task shortcuts

### Environment

The root [`/.env`](/Users/sm2101/Projects/crona-node/.env) file controls the local runtime mode.

```bash
CRONA_ENV=Prod
```

Set it to `Dev` to enable developer-only seed and clear helpers in the kernel and TUI.

### Project Tasks

The root [`Makefile`](/Users/sm2101/Projects/crona-node/Makefile) replaces the old JS `package.json` role for shared scripts and basic project metadata.

```bash
make help
make build
make test
make seed-dev
make clear-dev
```

### Run Kernel

```bash
make run-kernel
```

### Run TUI

```bash
make run-tui
```

The TUI will auto-start the kernel if `crona-kernel` is on your `PATH`.
When `CRONA_ENV=Dev`, the TUI exposes global hotkeys for developer data management:
- `f6` seeds sample data
- `f7` clears all local data

### Run Tests

```bash
make test
```

## Design Principles

- Local-first
- Authoritative data over derived state
- Everything is replayable
- No magic background jobs
- UIs are clients, not controllers
- Git-like mental model

## Status

Crona is in public beta and under active development. The project is open source, but the workflow, storage layout, and IPC/API details are still moving and may change between releases.

Current focus:
- TUI layout and navigation
- command palette
- exportable work logs
- daily planning workflows

## Philosophy

Your work already has structure. Crona just makes it explicit.

[License](LICENSE)
> Crona is an opinionated, experimental project. The code is MIT licensed, but the architecture and APIs may change without notice while the product is still settling.
