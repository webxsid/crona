import { sql } from "kysely";
import { SqliteDb } from "./db";

export async function dbPing(): Promise<boolean> {
  try {
    await SqliteDb.getDB()
      .selectFrom(sql`(SELECT 1)`.as("health_check"))
      .selectAll()
      .executeTakeFirst();
    return true;
  } catch {
    return false;
  }
}

