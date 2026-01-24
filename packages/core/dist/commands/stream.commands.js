"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.createStream = createStream;
exports.updateStream = updateStream;
exports.deleteStream = deleteStream;
exports.listStreamsByRepo = listStreamsByRepo;
const crypto_1 = require("crypto");
/**
 * Create a new stream under a repo
 */
async function createStream(ctx, input) {
    if (!input.name.trim()) {
        throw new Error("Stream name cannot be empty");
    }
    const stream = {
        id: (0, crypto_1.randomUUID)(),
        repoId: input.repoId,
        name: input.name.trim(),
        visibility: input.visibility ?? "personal",
    };
    const now = ctx.now();
    await ctx.streams.create(stream, {
        userId: ctx.userId,
        now,
    });
    await ctx.ops.append({
        id: (0, crypto_1.randomUUID)(),
        entity: "stream",
        entityId: stream.id,
        action: "create",
        payload: stream,
        timestamp: now,
        userId: ctx.userId,
        deviceId: ctx.deviceId,
    });
    return stream;
}
/**
 * Rename / change visibility of a stream
 */
async function updateStream(ctx, streamId, updates) {
    if (updates.name !== undefined && !updates.name.trim()) {
        throw new Error("Stream name cannot be empty");
    }
    const now = ctx.now();
    const updated = await ctx.streams.update(streamId, {
        name: updates.name?.trim(),
        visibility: updates.visibility,
    }, {
        userId: ctx.userId,
        now,
    });
    await ctx.ops.append({
        id: (0, crypto_1.randomUUID)(),
        entity: "stream",
        entityId: streamId,
        action: "update",
        payload: updates,
        timestamp: now,
        userId: ctx.userId,
        deviceId: ctx.deviceId,
    });
    return updated;
}
/**
 * Delete a stream (soft delete)
 */
async function deleteStream(ctx, streamId) {
    const now = ctx.now();
    await ctx.streams.softDelete(streamId, {
        userId: ctx.userId,
        now,
    });
    await ctx.ops.append({
        id: (0, crypto_1.randomUUID)(),
        entity: "stream",
        entityId: streamId,
        action: "delete",
        payload: null,
        timestamp: now,
        userId: ctx.userId,
        deviceId: ctx.deviceId,
    });
}
/**
 * List streams for a repo
 * Read-only → no ops
 */
async function listStreamsByRepo(ctx, repoId) {
    return ctx.streams.listByRepo(repoId, ctx.userId);
}
