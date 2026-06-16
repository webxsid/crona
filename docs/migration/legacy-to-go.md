# Legacy To Go

Use this guide when you are moving off the legacy install script and want to keep Crona installed from source with `go install`.

## Before You Start

Stop every running Crona process before you touch the install.

Close any open TUI windows, CLI sessions, and background Crona processes. If a process is still running after you close the app, stop it manually with your operating system tools.

## 1. Download The Latest Beta Installer

Open the latest beta release page on GitHub Releases and download the legacy installer script for your platform.

- macOS and Linux: `install-crona-tui.sh`
- Windows: `install-crona-tui.ps1`

## 2. Run The Legacy Installer Once

Make the script executable, then run it and allow it to replace the existing install when prompted.

```bash
chmod +x install-crona-tui.sh
./install-crona-tui.sh
```

Keep the existing install selected when the script asks whether to overwrite the current version.

## 3. Back Up Your Database

Run:

```bash
crona backup
```

The command prints the backup path. Keep that path for the restore step.

## 4. Remove The Runtime Directory

Remove the Crona runtime directory after the backup is complete.

- macOS prod: `~/Library/Application Support/Crona`
- macOS dev: `~/Library/Application Support/Crona Dev`
- Linux prod: `${XDG_DATA_HOME:-~/.local/share}/crona`
- Linux dev: `${XDG_DATA_HOME:-~/.local/share}/crona-dev`

## 5. Remove The Legacy Binaries

Remove the old binaries from your `PATH`.

Typical locations:

- `~/.local/bin/crona`
- `~/.local/bin/crona-daemon`
- `~/.local/bin/crona-tui`

## 6. Install From Source

Install the source build with:

```bash
go install github.com/webxsid/crona/...@latest
```

Make sure your `GOBIN` or `PATH` includes the directory where Go installs binaries.

## 7. Restore The Backup

Run:

```bash
crona restore <path-to-backup>
```

If the runtime directory already contains a `crona.db`, Crona prompts before overwriting it.

## After Migration

- Source installs now own their update path through `go install github.com/webxsid/crona/...@latest`.
- `crona backup` and `crona restore` only move `crona.db`.
- Keep the same `go install` command handy for future reinstalls or upgrades.
