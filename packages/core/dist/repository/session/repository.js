"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.SessionRepository = void 0;
const session_notes_1 = require("../../session_notes");
const storage_1 = require("../../storage");
class SessionRepository {
    async start(session, meta) {
        await storage_1.SqliteDb.getDB()
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
    async stop(sessionId, updates, meta) {
        const result = await storage_1.SqliteDb.getDB()
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
        const row = await storage_1.SqliteDb.getDB()
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
    async getActiveSession(userId) {
        const row = await storage_1.SqliteDb.getDB()
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
        if (!row)
            return null;
        return {
            id: row.id,
            issueId: row.issue_id,
            startTime: row.start_time,
            notes: row.notes ?? undefined,
        };
    }
    async listByIssue(issueId, userId) {
        const rows = await storage_1.SqliteDb.getDB()
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
    async getSessiobById(sessionId, userId) {
        const row = await storage_1.SqliteDb.getDB()
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
        if (!row)
            return null;
        return {
            id: row.id,
            issueId: row.issue_id,
            startTime: row.start_time,
            endTime: row.end_time ?? undefined,
            durationSeconds: row.duration_seconds ?? undefined,
            notes: row.notes ?? undefined,
        };
    }
    async getLastSessionForIssue(issueId, userId) {
        const row = await storage_1.SqliteDb.getDB()
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
        if (!row)
            return null;
        return {
            id: row.id,
            issueId: row.issue_id,
            startTime: row.start_time,
            endTime: row.end_time ?? undefined,
            durationSeconds: row.duration_seconds ?? undefined,
            notes: row.notes ?? undefined,
        };
    }
    async ammendSessionNotes(sessionId, newNotes, meta) {
        const result = await storage_1.SqliteDb.getDB()
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
        const row = await storage_1.SqliteDb.getDB()
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
    async getLastSessionForUser(userId) {
        const row = await storage_1.SqliteDb.getDB()
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
        if (!row)
            return null;
        return {
            id: row.id,
            issueId: row.issue_id,
            startTime: row.start_time,
            endTime: row.end_time ?? undefined,
            durationSeconds: row.duration_seconds ?? undefined,
            notes: row.notes ?? undefined,
        };
    }
    async listEnded(input) {
        let query = storage_1.SqliteDb.getDB()
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
            parsedNotes: session_notes_1.SessionNotesParser.parse(r.notes ?? null),
        }));
    }
}
exports.SessionRepository = SessionRepository;
