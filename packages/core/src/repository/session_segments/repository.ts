import { randomUUID } from "node:crypto";
import type { SessionSegment, SessionSegmentType } from "../../domain";
import type { SessionSegmentsTable } from "../../storage";
import { SqliteDb } from "../../storage";
import type { ISessionSegmentRepository } from "./interface";

export class SessionSegmentRepository
  implements ISessionSegmentRepository {
  async getActive(
    userId: string,
    deviceId: string,
    sessionId: string
  ): Promise<SessionSegment | null> {
    const row = await SqliteDb.getDB()
      .selectFrom("session_segments")
      .selectAll()
      .where("user_id", "=", userId)
      .where("device_id", "=", deviceId)
      .where("session_id", "=", sessionId)
      .where("end_time", "is", null)
      .executeTakeFirst();

    return row ? this.map(row) : null;
  }

  async startSegment(
    userId: string,
    deviceId: string,
    sessionId: string,
    type: SessionSegmentType
  ): Promise<SessionSegment> {
    // Safety: end any dangling active segment
    await this.endActiveSegment(userId, deviceId, sessionId);

    const now = new Date().toISOString();
    const id = randomUUID();

    await SqliteDb.getDB()
      .insertInto("session_segments")
      .values({
        id,
        user_id: userId,
        device_id: deviceId,
        session_id: sessionId,
        segment_type: type,
        start_time: now,
        end_time: null,
        created_at: now,
      })
      .execute();

    return {
      id,
      userId,
      deviceId,
      sessionId,
      segmentType: type,
      startTime: new Date(now),
      createdAt: new Date(now),
    };
  }

  async endActiveSegment(
    userId: string,
    deviceId: string,
    sessionId: string
  ): Promise<void> {
    const now = new Date().toISOString();

    await SqliteDb.getDB()
      .updateTable("session_segments")
      .set({ end_time: now })
      .where("user_id", "=", userId)
      .where("session_id", "=", sessionId)
      .where("end_time", "is", null)
      .where("device_id", "=", deviceId)
      .execute();
  }

  async listBySession(
    sessionId: string
  ): Promise<SessionSegment[]> {
    const rows = await SqliteDb.getDB()
      .selectFrom("session_segments")
      .selectAll()
      .where("session_id", "=", sessionId)
      .orderBy("start_time", "asc")
      .execute();

    return rows.map(this.map);
  }

  async countWorkSegments(
    sessionId: string
  ): Promise<number> {
    const row = await SqliteDb.getDB()
      .selectFrom("session_segments")
      .select(({ fn }) => fn.countAll().as("count"))
      .where("session_id", "=", sessionId)
      .where("segment_type", "=", "work")
      .where("end_time", "is not", null)
      .executeTakeFirst();

    return Number(row?.count ?? 0);
  }
  async applyElapsedOffset(
    sessionId: string,
    offsetSeconds: number
  ): Promise<void> {
    if (offsetSeconds <= 0) return;

    // Only apply to the CURRENT active segment
    await SqliteDb.getDB()
      .updateTable("session_segments")
      .set({
        elapsed_offset_seconds: offsetSeconds,
      })
      .where("session_id", "=", sessionId)
      .where("end_time", "is", null) // active only
      .execute();
  }

  private map(
    row: SessionSegmentsTable
  ): SessionSegment {
    return {
      id: row.id,
      userId: row.user_id,
      deviceId: row.device_id,
      sessionId: row.session_id,
      segmentType: row.segment_type,
      startTime: new Date(row.start_time),
      endTime: row.end_time ? new Date(row.end_time) : undefined,
      createdAt: new Date(row.created_at),
    };
  }
}
