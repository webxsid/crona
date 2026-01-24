"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.stashPush = stashPush;
exports.stashPop = stashPop;
exports.stashDrop = stashDrop;
const crypto_1 = require("crypto");
const time_1 = require("../timer/time");
const active_context_commands_1 = require("./active_context.commands");
const timer_1 = require("../timer");
/**
 * Stash current context (and session if running)
 */
async function stashPush(ctx, stashNote) {
    const now = ctx.now();
    const context = await ctx.activeContext.get(ctx.userId, ctx.deviceId);
    if (!context)
        throw new Error("No active context to stash");
    const activeSession = await ctx.sessions.getActiveSession(ctx.userId);
    let segmentSnapshot;
    if (activeSession) {
        const activeSegment = await ctx.sessionSegments.getActive(ctx.userId, ctx.deviceId, activeSession.id);
        if (activeSegment) {
            const elapsed = (0, time_1.elapsedSeconds)(activeSegment.startTime.toISOString(), now);
            segmentSnapshot = {
                sessionId: activeSession.id,
                pausedSegmentType: activeSegment.segmentType,
                elapsedSeconds: elapsed,
            };
            // close segment
            await ctx.sessionSegments.endActiveSegment(ctx.userId, ctx.deviceId, activeSession.id);
        }
        await ctx.sessions.stop(activeSession.id, {
            endTime: now,
            durationSeconds: (0, time_1.elapsedSeconds)(activeSession.startTime, now),
        }, {
            userId: ctx.userId,
            deviceId: ctx.deviceId,
            now,
        });
    }
    const stash = {
        id: (0, crypto_1.randomUUID)(),
        userId: ctx.userId,
        deviceId: ctx.deviceId,
        repoId: context.repoId,
        streamId: context.streamId,
        issueId: context.issueId,
        note: stashNote,
        ...segmentSnapshot,
        createdAt: new Date(now),
        updatedAt: new Date(now),
    };
    await ctx.stash.save(stash);
    await ctx.activeContext.clear(ctx.userId, ctx.deviceId);
    ctx.events.emit({ type: "stash.created", payload: stash });
    await (0, active_context_commands_1.emitContextChanged)(ctx);
    return stash;
}
async function stashPop(ctx, stashId) {
    const stash = await ctx.stash.get(stashId, ctx.userId);
    if (!stash)
        throw new Error("Stash not found");
    await ctx.activeContext.set(ctx.userId, ctx.deviceId, {
        repoId: stash.repoId,
        streamId: stash.streamId,
        issueId: stash.issueId,
    });
    if (stash.sessionId && stash.issueId && stash.pausedSegmentType) {
        const timer = timer_1.TimerService.getInstance(ctx);
        await timer.restoreFromStash({
            issueId: stash.issueId,
            segmentType: stash.pausedSegmentType,
            elapsedSeconds: stash.elapsedSeconds ?? 0,
        });
    }
    await ctx.stash.delete(stashId, ctx.userId);
    ctx.events.emit({ type: "stash.applied", payload: stash });
    await (0, active_context_commands_1.emitContextChanged)(ctx);
}
async function stashDrop(ctx, stashId) {
    await ctx.stash.delete(stashId, ctx.userId);
    await ctx.ops.append({
        id: (0, crypto_1.randomUUID)(),
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
