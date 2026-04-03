# Crona

Crona is a local-first work kernel for developers. It combines a local daemon, a terminal UI, and a CLI into one workflow for planning work, tracking focus sessions, and exporting structured artifacts.

This repository is a Go monorepo with four main modules:
- `kernel`: local daemon, SQLite store, timer, IPC, update checks
- `tui`: Bubble Tea terminal UI
- `cli`: scriptable commands and kernel control flows
- `shared`: shared types, config, protocol, and utilities

## Overview

Crona is built around a few simple ideas:
- local-first state, with the kernel as the source of truth
- terminal-native interaction through the TUI and CLI
- structured work objects instead of loose notes
- deterministic exports and automation hooks instead of cloud coupling

## Installation

### Public Beta

The beta ships as prebuilt binaries. End users do not need Go installed.

Installed binaries:
- `crona`
- `crona-tui`
- `crona-kernel`

Release assets now ship as:
- one platform bundle zip containing all three binaries
- one shared `crona-assets-<version>.tar.gz` archive for report/export assets
- installer scripts for Unix-like systems and Windows

### macOS And Linux

Install the current beta:

```bash
curl -fsSL https://github.com/webxsid/crona/releases/download/v0.4.0-beta.3/install-crona-tui.sh | sh
```

Force a non-interactive reinstall:

```bash
curl -fsSL https://github.com/webxsid/crona/releases/download/v0.4.0-beta.3/install-crona-tui.sh | CRONA_INSTALL_FORCE=1 sh
```

Run the installer from a downloaded file:

```bash
curl -fsSL -o /tmp/install-crona-tui.sh https://github.com/webxsid/crona/releases/download/v0.4.0-beta.3/install-crona-tui.sh
sh /tmp/install-crona-tui.sh
```

### Windows

Install from PowerShell:

```powershell
$version = "v0.4.0-beta.3"
Invoke-WebRequest "https://github.com/webxsid/crona/releases/download/$version/install-crona-tui.ps1" -OutFile "$env:TEMP\install-crona-tui.ps1"
powershell -NoProfile -ExecutionPolicy Bypass -File "$env:TEMP\install-crona-tui.ps1"
```

By default, Windows installs binaries into `%LocalAppData%\Programs\Crona\bin`, adds that directory to the user `PATH`, and stores runtime data under `%LocalAppData%\Crona`.

Override the binary install location with `CRONA_INSTALL_DIR`:

```powershell
$env:CRONA_INSTALL_DIR = "D:\tools\crona\bin"
powershell -NoProfile -ExecutionPolicy Bypass -File "$env:TEMP\install-crona-tui.ps1"
```

## Getting Started

Launch the TUI:

```bash
crona
```

The TUI starts the local kernel automatically when needed.
`crona-tui` remains available as a compatibility entrypoint.

For manual installation, download your platform bundle zip from the GitHub release, extract `crona`, `crona-tui`, and `crona-kernel`, then keep the shared assets archive from the same release if you want the bundled report templates and export assets.

You can inspect the kernel directly from the CLI:

```bash
crona kernel attach --json
crona kernel status --json
crona kernel info --json
```

## CLI

The `crona` binary exposes automation-friendly commands for the local kernel and runtime state.

Examples:

```bash
crona kernel attach --json
crona context get --json
crona timer status --json
crona update status --json
crona export calendar --repo-id 1 --json
```

## Support And Updates

Public support surfaces live on GitHub:

- Bugs: `Issues`
- Help and ideas: `Discussions`
- Release updates: `Releases`
- Roadmap reading: `ROADMAP.md`

Generate a support bundle from the TUI Support view before filing a bug when possible.

Generate shell completions:

```bash
crona completion zsh
crona completion bash
crona completion fish
```

## Runtime Layout

### Runtime Data

Default runtime directories:

- macOS prod: `~/Library/Application Support/Crona`
- macOS dev: `~/Library/Application Support/Crona Dev`
- Linux prod: `${XDG_DATA_HOME:-~/.local/share}/crona`
- Linux dev: `${XDG_DATA_HOME:-~/.local/share}/crona-dev`
- Windows prod: `%LocalAppData%\Crona`
- Windows dev: `%LocalAppData%\Crona Dev`

