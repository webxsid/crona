---
hosted: false
title: "Tags Spec"
description: "Proposed free-form tags on issues and habits, shared with notes, for filtering, grouping, and reporting across repos and streams."
order: 7.3
---

## Tags

Status: proposal. Not scheduled. This document describes what a tags feature would be if built; it is not an implementation plan.

### Goal

Let users label issues and habits with lightweight tags that cut across the repo → stream hierarchy, then use those tags to filter lists, plan the day, and slice reports. Example: `urgent`, `waiting`, `quick`, `review`, `health`, `client-acme`.

Repos and streams answer "where does this work live". Tags answer "what kind of work is this" or "what state is it in", independent of where it lives.

### Non-goals

- Replacing repos or streams. Tags never decide where an issue belongs.
- Hierarchical or nested tags. `client/acme` is just a tag name containing a slash.
- Tag-level permissions, colors per user, or sharing.
- Tags on sessions, check-ins, or habit completions. They inherit from their issue or habit where useful (see Reports).
- Automation rules triggered by tags. A future rules engine may use them; this spec does not.
- Tags on repos or streams. That would duplicate the hierarchy.
- The macOS companion. Showing or filtering by tags there is left to the companion repo.

### Concepts

- **Tag**: a short label. One global vocabulary shared by issues, habits, and notes.
- **Tagging**: the attachment of a tag to one issue or habit. An item may have any number of tags.
- **Tag filter**: a set of tags applied to a list; matching items are shown.
- **System tag**: a tag Crona ships with. It is always in the vocabulary and cannot be deleted.

### System tags

Crona seeds these on first run (and on upgrade, for any that are missing):

| Tag | Meaning |
|---|---|
| `urgent` | Needs attention soon. |
| `waiting` | Blocked on someone or something else. |
| `quick` | Small enough to do in a spare slot. |
| `review` | Needs reviewing, by you or someone else. |
| `someday` | Not planned; kept for later. |

Rules:
- System tags are marked `system` and listed under System in the Tags view.
- They cannot be deleted or renamed, and other tags cannot be merged into or renamed onto them. Their color can be changed.
- Removing them from items works like any other tag.
- "Remove unused tags" never touches system tags.
- If a user tag with the same name already exists at seed time, it is marked `system` in place rather than duplicated.

### Tag names

- Lowercase letters, digits, `-`, `_`, and `/`. Input is lowercased and spaces become `-`.
- 1–32 characters. Must start with a letter or digit.
- Unique across the vocabulary. `Urgent` and `urgent` are the same tag.
- Shown in the UI as `+urgent`, the same form used as an inline token (see Notes integration). `#` stays reserved for repos and streams.

### Storage

