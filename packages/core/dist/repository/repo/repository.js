"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.SqliteRepoRepository = void 0;
const storage_1 = require("../../storage");
class SqliteRepoRepository {
    async create(repo, meta) {
        await storage_1.SqliteDb.getDB().insertInto("repos")
            .values({
            id: repo.id,
            name: repo.name,
            color: repo.color ?? null,
            user_id: meta.userId,
            created_at: meta.now,
            updated_at: meta.now,
            deleted_at: null,
        })
            .execute();
        return repo;
    }
    async getById(repoId, userId) {
        const row = await storage_1.SqliteDb.getDB()
            .selectFrom("repos")
            .select(["id", "name", "color"])
            .where("id", "=", repoId)
            .where("user_id", "=", userId)
            .where("deleted_at", "is", null)
            .executeTakeFirst();
        if (!row)
            return null;
        return {
            id: row.id,
            name: row.name,
            color: row.color ?? undefined,
        };
    }
    async list(userId) {
        const rows = await storage_1.SqliteDb.getDB()
            .selectFrom("repos")
            .select(["id", "name", "color"])
            .where("user_id", "=", userId)
            .where("deleted_at", "is", null)
            .orderBy("created_at", "asc")
            .execute();
        return rows.map((r) => ({
            id: r.id,
            name: r.name,
            color: r.color ?? undefined,
        }));
    }
    async update(repoId, updates, meta) {
        const result = await storage_1.SqliteDb.getDB()
            .updateTable("repos")
            .set({
            name: updates.name,
            color: updates.color ?? null,
            updated_at: meta.now,
        })
            .where("id", "=", repoId)
            .where("user_id", "=", meta.userId)
            .where("deleted_at", "is", null)
            .executeTakeFirst();
        if (result.numUpdatedRows === BigInt(0)) {
            throw new Error("Repo not found or already deleted");
        }
        const updated = await this.getById(repoId, meta.userId);
        if (!updated || updated === null) {
            throw new Error("Repo disappeared after update");
        }
        return updated;
    }
    async softDelete(repoId, meta) {
        const result = await storage_1.SqliteDb.getDB()
            .updateTable("repos")
            .set({
            deleted_at: meta.now,
            updated_at: meta.now,
        })
            .where("id", "=", repoId)
            .where("user_id", "=", meta.userId)
            .where("deleted_at", "is", null)
            .executeTakeFirst();
        if (result.numUpdatedRows === BigInt(0)) {
            throw new Error("Repo not found or already deleted");
        }
    }
}
exports.SqliteRepoRepository = SqliteRepoRepository;
