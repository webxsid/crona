import { Kysely } from "kysely";
import type { DB } from "./schema";
declare class _SqliteDb {
    private static instance;
    private constructor();
    static getInstance(): _SqliteDb;
    private _db;
    initDb(dbPath?: string): void;
    private get _defaultDbPath();
    getDB(): Kysely<DB>;
}
export declare const SqliteDb: _SqliteDb;
export type Sq = Kysely<DB>;
export {};