- `tags`: id, name, optional color (one of the theme's palette names), system flag, created, updated.
- `issue_tags`: issue id, tag id.
- `habit_tags`: habit id, tag id.

Rules:
- A tag is created implicitly the first time it is applied. There is no separate "create tag" step.
- A tag with no taggings and no note references is kept until explicitly deleted, so it stays in autocomplete. The Tags view offers "remove unused tags".
- Deleting a tag removes all its taggings after confirmation that shows the counts.
- Renaming a tag updates every tagging. Renaming onto an existing name merges the two after confirmation.
- Archived, done, and abandoned issues keep their tags. Inactive habits keep theirs.
- Tag mutations are recorded in the op log like other entities, so they fit the sync design in Feature Design.

Shared types gain `Tags []string` on `Issue` and `Habit` (names, not ids), so clients never resolve tag ids themselves.

### Kernel and IPC

Proposed operations:

- list tags, with usage counts per issue, habit, and note
- set an item's tags (replace the full set), add tags, remove tags, for issues and habits
- bulk add or remove a tag across a list of issues or habits
- rename, merge, set color, delete a tag
- existing list operations (`issue.list`, `issue.list_all`, `habit.list`, `habit.list_all`, `habit.list_due`, daily plan) accept an optional tag filter

The kernel validates and normalizes names, so the TUI and CLI never disagree on what a valid tag is.

#### Filter semantics

A tag filter has two parts:
- `any`: item has at least one of these tags.
- `none`: item has none of these tags.

Either part may be empty. `any` with multiple tags is OR. Users who need AND can combine a tag filter with the existing text filter; a dedicated `all` mode is deferred until someone asks for it.

The special value `untagged` matches items with no tags.

### TUI

#### Showing tags

- Issue and habit rows show tags as compact chips after the title (`+urgent +quick`), truncated with `+2` when the row is narrow.
- Detail panes show the full tag list.
- Tags with a color render in that palette color; others use the muted style.

#### Editing tags

- In the existing issue and habit create and edit dialogs, a **Tags** field: comma- or space-separated input with autocomplete from the vocabulary. New names are created on save.
- On a selected issue or habit row, `+` opens a small tag picker: a checklist of existing tags with a text input at the top to filter or create. `space` toggles, `enter` saves.
- With multiple items selected (if multi-select exists on that list), `+` applies the picker to all of them; tags already on some items show as partially checked.

#### Filtering

- `/` keeps working as today. Text starting with `+` is treated as a tag: `/+urgent` filters to that tag, `/+urgent +quick` matches either, `/-waiting` excludes, and `/+urgent login` combines a tag with the existing text match.
- The active tag filter shows in the pane header (`+urgent −waiting`), and `esc` clears it as it clears the text filter today.
- The filter applies per view, like the existing UI-local filter. It does not change the checked-out context.
- The tag part of each view's filter persists across TUI restarts, stored with the other TUI settings. The free-text part does not. A saved filter referencing a deleted tag drops that tag silently on load.

This uses one existing key, `/`, plus `+` for editing. No other new keys.

#### Grouping

On Issues and Habits, the existing sort or group toggle (if present) gains a **by tag** option: one section per tag in name order, untagged last. An item with several tags appears under each of them, with a dim marker after its first appearance.

#### Tags view

Tab 2 is renamed **Work** and holds Issues, Habits and a **Tags** sub-tab. The Tags view lists the vocabulary, system tags first, with usage counts. `enter` renames (merging onto an existing name after a warning), `←/→` cycles the color, `D` deletes after a confirm, and `U` removes unused tags. System tags only change color.

### Daily plan and dashboard

- The Daily view supports the same `/+tag` filter.
- The daily summary and today summary include focused time per tag for the day.
- Dashboard distribution gains a **by tag** breakdown next to the existing ones. Time is attributed from sessions to their issue's tags. A session on an issue with several tags counts fully toward each tag, and the view says so, so totals across tags can exceed the day's total.

### Reports and exports

- Daily, weekly, repo, and stream reports include a per-tag time table when any tags were used in the period.
- CSV exports gain a `tags` column on issue and session rows (semicolon-separated names).
- Exports and reports accept the same tag filter (`--tag`, `--not-tag`) to limit output.
- Calendar export is unchanged.

### CLI

- `crona issue tag <id> +urgent -waiting` adds and removes in one call. `crona habit tag <id> ...` likewise.
- `crona issue ls --tag urgent --not-tag waiting`, `crona habit ls --tag health`.
- `crona tag ls` (with counts), `crona tag rename <old> <new>`, `crona tag delete <name>`.
- Create commands accept `--tag` (repeatable).

### Notes integration

Notes and tags share one vocabulary.

- A note's front-matter `tags` use the same names and normalization. The Notes view's tag filter uses the same `/+tag` syntax.
- Task lines gain a `+tag` token (repeatable): `- fix login redirect #crona/api +urgent ~45m`. Triage shows a **tags** column, editable with the same tag picker, and created issues get those tags.
- A note's own tags are not copied to the issues it creates; only `+tag` tokens on the line are. This keeps a note tagged `design` from tagging every task inside it.

The notes spec's inline token table should list `+tag` once this spec is accepted.

