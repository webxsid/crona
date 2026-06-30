# Legacy To Winget

Use this guide when you are moving off the legacy install script and want to keep Crona managed by winget on Windows.
Stop Crona, run the beta installer once, back up the database, clear the old runtime, remove the old binaries, install Winget, and restore the backup.

Estimated time: 20-30 minutes, mainly because the Windows install and restore steps can take a little longer to complete.

Use this guide when the final install should be managed by Winget rather than the legacy script.

## Before You Start

Stop every running Crona process before you touch the install.

Close any open TUI windows, CLI sessions, and background Crona processes. If a process is still running after you close the app, stop it manually with Task Manager or `Stop-Process`.

## 1. Download The Latest Beta Installer

Download `install-crona-tui.ps1` from the latest beta release page on GitHub Releases.

## 2. Run The Legacy Installer Once

Run the script and allow it to replace the existing install when prompted.

```powershell
$version = "<version>"
Invoke-WebRequest "https://github.com/webxsid/crona/releases/download/$version/install-crona-tui.ps1" -OutFile "$env:TEMP\install-crona-tui.ps1"
powershell -NoProfile -ExecutionPolicy Bypass -File "$env:TEMP\install-crona-tui.ps1"
```

Keep the existing install selected when the script asks whether to overwrite the current version.

## 3. Back Up Your Database

Run:

```powershell
crona backup
```

The command prints the backup path. Keep that path for the restore step.

## 4. Remove The Runtime Directory

Remove the Crona runtime directory after the backup is complete.

- Windows prod: `%LocalAppData%\Crona`
- Windows dev: `%LocalAppData%\Crona Dev`

## 5. Remove The Legacy Binaries

Remove the old binaries from the winget user install directory if they are still present.

Typical location:

- `%LocalAppData%\Programs\Crona\bin`

## 6. Install Winget

Install the package manager version you want to keep:

```powershell
winget install --id Webxsid.Crona -e
```

## 7. Restore The Backup

Run:

```powershell
crona restore <path-to-backup>
```

If the runtime directory already contains a `crona.db`, Crona prompts before overwriting it.

## After Migration

- Winget owns install, update, and uninstall.
- Updates use `winget upgrade --id Webxsid.Crona -e`.
- `crona backup` and `crona restore` only move `crona.db`.

## Contact

If you hit a Windows-specific install problem, contact me directly:

- Email: `me@webxsid.com`
- WhatsApp: `+91-6375728437`
