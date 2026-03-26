# Crona

**Crona** is a local-first, developer-centric work tracking system that blends ideas from Git, Pomodoro timers, and task managers into a single, opinionated workflow.

This repository is a Go monorepo containing the Crona kernel, shared contracts, terminal UI, and CLI.

## Public Beta Install

The current beta ships as two binaries:
- `crona-tui`
- `crona-kernel`

End users do not need Go installed. The installer downloads the correct TUI and kernel pair for the current machine.

```bash
curl -fsSL https://github.com/webxsid/crona/releases/download/v0.2.1/install-crona-tui.sh | sh
```

By default this installs into `~/.local/bin`.

If you want to replace an existing install non-interactively, set `CRONA_INSTALL_FORCE=1` on the `sh` side of the pipe:

```bash
curl -fsSL https://github.com/webxsid/crona/releases/download/v0.2.1/install-crona-tui.sh | CRONA_INSTALL_FORCE=1 sh
```

If you want the installer to prompt and wait for your `y/N` response, do not run it through `curl | sh`. Save it first and then execute it so the script can read from your terminal:

```bash
curl -fsSL -o /tmp/install-crona-tui.sh https://github.com/webxsid/crona/releases/download/v0.2.1/install-crona-tui.sh
sh /tmp/install-crona-tui.sh
```

## Manual Build And Install

If you want to build from source instead of using the release installer, clone the repo and build the workspace locally.

```bash
make build
```

To build the kernel binary into the repo `bin/` directory:

```bash
make install-kernel
```

To build the TUI binary locally into the repo `bin/` directory:

```bash
make install-tui
```

To build the CLI binary locally into the repo `bin/` directory:

```bash
make install-cli
```

If you prefer fully manual Go commands instead of the `Makefile` targets:

```bash
cd kernel && go build -o ../bin/crona-kernel ./cmd/crona-kernel
cd tui && go build -o ../bin/crona-tui .
cd cli && go build -o ../bin/crona ./cmd/crona
```

Make sure both `crona-kernel` and `crona-tui` are on your `PATH` before launching:

```bash
crona-tui
```

## Repository Structure

```text
crona/
├─ Makefile  # Project metadata and common tasks
├─ kernel/   # Local daemon: storage, commands, timer, IPC
├─ tui/      # Terminal UI (Bubble Tea)
├─ cli/      # Scriptable CLI and local automation entrypoints
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

The root `.env` file controls the local runtime mode.

```bash
CRONA_ENV=Prod
```

Set it to `Dev` to enable developer-only seed and clear helpers in the kernel and TUI.

Runtime storage defaults to:
- macOS prod: `~/Library/Application Support/Crona`
- macOS dev: `~/Library/Application Support/Crona Dev`
- Linux prod: `${XDG_DATA_HOME:-~/.local/share}/crona`
- Linux dev: `${XDG_DATA_HOME:-~/.local/share}/crona-dev`
- Windows prod: `%LocalAppData%\Crona`
- Windows dev: `%LocalAppData%\Crona Dev`

On macOS and Linux, existing `~/.crona` and `~/.crona-dev` runtime directories are migrated automatically on install or first kernel start.
Set `CRONA_HOME` to override the runtime directory and skip the default-path migration logic.

### Project Tasks

The root `Makefile` is the shared task runner for project metadata, builds, tests, and release helpers.

```bash
make help
make build
make test
make lint
make seed-dev
make clear-dev
```

When `CRONA_ENV=Dev`, local binary targets use `-dev` suffixes:

```bash
CRONA_ENV=Dev make install-kernel install-tui install-cli
# -> bin/crona-kernel-dev, bin/crona-tui-dev, bin/crona-dev
```

Install the linter once with:

```bash
make install-lint
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

### Windows Dev Run

Use PowerShell from the repo root.

Set dev mode:

```powershell
$env:CRONA_ENV = "Dev"
```

Build the dev kernel and TUI into `bin\`:

```powershell
go build -o .\bin\crona-kernel-dev.exe .\kernel\cmd\crona-kernel
go build -o .\bin\crona-tui-dev.exe .\tui
```

Add the repo `bin\` directory to `PATH` for the current shell:

```powershell
$env:PATH = "$PWD\bin;$env:PATH"
```

Run the TUI:

```powershell
.\bin\crona-tui-dev.exe
```

This is the Windows equivalent of `CRONA_ENV=Dev make install-kernel install-tui` followed by `CRONA_ENV=Dev make run-tui`.

If you want to run without building the TUI binary first, this also works once the dev kernel binary exists in `bin\`:

```powershell
go run .\tui
```

If kernel auto-start is not available, run it in a separate terminal:

```powershell
$env:CRONA_ENV = "Dev"
go run .\kernel\cmd\crona-kernel
```

### CLI

The `crona` CLI exposes local automation-friendly commands for kernel, context, timer, and calendar export flows.

Examples:

```bash
crona kernel attach
crona context get --json
crona timer status --json
crona export calendar --start 2026-03-17 --end 2026-03-23
```

Shell completions can be generated with:

```bash
crona completion zsh
crona completion bash
crona completion fish
```

### Notifications

Structured timer boundaries can trigger local OS notifications and optional audible cues.

- macOS: `osascript`
- Linux: `notify-send`, with `canberra-gtk-play` used for sound when available
- Windows: PowerShell fallback

If the local notification tooling is unavailable, Crona continues running and skips the notification instead of failing the timer flow.

### Calendar Automation

Crona can generate local `.ics` calendar exports for automation workflows.

The intended Phase 4 workflow is:

- Crona writes deterministic `.ics` files into a dedicated ICS export directory
- local tools such as Shortcuts, Folder Actions, or import/watch-folder automations react to those files
- Crona itself does not need direct Google or iCloud API access

### Run Tests

```bash
make test
```

### Lint

```bash
make lint
```

The repo ships with `.golangci.yml` as the shared lint baseline.

### Neovim

For good Go highlighting in Neovim, make sure `gopls` is installed and semantic tokens are enabled in your editor config:

```bash
go install golang.org/x/tools/gopls@latest
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
