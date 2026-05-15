# Socket API

This document describes Crona's local engine IPC surface.

The protocol and method names still use `kernel` for the internal engine process. Product docs usually call the same process the local engine or background engine.

It is an open-source reference for contributors and advanced users, not a promise of long-term network API stability before `1.0.0`.

## Transport

Crona uses a local request/response transport between clients and the local engine.

- Unix-domain sockets on Unix-like platforms
- named pipes on Windows

The transport is local-only. The canonical wire envelopes live in [`shared/protocol/ipc.go`](../../shared/protocol/ipc.go).

## Wire Envelopes

### Request

```json
{
  "id": "string",
  "method": "string",
  "params": {}
}
```

- `id`: client-generated correlation ID
- `method`: RPC method name
- `params`: optional JSON payload

### Response

```json
{
  "id": "string",
  "result": {},
  "error": {
    "code": "string",
    "message": "string",
    "data": {}
  }
}
```

- `result` is omitted on error
- `error` is omitted on success
- `error.data` is optional structured metadata for client-recoverable errors

### Event

```json
{
  "type": "string",
  "payload": {}
}
```

Event envelopes are pushed after `events.subscribe`.

## Source Of Truth

Use these files as the canonical contract:

- [`shared/protocol/ipc.go`](../../shared/protocol/ipc.go)
- [`shared/protocol/methods.go`](../../shared/protocol/methods.go)
- [`shared/dto/requests.go`](../../shared/dto/requests.go)
- [`shared/types/domain.go`](../../shared/types/domain.go)
- [`shared/types/events.go`](../../shared/types/events.go)

## Compatibility Note

- The local engine IPC surface is shared across Crona clients.
- It is intentionally documented because the project is open source.
- Before `1.0.0`, consumers should expect the shared Go types and method constants to be the source of truth over any prose doc.
- GUI compatibility should be checked against `kernel.info.get -> protocolVersion`.
- `protocolVersion` is independent from the Crona release version and only changes when the local IPC contract or its client-visible semantics change.

`kernel.info.get` is the expected GUI handshake:

- `protocolVersion` for compatibility checks
- runtime transport and endpoint details
- running release channel metadata

## RPC Methods

Request DTO names below refer to types in [`shared/dto/requests.go`](../../shared/dto/requests.go). Result payloads are returned as JSON objects or arrays matching the shared domain/DTO types used by the local engine handlers.

### Event Subscription

| Method | Request | Result | Notes |
| --- | --- | --- | --- |
| `events.subscribe` | `dto.Empty` | stream subscription ack | Starts the event stream. |

### Health And Kernel

| Method | Request | Result | Notes |
| --- | --- | --- | --- |
| `health.get` | `dto.Empty` | health status object | Kernel health and readiness. |
| `kernel.info.get` | `dto.Empty` | kernel info object | Runtime metadata, transport, endpoint, install metadata, and `protocolVersion`. |
| `kernel.shutdown` | `dto.Empty` | `dto.OKResponse` | Graceful local shutdown. |
| `kernel.restart` | `dto.Empty` | `dto.OKResponse` | Restarts the local engine. |
| `kernel.dev.seed` | `dto.Empty` | `dto.OKResponse` | Dev-only sample data seed. |
| `kernel.dev.clear` | `dto.Empty` | `dto.OKResponse` | Dev-only local data clear. |
| `kernel.dev.prepare_local_update` | `dto.Empty` | `dto.LocalUpdatePreparedResponse` | Dev-only local updater simulation setup. |
| `kernel.data.wipe` | `dto.ConfirmDangerousActionRequest` | `dto.OKResponse` | Wipes runtime data after explicit confirmation. |

### Updates

| Method | Request | Result | Notes |
| --- | --- | --- | --- |
| `update.status.get` | `dto.Empty` | update status object | Current version, channel, availability, notes metadata. |
| `update.check` | `dto.Empty` | update status object | Performs a refresh against the configured source. |
| `update.dismiss` | `dto.Empty` | update status object | Dismisses the current update prompt. |

### Alerts

