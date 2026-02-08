"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.createCommandContext = createCommandContext;
const health_1 = require("./health");
const repository_1 = require("./repository");
const storage_1 = require("./storage");
async function createCommandContext(input) {
    return {
        userId: input.userId,
        deviceId: input.deviceId,
        now: input.now,
        repos: new repository_1.SqliteRepoRepository(),
        issues: new repository_1.SqliteIssueRepository(),
        sessions: new repository_1.SessionRepository(),
        stash: new repository_1.StashRepository(),
        ops: new repository_1.SqliteOpRepository(),
        streams: new repository_1.SqliteStreamRepository(),
        health: new health_1.HealthService({
            dbPing: storage_1.dbPing,
        }),
        coreSettings: new repository_1.CoreSettingsRepository(),
        sessionSegments: new repository_1.SessionSegmentRepository(),
        activeContext: new repository_1.ActiveContextRepository(),
        scratchPads: new repository_1.ScratchRepo(),
        events: input.events,
        authToken: crypto.randomUUID(),
    };
}
