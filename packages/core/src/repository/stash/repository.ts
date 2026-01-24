import type { Stash } from "../../domain";
import type { StashTable } from "../../storage";
import { SqliteDb } from "../../storage";
import type { IStashRepository } from "./interface";

export class StashRepository implements IStashRepository {
  async list(userId: string): Promise<Stash[]> {
    const rows = await SqliteDb.getDB()
      .selectFrom("stash")
      .selectAll()
      .where("user_id", "=", userId)
      .orderBy("created_at", "desc")
      .execute();

    return rows.map(this.map);
  }

  async get(id: string, userId: string): Promise<Stash | null> {
    const row = await SqliteDb.getDB()
      .selectFrom("stash")
      .selectAll()
      .where("id", "=", id)
      .where("user_id", "=", userId)
      .executeTakeFirst();

    return row ? this.map(row) : null;
  }

  async save(stash: Stash): Promise<void> {
    await SqliteDb.getDB()
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

  async delete(id: string, userId: string): Promise<void> {
    await SqliteDb.getDB()
      .deleteFrom("stash")
      .where("id", "=", id)
      .where("user_id", "=", userId)
      .execute();
  }

  private map(row: StashTable): Stash {
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
