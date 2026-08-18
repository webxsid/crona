---
hosted: true
title: CLI and Local Engine
description: Execute Crona commands, manage the daemon process, and configure shell completions.
order: 5.13
---

Crona is controlled via the `crona` executable. It acts as a client that routes instructions to `crona-daemon` over local IPC Unix sockets or Windows named pipes.

## Launcher & CLI Entrypoint

Running `crona` without arguments:
1. Verifies if the daemon process (`crona-daemon`) is active.
2. Launches `crona-daemon` in the background if it is missing.
3. Launches the Bubble Tea TUI (`crona-tui`) in your active terminal session.

## Daemon Administration

The `crona daemon` command group monitors and manages the background daemon process:
```bash
# Get daemon process stats and connection details
crona daemon status

# Inspect running engine metadata and versions
crona daemon info

# Restart the local daemon process
crona daemon restart

# Wipes the local database and configuration (destructive)
crona daemon wipe-data
```

Add the `--json` flag to commands to output structured JSON data for scripting.

## Context Checkout Commands

Manage the active `{ repository -> stream -> issue }` selection path:
- `crona context get`: Prints the active context hierarchy.
- `crona context switch-repo --id <id>`: Checks out a repository.
- `crona context switch-stream --id <id>`: Checks out a stream.
- `crona context switch-issue --id <id>`: Checks out a task.
- `crona context clear`: Resets the active context.

## Timer & Focus Controls

Interact with active focus sessions:
- `crona timer status`: Inspects elapsed duration, segments, and active cycle states.
- `crona timer start --from-context`: Starts the session timer for the checked-out issue.
- `crona timer pause`: Suspends the timer.
- `crona timer resume`: Resumes the suspended timer.
- `crona timer end`: Terminates the session and prompts for the commit summary.

## Read-Only Summary

`crona summary` prints a compact snapshot without changing any workspace data. It accepts a date, a date range, or convenience periods:
```bash
crona summary
crona summary --yesterday
crona summary --week
crona summary --month
crona summary --last-x-days 7
```

Use `crona export summary` when you need a Markdown, PDF, or clipboard artifact instead.

## Shell Completions

Generate native shell tab-completion definitions:
```bash
# Output completion scripts for your active shell
crona completion zsh
crona completion bash
crona completion fish
```
To load them, source the output in your shell configuration profile (e.g., `source <(crona completion zsh)`).
