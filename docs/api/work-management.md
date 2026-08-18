---
title: Work Management
description: Repositories, streams, issues, habits, and daily planning methods.
hosted: true
order: 7.3
---

# Work Management

These methods mutate source-of-truth work state. After a successful mutation, update the local view from its response and listen for the corresponding event.

## Repositories and streams

| Resource | Methods |
| --- | --- |
| Repositories | `repo.list`, `repo.create`, `repo.update`, `repo.delete` |
| Streams | `stream.list`, `stream.create`, `stream.update`, `stream.delete` |

Create and update requests use the corresponding request DTO; list methods return collections and delete methods return an acknowledgement or error.

## Issues and daily planning

| Capability | Methods |
| --- | --- |
| Read | `issue.list`, `issue.list_all`, `issue.daily_summary`, `issue.today_summary` |
| Lifecycle | `issue.create`, `issue.update`, `issue.delete`, `issue.change_status`, `issue.status_transitions` |
| Todo state | `issue.set_todo`, `issue.clear_todo` |
| Planning | `daily_plan.get` |

Status changes must use a transition supported by `issue.status_transitions`. Daily summaries are date-boundary aware and should be requested using the configured logical day, not an arbitrary calendar midnight.

## Habits

`habit.list`, `habit.list_all`, and `habit.list_due` read habit state. Use `habit.create`, `habit.update`, and `habit.delete` for configuration, and `habit.complete` / `habit.uncomplete` for a day’s completion state. `habit.history` returns historical completion data.

## Source links

Request shapes are authoritative in [`shared/dto/requests.go`](../../shared/dto/requests.go); method constants are in [`shared/protocol/methods.go`](../../shared/protocol/methods.go).
