import { randomUUID } from "crypto";
import type { Session } from "../domain/session";
import type { ICommandContext } from "./context";
import { elapsedSeconds } from "../timer/time";
import type { SessionSegmentType } from "../domain";
import type { ParsedSessionNotes } from "../session_notes";
import { SessionNotesParser } from "../session_notes";

/**
 * Start a session for an issue
 * Enforces: only one active session per user
 */
export async function startSession(
  ctx: ICommandContext,
  issueId: string
): Promise<Session> {
  const existing = await ctx.sessions.getActiveSession(ctx.userId);
  if (existing) {
    throw new Error("A session is already running");
  }

  const now = ctx.now();

  const session: Session = {
    id: randomUUID(),
    issueId,
    startTime: now,
  };

  await ctx.sessions.start(session, {
    userId: ctx.userId,
    deviceId: ctx.deviceId,
    now,
  });

  // 🔑 Segment lifecycle starts here
  await ctx.sessionSegments.startSegment(
    ctx.userId,
    ctx.deviceId,
    session.id,
    "work"
  );

  await ctx.ops.append({
    id: randomUUID(),
    entity: "session",
    entityId: session.id,
    action: "create",
    payload: session,
    timestamp: now,
    userId: ctx.userId,
    deviceId: ctx.deviceId,
  });

  return session;
}


/**
 * Stop the currently active session
 * Idempotent: calling stop with no active session is a no-op
 */
export async function stopSession(
  ctx: ICommandContext,
  commitMessage?: string | undefined
): Promise<Session | null> {
  const active = await ctx.sessions.getActiveSession(ctx.userId);
  if (!active) return null;

  const now = ctx.now();

  // 🔑 Close active segment
  await ctx.sessionSegments.endActiveSegment(
    ctx.userId,
    ctx.deviceId,
    active.id
  );

  const segments = await ctx.sessionSegments.listBySession(
    active.id,
  );

  const workSummary = SessionNotesParser.computeWorkSummary(segments);
  const workSummaryLines = SessionNotesParser.formatWorkSummary(workSummary);

  // --- Notes / commit handling ---
  const existingNotes = active.notes ?? null;
  const parsedExistingNotes = SessionNotesParser.parse(existingNotes);

  const activeContext = await ctx.activeContext.get(ctx.userId, ctx.deviceId);
  const mergedNotes = SessionNotesParser.generateDefaultSessionNotes({
    commit: parsedExistingNotes.commit ? `${parsedExistingNotes.commit}\n${commitMessage ?? ""}`.trim() : commitMessage,
    workSummary: parsedExistingNotes.work
      ? [...workSummaryLines, "", parsedExistingNotes.work.split("\n")].flat()
      : workSummaryLines,
    repoId: activeContext?.repoId,
    streamId: activeContext?.streamId,
    issueId: active.issueId,
  });

  // Enforce commit presence
  SessionNotesParser.assertCommitMessage(mergedNotes);

  let offSetSeconds = 0;
  for (const segment of segments) {
    offSetSeconds += segment.elapsedOffsetSeconds ?? 0;
  }


  const stopped = await ctx.sessions.stop(
    active.id,
    {
      endTime: now,
      durationSeconds: elapsedSeconds(active.startTime, now) + offSetSeconds,
      notes: mergedNotes,
    },
    {
      userId: ctx.userId,
      deviceId: ctx.deviceId,
      now,
    }
  );

  await ctx.ops.append({
    id: randomUUID(),
    entity: "session",
    entityId: stopped.id,
    action: "update",
    payload: { endTime: stopped.endTime },
    timestamp: now,
    userId: ctx.userId,
    deviceId: ctx.deviceId,
  });

  return stopped;
}

export async function ammendSessionNotes(
  ctx: ICommandContext,
  message: string,
  sessionId?: string | undefined
): Promise<Session> {
  let session: Session | null = null;

  if (sessionId) {
    session = await ctx.sessions.getSessiobById(sessionId, ctx.userId);
    if (!session) {
      throw new Error("Session not found");
    }
  } else {
    session = await ctx.sessions.getLastSessionForUser(ctx.userId);
  }
  if (!session) {
    throw new Error("No session found to ammend");
  }
  const existingNotes = session.notes ?? null;

  const mergedNotes = SessionNotesParser.ammendCommitMessage(
    existingNotes,
    message
  );

  const updated = await ctx.sessions.ammendSessionNotes(
    session.id,
    mergedNotes,
    {
      userId: ctx.userId,
      deviceId: ctx.deviceId,
      now: ctx.now(),
    }
  )

  await ctx.ops.append({
    id: randomUUID(),
    entity: "session",
    entityId: updated.id,
    action: "update",
    payload: { notes: updated.notes },
    timestamp: ctx.now(),
    userId: ctx.userId,
    deviceId: ctx.deviceId,
  });

  return updated;
}


/**
 * Pause the current session
 * Stores elapsed time in stash
 */
export async function pauseSession(ctx: ICommandContext, nextSegmentType: SessionSegmentType = "rest"): Promise<void> {
  const active = await ctx.sessions.getActiveSession(ctx.userId);
  if (!active) return;

  const current = await ctx.sessionSegments.getActive(
    ctx.userId,
    ctx.deviceId,
    active.id
  );

  if (current?.segmentType === "rest") return;

  await ctx.sessionSegments.startSegment(
    ctx.userId,
    ctx.deviceId,
    active.id,
    nextSegmentType
  );
}

export async function resumeSession(ctx: ICommandContext): Promise<void> {
  const active = await ctx.sessions.getActiveSession(ctx.userId);
  if (!active) return;

  const current = await ctx.sessionSegments.getActive(
    ctx.userId,
    ctx.deviceId,
    active.id
  );

  if (current?.segmentType === "work") return;

  await ctx.sessionSegments.startSegment(
    ctx.userId,
    ctx.deviceId,
    active.id,
    "work"
  );
}

/**
 * List Session History
 * Read-only
 */
export async function listSessionHistory(
  ctx: ICommandContext,
  query: {
    repoId: string | undefined;
    streamId: string | undefined;
    issueId: string | undefined;
    since: string | undefined;
    until: string | undefined;
    limit: number | undefined;
    offset: number | undefined;
  },
  useContext: boolean = false
): Promise<Array<Session & { parsedNotes: ParsedSessionNotes }>> {
  if (useContext) {
    const activeContext = await ctx.activeContext.get(ctx.userId, ctx.deviceId);
    if (activeContext?.repoId) {
      query.repoId = activeContext.repoId;
    }
    if (activeContext?.streamId) {
      query.streamId = activeContext.streamId;
    }
    if (activeContext?.issueId) {
      query.issueId = activeContext.issueId;
    }
  }

  if (!query.limit) {
    console.warn("listSessionHistory called without limit, defaulting to 100");
    query.limit = 100;
  }


  return ctx.sessions.listEnded({
    userId: ctx.userId,
    repoId: query.repoId,
    streamId: query.streamId,
    issueId: query.issueId,
    since: query.since,
    until: query.until,
    limit: query.limit,
    offset: query.offset,
  });
}
/**
 * List all sessions for an issue
 * Read-only
 */
export async function listSessionsByIssue(
  ctx: ICommandContext,
  issueId: string
): Promise<Session[]> {
  return ctx.sessions.listByIssue(issueId, ctx.userId);
}
