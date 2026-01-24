"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.createIssue = createIssue;
exports.updateIssue = updateIssue;
exports.changeIssueStatus = changeIssueStatus;
exports.deleteIssue = deleteIssue;
exports.listIssuesByStream = listIssuesByStream;
const crypto_1 = require("crypto");
/**
 * Create a new issue under a stream
 */
async function createIssue(ctx, input) {
    if (!input.title.trim()) {
        throw new Error("Issue title cannot be empty");
    }
    if (input.estimateMinutes !== undefined &&
        input.estimateMinutes < 0) {
        throw new Error("Estimate must be >= 0");
    }
    const issue = {
        id: (0, crypto_1.randomUUID)(),
        streamId: input.streamId,
        title: input.title.trim(),
        status: "todo",
        estimateMinutes: input.estimateMinutes,
        notes: input.notes,
    };
    const now = ctx.now();
    await ctx.issues.create(issue, {
        userId: ctx.userId,
        now,
    });
    await ctx.ops.append({
        id: (0, crypto_1.randomUUID)(),
        entity: "issue",
        entityId: issue.id,
        action: "create",
        payload: issue,
        timestamp: now,
        userId: ctx.userId,
        deviceId: ctx.deviceId,
    });
    return issue;
}
/**
 * Update issue metadata (title, estimate, notes)
 * Does NOT handle status transitions
 */
async function updateIssue(ctx, issueId, updates) {
    if (updates.title !== undefined && !updates.title.trim()) {
        throw new Error("Issue title cannot be empty");
    }
    if (updates.estimateMinutes !== undefined &&
        updates.estimateMinutes !== null &&
        updates.estimateMinutes < 0) {
        throw new Error("Estimate must be >= 0");
    }
    const now = ctx.now();
    const updated = await ctx.issues.update(issueId, {
        title: updates.title?.trim(),
        estimateMinutes: updates.estimateMinutes,
        notes: updates.notes,
    }, {
        userId: ctx.userId,
        now,
    });
    await ctx.ops.append({
        id: (0, crypto_1.randomUUID)(),
        entity: "issue",
        entityId: issueId,
        action: "update",
        payload: updates,
        timestamp: now,
        userId: ctx.userId,
        deviceId: ctx.deviceId,
    });
    return updated;
}
/**
 * Change issue status
 * Explicit command to keep state machine clean
 */
async function changeIssueStatus(ctx, issueId, nextStatus) {
    const issue = await ctx.issues.getById(issueId, ctx.userId);
    if (!issue) {
        throw new Error("Issue not found");
    }
    // Simple, explicit state machine
    const allowed = {
        todo: ["active"],
        active: ["done", "todo"],
        done: ["todo"],
    };
    if (!allowed[issue.status].includes(nextStatus)) {
        throw new Error(`Invalid status transition: ${issue.status} → ${nextStatus}`);
    }
    const now = ctx.now();
    const updated = await ctx.issues.update(issueId, { status: nextStatus }, {
        userId: ctx.userId,
        now,
    });
    await ctx.ops.append({
        id: (0, crypto_1.randomUUID)(),
        entity: "issue",
        entityId: issueId,
        action: "update",
        payload: { status: nextStatus },
        timestamp: now,
        userId: ctx.userId,
        deviceId: ctx.deviceId,
    });
    return updated;
}
/**
 * Delete an issue (soft delete)
 */
async function deleteIssue(ctx, issueId) {
    const now = ctx.now();
    await ctx.issues.softDelete(issueId, {
        userId: ctx.userId,
        now,
    });
    await ctx.ops.append({
        id: (0, crypto_1.randomUUID)(),
        entity: "issue",
        entityId: issueId,
        action: "delete",
        payload: null,
        timestamp: now,
        userId: ctx.userId,
        deviceId: ctx.deviceId,
    });
}
/**
 * List issues in a stream
 * Read-only
 */
async function listIssuesByStream(ctx, streamId) {
    return ctx.issues.listByStream(streamId, ctx.userId);
}
