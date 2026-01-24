"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.SessionSegmentRepository = void 0;
const node_crypto_1 = require("node:crypto");
const storage_1 = require("../../storage");
class SessionSegmentRepository {
    async getActive(userId, deviceId, sessionId) {
        const row = await storage_1.SqliteDb.getDB()
            .selectFrom("session_segments")
            .selectAll()
            .where("user_id", "=", userId)
            .where("device_id", "=", deviceId)
            .where("session_id", "=", sessionId)
            .where("end_time", "is", null)
            .executeTakeFirst();
        return row ? this.map(row) : null;
    }
    async startSegment(userId, deviceId, sessionId, type) {
        // Safety: end any dangling active segment
        await this.endActiveSegment(userId, deviceId, sessionId);
        const now = new Date().toISOString();
        const id = (0, node_crypto_1.randomUUID)();
        await storage_1.SqliteDb.getDB()
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
    async endActiveSegment(userId, deviceId, sessionId) {
        const now = new Date().toISOString();
        await storage_1.SqliteDb.getDB()
            .updateTable("session_segments")
            .set({ end_time: now })
            .where("user_id", "=", userId)
            .where("session_id", "=", sessionId)
            .where("end_time", "is", null)
            .where("device_id", "=", deviceId)
            .execute();
    }
    async listBySession(sessionId) {
        const rows = await storage_1.SqliteDb.getDB()
            .selectFrom("session_segments")
            .selectAll()
            .where("session_id", "=", sessionId)
            .orderBy("start_time", "asc")
            .execute();
        return rows.map(this.map);
    }
    async countWorkSegments(sessionId) {
        const row = await storage_1.SqliteDb.getDB()
            .selectFrom("session_segments")
            .select(({ fn }) => fn.countAll().as("count"))
            .where("session_id", "=", sessionId)
            .where("segment_type", "=", "work")
            .where("end_time", "is not", null)
            .executeTakeFirst();
        return Number(row?.count ?? 0);
    }
    async applyElapsedOffset(sessionId, offsetSeconds) {
        if (offsetSeconds <= 0)
            return;
        // Only apply to the CURRENT active segment
        await storage_1.SqliteDb.getDB()
            .updateTable("session_segments")
            .set({
            elapsed_offset_seconds: offsetSeconds,
        })
            .where("session_id", "=", sessionId)
            .where("end_time", "is", null) // active only
            .execute();
    }
    map(row) {
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
exports.SessionSegmentRepository = SessionSegmentRepository;
