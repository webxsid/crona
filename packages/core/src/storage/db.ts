import Database from "better-sqlite3";
import { Kysely, SqliteDialect } from "kysely";
import type { DB } from "./schema";
import path from "path";
import os from "os";


class _SqliteDb {

  // singleton instance
  private static instance: _SqliteDb;

  private constructor() { }

  public static getInstance() {
    return this.instance || (this.instance = new this());
  }

  private _db: Kysely<DB> | null = null;


  public initDb(dbPath?: string) {
    if (this._db) {
      throw new Error("Database is already initialized");
    }
    // db path will only be used for testing purposes
    console.debug("Initializing SQLite DB at path:", dbPath || this._defaultDbPath);
    if (dbPath) {
      const testSqlite = new Database(dbPath);
      this._db = new Kysely<DB>({
        dialect: new SqliteDialect({
          database: testSqlite,
        }),
      });
    } else {
      const defaultSqlite = new Database(this._defaultDbPath);
      this._db = new Kysely<DB>({
        dialect: new SqliteDialect({
          database: defaultSqlite,
        }),
      });
    }
  }

  private get _defaultDbPath() {
    return path.join(os.homedir(), ".crona", "crona.db");
  }

  public getDB(): Kysely<DB> {
    if (!this._db) {
      throw new Error("Database is not initialized");
    }
    return this._db;
  }

}

export const SqliteDb = _SqliteDb.getInstance();
export type Sq = Kysely<DB>;
