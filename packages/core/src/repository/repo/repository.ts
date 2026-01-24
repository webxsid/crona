import type { Repo } from "../../domain";
import { SqliteDb } from "../../storage";
import type { IRepoRepository } from "./interface";

export class SqliteRepoRepository implements IRepoRepository {
  async create(
    repo: Repo,
    meta: { userId: string; now: string }
  ): Promise<Repo> {
    await SqliteDb.getDB().insertInto("repos")
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

  async getById(repoId: string, userId: string): Promise<Repo | null> {
    const row = await SqliteDb.getDB()
      .selectFrom("repos")
      .select(["id", "name", "color"])
      .where("id", "=", repoId)
      .where("user_id", "=", userId)
      .where("deleted_at", "is", null)
      .executeTakeFirst();

    if (!row) return null;

    return {
      id: row.id,
      name: row.name,
      color: row.color ?? undefined,
    };
  }

  async list(userId: string): Promise<Repo[]> {
    const rows = await SqliteDb.getDB()
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

  async update(
    repoId: string,
    updates: { name?: string; color?: string },
    meta: { userId: string; deviceId: string; now: string }
  ): Promise<Repo> {
    const result = await SqliteDb.getDB()
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

  async softDelete(
    repoId: string,
    meta: { userId: string; now: string }
  ): Promise<void> {
    const result = await SqliteDb.getDB()
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
