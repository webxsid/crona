---
hosted: true
title: Focus Sessions
description: Timed work blocks, Pomodoro profiles, stash management, and session logs.
order: 5.2
---

Focus sessions are timed intervals tied directly to specific issues. The daemon manages active timers, ensuring that tracking continues even if clients disconnect.

## Starting a Session

Focus sessions can be initiated for issues in the **Planned**, **Ready**, or **In Progress** states:
- Starting a timer on a **Planned** or **Ready** issue automatically promotes its status to **In Progress**.
- The timer configuration (`Pomodoro` cadence or no-break `Timer` countdown mode) is loaded directly from the issue profile.
- Start dialogs show worked time, estimate context, total duration, and an `Ends At` preview before the session begins.
- Timer defaults use the remaining estimate instead of ignoring already logged work.
- The interactive client owns the terminal tab/window title, updating it with elapsed time and task context.

When a countdown expires, you can commit the session or add more time. Extending a countdown adds duration only; it does not add Pomodoro breaks or cycles.

## Interruptions & Stashes

If interrupted during a session, you can create a **Stash** to suspend the current work state:
- The stash preserves elapsed timer segments, cycle counts, and task context.
- Resuming a task with an existing stash prompts the user to either resume the stashed session or start a fresh timer while keeping the stash intact.
- Stashes prevent incomplete or fragmented logs when switching tasks mid-session.

## Completing and Amending Sessions

Stopping a session triggers the completion flow:
- **Session Summaries**: Users write a commit-style summary detailing what was accomplished.
- **Session History**: Completed sessions are written to the database. You can browse history or amend logs (such as editing descriptions or correcting durations) retroactively.
- **Manual Logging**: Work completed offline or away from the terminal can be logged manually.

## Focus Inactivity Alerts

The local daemon runs focus inactivity checks:
- The TUI client periodically reports keyboard activity to the daemon.
- If no client activity is reported for a configured duration, the daemon triggers an OS-level notification asking the user to review, pause, or end the active session.