| Method | Request | Result | Notes |
| --- | --- | --- | --- |
| `alerts.status.get` | `dto.Empty` | alert status object | Active backend, capability flags, and runtime support for alerts/sound. |
| `alerts.test_notification` | `dto.Empty` | `dto.OKResponse` | Sends a sample alert through the current backend. |
| `alerts.test_sound` | `dto.Empty` | `dto.OKResponse` | Plays the selected bundled alert preset when supported. |
| `alerts.notify` | `types.AlertRequest` | `dto.OKResponse` | Delivers one structured alert request through the local engine alerts layer. |
| `alerts.reminders.list` | `dto.Empty` | reminder list | Lists scheduled local alert reminders. |
| `alerts.reminders.create` | `dto.AlertReminderCreateRequest` | reminder object | Creates a scheduled reminder rule. |
| `alerts.reminders.update` | `dto.AlertReminderUpdateRequest` | reminder object | Updates one scheduled reminder rule. |
| `alerts.reminders.delete` | `dto.AlertReminderIDRequest` | `dto.OKResponse` | Deletes one scheduled reminder rule. |
| `alerts.reminders.toggle` | `dto.AlertReminderToggleRequest` | reminder object | Enables or disables one scheduled reminder rule. |

Alert behavior notes:

- the local engine, not the TUI, decides when alerts fire
- scheduled reminders are local-only and only fire while the local engine is running
- focus inactivity alerts are local-engine-owned; TUI clients may call `timer.activity.touch` to report recent user input while a focus session is active
- `AlertStatus` reflects the current OS helper/backend that the local engine detected at runtime

### Repositories

| Method | Request | Result | Notes |
| --- | --- | --- | --- |
| `repo.list` | `dto.Empty` | `[]types.Repo` | Lists repos. |
| `repo.create` | `dto.CreateRepoRequest` | repo object | Creates a repo. |
| `repo.update` | `dto.UpdateRepoRequest` | repo object | Updates a repo. |
| `repo.delete` | `dto.NumericIDRequest` | `dto.OKResponse` | Deletes a repo. |

### Streams

| Method | Request | Result | Notes |
| --- | --- | --- | --- |
| `stream.list` | `dto.ListStreamsQuery` | `[]types.Stream` | Lists streams for a repo. |
| `stream.create` | `dto.CreateStreamRequest` | stream object | Creates a stream. |
| `stream.update` | `dto.UpdateStreamRequest` | stream object | Updates a stream. |
| `stream.delete` | `dto.NumericIDRequest` | `dto.OKResponse` | Deletes a stream. |

### Issues And Daily Planning

| Method | Request | Result | Notes |
| --- | --- | --- | --- |
| `issue.list` | `dto.ListIssuesQuery` | issue list | Lists issues for a stream. |
| `issue.list_all` | `dto.Empty` | issue-with-meta list | Lists issues across the workspace. |
| `issue.create` | `dto.CreateIssueRequest` | issue object | Creates an issue. |
| `issue.update` | `dto.UpdateIssueRequest` | issue object | Updates an issue. |
| `issue.delete` | `dto.NumericIDRequest` | `dto.OKResponse` | Deletes an issue. |
| `issue.change_status` | `dto.ChangeIssueStatusRequest` | issue object | Applies a lifecycle transition. |
| `issue.set_todo` | `dto.SetIssueTodoRequest` | issue object | Sets a todo date. |
| `issue.clear_todo` | `dto.SetIssueTodoRequest` | issue object | Clears a todo date. |
| `issue.daily_summary` | `dto.DailyIssueSummaryQuery` | daily summary object | Summary for an arbitrary date. |
| `issue.today_summary` | `dto.Empty` | daily summary object | Today's summary shortcut. |
| `daily_plan.get` | `dto.DailyPlanQuery` | daily plan object | Returns planned issues and supporting daily data. |

### Habits

