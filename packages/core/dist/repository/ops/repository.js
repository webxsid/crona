"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.SqliteOpRepository = void 0;
const storage_1 = require("../../storage");
class SqliteOpRepository {
    async append(op) {
        await storage_1.SqliteDb.getDB()
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
    async latest(limit) {
        const rows = await storage_1.SqliteDb.getDB()
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
    async listSince(userId, sinceTimestamp) {
        const rows = await storage_1.SqliteDb.getDB()
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
    async listByEntity(entity, entityId, userId, limit) {
        const rows = await storage_1.SqliteDb.getDB()
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
    mapRow(row) {
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
exports.SqliteOpRepository = SqliteOpRepository;
