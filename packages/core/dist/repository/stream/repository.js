"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.SqliteStreamRepository = void 0;
const storage_1 = require("../../storage");
class SqliteStreamRepository {
    async create(stream, meta) {
        await storage_1.SqliteDb.getDB()
            .insertInto("streams")
            .values({
            id: stream.id,
            repo_id: stream.repoId,
            name: stream.name,
            visibility: stream.visibility,
            user_id: meta.userId,
            created_at: meta.now,
            updated_at: meta.now,
            deleted_at: null,
        })
            .execute();
        return stream;
    }
    async getById(streamId, userId) {
        const row = await storage_1.SqliteDb.getDB()
            .selectFrom("streams")
            .select(["id", "repo_id", "name", "visibility"])
            .where("id", "=", streamId)
            .where("user_id", "=", userId)
            .where("deleted_at", "is", null)
            .executeTakeFirst();
        if (!row)
            return null;
        return {
            id: row.id,
            repoId: row.repo_id,
            name: row.name,
            visibility: row.visibility,
        };
    }
    async listByRepo(repoId, userId) {
        const rows = await storage_1.SqliteDb.getDB()
            .selectFrom("streams")
            .select(["id", "repo_id", "name", "visibility"])
            .where("repo_id", "=", repoId)
            .where("user_id", "=", userId)
            .where("deleted_at", "is", null)
            .orderBy("created_at", "asc")
            .execute();
        return rows.map((r) => ({
            id: r.id,
            repoId: r.repo_id,
            name: r.name,
            visibility: r.visibility,
        }));
    }
    async update(streamId, updates, meta) {
        const result = await storage_1.SqliteDb.getDB()
            .updateTable("streams")
            .set({
            name: updates.name,
            visibility: updates.visibility,
            updated_at: meta.now,
        })
            .where("id", "=", streamId)
            .where("user_id", "=", meta.userId)
            .where("deleted_at", "is", null)
            .executeTakeFirst();
        if (result.numUpdatedRows === BigInt(0)) {
            throw new Error("Stream not found or already deleted");
        }
        const updated = await this.getById(streamId, meta.userId);
        if (!updated) {
            throw new Error("Stream disappeared after update");
        }
        return updated;
    }
    async softDelete(streamId, meta) {
        const result = await storage_1.SqliteDb.getDB()
            .updateTable("streams")
            .set({
            deleted_at: meta.now,
            updated_at: meta.now,
        })
            .where("id", "=", streamId)
            .where("user_id", "=", meta.userId)
            .where("deleted_at", "is", null)
            .executeTakeFirst();
        if (result.numUpdatedRows === BigInt(0)) {
            throw new Error("Stream not found or already deleted");
        }
    }
}
exports.SqliteStreamRepository = SqliteStreamRepository;