| Method | Request | Result | Notes |
| --- | --- | --- | --- |
| `habit.list` | `dto.ListHabitsQuery` | habit list | Lists habits for a stream. |
| `habit.list_due` | `dto.ListHabitsDueQuery` | due-habit list | Lists habits due for a date. |
| `habit.create` | `dto.CreateHabitRequest` | habit object | Creates a habit. |
| `habit.update` | `dto.UpdateHabitRequest` | habit object | Updates a habit. |
| `habit.delete` | `dto.NumericIDRequest` | `dto.OKResponse` | Deletes a habit. |
| `habit.complete` | `dto.HabitCompletionUpsertRequest` | completion object | Marks or logs habit completion. |
| `habit.uncomplete` | `dto.HabitCompletionUpsertRequest` | completion object | Removes completion status. |
| `habit.history` | `dto.HabitHistoryQuery` | history list | Habit completion history. |

### Check-Ins And Metrics

| Method | Request | Result | Notes |
| --- | --- | --- | --- |
| `checkin.get` | `dto.DailyCheckInQuery` | check-in object | Gets one day's check-in. |
| `checkin.upsert` | `dto.DailyCheckInUpsertRequest` | check-in object | Creates or updates a check-in. |
| `checkin.delete` | `dto.DeleteByDateRequest` | `dto.OKResponse` | Deletes a daily check-in. |
| `checkin.range` | `dto.DateRangeQuery` | check-in list | Lists check-ins in a range. |
| `metrics.range` | `dto.DateRangeQuery` | metrics range object | Per-day metrics for a date window. |
| `metrics.rollup` | `dto.DateRangeQuery` | metrics rollup object | Aggregate rollups for a date window. |
| `metrics.streaks` | `dto.DateRangeQuery` | streak summary object | Streak calculations over a date window. |
| `metrics.streaks_lifetime` | `dto.DailyCheckInQuery` | streak summary object | Streak calculations across stored history through one date. |

Streak behavior notes:

- `metrics.streaks` keeps range-based semantics for callers that need streaks constrained to a specific date window.
- `metrics.streaks_lifetime` computes the same streak summary shape across stored local history through the requested date, excluding future records.
- The lifetime start date is derived from the earliest stored ended focus session, check-in, or habit completion at or before the requested date. If no history exists, the requested date is used as the start date.
- The TUI Wellbeing dashboard uses `metrics.streaks_lifetime` for Momentum while retaining 7-day `metrics.range`, `metrics.rollup`, burnout, and dashboard summary calls.
- For weekly and monthly custom habit streaks, an incomplete current week/month does not break the current streak while that bucket is still open.

### Dashboards

| Method | Request | Result | Notes |
| --- | --- | --- | --- |
| `dashboard.window` | `dto.DashboardWindowQuery` | dashboard window object | Shared dashboard data for a range and optional scope. |
| `dashboard.focus_score` | `dto.DashboardSummaryQuery` | focus score summary | Focus scoring summary. |
| `dashboard.distribution` | `dto.DashboardSummaryQuery` | distribution summary | Time distribution summary. |
| `dashboard.goal_progress` | `dto.DashboardSummaryQuery` | goal progress summary | Estimate and execution progress. |

### Exports

| Method | Request | Result | Notes |
| --- | --- | --- | --- |
| `export.daily` | `dto.DailyReportRequest` | export result | Daily report generation. |
| `export.weekly` | `dto.ExportReportRequest` | export result | Weekly report generation. |
| `export.repo` | `dto.ExportReportRequest` | export result | Repo report generation. |
| `export.stream` | `dto.ExportReportRequest` | export result | Stream report generation. |
| `export.issue_rollup` | `dto.ExportReportRequest` | export result | Issue rollup generation. |
| `export.csv` | `dto.ExportReportRequest` | export result | CSV export generation. |
| `export.calendar` | `dto.ExportCalendarRequest` | calendar export result | Writes deterministic `.ics` artifacts. |
| `export.assets.get` | `dto.Empty` | export assets metadata | Export templates, docs, preset metadata, directories. |
| `export.reports_dir.set` | `dto.ExportReportsDirUpdateRequest` | `dto.OKResponse` | Sets the reports directory. |
| `export.ics_dir.set` | `dto.ExportICSDirUpdateRequest` | `dto.OKResponse` | Sets the ICS directory. |
| `export.reports.list` | `dto.Empty` | generated report list | Lists generated report artifacts. |
| `export.reports.delete` | `dto.ExportReportDeleteRequest` | `dto.OKResponse` | Deletes a generated report artifact. |
| `export.template.reset` | `dto.ExportTemplateResetRequest` | asset reset result | Resets a template/spec to bundled defaults. |
| `export.template.apply` | `dto.ExportTemplatePresetApplyRequest` | asset preset result | Applies a built-in preset. |

