---
hosted: true
title: Calendar and File Automation
description: Integrate Crona exports, reports, and calendar feeds with external tools.
order: 5.11
---

Crona is local-first. Instead of building direct cloud connectors, the system outputs structured local files (Markdown, CSV, and iCalendar) that can be integrated into external desktop workflows, personal scripts, or calendar clients.

## iCalendar (.ics) Automation

Crona generates deterministic `.ics` files representing planned work and recorded sessions:
- **Separation**: Exposes separate feeds for issues (`issues.ics`) and sessions (`sessions.ics`).
- **Scoping**: Calendar exports target a specific repository.
- **Configurable Paths**: Specify a dedicated calendar directory in the **Config** view.

### Automated Importing Patterns

Because the `.ics` files are updated deterministically, local calendar clients or automation helpers can ingest them:
- **macOS/Linux**: Sync with native calendars via loopback servers (e.g., using a local HTTP server exposing the ICS directory to Apple Calendar or Thunderbird).
- **Directory Watchers**: Run background services (e.g., `entr` or `fswatch`) that trigger synchronization or push tasks when `issues.ics` updates.

## Report File Watchers

Narrative reports are written to the configured reports folder. You can configure local documentation wikis or personal knowledge bases (like Obsidian or Logseq) to index these folders directly.

### Syncing to Obsidian
Configure Crona to write Markdown reports directly into your Obsidian Vault:
1. Open the TUI **Config** view.
2. Edit the `Reports Directory` path to target your vault directory (e.g., `/Users/username/Vault/Crona-Reports`).
3. Generated reports will instantly appear as notes, utilizing the Obsidian-friendly YAML frontmatter emitted by Crona.

## Output Separation

Keep report files and calendar files in separate directories:
- **Reports Folder**: Point this to human-facing document stores or note vaults.
- **ICS Folder**: Point this to system directories monitored by sync daemons or local calendar services.
