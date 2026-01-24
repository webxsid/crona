"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.StashRoutes = void 0;
const core_1 = require("@crona/core");
/**
 * Stash HTTP APIs
 * Stash = suspended active_context (+ session cleanup)
 */
class StashRoutes {
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
         * List all stashes
         * GET /stash
         */
        this.app.get("/stash", async () => {
            return this.ctx.stash.list(this.ctx.userId);
        });
        /**
         * Get a specific stash
         * GET /stash/:id
         */
        this.app.get("/stash/:id", async (req) => {
            const { id } = req.params;
            return this.ctx.stash.get(id, this.ctx.userId);
        });
    }
    // ---------- Commands ----------
    registerCommands() {
        /**
         * Create stash from current active_context
         * POST /stash
         */
        this.app.post("/stash", async (req) => {
            const body = req.body;
            return (0, core_1.stashPush)(this.ctx, body.stashNote);
        });
        /**
         * Apply stash (restore context, delete stash)
         * POST /stash/:id/apply
         */
        this.app.post("/stash/:id/apply", async (req) => {
            const { id } = req.params;
            await (0, core_1.stashPop)(this.ctx, id);
            return { ok: true };
        });
        /**
         * Drop stash (delete without applying)
         * DELETE /stash/:id
         */
        this.app.delete("/stash/:id", async (req) => {
            const { id } = req.params;
            await (0, core_1.stashDrop)(this.ctx, id);
            return { ok: true };
        });
    }
}
exports.StashRoutes = StashRoutes;
