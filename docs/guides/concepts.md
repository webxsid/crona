---
hosted: true
title: Concepts
description: The simple work model behind Crona's planning, focus, and history.
order: 1.1
---

# How Crona thinks about work

Crona is built around a simple loop: decide what matters, work on one thing, and leave enough context to pick it up again.

## The work model

Imagine you are working on a product:

- A **repository** is the broad place where the work belongs: `work`, `personal`, or `research`.
- A **stream** is a long-running lane inside it: `crona-site`, `backend`, or `experiments`.
- An **issue** is one piece of work you can intend, schedule, and complete.
- A **session** is a focused interval spent on an issue.

This gives Crona more context than a flat task list without forcing you into a project-management ceremony.

## Active context

Your active context is the repository, stream, and issue you are currently working in. The TUI and CLI share it through the local daemon, so a command can usually operate on the thing you already selected.

Think of context as a working pointer, not a permanent assignment. Change it when the work changes.

## Sessions are records, not just timers

A timer answers “how long have I been doing this?” A Crona session also answers “what was I doing, and what did I leave behind?”

Sessions can include work and break segments, be stashed when you are interrupted, and end with a short summary. The result is a readable history connected to the issue that motivated it.

## The local pieces

Crona has three surfaces:

- `crona` is the launcher and scriptable CLI.
- `crona-tui` is the terminal interface for planning and focused work.
- `crona-daemon` is the local background service that owns the database, timers, reminders, and client communication.

The TUI, CLI, and macOS Companion are clients of the same local state. Switching surfaces does not require an account or a sync step.

## Planning, focus, and review

Most days can be understood as three moves:

1. **Plan**: choose issues, habits, and check-ins for the day.
2. **Focus**: start a session from the issue you are moving forward.
3. **Review**: read the day, session summaries, habits, and Momentum to see what happened.

Read the workflow guides when you are ready to go deeper:

- [Issues and Planning](../workflows/issues-and-planning.md)
- [Focus Sessions](../workflows/focus-sessions.md)
- [Habits](../workflows/habits.md)
- [Check-Ins and Wellbeing](../workflows/check-ins-and-wellbeing.md)