Export behavior notes:

- markdown export does not require extra renderer tooling
- daily and weekly PDF export require `weasyprint`
- repo, stream, and issue-rollup PDF export require `pandoc` plus a supported engine
- `export.assets.get` is the runtime capability/status surface for renderer availability, active template paths, and reports directories

### Sessions And Timer

| Method | Request | Result | Notes |
| --- | --- | --- | --- |
| `session.list_by_issue` | `dto.ListSessionsQuery` | session list | Lists sessions by issue. |
| `session.get` | `dto.SessionIDRequest` | session object | Gets one session. |
| `session.detail` | `dto.SessionIDRequest` | session detail object | Rich session detail for overlays and history. |
| `session.get_active` | `dto.Empty` | active session or `null` | Current active session. |
| `session.start` | `dto.StartSessionRequest` | active session object | Starts a session for an issue. |
| `session.pause` | `dto.Empty` | timer/session state | Pauses the active session. |
| `session.resume` | `dto.Empty` | timer/session state | Resumes the active session. |
| `session.end` | `dto.EndSessionRequest` | ended session object | Ends the active session. |
| `session.log_manual` | `dto.ManualSessionLogRequest` | session object | Logs a manual session entry. |
| `session.amend_note` | `dto.AmendSessionNoteRequest` | session object | Rewrites the stored session note. |
| `session.history` | `dto.SessionHistoryQuery` | session history result | History queries with scope and paging controls. |
| `timer.get_state` | `dto.Empty` | timer state object | Current timer state. |
| `timer.start` | `dto.TimerStartRequest` | timer/session state | Starts a timer, optionally from context or an explicit repo/stream/issue path. |
| `timer.activity.touch` | `dto.Empty` | `dto.OKResponse` | Records recent client activity for active-session inactivity alert suppression. |
| `timer.pause` | `dto.Empty` | timer/session state | Pauses the timer. |
| `timer.resume` | `dto.Empty` | timer/session state | Resumes the timer. |
| `timer.end` | `dto.EndSessionRequest` | ended session object | Ends the active timer/session. |

Timer start behavior notes:

- `TimerStartRequest` can carry `repoId`, `streamId`, and `issueId` so clients can start focus from a selected issue without first mutating the shared active context.
- If `issueId` is omitted, the local engine resolves the current active context issue.
- If the target issue already has saved stashes, `timer.start` fails with `error.code = "stash_conflict"` unless `ignoreExistingStashes` is true.
- `stash_conflict` responses include `error.data` shaped as `types.StashConflict`, with the target issue ID and matching stash list.
- Clients should offer an explicit resume-vs-continue choice. Resuming should call `stash.apply`; continuing fresh should retry `timer.start` with the same repo/stream/issue path and `ignoreExistingStashes = true`.
- Inactivity alerts use core settings for enablement, first-alert threshold, and repeat interval. The default is enabled, 60 minutes to first alert, and 60 minutes between repeats.

### Context

| Method | Request | Result | Notes |
| --- | --- | --- | --- |
| `context.get` | `dto.Empty` | active context object | Gets the current shared context. |
| `context.set` | `dto.UpdateContextRequest` | active context object | Sets repo, stream, and issue together. |
| `context.switch_repo` | `dto.SwitchRepoRequest` | active context object | Switches repo only. |
| `context.switch_stream` | `dto.SwitchStreamRequest` | active context object | Switches stream only. |
| `context.switch_issue` | `dto.SwitchIssueRequest` | active context object | Switches issue only. |
| `context.clear_issue` | `dto.Empty` | active context object | Clears the current issue selection. |
| `context.clear` | `dto.Empty` | active context object | Clears the entire active context. |

