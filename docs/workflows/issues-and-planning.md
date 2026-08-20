---
hosted: true
title: Issues and Planning
description: Turn a backlog into a day you can see and act on.
order: 5.1
---

# Plan the workday

Crona uses issues as the bridge between intention and focused work. Capture the work first; decide when to do it without losing the original context.

## Capture an issue

Give an issue a concrete title, then add an estimate, notes, repository, and stream when they are useful. An issue should describe a move you can finish, not an entire project.

Good: `replace the docs install instructions`.

Less useful: `improve documentation`.

## Put work on the day

Assign a planning date when you want an issue to appear in Daily. Crona promotes it from **Backlog** to **Planned** and keeps it with that date until you move or finish it.

The Daily view brings together:

- planned issues
- pinned work
- overdue issues
- habits due today
- wellbeing and planning prompts

## Use the issue lifecycle

Statuses tell you where an issue is in its journey:

`Backlog → Planned → Ready → In Progress → In Review → Done`

Use **Blocked** when something outside the issue is holding it up. Use **Abandoned** when you have deliberately decided not to continue.

Starting a focus session promotes a Planned or Ready issue to **In Progress**. The daemon validates transitions so different Crona clients do not silently disagree.

## Set the active context

Check out a repository, stream, or issue when you want commands and reports to default to it. The active context is shared by the TUI and CLI through the local daemon.

It is a pointer for the work you are doing now. It is not a replacement for planning dates or issue history.

## Next step

Select an issue and [start a focus session](focus-sessions.md). When the work is done, Crona will ask you to leave a short summary.
