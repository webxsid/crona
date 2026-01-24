"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.createRepo = createRepo;
exports.updateRepo = updateRepo;
exports.deleteRepo = deleteRepo;
exports.listRepos = listRepos;
const crypto_1 = require("crypto");
/**
 * Create a new repo
 */
async function createRepo(ctx, input) {
    if (!input.name.trim()) {
        throw new Error("Repo name cannot be empty");
    }
    const repo = {
        id: (0, crypto_1.randomUUID)(),
        name: input.name.trim(),
        color: input.color,
    };
    const now = ctx.now();
    await ctx.repos.create(repo, {
        userId: ctx.userId,
        now,
    });
    await ctx.ops.append({
        id: (0, crypto_1.randomUUID)(),
        entity: "repo",
        entityId: repo.id,
        action: "create",
        payload: repo,
        timestamp: now,
        userId: ctx.userId,
        deviceId: ctx.deviceId,
    });
    return repo;
}
/**
 * Rename / recolor a repo
 */
async function updateRepo(ctx, repoId, updates) {
    if (updates.name !== undefined && !updates.name.trim()) {
        throw new Error("Repo name cannot be empty");
    }
    const now = ctx.now();
    const updateObj = {};
    if (updates.name !== undefined) {
        updateObj["name"] = updates.name.trim();
    }
    if (Object.prototype.hasOwnProperty.call(updates, "color")) {
        updateObj["color"] = updates.color;
    }
    const updated = await ctx.repos.update(repoId, updateObj, {
        userId: ctx.userId,
        now,
    });
    await ctx.ops.append({
        id: (0, crypto_1.randomUUID)(),
        entity: "repo",
        entityId: repoId,
        action: "update",
        payload: updates,
        timestamp: now,
        userId: ctx.userId,
        deviceId: ctx.deviceId,
    });
    return updated;
}
/**
 * Delete a repo (soft delete)
 * NOTE: cascading deletes are handled at storage or command level later
 */
async function deleteRepo(ctx, repoId) {
    const now = ctx.now();
    await ctx.repos.softDelete(repoId, {
        userId: ctx.userId,
        now,
    });
    await ctx.ops.append({
        id: (0, crypto_1.randomUUID)(),
        entity: "repo",
        entityId: repoId,
        action: "delete",
        payload: null,
        timestamp: now,
        userId: ctx.userId,
        deviceId: ctx.deviceId,
    });
}
/**
 * List all repos for current user
 * Read-only, no ops emitted
 */
async function listRepos(ctx) {
    return ctx.repos.list(ctx.userId);
}
