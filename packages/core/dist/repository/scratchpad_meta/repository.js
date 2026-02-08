"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.ScratchRepo = void 0;
const storage_1 = require("../../storage");
class ScratchRepo {
    async upsert(meta, userMeta) {
        await storage_1.SqliteDb.getDB()
            .insertInto("scratch_pad_meta")
            .values({
            id: meta.id,
            name: meta.name,
            path: meta.path,
            last_opened_at: meta.lastOpenedAt.toISOString(),
            pinned: meta.pinned ? 1 : 0,
            user_id: userMeta.userId,
            device_id: userMeta.deviceId,
        })
            .onConflict((oc) => oc
            .column("path")
            .doUpdateSet({
            name: meta.name,
            last_opened_at: meta.lastOpenedAt.toISOString(),
            pinned: meta.pinned ? 1 : 0,
        })
            .where("user_id", "=", userMeta.userId))
            .execute();
    }
    async list(userId, deviceId, options) {
        let query = storage_1.SqliteDb.getDB()
            .selectFrom("scratch_pad_meta")
            .selectAll()
            .where("user_id", "=", userId)
            .where("device_id", "=", deviceId);
        if (options?.pinnedOnly) {
            query = query.where("pinned", "=", 1);
        }
        const rows = await query.orderBy("last_opened_at", "desc").execute();
        return rows.map((row) => ({
            id: row.id,
            userId: row.user_id,
            deviceId: row.device_id,
            name: row.name,
            lastOpenedAt: new Date(row.last_opened_at),
            path: row.path,
            pinned: row.pinned === 1,
        }));
    }
    async get(path, meta) {
        const row = await storage_1.SqliteDb.getDB()
            .selectFrom("scratch_pad_meta")
            .selectAll()
            .where("path", "=", path)
            .where("user_id", "=", meta.userId)
            .where("device_id", "=", meta.deviceId)
            .executeTakeFirst();
        if (!row)
            return null;
        return {
            id: row.id,
            name: row.name,
            lastOpenedAt: new Date(row.last_opened_at),
            path: row.path,
            pinned: row.pinned === 1,
        };
    }
    async remove(path, meta) {
        await storage_1.SqliteDb.getDB()
            .deleteFrom("scratch_pad_meta")
            .where("path", "=", path)
            .where("user_id", "=", meta.userId)
            .where("device_id", "=", meta.deviceId)
            .execute();
    }
}
exports.ScratchRepo = ScratchRepo;
