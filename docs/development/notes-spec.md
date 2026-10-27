---
hosted: false
title: "Notes Spec"
description: "Proposed markdown notes with YAML front matter and triageable task lines, indexed by the kernel and surfaced in a Notes view."
order: 7.2
---

## Notes

Status: proposal. Not scheduled. This document describes what a notes feature would be if built; it is not an implementation plan.

### Goal

Let users keep longer-form markdown notes inside Crona that can stand alone or attach to issues, streams, repos, and habits, and that stay reachable from the screens where that work happens.

Notes also replace the old scratchpad: a quick brain dump is just a note, with any tasks put in an `:::issues` block. Users write freely and save; saving triages the new task lines into issues through one review step instead of a form per task. There is one place to write (`n`), one place to look (the Notes view and its Inbox), and saving is what turns writing into work.

### Non-goals

- Replacing the existing short `notes` fields on issues, sessions, habit completions, habit focus sessions, and daily check-ins. Those stay as they are.
- Replacing the wellbeing check-in as the daily reflection.
- Sync, sharing, or any network dependency.
- A full markdown IDE. Editing is deliberately simple.
- A separate scratchpad feature. Dumps are notes; there is no dump/note mode switch.
- Natural-language or AI parsing of prose into tasks. Task detection is deterministic.

### Concepts

- **Note**: one `.md` file with YAML front matter.
- **Link**: an attachment from a note to a Crona entity (`repo`, `stream`, `issue`, `habit`). A note may have zero or more links. A note with no links is independent.
- **Pin**: a link flagged as important, so the note sorts first wherever that entity's notes appear.
- **Notes directory**: a user-configurable folder that holds all note files.
- **Task line**: a list item inside an `:::issues` block that Crona can turn into an issue. Lists outside a block are never tasks.
- **Triage**: the single review step that turns a note's pending task lines into issues. It runs automatically when a note is saved from Crona.
- **Inbox**: a built-in filter, not a separate entity: notes that are unlinked or still have pending task lines.

### Task lines

Line grammar (annotations, natural phrases, descriptions) and the review list are defined in the Issue Capture spec, which notes share with the `I` capture dialog and `crona issue add`. The sections below cover what is specific to notes.

#### Issues blocks

Only lines inside an issues block are task lines. Everything else in a note, including ordinary lists, is plain markdown and never becomes an issue.

```
Sync retry ideas, see the incident thread.

:::issues
- fix login redirect #api ~45m @today
  happens after the SSO change
- call bank friday
:::

- this list is just notes
```

Rules:
- A block opens with a line that is exactly `:::issues` and closes with a line that is exactly `:::`. A note may hold any number of blocks. An unclosed block runs to the end of the note.
- Inside a block, task lines and their indented descriptions follow the Issue Capture spec. Other lines in the block (blank lines, prose) are ignored.
- Blocks inside fenced code blocks are ignored.
- A line already carrying a write-back marker (see below) is triaged and is skipped.
- Markdown renderers that do not know the syntax show the block as plain text with its list intact, so notes stay readable in other editors.

#### Default target

Without a `#` annotation, a line targets the note's linked stream, the stream of its linked issue, or the note's linked repo (stream unresolved), in that order; otherwise the Inbox stream.

#### Write-back

After triage creates an issue, Crona rewrites that line in the file to append its public id:

```
- fix login redirect bug → CRN-142
```

Rules:
- Tokens are kept in the file as written; only the `→ ID` marker is appended. If a field was changed during triage, the line is not rewritten to match; the issue is the source of truth from then on.
- The created issue gets a link back to the note (`note_links` from the note to the issue), so the note shows up under that issue.
- Lines skipped during triage are left unchanged and stay pending.
- Once a line is triaged, the issue is updated only through the normal issue flows. Editing or checking the line later never changes the issue, and issue changes are never written back to the line.

#### When triage runs

- **Saving from Crona** (TUI capture dialog, TUI editor, `$EDITOR` from the Notes view, CLI `note add`) triages automatically. Only lines that are new or still pending are considered.
- **Edits made outside Crona** (another editor, Obsidian, sync) are indexed but never create issues on their own. Their new task lines stay pending, the note appears in Inbox, and the next save from Crona or an explicit `t` triages them.
- `t` in the Notes view triages a note's pending lines on demand, for example after skipping some during an earlier save.

