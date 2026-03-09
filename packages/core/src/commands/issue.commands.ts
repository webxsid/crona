import { randomUUID } from "crypto";
import type { DailyIssueSummary, Issue } from "../domain/issue";
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
    todoForDate?: string | undefined;
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
    todoForDate: input.todoForDate,
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

  ctx.events.emit({ type: "issue.created", payload: issue });

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

  ctx.events.emit({ type: "issue.updated", payload: updated });

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
    todo: ["active", "abandoned"],
    active: ["done", "todo", "abandoned"],
    done: ["todo"],
    abandoned: ["todo"],
  };

  if (!allowed[issue.status].includes(nextStatus)) {
    throw new Error(
      `Invalid status transition: ${issue.status} → ${nextStatus}`
    );
  }

  const now = ctx.now();
  let completedAt: string | null | undefined;
  let abandonedAt: string | null | undefined;

  switch (nextStatus) {
    case "done":
      completedAt = now;
      abandonedAt = null;
      break;
    case "abandoned":
      completedAt = null;
      abandonedAt = now;
      break;
    default:
      completedAt = null;
      abandonedAt = null;
      break;
  }

  const updated = await ctx.issues.update(
    issueId,
    {
      status: nextStatus,
      completedAt,
      abandonedAt,
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
    payload: { status: nextStatus },
    timestamp: now,
    userId: ctx.userId,
    deviceId: ctx.deviceId,
  });

  ctx.events.emit({ type: "issue.updated", payload: updated });

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

  ctx.events.emit({ type: "issue.deleted", payload: { id: issueId } });
}

export async function restoreIssue(
  ctx: ICommandContext,
  issueId: string
): Promise<void> {
  const now = ctx.now();

  await ctx.issues.restoreDeletedById(issueId, {
    userId: ctx.userId,
    now,
  });

  await ctx.ops.append({
    id: randomUUID(),
    entity: "issue",
    entityId: issueId,
    action: "restore",
    payload: null,
    timestamp: now,
    userId: ctx.userId,
    deviceId: ctx.deviceId,
  });
}

export async function cacadeSoftDeleteIssuesByStreamId(
  ctx: ICommandContext,
  streamId: string
): Promise<void> {
  const now = ctx.now();

  await ctx.issues.cascadeSoftDeleteByStreamId(streamId, {
    userId: ctx.userId,
    now,
  });

}

export async function cascadeSoftDeleteIssuesByRepoId(
  ctx: ICommandContext,
  repoId: string
): Promise<void> {
  const now = ctx.now();

  await ctx.issues.cascadeSoftDeleteByRepoId(repoId, {
    userId: ctx.userId,
    now,
  });

}

export async function restoreIssuesByStreamId(
  ctx: ICommandContext,
  streamId: string
): Promise<void> {
  const now = ctx.now();

  await ctx.issues.restoreDeletedByStreamId(streamId, {
    userId: ctx.userId,
    now,
  });

}

export async function restoreIssuesByRepoId(
  ctx: ICommandContext,
  repoId: string
): Promise<void> {
  const now = ctx.now();

  await ctx.issues.restoreDeletedByRepoId(repoId, {
    userId: ctx.userId,
    now,
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

export async function markIssueTodoForDate(
  ctx: ICommandContext,
  issueId: string,
  todoForDate: string
): Promise<Issue> {
  const now = ctx.now();

  const updated = await ctx.issues.update(
    issueId,
    {
      todoForDate,
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
    payload: { todoForDate },
    timestamp: now,
    userId: ctx.userId,
    deviceId: ctx.deviceId,
  });

  ctx.events.emit({ type: "issue.updated", payload: updated });

  return updated;
}

export async function markIssueTodoForToday(
  ctx: ICommandContext,
  issueId: string
): Promise<Issue> {
  const today = ctx.now().split("T")[0];
  if (!today) {
    throw new Error("Invalid date");
  }
  return markIssueTodoForDate(ctx, issueId, today);
}

export async function clearIssueTodoForDate(
  ctx: ICommandContext,
  issueId: string
): Promise<Issue> {
  const now = ctx.now();

  const updated = await ctx.issues.update(
    issueId,
    {
      todoForDate: null,
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
    payload: { todoForDate: null },
    timestamp: now,
    userId: ctx.userId,
    deviceId: ctx.deviceId,
  });

  ctx.events.emit({ type: "issue.updated", payload: updated });

  return updated;
}

export async function clearTodayTodos(
  ctx: ICommandContext,
): Promise<void> {
  const today = ctx.now().split("T")[0]; // YYYY-MM-DD

  if (!today) {
    throw new Error("Invalid date");
  }

  const issues = await ctx.issues.listByTodoForDate(today, ctx.userId);

  for (const issue of issues) {
    await clearIssueTodoForDate(ctx, issue.id);
  }
}

export async function computeDailyIssueSummaryForDate(
  ctx: ICommandContext,
  date: string,
): Promise<DailyIssueSummary> {
  if (!date) {
    throw new Error("Invalid date");
  }

  const issues = await ctx.issues.listByTodoForDate(date, ctx.userId);

  const totalEstimatedMinutes = issues.reduce((sum, issue) => {
    return sum + (issue.estimateMinutes ?? 0);
  }, 0);

  const completedIssues = issues.filter((issue) =>
    issue.completedAt?.startsWith(date)
  ).length;

  const abandonedIssues = issues.filter((issue) =>
    issue.abandonedAt?.startsWith(date)
  ).length;

  const dayStart = `${date}T00:00:00.000Z`;
  const dayEnd = `${date}T23:59:59.999Z`;
  const endedSessions = await ctx.sessions.listEnded({
    userId: ctx.userId,
    since: dayStart,
    until: dayEnd,
  });
  const issueIDs = new Set(issues.map((issue) => issue.id));
  const workedSeconds = endedSessions.reduce((sum, session) => {
    if (!issueIDs.has(session.issueId)) {
      return sum;
    }
    return sum + (session.durationSeconds ?? 0);
  }, 0);

  return {
    date,
    totalIssues: issues.length,
    issues,
    totalEstimatedMinutes,
    completedIssues,
    abandonedIssues,
    workedSeconds,
  };
}

export async function computeDailyIssueSummaryForToday(
  ctx: ICommandContext,
): Promise<DailyIssueSummary> {
  const today = ctx.now().split("T")[0];
  if (!today) {
    throw new Error("Invalid date");
  }
  return computeDailyIssueSummaryForDate(ctx, today);
}
