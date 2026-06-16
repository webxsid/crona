# Legacy To Homebrew

Use this guide when you are moving off the legacy install script and want to keep Crona on Homebrew afterward.

If you are on a beta build and want to stay on the beta track, install `crona-beta` instead of `crona`. The migration flow is the same either way.

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

If you only use the production install, the prod path is the one to remove.

## 5. Remove The Legacy Binaries

Remove the old binaries from your `PATH`.

Typical locations:

- `~/.local/bin/crona`
- `~/.local/bin/crona-daemon`
- `~/.local/bin/crona-tui`

If you installed somewhere else, remove those copies too.

## 6. Install Homebrew

Install the package manager version you want to keep:

```bash
brew tap webxsid/tap
brew install crona
```

For the beta track, use:

```bash
brew install crona-beta
```

## 7. Restore The Backup

Run:

```bash
crona restore <path-to-backup>
```

If the runtime directory already contains a `crona.db`, Crona prompts before overwriting it.

## After Migration

- Homebrew now owns install, update, and uninstall.
- Stable updates use `brew upgrade crona`.
- Beta updates use `brew upgrade crona-beta`.
- `crona backup` and `crona restore` only move `crona.db`.
