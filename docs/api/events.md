---
title: Events
description: Subscribe to Crona daemon events and keep client views synchronized.
hosted: true
badge: Beta
order: 7.6
---

# Events

:::note[Protocol compatibility]
The event stream is shared by Crona clients. Logical-date and day-boundary events require protocol `1.5`; clients that connect to an older daemon should tolerate missing events and reload state after reconnecting.
:::

Call `events.subscribe` once a client has connected. The daemon then sends event envelopes with a `type` and `payload`. Events are hints that state changed: clients should update the affected view from the payload when possible and refetch when the event is unknown or incomplete.

## Event groups

| Group | Examples |
| --- | --- |
| Entity lifecycle | Issue, habit, repository, and stream created, updated, deleted, or status-changed events. |
| Sessions and timers | Session start/pause/resume/end and timer state changes. |
| Settings | Settings changes, including away-mode and day-boundary configuration. |
| Context | Repository, stream, and issue context changes. |
| Updates | Update availability and update status changes. |

The exact event constants and payload types are defined in [`shared/types/events.go`](../../shared/types/events.go). Clients must ignore unknown event types for forward compatibility.

## Synchronization rules

- Apply events only after the connection is established and the initial snapshot has loaded.
- Keep one logical-date value shared by summary, issues, habits, calendar, and metrics views.
- Recompute that logical date at the configured start-of-day boundary, not at calendar midnight.
- On reconnect, discard assumptions about missed events and reload the affected snapshot.

For the wire shape and compatibility handshake, see [Transport and Envelopes](./transport.md).
