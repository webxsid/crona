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

## Adding Issues

`crona issue add` turns `- ` lines into issues in one step. Each line can carry annotations: `#repo/stream` or `#name` (prefix match), `~45m` estimate, `@friday` or `@2026-10-30` plan date, and `+tag`. Trailing phrases such as `for 45 min`, `tomorrow`, `next friday` or `in 3 days` also set the estimate and plan date. Indented lines under a `- ` line become its description.
- `crona issue add "fix login #api ~45m @today"`: Parses one line, prints the result, and asks before importing.
- `pbpaste | crona issue add -`: Reads every `- ` line from stdin. The confirmation is asked on the terminal.
- `--repo <name>` / `--stream <name>`: Default target for lines without `#`. Without them, lines go to the built-in Inbox stream, which is created the first time it is needed. Crona marks it `system`, so it cannot be renamed or deleted; move its issues out from the TUI with `e`.
- `--yes`: Imports every unflagged row without asking. `--dry-run` prints the parse as JSON and imports nothing.

Rows the parser cannot settle (an unknown or ambiguous `#` target, a repo with several streams, a bad `@` date or tag) are shown with `?` and left out of the import.

## Tags

Tags are shared by issues and habits. Names are lowercased, and spaces become `-`.
- `crona tag ls [--json]`: Lists every tag with its issue and habit counts, color, and whether it is a system tag.
- `crona tag rename <old> <new> [--merge]`: Renames a tag. Pass `--merge` to fold it into a tag that already exists.
- `crona tag color <name> <color|none>`: Sets a palette color (`blue`, `cyan`, `green`, `magenta`, `yellow`, `red`, `orange`, `lavender`) or clears it.
- `crona tag delete <name>`: Deletes a tag and removes it from every item.
- `crona tag prune`: Deletes tags no issue or habit uses.
- `crona issue tag <id> +urgent -waiting`: Adds and removes tags on an issue in one call. `crona habit tag <id> ...` does the same for a habit.

System tags (`urgent`, `waiting`, `quick`, `review`, `someday`) can't be renamed or deleted, but their color can change.

## Read-Only Summary

`crona summary` prints a compact snapshot without changing any workspace data. It accepts a date, a date range, or convenience periods:
```bash
crona summary
crona summary --yesterday
crona summary --week
crona summary --month
crona summary --last-x-days 7
```

When any worked issue was tagged, the summary adds focused time per tag. A session counts toward every tag on its issue, so the shares can add up to more than 100%.

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
