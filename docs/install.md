# Install

## Release Channels

- `stable` is the preferred channel for general users.
- `beta` is for testers who want pre-release builds and faster iteration.
- `v1.0.2` is the current stable release.

See the published builds on [GitHub Releases](https://github.com/webxsid/crona/releases).

## Release Artifacts

End users do not need Go installed.

Installed binaries:
- `crona`
- `crona-tui`
- `crona-kernel`

`crona-kernel` is the internal binary name for Crona's background local engine. Most user-facing docs call it the local engine because it owns storage, timers, reminders, update checks, and local IPC.

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

The TUI starts the local engine automatically when needed. `crona-tui` remains available as a compatibility entrypoint.

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

On macOS and Linux, legacy `~/.crona` and `~/.crona-dev` directories migrate automatically on install or first local engine start unless `CRONA_HOME` is set.

### Binary Install Location

Default binary install directories:

- macOS/Linux: `~/.local/bin`
- Windows: `%LocalAppData%\Programs\Crona\bin`

Override the binary install directory with `CRONA_INSTALL_DIR`.

## Updates

- General users should stay on the `stable` update channel.
- Beta testers can opt into the `beta` channel from the TUI settings.
- Stable users receive normal releases; beta users can opt into future prerelease validation builds.

Use the in-app `Updates` view to check, read notes, and install supported updates.

## Notifications And Alerts

Alerts are emitted by the local engine. The TUI configures and tests them, but the background engine is the process that decides when to fire:

- timer boundary alerts
- focus inactivity alerts when an active work session runs too long without TUI activity
- update-available alerts
- support/export completion alerts
- scheduled reminders such as nightly check-in reminders

Scheduled reminders and inactivity alerts are local-only and only fire while the local engine is running.

Supported notification helpers by OS:

- macOS:
  - notifications: `terminal-notifier`, fallback `osascript`
  - sound playback: `afplay`
- Linux:
  - notifications: `notify-send`
  - sound playback: `paplay`, `aplay`, `play`, fallback `canberra-gtk-play`
- Windows:
  - notifications: `BurntToast` when installed, fallback PowerShell toast delivery
  - sound playback: PowerShell `SoundPlayer`

The `Alerts` view shows the active backend and whether subtitle, urgency, icon, and bundled sound support are currently available on the running machine.

## PDF Rendering

Markdown export works without extra tooling. PDF export requires local renderer support.

Current renderer expectations:

- Daily and weekly narrative PDF exports require `weasyprint`
- Repo, stream, and issue-rollup PDF exports require `pandoc` plus one supported PDF engine:
  - `tectonic`
  - `weasyprint`
  - `wkhtmltopdf`
  - `xelatex`
  - `pdflatex`

Renderer availability is detected at runtime and surfaced in the TUI `Config` view and through `export.assets.get`.

If the required renderer chain is missing, PDF export remains unavailable but markdown export still works.
