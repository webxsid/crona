import type { Op } from "../../domain";
import { SqliteDb } from "../../storage";
import type { IOpRepository } from "./interface";

export class SqliteOpRepository implements IOpRepository {
  async append(op: Op): Promise<void> {
    await SqliteDb.getDB()
      .insertInto("ops")
      .values({
        id: op.id,
        entity: op.entity,
        entity_id: op.entityId,
        action: op.action,
        payload: JSON.stringify(op.payload),
        timestamp: op.timestamp,
        user_id: op.userId,
        device_id: op.deviceId,
      })
      .execute();
  }

  async latest(limit: number): Promise<Op[]> {
    const rows = await SqliteDb.getDB()
      .selectFrom("ops")
      .select([
        "id",
        "entity",
        "entity_id",
        "action",
        "payload",
        "timestamp",
        "user_id",
        "device_id",
      ])
      .orderBy("timestamp", "desc")
      .limit(limit)
      .execute();

    return rows.map(this.mapRow).reverse();
  }

  async listSince(
    userId: string,
    sinceTimestamp: string
  ): Promise<Op[]> {
    const rows = await SqliteDb.getDB()
      .selectFrom("ops")
      .select([
        "id",
        "entity",
        "entity_id",
        "action",
        "payload",
        "timestamp",
        "user_id",
        "device_id",
      ])
      .where("user_id", "=", userId)
      .where("timestamp", ">", sinceTimestamp)
      .orderBy("timestamp", "asc")
      .execute();

    return rows.map(this.mapRow);
  }

  async listByEntity(
    entity: Op["entity"],
    entityId: string,
    userId: string,
    limit: number | undefined
  ): Promise<Op[]> {
    const rows = await SqliteDb.getDB()
      .selectFrom("ops")
      .select([
        "id",
        "entity",
        "entity_id",
        "action",
        "payload",
        "timestamp",
        "user_id",
        "device_id",
      ])
      .where("entity", "=", entity)
      .where("entity_id", "=", entityId)
      .where("user_id", "=", userId)
      .orderBy("timestamp", "asc")
      .limit(limit ?? 100)
      .execute();

    return rows.map(this.mapRow);
  }

  private mapRow(row: {
    id: string;
    entity: Op["entity"];
    entity_id: string;
    action: Op["action"];
    payload: string;
    timestamp: string;
    user_id: string;
    device_id: string;
  }): Op {
    return {
      id: row.id,
      entity: row.entity,
      entityId: row.entity_id,
      action: row.action,
      payload: JSON.parse(row.payload),
      timestamp: row.timestamp,
      userId: row.user_id,
      deviceId: row.device_id,
    };
  }
}