### Storage

#### Files are the source of truth

Notes live as plain markdown files in the configured notes directory. Users can edit, version, or sync that folder with any tool (git, Obsidian, iCloud). Crona never requires its own copy to be authoritative.

File naming: `<slug>--<id>.md`, for example `retry-strategy-for-sync--n_7f3k2.md`. The `id` in front matter is what identifies a note; renaming the file does not break links.

Subfolders are allowed and scanned recursively. Crona writes new notes to the root of the directory.

#### Front matter

```yaml
---
id: n_7f3k2               # required, stable public id; generated if missing
title: Retry strategy for sync
kind: note                # note | log
date: 2026-10-20          # only for kind: log
links:
  - issue: CRN-142
  - stream: sync
  - habit: reading
pinned:
  - issue: CRN-142
tags: [design]
created: 2026-10-20T09:12:00Z
updated: 2026-10-20T10:40:00Z
---
```

Rules:
- Links reference entities by their public ids, not internal row ids.
- Unknown keys are preserved on write.
- A file without front matter is still indexed; Crona adds an `id` and `title` (from the first heading or filename) the first time it saves the note.
- A link to an entity that no longer exists is kept in the file and shown as a broken link, never silently dropped.

#### Database index

The kernel keeps an index that can always be rebuilt from the files:

- `notes`: id, relative path, title, kind, date, tags, content hash, file mtime, created, updated, excerpt, pending task count.

Exports and reports do not include note ids for now; there is no use case yet.
- `note_links`: note id, target type, target public id, pinned.

Indexing:
- Full rescan on daemon start and whenever the notes directory setting changes.
- File watching while the daemon runs; changes on disk update the index.
- When a file and the index disagree, the file wins.
- Duplicate ids across files are reported as an index warning; the most recently modified file wins until the user resolves it.

#### Configuration

- `notes_dir` in core settings. Defaults to a `notes/` folder under the resolved runtime directory. Any absolute path is allowed.
- Changing `notes_dir` re-indexes the new location. Crona never moves or deletes files when the setting changes.
- `daily_note`: `off` (default) or `on`. When on, Crona creates today's `kind: log` note on first access. An explicit action in the Notes view creates one regardless of the setting.

### Kernel and IPC

The kernel owns the index and all file writes, so the TUI and CLI stay thin. Proposed operations:

- list notes, filtered by link target, kind, tag, pinned, unlinked, inbox, or text query
- get a note (front matter plus body)
- parse a note's pending task lines into proposed issues (no writes)
- triage: create issues from an accepted list of proposals (each proposal may carry user-edited fields), write markers back, and link the note to each created issue, as one operation
- promote a short `notes` field (issue, session, habit completion, habit focus session, check-in) to a full note linked to its owner, reusing create and link
- create, save, rename, delete a note
- add, remove, pin, and unpin links
- reindex

Deletes are explicit and confirmed in the TUI. The kernel records note mutations in the op log like other entities.

The task-line parser lives in the kernel so the TUI and CLI produce identical results.

The CLI gets matching commands:
- `crona note add "text"` creates a note; `crona note add -` reads multi-line text from stdin (for example `pbpaste | crona note add -`); `crona note add` with no text opens `$EDITOR` on a new note.
- `--issue CRN-142`, `--repo crona`, `--stream sync`, etc. link the new note; without them it is unlinked and lands in the Inbox.
- `note add` triages on save. On a terminal it shows the proposals and asks for confirmation; non-interactive runs (or `--yes`) create every resolved proposal and leave flagged ones pending. `--no-triage` saves without triaging.
- `crona note ls --issue CRN-142`, `crona note ls --inbox`.
- `crona note triage <id>` triages pending lines with the same prompt, `--yes` for scripts.

### TUI

#### Notes view

A new `ViewNotes` in the Issues tab, after Issues and Habits. In the view-jump menu it uses `e`, because `n` already jumps to Session.

