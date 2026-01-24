"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.getActiveContext = getActiveContext;
exports.switchRepo = switchRepo;
exports.switchStream = switchStream;
exports.switchIssue = switchIssue;
exports.clearIssue = clearIssue;
exports.clearContext = clearContext;
exports.emitContextChanged = emitContextChanged;
const crypto_1 = require("crypto");
/**
 * Get current active context (read-only)
 */
async function getActiveContext(ctx) {
    return ctx.activeContext.get(ctx.userId, ctx.deviceId);
}
/**
 * Switch active repo
 * Clears downstream context (stream, issue)
 */
async function switchRepo(ctx, repoId) {
    const now = ctx.now();
    const context = await ctx.activeContext.set(ctx.userId, ctx.deviceId, {
        repoId,
        streamId: undefined,
        issueId: undefined,
    });
    await ctx.ops.append({
        id: (0, crypto_1.randomUUID)(),
        entity: "active_context",
        entityId: ctx.userId,
        action: "update",
        payload: { repoId },
        timestamp: now,
        userId: ctx.userId,
        deviceId: ctx.deviceId,
    });
    await emitContextChanged(ctx);
    return context;
}
/**
 * Switch active stream
 * Requires repo to be set
 * Clears issue
 */
async function switchStream(ctx, streamId) {
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
        id: (0, crypto_1.randomUUID)(),
        entity: "active_context",
        entityId: ctx.userId,
        action: "update",
        payload: { streamId },
        timestamp: now,
        userId: ctx.userId,
        deviceId: ctx.deviceId,
    });
    await emitContextChanged(ctx);
    return context;
}
/**
 * Switch active issue
 * Requires stream to be set
 */
async function switchIssue(ctx, issueId) {
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
        id: (0, crypto_1.randomUUID)(),
        entity: "active_context",
        entityId: ctx.userId,
        action: "update",
        payload: { issueId },
        timestamp: now,
        userId: ctx.userId,
        deviceId: ctx.deviceId,
    });
    await emitContextChanged(ctx);
    return context;
}
/**
 * Clear active issue only
 * (equivalent to git checkout --)
 */
async function clearIssue(ctx) {
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
        id: (0, crypto_1.randomUUID)(),
        entity: "active_context",
        entityId: ctx.userId,
        action: "update",
        payload: { issueId: null },
        timestamp: now,
        userId: ctx.userId,
        deviceId: ctx.deviceId,
    });
    await emitContextChanged(ctx);
    return context;
}
/**
 * Clear entire context
 * (equivalent to git checkout --detach)
 */
async function clearContext(ctx) {
    const now = ctx.now();
    await ctx.activeContext.clear(ctx.userId, ctx.deviceId);
    await ctx.ops.append({
        id: (0, crypto_1.randomUUID)(),
        entity: "active_context",
        entityId: ctx.userId,
        action: "delete",
        payload: {},
        timestamp: now,
        userId: ctx.userId,
        deviceId: ctx.deviceId,
    });
    await emitContextChanged(ctx);
}
async function emitContextChanged(ctx) {
    const context = await ctx.activeContext.get(ctx.userId, ctx.deviceId);
    ctx.events.emit({
        type: "context.changed",
        payload: {
            deviceId: ctx.deviceId,
            repoId: context?.repoId,
            streamId: context?.streamId,
            issueId: context?.issueId,
        },
    });
}
