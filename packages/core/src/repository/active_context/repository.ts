import type { ActiveContext } from "../../domain";
import { SqliteDb } from "../../storage";
import type { IActiveContextRepository } from "./interface";

export class ActiveContextRepository
  implements IActiveContextRepository {
  async get(userId: string, deviceId: string): Promise<ActiveContext | null> {
    const row = await SqliteDb.getDB()
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

  async set(
    userId: string,
    deviceId: string,
    context: {
      repoId?: string;
      streamId?: string;
      issueId?: string;
    }
  ): Promise<ActiveContext> {
    const now = new Date().toISOString();

    await SqliteDb.getDB()
      .insertInto("active_context")
      .values({
        user_id: userId,
        device_id: deviceId,
        repo_id: context.repoId ?? null,
        stream_id: context.streamId ?? null,
        issue_id: context.issueId ?? null,
        updated_at: now,
      })
      .onConflict((oc) =>
        oc.column("user_id").doUpdateSet({
          repo_id: context.repoId ?? null,
          stream_id: context.streamId ?? null,
          issue_id: context.issueId ?? null,
          updated_at: now,
        })
      )
      .execute();

    return (await this.get(userId, deviceId))!;
  }

  async clear(userId: string, deviceId: string): Promise<void> {
    await SqliteDb.getDB()
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

  async initializeDefaults(userId: string, deviceId: string): Promise<void> {
    const existing = await this.get(userId, deviceId);
    if (existing) {
      return;
    }

    await SqliteDb.getDB()
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
