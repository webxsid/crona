# Development

## Prerequisites

- Go `1.26+`
- `make`

## Common Tasks

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

## Dev Mode

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
- `f8` prepares a local update simulation

## Running From Source

### macOS And Linux

Run the kernel:

```bash
make run-kernel
```

Run the TUI:

```bash
make run-tui
```

`make run-tui` prepends repo-local `bin/` to `PATH`, so a built kernel in `bin/` is discoverable automatically.

The TUI writes the terminal tab/window title while it is running. During active focus sessions the title includes the issue/session context and elapsed timer state; otherwise it includes the current repo/stream context and view. On normal exit the TUI resets the title with a best-effort terminal control sequence.

### Windows

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

## Testing And Linting

Run the full test suite:

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

Targeted module checks are also used heavily during refactors:

```bash
go test ./kernel/internal/...
go test ./tui/internal/tui/... ./tui/internal/api
go test ./cli/...
```

## Notifications And Alerts In Development

The kernel owns alert delivery. Running the TUI alone is not enough for scheduled reminders; the local kernel must stay up.

Current backend expectations by OS:

- macOS: `terminal-notifier`, fallback `osascript`, sound via `afplay`
- Linux: `notify-send`, sound via `paplay`, `aplay`, `play`, or `canberra-gtk-play`
- Windows: `BurntToast` preferred, fallback PowerShell toast delivery

The TUI `Alerts` view is the easiest smoke-test surface:

- `Test Notification`
- `Test Sound`
- adjust focus inactivity alert threshold/repeat controls
- create a check-in reminder for a near-future time

Focus inactivity alerts are kernel-owned. During active focus sessions the TUI reports throttled keypress activity with `timer.activity.touch`; if no activity is reported for the configured threshold, the kernel sends a review-session alert and repeats on the configured interval.

## PDF Rendering In Development

The export layer uses two PDF paths:

- daily/weekly narrative PDF: `weasyprint`
- repo/stream/issue-rollup PDF: `pandoc` plus `tectonic`, `weasyprint`, `wkhtmltopdf`, `xelatex`, or `pdflatex`

Useful local checks:

```bash
go test ./kernel/internal/export
go test ./kernel/internal/... ./shared/... ./tui/internal/api
```

If renderer tooling is missing, markdown export should still work and the runtime asset status should mark PDF rendering as unavailable.

## Repository Layout

```text
crona/
├─ cli/
├─ kernel/
├─ shared/
├─ tui/
├─ docs/
├─ Makefile
├─ go.work
└─ README.md
```
