"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.SessionRoutes = void 0;
const core_1 = require("@crona/core");
/**
 * Session HTTP APIs
 * Owns session lifecycle + segment transitions
 */
class SessionRoutes {
    app;
    ctx;
    constructor(app, ctx) {
        this.app = app;
        this.ctx = ctx;
    }
    register() {
        this.registerQueries();
        this.registerCommands();
    }
    // ---------- Queries ----------
    registerQueries() {
        /**
         * List sessions for an issue
         * GET /sessions?issueId=...
         */
        this.app.get("/sessions", async (req) => {
            const { issueId } = req.query;
            if (!issueId) {
                throw new Error("issueId is required");
            }
            return this.ctx.sessions.listByIssue(issueId, this.ctx.userId);
        });
        /**
         * Get session by ID
         * GET /sessions/:sessionId
         */
        this.app.get("/sessions/:sessionId", async (req) => {
            const { sessionId } = req.params;
            return this.ctx.sessions.getSessiobById(sessionId, this.ctx.userId);
        });
        /**
         * Get active session (if any)
         * GET /sessions/active
         */
        this.app.get("/sessions/active", async () => {
            return this.ctx.sessions.getActiveSession(this.ctx.userId);
        });
    }
    // ---------- Commands ----------
    registerCommands() {
        /**
         * Start a session for an issue
         * POST /sessions/start?issueId=...
         */
        this.app.post("/sessions/start", async (req) => {
            const { issueId } = req.query;
            if (!issueId) {
                throw new Error("issueId is required");
            }
            return (0, core_1.startSession)(this.ctx, issueId);
        });
        /**
         * Pause current session (work → rest segment)
         * POST /sessions/pause
         */
        this.app.post("/sessions/pause", async () => {
            const active = await this.ctx.sessions.getActiveSession(this.ctx.userId);
            if (!active)
                return { ok: true };
            await (0, core_1.pauseSession)(this.ctx);
            return { ok: true };
        });
        /**
         * Resume session (rest → work segment)
         * POST /sessions/resume
         */
        this.app.post("/sessions/resume", async () => {
            const active = await this.ctx.sessions.getActiveSession(this.ctx.userId);
            if (!active)
                return { ok: true };
            await (0, core_1.resumeSession)(this.ctx);
            return { ok: true };
        });
        /**
         * End the active session
         * POST /sessions/end
         */
        this.app.post("/sessions/end", async (req) => {
            const active = await this.ctx.sessions.getActiveSession(this.ctx.userId);
            if (!active)
                return { ok: true };
            const body = req.body;
            return (0, core_1.stopSession)(this.ctx, body.commitMessage);
        });
        /**
         * Ammend a session commit message
         * PATCH /sessions/note?id=...
         * body: { note: string }
         */
        this.app.patch("/sessions/note", async (req) => {
            const { id } = req.query;
            const body = req.body;
            return (0, core_1.ammendSessionNotes)(this.ctx, body.note, id);
        });
        /**
         * List Session History
         * GET /sessions/history?repoId=...
         */
        this.app.get("/sessions/history", async (req) => {
            const { repoId, streamId, issueId, since, until, limit, offset, context, } = req.query;
            return await (0, core_1.listSessionHistory)(this.ctx, {
                repoId,
                streamId,
                issueId,
                since,
                until,
                limit,
                offset,
            }, context === "1");
        });
    }
}
exports.SessionRoutes = SessionRoutes;
