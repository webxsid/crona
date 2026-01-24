"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.TimerRoutes = void 0;
const core_1 = require("@crona/core");
/**
 * Timer HTTP APIs
 * Read-only derived timer state
 */
class TimerRoutes {
    app;
    _ctx;
    timer;
    constructor(app, _ctx) {
        this.app = app;
        this._ctx = _ctx;
        this.timer = new core_1.TimerService(_ctx);
    }
    register() {
        this.registerQueries();
        this.registerCommands();
    }
    // ---------- Queries ----------
    registerQueries() {
        /**
         * Get current timer state
         * GET /timer/state
         */
        this.app.get("/timer/state", async () => {
            return this.timer.getState();
        });
    }
    // ---------- Commands ----------
    registerCommands() {
        /**
         * Start timer
         * POST /timer/start?issueId=...
         * issueId optional (falls back to active context)
         */
        this.app.post("/timer/start", async (req) => {
            const { issueId } = req.query;
            return this.timer.start(issueId);
        });
        /**
         * Pause timer (work → rest)
         * POST /timer/pause
         */
        this.app.post("/timer/pause", async () => {
            return this.timer.pause();
        });
        /**
         * Resume timer (rest → work)
         * POST /timer/resume
         */
        this.app.post("/timer/resume", async () => {
            return this.timer.resume();
        });
        /**
         * End timer (stop session)
         * POST /timer/end
         */
        this.app.post("/timer/end", async (req) => {
            const body = req.body;
            return this.timer.end(body?.commitMessage);
        });
    }
}
exports.TimerRoutes = TimerRoutes;