### Settings

| Method | Request | Result | Notes |
| --- | --- | --- | --- |
| `settings.get_all` | `dto.Empty` | settings object | Full core-settings payload. |
| `settings.get` | `dto.GetCoreSettingRequest` | single setting result | Gets one core setting. |
| `settings.patch` | `dto.PatchCoreSettingRequest` | settings object | Patches one setting. |
| `settings.put` | `dto.PutCoreSettingsRequest` | settings object | Replaces multiple settings at once. |

### Stash

| Method | Request | Result | Notes |
| --- | --- | --- | --- |
| `stash.list` | `dto.Empty` | stash list | Lists saved stashes. |
| `stash.get` | `dto.StashIDRequest` | stash object | Gets one stash. |
| `stash.push` | `dto.CreateStashRequest` | stash object | Creates a stash from current context. |
| `stash.apply` | `dto.StashIDRequest` | active context object | Applies a stash. |
| `stash.drop` | `dto.StashIDRequest` | `dto.OKResponse` | Deletes a stash. |

### Scratchpads

| Method | Request | Result | Notes |
| --- | --- | --- | --- |
| `scratchpad.list` | `dto.ListScratchpadsQuery` | scratchpad metadata list | Lists scratchpads. |
| `scratchpad.register` | `dto.RegisterScratchpadRequest` | scratchpad metadata | Registers or updates a scratchpad path. |
| `scratchpad.get_meta` | `dto.ScratchpadIDRequest` | scratchpad metadata | Gets one scratchpad's metadata. |
| `scratchpad.read` | `dto.ScratchpadIDRequest` | scratchpad content result | Reads scratchpad contents. |
| `scratchpad.pin` | `dto.PinScratchpadRequest` | scratchpad metadata | Pins or unpins a scratchpad. |
| `scratchpad.delete` | `dto.ScratchpadIDRequest` | `dto.OKResponse` | Deletes a scratchpad registration and backing file where supported. |

### Operations Log

| Method | Request | Result | Notes |
| --- | --- | --- | --- |
| `ops.list` | `dto.ListOpsQuery` | ops list | Lists ops with optional filters. |
| `ops.latest` | `dto.ListLatestOpsQuery` | ops list | Latest ops shortcut. |
| `ops.since` | `dto.ListOpsSinceQuery` | ops list | Lists ops since a timestamp. |

## Events

Event types live in [`shared/types/events.go`](../../shared/types/events.go).

### Entity Lifecycle Events

- `repo.created`
- `repo.updated`
- `repo.deleted`
- `stream.created`
- `stream.updated`
- `stream.deleted`
- `issue.created`
- `issue.updated`
- `issue.deleted`
- `habit.created`
- `habit.updated`
- `habit.deleted`
- `habit.completed`
- `habit.uncompleted`
- `checkin.updated`
- `checkin.deleted`
- `scratchpad.created`
- `scratchpad.updated`
- `scratchpad.deleted`

Typical payload:
- `types.IDEventPayload`

### Session And Timer Events

- `session.started`
- `session.stopped`
- `timer.state`
- `timer.boundary`
- `timer.tick`

Payload notes:
- `timer.boundary` uses `types.TimerBoundaryPayload`
- `timer.tick` uses `types.TimerTickPayload`
- `timer.state` carries the current timer/session state snapshot

### Context And Stash Events

- `context.repo.changed`
- `context.stream.changed`
- `context.issue.changed`
- `context.cleared`
- `stash.created`
- `stash.applied`
- `stash.dropped`

Payload notes:
- context change events use `types.ContextChangedPayload`
- `context.cleared` uses `types.ContextClearedPayload`
- stash events use `types.StashEventPayload`

### Update Events

- `update.status`

Payload notes:
- carries the current shared update status snapshot used by the TUI and CLI
