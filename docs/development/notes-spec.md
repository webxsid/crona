---
hosted: false
title: "Notes Spec"
description: "Proposed markdown notes with YAML front matter, indexed by the kernel and surfaced in a Notes view."
order: 7.2
---

## Notes

Status: proposal. Not scheduled. This document describes what a notes feature would be if built; it is not an implementation plan.

### Goal

Let users keep longer-form markdown notes inside Crona that can stand alone or attach to issues, streams, repos, and habits, and that stay reachable from the screens where that work happens.

### Non-goals

- Replacing the existing short `notes` fields on issues, sessions, habit completions, habit focus sessions, and daily check-ins. Those stay as they are.
- Replacing the wellbeing check-in as the daily reflection.
- Sync, sharing, or any network dependency.
- A full markdown IDE. Editing is deliberately simple.

### Concepts

- **Note**: one `.md` file with YAML front matter.
- **Link**: an attachment from a note to a Crona entity (`repo`, `stream`, `issue`, `habit`). A note may have zero or more links. A note with no links is independent.
- **Pin**: a link flagged as important, so the note sorts first wherever that entity's notes appear.
- **Notes directory**: a user-configurable folder that holds all note files.

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

- `notes`: id, relative path, title, kind, date, tags, content hash, file mtime, created, updated, excerpt.
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

- list notes, filtered by link target, kind, tag, pinned, unlinked, or text query
- get a note (front matter plus body)
- create, save, rename, delete a note
- add, remove, pin, and unpin links
- reindex

Deletes are explicit and confirmed in the TUI. The kernel records note mutations in the op log like other entities.

The CLI gets matching commands, for example `crona note add --issue CRN-142 "..."` and `crona note ls --issue CRN-142`.

### TUI

#### Notes view

A new `ViewNotes` in the Issues tab, after Issues and Habits. In the view-jump menu it uses `e`, because `n` already jumps to Session.

Layout, three panes:
- **Filters**: All, Unlinked, Pinned, Logs, then grouped by repo/stream, issue, and habit.
- **List**: title, link chips, updated time, one-line excerpt. Pinned first, then most recently updated.
- **Detail**: rendered markdown preview that switches to an inline editor in place.

On narrow terminals the filters pane collapses and the detail replaces the list.

#### Editing

- Inline editor based on `bubbles/textarea`, showing the body only. Front matter is never edited as raw text in the TUI.
- A toggle switches between edit and rendered preview. On wide terminals edit and preview may sit side by side.
- Title, links, pins, and tags are edited through dialogs that follow the existing picker patterns (choose an issue, habit, or stream).
- Saving writes the file through the kernel and updates `updated`.
- Opening the note in `$EDITOR` stays available as a secondary key inside the Notes view only (not `E`, which already means export on Issues and Habits). After the editor exits the note is re-read.

#### Quick capture

- `n` (global): new note linked to the current selection. On Issues or Habits that is the selected issue or habit. On Session it is the active issue. Elsewhere it falls back to the active context.
- `N` (global): new independent note.

Both open a small capture dialog (title and body). Saving stays on the current screen. Neither key is bound globally or per view today. Inside dialogs `n` keeps its existing meanings (the "no" answer, the Session jump), since dialogs take focus.

#### Other screens

Other screens get a light touch only: a notes count indicator (for example `✎2`) on issue rows, habit rows, and the active session header, and a key on the selected item that opens the Notes view filtered to it. Dedicated notes panes on Daily, Issues, Habits, and Session are deferred until the Notes view has proven useful.

Notes linked to a repo or stream also count toward the issues beneath them, shown separately as inherited.

### Open questions

- Which key in the Notes view opens `$EDITOR`.
- Whether glamour rendering is affordable in narrow panes or a lighter renderer is needed.
- Whether review and momentum should list notes created in the period.
- Whether to offer "promote to note" for the existing short `notes` fields.
- How note ids surface in exports and reports, if at all.
