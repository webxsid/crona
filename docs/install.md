# Install

## Release Artifacts

End users do not need Go installed.

Installed binaries:

- `crona`
- `crona-daemon`
- `crona-tui`

`crona-daemon` is the background local engine binary. Most user-facing docs call it the daemon or local engine because it owns storage, timers, reminders, update checks, and local IPC.

Release assets ship as:

- one platform bundle zip containing all three binaries
- one shared `crona-assets-<version>.tar.gz` archive for legacy script installers and release compatibility
- installer scripts for Unix-like systems and Windows

> [!IMPORTANT]
> Crona 1.6.x shows an install-script deprecation banner in the Updates view.
> GitHub install scripts are legacy fallback and will stop being exposed in-app moving forward.
> Prefer a managed package installer when possible, and use [migration.md](migration.md) if you need to switch install methods or release channels.

## macOS And Linux

Prefer Homebrew:

```bash
brew tap webxsid/tap
brew install crona
```

or:

```bash
brew install webxsid/tap/crona
```

Prerelease beta tags publish a separate opt-in formula:

```bash
brew install webxsid/tap/crona-beta
```

Homebrew users update with:

```bash
brew upgrade crona
```

Beta users update with:

```bash
brew upgrade crona-beta
```

Legacy install script fallback:

```bash
curl -fsSL https://crona.work/install.sh | sh
```

This script is a legacy fallback, not the preferred path. Crona 1.6.x warns about it in the Updates view, and later releases stop exposing it there.

If you need to switch install methods or release channels, use the migration guide:

```bash
https://crona.work/migration
```

For step-by-step destination-specific migration flows, see:

- [Legacy to Homebrew](migration/legacy-to-brew.md)
- [Legacy to Go](migration/legacy-to-go.md)
- [Legacy to Winget](migration/legacy-to-winget.md)

The guide uses `crona backup` and `crona restore <path>` to preserve your database while you reinstall.
When you remove Crona with your package manager, only the binaries go away. Your runtime data stays behind until you remove it yourself or run `crona daemon wipe-data --force` before uninstalling.
If you want a fully clean switch after uninstalling, remove the runtime directory manually using the paths listed in the migration guide.

Legacy direct GitHub install:

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

Prefer winget:

```powershell
winget install --id Webxsid.Crona -e
```

The winget package installs the Crona bundle and exposes `crona`, `crona-daemon`, and `crona-tui`.

Winget users update with:

```powershell
winget upgrade --id Webxsid.Crona -e
```

Legacy install script fallback:

```powershell
$version = "<version>"
Invoke-WebRequest "https://github.com/webxsid/crona/releases/download/$version/install-crona-tui.ps1" -OutFile "$env:TEMP\install-crona-tui.ps1"
powershell -NoProfile -ExecutionPolicy Bypass -File "$env:TEMP\install-crona-tui.ps1"
```

For managed installs, winget owns the update command, and Crona only surfaces it in the Updates view.

By default, Windows installs binaries into `%LocalAppData%\Programs\Crona\bin`, adds that directory to the user `PATH`, and stores runtime data under `%LocalAppData%\Crona`.

Override the binary install location with `CRONA_INSTALL_DIR`:

```powershell
$env:CRONA_INSTALL_DIR = "D:\tools\crona\bin"
powershell -NoProfile -ExecutionPolicy Bypass -File "$env:TEMP\install-crona-tui.ps1"
```

## Manual Installation

Download your platform bundle zip from the release page and extract `crona`, `crona-daemon`, and `crona-tui`. The embedded templates and alert assets ship inside the binaries, so the shared assets archive is only needed for the legacy script installers.

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

Users can switch release tracks from the TUI settings.
The default track follows normal releases, and testers can opt into upcoming builds.

Use the in-app `Updates` view to check release status, read notes, and get the right migration or package-manager command.

- Homebrew installs never self-update from inside Crona.
- Winget installs never self-update from inside Crona.
- The TUI shows the source-aware update command for the current install type, but does not execute it.
- When Crona asks you to migrate, back up with `crona backup`, uninstall with your package manager, remove runtime data if you want a clean reset, then reinstall and restore with `crona restore <path>`.
- The migration guide at [docs/migration.md](migration.md) is the canonical handoff for switching install methods or release channels.
- Script installs rerun the install script.
- Winget installs use `winget upgrade --id Webxsid.Crona -e`.
- Source installs show the `go install` command.
- Manual installs and unknown installs are directed to the GitHub release page.

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

Bundled alert sounds include royalty-free effects by Universfield on Pixabay:

> Sound Effect by [Universfield](https://pixabay.com/users/universfield-28281460/?utm_source=link-attribution&utm_medium=referral&utm_campaign=music&utm_content=494248) from [Pixabay](https://pixabay.com//?utm_source=link-attribution&utm_medium=referral&utm_campaign=music&utm_content=494248)

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
