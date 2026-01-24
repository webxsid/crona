"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.initDb = initDb;
const schema_1 = require("./schema");
async function initDb() {
    try {
        await (0, schema_1.initSchema)();
        console.log("Database schema initialized successfully.");
    }
    catch (error) {
        console.error("Error initializing database schema:", error);
        throw error;
    }
}
