import { randomUUID } from "crypto";
import type { ICommandContext } from "./context";
import type { ActiveContext } from "../domain";

/**
 * Get current active context (read-only)
 */
export async function getActiveContext(
  ctx: ICommandContext
): Promise<ActiveContext | null> {
  return ctx.activeContext.get(ctx.userId, ctx.deviceId);
}

/**
 * Switch active repo
 * Clears downstream context (stream, issue)
 */
export async function switchRepo(
  ctx: ICommandContext,
  repoId: string
): Promise<ActiveContext> {
  const now = ctx.now();

  const context = await ctx.activeContext.set(ctx.userId, ctx.deviceId, {
    repoId,
    streamId: undefined,
    issueId: undefined,
  });

  await ctx.ops.append({
    id: randomUUID(),
    entity: "active_context",
    entityId: ctx.userId,
    action: "update",
    payload: { repoId },
    timestamp: now,
    userId: ctx.userId,
    deviceId: ctx.deviceId,
  });

  await emitRepoChanged(ctx);

  return context;
}

/**
 * Switch active stream
 * Requires repo to be set
 * Clears issue
 */
export async function switchStream(
  ctx: ICommandContext,
  streamId: string
): Promise<ActiveContext> {
  const existing = await ctx.activeContext.get(ctx.userId, ctx.deviceId);
  if (!existing?.repoId) {
    throw new Error("No active repo. Switch repo first.");
  }

  const now = ctx.now();

  const context = await ctx.activeContext.set(ctx.userId, ctx.deviceId, {
    repoId: existing.repoId,
    streamId,
    issueId: undefined,
  });

  await ctx.ops.append({
    id: randomUUID(),
    entity: "active_context",
    entityId: ctx.userId,
    action: "update",
    payload: { streamId },
    timestamp: now,
    userId: ctx.userId,
    deviceId: ctx.deviceId,
  });

  await emitStreamChanged(ctx);

  return context;
}

/**
 * Switch active issue
 * Requires stream to be set
 */
export async function switchIssue(
  ctx: ICommandContext,
  issueId: string
): Promise<ActiveContext> {
  const existing = await ctx.activeContext.get(ctx.userId, ctx.deviceId);
  if (!existing?.streamId) {
    throw new Error("No active stream. Switch stream first.");
  }

  const now = ctx.now();

  const context = await ctx.activeContext.set(ctx.userId, ctx.deviceId, {
    repoId: existing.repoId,
    streamId: existing.streamId,
    issueId,
  });

  await ctx.ops.append({
    id: randomUUID(),
    entity: "active_context",
    entityId: ctx.userId,
    action: "update",
    payload: { issueId },
    timestamp: now,
    userId: ctx.userId,
    deviceId: ctx.deviceId,
  });

  await emitIssueChanged(ctx);

  return context;
}

/**
 * Clear active issue only
 * (equivalent to git checkout --)
 */
export async function clearIssue(
  ctx: ICommandContext
): Promise<ActiveContext> {
  const existing = await ctx.activeContext.get(ctx.userId, ctx.deviceId);
  if (!existing) {
    throw new Error("No active context");
  }

  const now = ctx.now();

  const context = await ctx.activeContext.set(ctx.userId, ctx.deviceId, {
    repoId: existing.repoId,
    streamId: existing.streamId,
    issueId: undefined,
  });

  await ctx.ops.append({
    id: randomUUID(),
    entity: "active_context",
    entityId: ctx.userId,
    action: "update",
    payload: { issueId: null },
    timestamp: now,
    userId: ctx.userId,
    deviceId: ctx.deviceId,
  });

  await emitIssueChanged(ctx);

  return context;
}

/**
 * Clear entire context
 * (equivalent to git checkout --detach)
 */
export async function clearContext(
  ctx: ICommandContext
): Promise<void> {
  const now = ctx.now();

  await ctx.activeContext.clear(ctx.userId, ctx.deviceId);

  await ctx.ops.append({
    id: randomUUID(),
    entity: "active_context",
    entityId: ctx.userId,
    action: "delete",
    payload: {},
    timestamp: now,
    userId: ctx.userId,
    deviceId: ctx.deviceId,
  });

  await emitContextCleared(ctx);
}

type ContextSnapshot = {
  deviceId: string;
  repoId?: string | undefined;
  streamId?: string | undefined;
  issueId?: string | undefined;
};

async function snapshot(ctx: ICommandContext): Promise<ContextSnapshot> {
  const context = await ctx.activeContext.get(ctx.userId, ctx.deviceId);
  return {
    deviceId: ctx.deviceId,
    repoId: context?.repoId,
    streamId: context?.streamId,
    issueId: context?.issueId,
  };
}

export async function emitRepoChanged(ctx: ICommandContext) {
  ctx.events.emit({ type: "context.repo.changed", payload: await snapshot(ctx) });
}

export async function emitStreamChanged(ctx: ICommandContext) {
  ctx.events.emit({ type: "context.stream.changed", payload: await snapshot(ctx) });
}

export async function emitIssueChanged(ctx: ICommandContext) {
  ctx.events.emit({ type: "context.issue.changed", payload: await snapshot(ctx) });
}

export async function emitContextCleared(ctx: ICommandContext) {
  ctx.events.emit({ type: "context.cleared", payload: { deviceId: ctx.deviceId } });
}
