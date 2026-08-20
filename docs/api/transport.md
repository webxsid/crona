---
title: Transport and Envelopes
description: Connect to the local Crona daemon and exchange IPC messages safely.
hosted: true
badge: Beta
order: 7.1
---

# Transport and Envelopes

:::note[Protocol compatibility]
Protocol `1.5` is used by the current beta line. Clients should read `protocolVersion` from `kernel.info.get` instead of assuming that every daemon supports the newest envelopes or events.
:::

Crona uses Unix-domain sockets on Unix-like systems and named pipes on Windows. The transport is local-only; clients should obtain the active endpoint from the daemon runtime metadata rather than hardcoding a path.

## Request

```json
{"id":"client-123","method":"health.get","params":{}}
```

`id` is a client-generated correlation value. `method` is one of the canonical names in `shared/protocol/methods.go`. `params` is an object matching the method request DTO and may be empty.

## Response

```json
{"id":"client-123","result":{}}
```

Errors use the same correlation ID:

```json
{"id":"client-123","error":{"code":"invalid_request","message":"...","data":{}}}
```

Success responses contain `result`; error responses contain `error`. Treat `error.data` as optional structured metadata.

## Events

After `events.subscribe`, the daemon may push:

```json
{"type":"issue.updated","payload":{}}
```

Events are notifications, not request responses, and do not use the request `id` field. Clients should reconcile event payloads with a fresh domain read when they receive an event they do not understand.

## Compatibility handshake

Call `kernel.info.get` during startup. Its result includes `protocolVersion`, runtime transport details, endpoint information, and release metadata. Refuse or degrade gracefully when the protocol is outside the client’s supported range.

The current protocol is **1.5**. The canonical source files are [`shared/protocol/ipc.go`](../../shared/protocol/ipc.go), [`shared/protocol/methods.go`](../../shared/protocol/methods.go), and [`shared/protocol/version.go`](../../shared/protocol/version.go).
