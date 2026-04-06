# Install

## Release Channels

- `stable` is the preferred channel for general users.
- `beta` is for testers who want pre-release builds and faster iteration.
- `1.0.0` will be the first stable milestone.
- `v1.0.0-beta.2` is the current prerelease build for tester validation before the first stable release.

See the published builds on [GitHub Releases](https://github.com/webxsid/crona/releases).

## Release Artifacts

End users do not need Go installed.

Installed binaries:
- `crona`
- `crona-tui`
- `crona-kernel`

Release assets ship as:
- one platform bundle zip containing all three binaries
- one shared `crona-assets-<version>.tar.gz` archive for report/export assets
- installer scripts for Unix-like systems and Windows

## macOS And Linux

Install a specific release:

```bash
curl -fsSL https://github.com/webxsid/crona/releases/download/<version>/install-crona-tui.sh | sh
```

Force a non-interactive reinstall:

```bash
curl -fsSL https://github.com/webxsid/crona/releases/download/<version>/install-crona-tui.sh | CRONA_INSTALL_FORCE=1 sh
```

Run the installer from a downloaded file:

```bash
curl -fsSL -o /tmp/install-crona-tui.sh https://github.com/webxsid/crona/releases/download/<version>/install-crona-tui.sh
sh /tmp/install-crona-tui.sh
```

## Windows

Install from PowerShell:

```powershell
$version = "<version>"
Invoke-WebRequest "https://github.com/webxsid/crona/releases/download/$version/install-crona-tui.ps1" -OutFile "$env:TEMP\install-crona-tui.ps1"
powershell -NoProfile -ExecutionPolicy Bypass -File "$env:TEMP\install-crona-tui.ps1"
```

By default, Windows installs binaries into `%LocalAppData%\Programs\Crona\bin`, adds that directory to the user `PATH`, and stores runtime data under `%LocalAppData%\Crona`.

Override the binary install location with `CRONA_INSTALL_DIR`:

```powershell
$env:CRONA_INSTALL_DIR = "D:\tools\crona\bin"
powershell -NoProfile -ExecutionPolicy Bypass -File "$env:TEMP\install-crona-tui.ps1"
```

## Manual Installation

Download your platform bundle zip from the release page, extract `crona`, `crona-tui`, and `crona-kernel`, and keep the shared assets archive from the same release if you want the bundled report templates and export assets.

The TUI starts the local kernel automatically when needed. `crona-tui` remains available as a compatibility entrypoint.

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

## Updates

- General users should stay on the `stable` update channel.
- Beta testers can opt into the `beta` channel from the TUI settings.
- The current beta train is the `1.0.0` prerelease validation path.

Use the in-app `Updates` view to check, read notes, and install supported updates.
