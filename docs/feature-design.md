# Feature Design

## Multi-Device Sync

This document is about future sync architecture only.

Current local architecture:
- the local engine is a Go background process
- local clients talk to it over a Unix socket
- this document does not change that local IPC model

Future sync is a separate concern layered on top of the local engine.

### Problem

Crona is local-first with a SQLite database per device. Multiple devices create divergent state. Direct DB sync isn't viable because SQLite files are not merge-friendly, and there is no natural conflict resolution at the row level.

### Why Op Logs Are The Right Foundation

Every mutation in Crona already produces an immutable `Op` record:

```text
{ id, user_id, device_id, entity, entity_id, action, payload, timestamp }
```

This is an append-only log of all state changes, ordered by timestamp, tagged by device. Syncing Crona across devices reduces to: share the op log, replay it on each device, derive consistent state.

This is structurally similar to how distributed databases, CRDTs, and event-sourced systems handle sync. No central authority is required.

### Proposed Solution: Layered Sync

Three sync modes, each building on the previous:

#### Layer 1: File-Based Sync

Export the op log as an append-only NDJSON file to a shared folder watched by any cloud sync provider:

```text
~/<cloud-provider>/crona-sync/
  ops-<device-id>.ndjson
  ops-<device-id-2>.ndjson
```

On startup and periodically, each device:
1. reads all `ops-*.ndjson` files from other devices
2. filters ops it has not seen yet
3. replays them against the local DB in timestamp order
4. appends its own new ops to `ops-<its-device-id>.ndjson`

Works with iCloud Drive, Dropbox, Google Drive, Syncthing, or any folder-sync tool.

#### Layer 2: Self-Hosted Sync Relay

A lightweight relay server that acts as a dumb op store:

```text
POST /ops
GET  /ops?since=
```

Devices push their ops on mutation and pull on startup or reconnect.

#### Layer 3: Local Network P2P

When devices are on the same network, sync directly without a relay:

- kernel advertises itself via mDNS
- devices discover peers
- peers exchange ops directly

### Conflict Resolution

Since ops are immutable and timestamped, conflicts are resolved deterministically:

| Entity | Strategy | Rationale |
| --- | --- | --- |
| Repo / Stream / Issue fields | Last-write-wins by `timestamp` | Simple fields, low conflict risk |
| Issue status transitions | Ordered by timestamp | State-machine style updates |
| Active context | Per device, never merged | Context is device-local |
| Sessions | Device-scoped, no merge needed | Sessions are owned by one device |
| Stash | Namespaced by device | Stashes are device-local |
| Deletes | Soft-delete wins | `deleted_at` is terminal |

### What Does Not Sync

- scratchpad file contents
- active context
- timer state
- auth tokens

### Implementation Notes

- Op replay must be idempotent.
- Minor clock skew is acceptable for last-write-wins.
- The local DB should remain derivable from the op log.
- Every mutation must produce an op.
