"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.dbPing = dbPing;
const kysely_1 = require("kysely");
const db_1 = require("./db");
async function dbPing() {
    try {
        await db_1.SqliteDb.getDB()
            .selectFrom((0, kysely_1.sql) `(SELECT 1)`.as("health_check"))
            .selectAll()
            .executeTakeFirst();
        return true;
    }
    catch {
        return false;
    }
}
