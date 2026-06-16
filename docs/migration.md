# Migration Guide

Use this guide when switching Crona install methods or release channels.

This is the canonical handoff for moving between Homebrew, winget, the legacy install script, or the `crona-beta`/stable channels. Crona's Updates view points here once the migration banner appears.

The safe flow is:

1. Create a backup of your database.
2. Remove the old install.
3. Remove the old runtime directory if you want a clean switch.
4. Install Crona again with the method you want to keep.
5. Restore the database into the fresh runtime directory.

After migration, the package manager owns updates:

- Homebrew: `brew upgrade crona` or `brew upgrade crona-beta`
- Winget: `winget upgrade --id Webxsid.Crona -e`
- Source installs: rerun the `go install` command

## 1. Back Up

Run:

```bash
crona backup
```

The command prints the absolute backup path, for example:

```text
backup file: /Users/alice/Library/Application Support/Crona Backups/crona-db-20260616-062010.db
```

Keep that file. It is stored outside Crona's runtime directory, so uninstalling Crona will not delete it.

## 2. Remove The Old Install

Use the package manager or installer you are migrating away from:

- Homebrew stable:
  ```bash
  brew uninstall crona
  ```
- Homebrew beta:
  ```bash
  brew uninstall crona-beta
  ```
- Winget:
  ```powershell
  winget uninstall --id Webxsid.Crona -e
  ```
- Script/manual installs:
  remove the installed binaries from your `PATH`, or rerun your package manager install after cleanup.
  If you came from the legacy install script, move to a managed package installer after cleanup.

## 3. Remove The Old Runtime Dir

If you want a clean reinstall, remove the runtime directory after backing up the database:

- macOS prod: `~/Library/Application Support/Crona`
- macOS dev: `~/Library/Application Support/Crona Dev`
- Linux prod: `${XDG_DATA_HOME:-~/.local/share}/crona`
- Linux dev: `${XDG_DATA_HOME:-~/.local/share}/crona-dev`
- Windows prod: `%LocalAppData%\Crona`
- Windows dev: `%LocalAppData%\Crona Dev`

If you keep the runtime directory in place, `crona restore` still only restores `crona.db`, but a full clean switch usually starts from a blank runtime dir.

## 4. Reinstall

Use the preferred managed installer:

- Homebrew stable:
  ```bash
  brew install webxsid/tap/crona
  ```
- Homebrew beta:
  ```bash
  brew install webxsid/tap/crona-beta
  ```
- Winget:
  ```powershell
  winget install --id Webxsid.Crona -e
  ```

## 5. Restore

Run:

```bash
crona restore <path-to-backup>
```

If Crona already has a `crona.db` at the target runtime dir, the command asks before overwriting it.

## Notes

- `crona backup` only copies `crona.db`.
- `crona restore` only restores `crona.db`.
- Diagnostics, runtime state, and install metadata are intentionally left alone.
- If you are switching between package managers, use the new installer-specific update command after the reinstall:
  - Homebrew: `brew upgrade crona`
  - Homebrew beta: `brew upgrade crona-beta`
  - Winget: `winget upgrade --id Webxsid.Crona -e`