Layout, three panes:
- **Filters**: Inbox (default), All, Pinned, Logs, then grouped by repo/stream, issue, and habit. Inbox shows notes that are unlinked or have pending task lines, and replaces a separate Unlinked filter.
- **List**: title, link chips, pending task badge (for example `☐3`), updated time, one-line excerpt. Pinned first, then most recently updated.
- **Detail**: rendered markdown preview (glamour, including in narrow panes) that switches to an inline editor in place.

On narrow terminals the filters pane collapses and the detail replaces the list.

#### Editing

- Inline editor based on `bubbles/textarea`, showing the body only. Front matter is never edited as raw text in the TUI.
- A toggle switches between edit and rendered preview. On wide terminals edit and preview may sit side by side.
- Title, links, pins, and tags are edited through dialogs that follow the existing picker patterns (choose an issue, habit, or stream).
- Saving writes the file through the kernel, updates `updated`, and triages new task lines.
- `ctrl+g` opens the note in `$EDITOR`, inside the Notes view only (not `E`, which already means export on Issues and Habits). After the editor exits the note is re-read and triaged like any other save.

#### Quick capture

`n` (global) opens a single multi-line capture dialog:

```
┌ New note ─────────────────────────────┐
│ Link: CRN-142 (current)   [tab: none]  │
│ :::issues                              │
│ - fix login redirect #api ~45m @today  │
│ - call bank                            │
│ :::                                    │
│                                        │
│ ctrl+s save · esc cancel               │
└────────────────────────────────────────┘
```

- The link defaults to the current selection: the selected issue or habit on Issues or Habits, the active issue on Session, otherwise the active context. `tab` cycles current → none (Inbox) → pick, using the existing picker.
- One textarea, no separate title field. The title comes from the first heading, else the first line.
- `ctrl+s` saves. If the note has task lines, the triage review opens immediately; otherwise the dialog closes. Either way you return to the current screen.

There is no separate key for an independent note; choosing "none" in the dialog covers it. `n` is not bound globally or per view today. Inside dialogs `n` keeps its existing meanings (the "no" answer, the Session jump), since dialogs take focus.

#### Triage

Saving a note with new task lines opens one review list for all of them. `t` on a note in the Notes view (list or detail) opens the same list for its pending lines.

```
   title                       repo    stream  est  plan
[x] fix login redirect bug      crona   api     45m  today  
[x] release notes for beta.3    crona   docs    —    —      
[?] call bank                   Inbox   Inbox   —    —      
[?] refactor sync               crona   ?       —    fri    
space toggle · ←/→ field · enter edit field · ctrl+s create · esc skip
```

- Every field is editable: title, repo, stream, estimate, and plan date. `←`/`→` moves between fields and `enter` edits the focused one: title, estimate, and plan date edit inline; repo and stream use the existing pickers. Changing the repo clears the stream unless the repo has exactly one.
- All proposals start selected except flagged ones (`?`: ambiguous, unknown, or unresolved repo or stream). A flagged row becomes selected once its target is resolved.
- `ctrl+s` creates the selected issues in one kernel call, writes markers back, and refreshes the note.
- `esc` skips triage. The note is still saved, unselected lines stay pending, and the note stays in Inbox.
- A `t` on a note with no pending task lines shows a short status message instead of the dialog.

#### Key summary

| Key | Where | Does |
|---|---|---|
| `n` | global | Quick capture dialog |
| `e` | view jump | Open the Notes view |
| `t` | Notes view | Re-triage pending lines of the selected note |
| `ctrl+g` | Notes view | Open the note in `$EDITOR` |

Everything else reuses existing keys (`enter` to open and edit, `ctrl+s` to save, the existing delete and confirm flow, `/` to filter).

#### Other screens

Other screens get a light touch only: a notes count indicator (for example `✎2`) on issue rows, habit rows, and the active session header, and a key on the selected item that opens the Notes view filtered to it. Dedicated notes panes on Daily, Issues, Habits, and Session are deferred until the Notes view has proven useful.

Notes linked to a repo or stream also count toward the issues beneath them, shown separately as inherited.

Additional touches:
- Review and momentum list the notes created in their period.
- The morning brief and end-of-day flow show a nudge when Inbox is non-empty, with a key to jump to the Notes view on Inbox.
- Wherever a short `notes` field is shown (issue, session, habit completion, habit focus session, check-in), an action promotes it to a full note linked to its owner.
