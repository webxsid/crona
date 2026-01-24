"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.ActiveContextRepository = void 0;
const storage_1 = require("../../storage");
class ActiveContextRepository {
    async get(userId, deviceId) {
        const row = await storage_1.SqliteDb.getDB()
            .selectFrom("active_context")
            .selectAll()
            .where("user_id", "=", userId)
            .where("device_id", "=", deviceId)
            .executeTakeFirst();
        return row
            ? {
                userId: row.user_id,
                deviceId: row.device_id,
                repoId: row.repo_id ?? undefined,
                streamId: row.stream_id ?? undefined,
                issueId: row.issue_id ?? undefined,
                updatedAt: new Date(row.updated_at),
            }
            : null;
    }
    async set(userId, deviceId, context) {
        const now = new Date().toISOString();
        await storage_1.SqliteDb.getDB()
            .insertInto("active_context")
            .values({
            user_id: userId,
            device_id: deviceId,
            repo_id: context.repoId ?? null,
            stream_id: context.streamId ?? null,
            issue_id: context.issueId ?? null,
            updated_at: now,
        })
            .onConflict((oc) => oc.column("user_id").doUpdateSet({
            repo_id: context.repoId ?? null,
            stream_id: context.streamId ?? null,
            issue_id: context.issueId ?? null,
            updated_at: now,
        }))
            .execute();
        return (await this.get(userId, deviceId));
    }
    async clear(userId, deviceId) {
        await storage_1.SqliteDb.getDB()
            .updateTable("active_context")
            .set({
            repo_id: null,
            stream_id: null,
            issue_id: null,
            updated_at: new Date().toISOString(),
        })
            .where("user_id", "=", userId)
            .where("device_id", "=", deviceId)
            .execute();
    }
    async initializeDefaults(userId, deviceId) {
        const existing = await this.get(userId, deviceId);
        if (existing) {
            return;
        }
        await storage_1.SqliteDb.getDB()
            .insertInto("active_context")
            .values({
            user_id: userId,
            device_id: deviceId,
            repo_id: null,
            stream_id: null,
            issue_id: null,
            updated_at: new Date().toISOString(),
        })
            .execute();
    }
}
exports.ActiveContextRepository = ActiveContextRepository;
