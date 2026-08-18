---
hosted: true
title: macOS Shortcuts Workflows
description: Integrate Crona with macOS Shortcuts for TUI launchers, fast context switching, and automated exports.
order: 5.12
---

On macOS, you can use the native **Shortcuts** app to execute Crona CLI commands, linking them to global hotkeys, menu bar items, or automated triggers.

## General Shortcut Setup

Every workflow below utilizes the same Shortcuts action sequence:
1. Open the macOS **Shortcuts** app and create a new Shortcut.
2. Add a **Run Shell Script** action.
3. Configure the shell (e.g., `/bin/zsh` or `/bin/bash`).
4. Ensure the `crona` binary is accessible on the script's `PATH`. If needed, prepend the path override:
   ```bash
   export PATH="$HOME/.local/bin:$PATH"
   crona <command>
   ```

## Workflow 1: Quick Workspace Launchers

Create quick launchers that switch repositories or streams and then open the TUI immediately:

### Open Work Repository
```bash
export PATH="$HOME/.local/bin:$PATH"
crona context switch-repo --id 1 --json
# Launches the TUI in the terminal
open -a Terminal "$(which crona)"
```

### Open Personal Stream
```bash
export PATH="$HOME/.local/bin:$PATH"
crona context switch-stream --id 3 --json
open -a Terminal "$(which crona)"
```

## Workflow 2: Hotkey-Triggered Focus Sessions

Set up global keyboard shortcuts to start or stop focus timers on the currently checked-out issue:

### Start Active Focus
```bash
export PATH="$HOME/.local/bin:$PATH"
crona timer start --from-context --json
```

### Stop Focus and Log Session
Create a Shortcut that prompts for input (the commit message), then ends the session:
```bash
export PATH="$HOME/.local/bin:$PATH"
# Reads the text input from the previous Shortcut action
crona timer end --message "$1" --json
```

## Workflow 3: Automated iCalendar Updates

Create an automated calendar sync Shortcut that exports your records and triggers a folder action:
```bash
export PATH="$HOME/.local/bin:$PATH"
crona export calendar --json
```
Pin this Shortcut to run hourly via a launchd daemon, or trigger it automatically when opening your email client.
