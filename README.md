# Crona Node

**Crona** is a local-first, developer-centric work tracking system that blends ideas from Git, Pomodoro timers, and task managers into a single, opinionated workflow.

This repository contains the **Crona Node monorepo**, which powers the Crona kernel, core logic, and terminal UI.

> Crona is not a to-do app.  
> It is a **personal work kernel** that makes your work structure explicit, measurable, and reviewable.

---

## Repository Structure
```md
crona-node/
├─ packages/
│  ├─ core/        # Domain logic, commands, persistence, events
│  ├─ kernel/      # Local kernel process + HTTP/SSE APIs
│  └─ tui/         # Terminal UI (Ink-based)
│
├─ pnpm-workspace.yaml
├─ package.json
└─ README.md
```

---

## Core Concepts

Crona models work the way developers actually think about it.

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

Think: **personal Git branches**, not ephemeral tasks.

### Issue
The smallest unit of intentional work.

Each issue can have:
- title
- estimate
- notes
- status (todo, active, done, abandoned)

### Session
A focused work interval tied to an issue.

Sessions:
- are started/stopped via the timer
- contain one or more segments
- end with a **commit-style message**

### Session Segments
A session is composed of segments:
- `work`
- `short_break`
- `long_break`
- `rest`

Segments are authoritative for timing and boundaries.

### Timer
The timer is **derived state**, not stored state.

It:
- starts/stops sessions
- transitions segments
- enforces Pomodoro-style boundaries
- emits events for UIs to subscribe to

### Stash
A stash suspends your current context and timer state.

Similar to `git stash`:
- captures active context
- optionally snapshots timer state
- can be applied or dropped later

### Active Context
The current `{ repo → stream → issue }` selection.
Shared across kernel clients.

### Scratchpads
Scratchpads are **filesystem-backed notes**, not scoped metadata.

- Arbitrary files
- Multiple buffers
- Variable paths supported:
  - `[[date]]`
  - `[[time]]`
  - `[[datetime]]`
  - `[[timestamp]]`
  - `[[random]]`

Example:

```md
notes/[[date]]-daily.md
```

---

## Packages

### `@crona/core`

Contains:
- domain models
- SQLite schema (via Kysely)
- commands (repo, stream, issue, session, stash, scratchpad)
- timer logic and boundaries
- event bus

No HTTP. No UI. Pure logic.

---

### `@crona/kernel`

A **local kernel process** that:
- owns the database
- exposes HTTP + SSE APIs
- manages authentication
- emits real-time events
- auto-starts when needed

Endpoints include:
- `/repos`, `/streams`, `/issues`
- `/timer/*`
- `/context`
- `/stash`
- `/scratchpads`
- `/events` (SSE)

The kernel is **single-user, local-first**, and meant to be trusted.

---

### `@crona/tui`

An Ink-based terminal UI inspired by tools like:
- lazygit
- vim
- tmux

Features:
- persistent terminal takeover
- kernel auto-launch & discovery
- real-time updates via SSE
- contextual views (pre-session vs active session)
- future command palette + editor integration

---

## Development

### Prerequisites
- Node.js ≥ 20
- pnpm ≥ 10

### Install

```bash
  pnpm install
```


### Build all packages

```bash
  pnpm -r dev
```

### Run Kernel (Dev)

```bash
pnpm --filter @crona/kernel dev
```

### Run TUI (Dev)

```bash
pnpm --filter @crona/tui dev
```

The TUI will automatically start the kernel if it is not running.

⸻

## Testing

Crona uses end-to-end tests extensively to validate real behavior.

Tests cover:
- repos / streams / issues
- timer lifecycle
- session boundaries
- stash behavior
- scratchpads
- kernel startup & isolation

### Running tests

```bash
pnpm --filter @crona/kernel-e2e test:<module>
```

> refer to the test/package.json for different available modules

Design Principles
- Local-first
- Authoritative data over derived state
- Everything is replayable
- No magic background jobs
- UIs are clients, not controllers
- Git-like mental model

⸻

## Status

Crona is under active development.

Current focus:
- TUI layout & navigation
- command palette
- scratchpad editor integration
- exportable work logs
- daily planning workflows

⸻

## Philosophy

Your work already has structure.
Crona just makes it explicit.

⸻

[License](LICENSE.md)
> Crona is an opinionated, experimental project. The MIT license allows reuse, but the architecture and APIs may change without notice.