Override the runtime directory with `CRONA_HOME`.

On macOS and Linux, legacy `~/.crona` and `~/.crona-dev` directories migrate automatically on install or first kernel start unless `CRONA_HOME` is set.

### Binary Install Location

Default binary install directories:

- macOS/Linux: `~/.local/bin`
- Windows: `%LocalAppData%\Programs\Crona\bin`

Override the binary install directory with `CRONA_INSTALL_DIR`.

## Development

### Prerequisites

- Go `1.26+`
- `make`

### Common Tasks

```bash
make help
make build
make test
make lint
make install-kernel
make install-tui
make install-cli
```

Manual builds from the repo root:

```bash
cd kernel && go build -o ../bin/crona-kernel ./cmd/crona-kernel
cd tui && go build -o ../bin/crona-tui .
cd cli && go build -o ../bin/crona ./cmd/crona
```

### Dev Mode

The root `.env` file controls runtime mode:

```bash
CRONA_ENV=Prod
```

Set `CRONA_ENV=Dev` to enable developer helpers and `-dev` binary names.

In dev mode, install targets become:

```bash
CRONA_ENV=Dev make install-kernel install-tui install-cli
```

This produces:

- `bin/crona-kernel-dev`
- `bin/crona-tui-dev`
- `bin/crona-dev`

Developer-only TUI hotkeys:

- `f6` seeds sample data
- `f7` clears local data

### Running From Source

#### macOS And Linux

Run the kernel:

```bash
make run-kernel
```

Run the TUI:

```bash
make run-tui
```

`make run-tui` prepends repo-local `bin/` to `PATH`, so a built kernel in `bin/` is discoverable automatically.

#### Windows

Use PowerShell from the repo root:

```powershell
$env:CRONA_ENV = "Dev"
go build -o .\bin\crona-kernel-dev.exe .\kernel\cmd\crona-kernel
go build -o .\bin\crona-tui-dev.exe .\tui
$env:PATH = "$PWD\bin;$env:PATH"
.\bin\crona-dev.exe
```

If you want to start the kernel explicitly in a separate terminal:

```powershell
$env:CRONA_ENV = "Dev"
go run .\kernel\cmd\crona-kernel
```

### Testing And Linting

Run the test suite:

```bash
make test
```

Run linting:

```bash
make lint
```

Install the linter once:

```bash
make install-lint
```

## Core Concepts

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

The smallest intentional unit of work.

An issue can carry a title, estimate, notes, and lifecycle state.

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
- enforces Pomodoro-style boundaries
- emits events for subscribed clients

### Stash

A stash suspends the current context and can preserve timer state.

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

Crona can trigger local OS notifications and optional sounds at timer boundaries.

- macOS: `osascript`
- Linux: `notify-send`, with `canberra-gtk-play` when available
- Windows: PowerShell fallback

If notification tooling is unavailable, Crona skips the notification and continues running.

### Calendar Export

Crona can generate deterministic local `.ics` files for external automations.

Typical workflow:
- Crona writes `.ics` files into the configured export directory
- local automations watch that directory
- external tools import or react to those files

Crona does not require direct Google Calendar or iCloud API integration for this flow.

## Repository Layout

```text
crona/
├─ Makefile
├─ kernel/
├─ tui/
├─ cli/
├─ shared/
├─ go.work
└─ README.md
```

## Design Principles

- local-first
- authoritative data over derived state
- replayable operations
- no hidden background jobs
- UIs are clients, not controllers
- a git-like mental model for work state

## Project Status

Crona is in public beta. The core workflow is usable, but the product is still settling and some APIs, storage details, and update flows may continue to change between releases.

Current focus areas:
- kernel and TUI polish
- CLI automation
- Windows packaging and runtime support
- planning and export workflows

## License

[MIT](LICENSE)

> Crona is an opinionated, experimental project. The code is MIT licensed, but the architecture and APIs may continue to evolve quickly while the product is in beta.
