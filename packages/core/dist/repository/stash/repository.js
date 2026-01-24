"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.StashRepository = void 0;
const storage_1 = require("../../storage");
class StashRepository {
    async list(userId) {
        const rows = await storage_1.SqliteDb.getDB()
            .selectFrom("stash")
            .selectAll()
            .where("user_id", "=", userId)
            .orderBy("created_at", "desc")
            .execute();
        return rows.map(this.map);
    }
    async get(id, userId) {
        const row = await storage_1.SqliteDb.getDB()
            .selectFrom("stash")
            .selectAll()
            .where("id", "=", id)
            .where("user_id", "=", userId)
            .executeTakeFirst();
        return row ? this.map(row) : null;
    }
    async save(stash) {
        await storage_1.SqliteDb.getDB()
            .insertInto("stash")
            .values({
            id: stash.id,
            user_id: stash.userId,
            device_id: stash.deviceId,
            repo_id: stash.repoId ?? null,
            stream_id: stash.streamId ?? null,
            issue_id: stash.issueId ?? null,
            session_id: stash.sessionId ?? null,
            created_at: stash.createdAt.toISOString(),
            updated_at: new Date().toISOString(),
        })
            .execute();
    }
    async delete(id, userId) {
        await storage_1.SqliteDb.getDB()
            .deleteFrom("stash")
            .where("id", "=", id)
            .where("user_id", "=", userId)
            .execute();
    }
    map(row) {
        return {
            id: row.id,
            userId: row.user_id,
            deviceId: row.device_id,
            repoId: row.repo_id ?? undefined,
            streamId: row.stream_id ?? undefined,
            issueId: row.issue_id ?? undefined,
            sessionId: row.session_id ?? undefined,
            createdAt: new Date(row.created_at),
            updatedAt: new Date(row.updated_at),
        };
    }
}
exports.StashRepository = StashRepository;
