---
title: Runtime and Operations
description: Daemon health, lifecycle, updates, alerts, and operation history methods.
hosted: true
order: 7.2
---

# Runtime and Operations

## Health and lifecycle

| Method | Request | Result / behavior |
| --- | --- | --- |
| `health.get` | `dto.Empty` | Readiness and health status. |
| `kernel.info.get` | `dto.Empty` | Runtime metadata and `protocolVersion`; use for startup compatibility. |
| `kernel.shutdown` | `dto.Empty` | Gracefully stops the daemon. |
| `kernel.restart` | `dto.Empty` | Restarts the daemon. |
| `kernel.dev.seed` | `dto.Empty` | Seeds development data; development use only. |
| `kernel.dev.clear` | `dto.Empty` | Clears development data; development use only. |
| `kernel.data.wipe` | `dto.ConfirmDangerousActionRequest` | Destructively removes local data after explicit confirmation. |

## Updates and alerts

| Method family | Methods | Notes |
| --- | --- | --- |
| Updates | `update.status.get`, `update.check` | Read or refresh update metadata. |
| Alert status | `alerts.status.get` | Read backend capability and delivery status. |
| Alert tests | `alerts.test_notification`, `alerts.test_sound`, `alerts.notify` | Test or request delivery through the configured backend. |
| Alert delivery | `alerts.delivery.subscribe`, `alerts.delivery.ack` | Subscribe to and acknowledge delivery records. |
| Reminders | `alerts.reminders.list`, `.create`, `.update`, `.delete`, `.toggle` | Manage configured reminders. |

## Operations history

| Method | Purpose |
| --- | --- |
| `ops.list` | List operation records. |
| `ops.latest` | Read the newest operation record. |
| `ops.since` | Read records after a cursor or timestamp. |

Lifecycle and alert changes can also arrive through the [event stream](./events.md).
