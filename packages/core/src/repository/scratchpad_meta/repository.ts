import type { IScratchRepo } from "./interface";
import type { ScratchPadMeta } from "../../domain";
import { SqliteDb } from "../../storage";

export class ScratchRepo implements IScratchRepo {
  async upsert(meta: ScratchPadMeta, userMeta: {
    userId: string;
    deviceId: string;
  }): Promise<void> {
    await SqliteDb.getDB()
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
      .onConflict((oc) =>
        oc
          .column("path")
          .doUpdateSet({
            name: meta.name,
            last_opened_at: meta.lastOpenedAt.toISOString(),
            pinned: meta.pinned ? 1 : 0,
          })
          .where("user_id", "=", userMeta.userId)
      )
      .execute();
  }

  async list(
    userId: string,
    deviceId: string,
    options?: {
      pinnedOnly?: boolean;
    }
  ): Promise<ScratchPadMeta[]> {
    let query = SqliteDb.getDB()
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

  async get(path: string, meta: {
    userId: string;
    deviceId: string;
  }): Promise<ScratchPadMeta | null> {
    const row = await SqliteDb.getDB()
      .selectFrom("scratch_pad_meta")
      .selectAll()
      .where("path", "=", path)
      .where("user_id", "=", meta.userId)
      .where("device_id", "=", meta.deviceId)
      .executeTakeFirst();

    if (!row) return null;

    return {
      id: row.id,
      name: row.name,
      lastOpenedAt: new Date(row.last_opened_at),
      path: row.path,
      pinned: row.pinned === 1,
    };
  }

  async remove(path: string, meta: {
    userId: string;
    deviceId: string;
  }): Promise<void> {
    await SqliteDb.getDB()
      .deleteFrom("scratch_pad_meta")
      .where("path", "=", path)
      .where("user_id", "=", meta.userId)
      .where("device_id", "=", meta.deviceId)
      .execute();
  }
}
