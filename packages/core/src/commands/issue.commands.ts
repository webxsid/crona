import { randomUUID } from "crypto";
import type { Issue } from "../domain/issue";
import type { IssueStatus } from "../domain/issue";
import type { ICommandContext } from "./context";

/**
 * Create a new issue under a stream
 */
export async function createIssue(
  ctx: ICommandContext,
  input: {
    streamId: string;
    title: string;
    estimateMinutes?: number | undefined;
    notes?: string | undefined;
  }
): Promise<Issue> {
  if (!input.title.trim()) {
    throw new Error("Issue title cannot be empty");
  }

  if (
    input.estimateMinutes !== undefined &&
    input.estimateMinutes < 0
  ) {
    throw new Error("Estimate must be >= 0");
  }

  const issue: Issue = {
    id: randomUUID(),
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
    id: randomUUID(),
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
export async function updateIssue(
  ctx: ICommandContext,
  issueId: string,
  updates: {
    title?: string | undefined;
    estimateMinutes?: number | null | undefined;
    notes?: string | null | undefined;
  }
): Promise<Issue> {
  if (updates.title !== undefined && !updates.title.trim()) {
    throw new Error("Issue title cannot be empty");
  }

  if (
    updates.estimateMinutes !== undefined &&
    updates.estimateMinutes !== null &&
    updates.estimateMinutes < 0
  ) {
    throw new Error("Estimate must be >= 0");
  }

  const now = ctx.now();

  const updated = await ctx.issues.update(
    issueId,
    {
      title: updates.title?.trim(),
      estimateMinutes: updates.estimateMinutes,
      notes: updates.notes,
    },
    {
      userId: ctx.userId,
      now,
    }
  );

  await ctx.ops.append({
    id: randomUUID(),
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
export async function changeIssueStatus(
  ctx: ICommandContext,
  issueId: string,
  nextStatus: IssueStatus
): Promise<Issue> {
  const issue = await ctx.issues.getById(issueId, ctx.userId);
  if (!issue) {
    throw new Error("Issue not found");
  }

  // Simple, explicit state machine
  const allowed: Record<IssueStatus, IssueStatus[]> = {
    todo: ["active"],
    active: ["done", "todo"],
    done: ["todo"],
  };

  if (!allowed[issue.status].includes(nextStatus)) {
    throw new Error(
      `Invalid status transition: ${issue.status} → ${nextStatus}`
    );
  }

  const now = ctx.now();

  const updated = await ctx.issues.update(
    issueId,
    { status: nextStatus },
    {
      userId: ctx.userId,
      now,
    }
  );

  await ctx.ops.append({
    id: randomUUID(),
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
export async function deleteIssue(
  ctx: ICommandContext,
  issueId: string
): Promise<void> {
  const now = ctx.now();

  await ctx.issues.softDelete(issueId, {
    userId: ctx.userId,
    now,
  });

  await ctx.ops.append({
    id: randomUUID(),
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
export async function listIssuesByStream(
  ctx: ICommandContext,
  streamId: string
): Promise<Issue[]> {
  return ctx.issues.listByStream(streamId, ctx.userId);
}
