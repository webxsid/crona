"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.SqliteDb = void 0;
const better_sqlite3_1 = __importDefault(require("better-sqlite3"));
const kysely_1 = require("kysely");
const path_1 = __importDefault(require("path"));
const os_1 = __importDefault(require("os"));
class _SqliteDb {
    // singleton instance
    static instance;
    constructor() { }
    static getInstance() {
        return this.instance || (this.instance = new this());
    }
    _db = null;
    initDb(dbPath) {
        if (this._db) {
            throw new Error("Database is already initialized");
        }
        // db path will only be used for testing purposes
        console.debug("Initializing SQLite DB at path:", dbPath || this._defaultDbPath);
        if (dbPath) {
            const testSqlite = new better_sqlite3_1.default(dbPath);
            this._db = new kysely_1.Kysely({
                dialect: new kysely_1.SqliteDialect({
                    database: testSqlite,
                }),
            });
        }
        else {
            const defaultSqlite = new better_sqlite3_1.default(this._defaultDbPath);
            this._db = new kysely_1.Kysely({
                dialect: new kysely_1.SqliteDialect({
                    database: defaultSqlite,
                }),
            });
        }
    }
    get _defaultDbPath() {
        return path_1.default.join(os_1.default.homedir(), ".crona", "crona.db");
    }
    getDB() {
        if (!this._db) {
            throw new Error("Database is not initialized");
        }
        return this._db;
    }
}
exports.SqliteDb = _SqliteDb.getInstance();
