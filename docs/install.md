---
title: "Install Crona"
description: "Install Crona on macOS, Linux, or Windows and start your first local workday."
hosted: true
order: 0.5
---

# Install Crona

Crona is installed as a small local toolkit: the `crona` launcher, the terminal UI, and the background daemon that keeps timers and local state running.

You do not need Go installed for a normal package-manager install.

## macOS and Linux

Install with Homebrew:

```bash
brew install webxsid/tap/crona
```

Then start Crona:

```bash
crona
```

Homebrew handles updates:

```bash
brew upgrade crona
```

## Windows

Install with Scoop:

```powershell
scoop bucket add webxsid https://github.com/webxsid/scoop-bucket
scoop install crona
```

Start Crona from PowerShell:

```powershell
crona
```

Update it with:

```powershell
scoop update crona
```

## macOS Companion

The macOS Companion adds quick issue creation, check-ins, and a floating timer without replacing the TUI. It uses the same local daemon and data as Crona's terminal surfaces.

[Open the macOS Companion](https://crona.work/companions/)

## Where Crona keeps data

Crona stores its runtime data locally:

- macOS: `~/Library/Application Support/Crona`
- Linux: `${XDG_DATA_HOME:-~/.local/share}/crona`
- Windows: `%LocalAppData%\Crona`

Set `CRONA_HOME` when you need a different location.

## Next step

Open [Getting Started](guides/getting-started.md) and create one issue, run one session, and review the result.
