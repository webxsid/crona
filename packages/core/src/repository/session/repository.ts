import type { Session } from "../../domain";
import type { ParsedSessionNotes } from "../../session_notes";
import { SessionNotesParser } from "../../session_notes";
import { SqliteDb } from "../../storage";
import type { ISessionRepository } from "./interface";

export class SessionRepository
  implements ISessionRepository {
  async start(
    session: Session,
    meta: { userId: string; deviceId: string; now: string }
  ): Promise<Session> {
    await SqliteDb.getDB()
      .insertInto("sessions")
      .values({
        id: session.id,
        issue_id: session.issueId,
        start_time: session.startTime,
        end_time: null,
        duration_seconds: null,
        notes: session.notes ?? null,
        user_id: meta.userId,
        device_id: meta.deviceId,
        created_at: meta.now,
        updated_at: meta.now,
        deleted_at: null,
      })
      .execute();

    return session;
  }

  async stop(
    sessionId: string,
    updates: {
      endTime: string,
      durationSeconds: number,
      notes: string | undefined,
    },
    meta: { userId: string; deviceId: string; now: string }
  ): Promise<Session> {
    const result = await SqliteDb.getDB()
      .updateTable("sessions")
      .set({
        end_time: updates.endTime,
        duration_seconds: updates.durationSeconds,
        updated_at: meta.now,
        device_id: meta.deviceId,
        notes: updates.notes ?? null,
      })
      .where("id", "=", sessionId)
      .where("user_id", "=", meta.userId)
      .where("end_time", "is", null)
      .where("deleted_at", "is", null)
      .executeTakeFirst();

    if (result.numUpdatedRows === BigInt(0)) {
      throw new Error("Active session not found");
    }

    const row = await SqliteDb.getDB()
      .selectFrom("sessions")
      .select([
        "id",
        "issue_id",
        "start_time",
        "end_time",
        "duration_seconds",
        "notes",
      ])
      .where("id", "=", sessionId)
      .where("user_id", "=", meta.userId)
      .executeTakeFirst();

    if (!row) {
      throw new Error("Session disappeared after stop");
    }

    return {
      id: row.id,
      issueId: row.issue_id,
      startTime: row.start_time,
      endTime: row.end_time ?? undefined,
      durationSeconds: row.duration_seconds ?? undefined,
      notes: row.notes ?? undefined,
    };
  }

  async getActiveSession(
    userId: string
  ): Promise<Session | null> {
    const row = await SqliteDb.getDB()
      .selectFrom("sessions")
      .select([
        "id",
        "issue_id",
        "start_time",
        "notes",
      ])
      .where("user_id", "=", userId)
      .where("end_time", "is", null)
      .where("deleted_at", "is", null)
      .orderBy("start_time", "desc")
      .executeTakeFirst();

    if (!row) return null;

    return {
      id: row.id,
      issueId: row.issue_id,
      startTime: row.start_time,
      notes: row.notes ?? undefined,
    };
  }

  async listByIssue(
    issueId: string,
    userId: string
  ): Promise<Session[]> {
    const rows = await SqliteDb.getDB()
      .selectFrom("sessions")
      .select([
        "id",
        "issue_id",
        "start_time",
        "end_time",
        "duration_seconds",
        "notes",
      ])
      .where("issue_id", "=", issueId)
      .where("user_id", "=", userId)
      .where("deleted_at", "is", null)
      .orderBy("start_time", "asc")
      .execute();

    return rows.map((r) => ({
      id: r.id,
      issueId: r.issue_id,
      startTime: r.start_time,
      endTime: r.end_time ?? undefined,
      durationSeconds: r.duration_seconds ?? undefined,
      notes: r.notes ?? undefined,
    }));
  }

  async getSessiobById(sessionId: string, userId: string): Promise<Session | null> {
    const row = await SqliteDb.getDB()
      .selectFrom("sessions")
      .select([
        "id",
        "issue_id",
        "start_time",
        "end_time",
        "duration_seconds",
        "notes",
      ])
      .where("id", "=", sessionId)
      .where("user_id", "=", userId)
      .where("deleted_at", "is", null)
      .executeTakeFirst();

    if (!row) return null;

    return {
      id: row.id,
      issueId: row.issue_id,
      startTime: row.start_time,
      endTime: row.end_time ?? undefined,
      durationSeconds: row.duration_seconds ?? undefined,
      notes: row.notes ?? undefined,
    };
  }

  async getLastSessionForIssue(issueId: string, userId: string): Promise<Session | null> {
    const row = await SqliteDb.getDB()
      .selectFrom("sessions")
      .select([
        "id",
        "issue_id",
        "start_time",
        "end_time",
        "duration_seconds",
        "notes",
      ])
      .where("issue_id", "=", issueId)
      .where("user_id", "=", userId)
      .where("deleted_at", "is", null)
      .orderBy("start_time", "desc")
      .executeTakeFirst();

    if (!row) return null;

    return {
      id: row.id,
      issueId: row.issue_id,
      startTime: row.start_time,
      endTime: row.end_time ?? undefined,
      durationSeconds: row.duration_seconds ?? undefined,
      notes: row.notes ?? undefined,
    };
  }

  async ammendSessionNotes(sessionId: string, newNotes: string, meta: { userId: string; deviceId: string; now: string; }): Promise<Session> {
    const result = await SqliteDb.getDB()
      .updateTable("sessions")
      .set({
        notes: newNotes,
        updated_at: meta.now,
        device_id: meta.deviceId,
      })
      .where("id", "=", sessionId)
      .where("user_id", "=", meta.userId)
      .where("deleted_at", "is", null)
      .executeTakeFirst();

    if (result.numUpdatedRows === BigInt(0)) {
      throw new Error("Session not found for ammending notes");
    }

    const row = await SqliteDb.getDB()
      .selectFrom("sessions")
      .select([
        "id",
        "issue_id",
        "start_time",
        "end_time",
        "duration_seconds",
        "notes",
      ])
      .where("id", "=", sessionId)
      .where("user_id", "=", meta.userId)
      .executeTakeFirst();

    if (!row) {
      throw new Error("Session disappeared after ammending notes");
    }

    return {
      id: row.id,
      issueId: row.issue_id,
      startTime: row.start_time,
      endTime: row.end_time ?? undefined,
      durationSeconds: row.duration_seconds ?? undefined,
      notes: row.notes ?? undefined,
    };
  }

  async getLastSessionForUser(userId: string): Promise<Session | null> {
    const row = await SqliteDb.getDB()
      .selectFrom("sessions")
      .select([
        "id",
        "issue_id",
        "start_time",
        "end_time",
        "duration_seconds",
        "notes",
      ])
      .where("user_id", "=", userId)
      .where("deleted_at", "is", null)
      .orderBy("start_time", "desc")
      .executeTakeFirst();

    if (!row) return null;

    return {
      id: row.id,
      issueId: row.issue_id,
      startTime: row.start_time,
      endTime: row.end_time ?? undefined,
      durationSeconds: row.duration_seconds ?? undefined,
      notes: row.notes ?? undefined,
    };
  }

  async listEnded(
    input: {
      userId: string;
      repoId?: string | undefined;
      streamId?: string | undefined;
      issueId?: string | undefined;
      since?: string | undefined;
      until?: string | undefined;
      limit?: number | undefined;
      offset?: number | undefined;
    }
  ): Promise<Array<Session & { parsedNotes: ParsedSessionNotes }>> {
    let query = SqliteDb.getDB()
      .selectFrom("sessions")
      .select([
        "id",
        "issue_id",
        "start_time",
        "end_time",
        "duration_seconds",
        "notes",
      ])
      .where("user_id", "=", input.userId)
      .where("end_time", "is not", null)
      .where("deleted_at", "is", null)
      .orderBy("start_time", "desc");

    if (input.issueId) {
      query = query.where("issue_id", "=", input.issueId);
    }

    if (input.since) {
      query = query.where("start_time", ">=", input.since);
    }

    if (input.until) {
      query = query.where("start_time", "<=", input.until);
    }

    if (input.limit) {
      query = query.limit(input.limit);
    }

    if (input.offset) {
      query = query.offset(input.offset);
    }

    const rows = await query.execute();

    return rows.map((r) => ({
      id: r.id,
      issueId: r.issue_id,
      startTime: r.start_time,
      endTime: r.end_time ?? undefined,
      durationSeconds: r.duration_seconds ?? undefined,
      notes: r.notes ?? undefined,
      parsedNotes: SessionNotesParser.parse(r.notes ?? null),
    }));
  }
}
