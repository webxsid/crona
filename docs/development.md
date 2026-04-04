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
