import { randomUUID } from "crypto";
import type { Stash } from "../domain/stash";
import type { ICommandContext } from "./context";
import { elapsedSeconds } from "../timer/time";
import { emitContextChanged } from "./active_context.commands";
import { TimerService } from "../timer";


/**
 * Stash current context (and session if running)
 */
export async function stashPush(
  ctx: ICommandContext,
  stashNote?: string
): Promise<Stash> {
  const now = ctx.now()

  const context = await ctx.activeContext.get(ctx.userId, ctx.deviceId)
  if (!context) throw new Error("No active context to stash")

  const activeSession = await ctx.sessions.getActiveSession(ctx.userId)
  let segmentSnapshot

  if (activeSession) {
    const activeSegment = await ctx.sessionSegments.getActive(
      ctx.userId,
      ctx.deviceId,
      activeSession.id
    )

    if (activeSegment) {
      const elapsed = elapsedSeconds(activeSegment.startTime.toISOString(), now)
      segmentSnapshot = {
        sessionId: activeSession.id,
        pausedSegmentType: activeSegment.segmentType,
        elapsedSeconds: elapsed,
      }
      // close segment
      await ctx.sessionSegments.endActiveSegment(
        ctx.userId,
        ctx.deviceId,
        activeSession.id
      );

    }
    await ctx.sessions.stop(
      activeSession.id,
      {
        endTime: now,
        durationSeconds: elapsedSeconds(activeSession.startTime, now),
      },
      {
        userId: ctx.userId,
        deviceId: ctx.deviceId,
        now,
      }
    );
  }

  const stash: Stash = {
    id: randomUUID(),
    userId: ctx.userId,
    deviceId: ctx.deviceId,

    repoId: context.repoId,
    streamId: context.streamId,
    issueId: context.issueId,

    note: stashNote,

    ...segmentSnapshot,

    createdAt: new Date(now),
    updatedAt: new Date(now),
  }

  await ctx.stash.save(stash)
  await ctx.activeContext.clear(ctx.userId, ctx.deviceId)

  ctx.events.emit({ type: "stash.created", payload: stash })
  await emitContextChanged(ctx)

  return stash
}

export async function stashPop(
  ctx: ICommandContext,
  stashId: string
): Promise<void> {
  const stash = await ctx.stash.get(stashId, ctx.userId)
  if (!stash) throw new Error("Stash not found")

  await ctx.activeContext.set(ctx.userId, ctx.deviceId, {
    repoId: stash.repoId,
    streamId: stash.streamId,
    issueId: stash.issueId,
  })

  if (stash.sessionId && stash.issueId && stash.pausedSegmentType) {
    const timer = TimerService.getInstance(ctx)

    await timer.restoreFromStash({
      issueId: stash.issueId,
      segmentType: stash.pausedSegmentType,
      elapsedSeconds: stash.elapsedSeconds ?? 0,
    })
  }

  await ctx.stash.delete(stashId, ctx.userId)

  ctx.events.emit({ type: "stash.applied", payload: stash })
  await emitContextChanged(ctx)
}


export async function stashDrop(
  ctx: ICommandContext,
  stashId: string
): Promise<void> {
  await ctx.stash.delete(stashId, ctx.userId);

  await ctx.ops.append({
    id: randomUUID(),
    entity: "stash",
    entityId: stashId,
    action: "delete",
    payload: {},
    timestamp: ctx.now(),
    userId: ctx.userId,
    deviceId: ctx.deviceId,
  });

  ctx.events.emit({
    type: "stash.dropped",
    payload: {
      id: stashId,
      deviceId: ctx.deviceId
    },
  });
}
