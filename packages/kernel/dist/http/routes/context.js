"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.ContextRoutes = void 0;
const core_1 = require("@crona/core");
/**
 * Active context HTTP APIs
 * Git-like working context:
 * repo → stream → issue
 */
class ContextRoutes {
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
         * Get current active context
         * GET /context
         */
        this.app.get("/context", async () => {
            return (0, core_1.getActiveContext)(this.ctx);
        });
    }
    // ---------- Commands ----------
    registerCommands() {
        /**
         * Edit Active context object
         * PUT /context
         * Body: { repoId?, streamId?, issueId? }
         * If repoId changes, streamId and issueId are cleared
         * If streamId changes, issueId is cleared
         */
        this.app.put("/context", async (req) => {
            const { repoId, streamId, issueId } = req.body;
            if (repoId) {
                await (0, core_1.switchRepo)(this.ctx, repoId);
            }
            if (streamId) {
                await (0, core_1.switchStream)(this.ctx, streamId);
            }
            if (issueId) {
                await (0, core_1.switchIssue)(this.ctx, issueId);
            }
            return (0, core_1.getActiveContext)(this.ctx);
        });
        /**
         * Switch repo
         * PUT /context/repo?repoId=...
         * Clears stream + issue
         */
        this.app.put("/context/repo", async (req) => {
            const { repoId } = req.query;
            if (!repoId) {
                throw new Error("repoId is required");
            }
            return (0, core_1.switchRepo)(this.ctx, repoId);
        });
        /**
         * Switch stream (within current repo)
         * PUT /context/stream?streamId=...
         * Clears issue
         */
        this.app.put("/context/stream", async (req) => {
            const { streamId } = req.query;
            if (!streamId) {
                throw new Error("streamId is required");
            }
            return (0, core_1.switchStream)(this.ctx, streamId);
        });
        /**
         * Switch issue (within current stream)
         * PUT /context/issue?issueId=...
         */
        this.app.put("/context/issue", async (req) => {
            const { issueId } = req.query;
            if (!issueId) {
                throw new Error("issueId is required");
            }
            return (0, core_1.switchIssue)(this.ctx, issueId);
        });
        /**
         * Clear active issue
         * DELETE /context/issue
         */
        this.app.delete("/context/issue", async () => {
            return (0, core_1.clearIssue)(this.ctx);
        });
        /**
         * Clear entire context
         * DELETE /context
         */
        this.app.delete("/context", async () => {
            await (0, core_1.clearContext)(this.ctx);
            return { ok: true };
        });
    }
}
exports.ContextRoutes = ContextRoutes;
