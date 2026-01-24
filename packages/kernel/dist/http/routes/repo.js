"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.RepoRoutes = void 0;
const core_1 = require("@crona/core");
/**
 * Repo HTTP APIs
 * - Queries
 * - Commands
 */
class RepoRoutes {
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
        this.app.get("/repos", async () => {
            return this.ctx.repos.list(this.ctx.userId);
        });
    }
    // ---------- Commands ----------
    registerCommands() {
        this.app.post("/commands/repo", async (req) => {
            const { name } = req.body;
            return (0, core_1.createRepo)(this.ctx, { name });
        });
        this.app.put("/commands/repo/:id", async (req) => {
            const { id } = req.params;
            const { name } = req.body;
            return (0, core_1.updateRepo)(this.ctx, id, {
                name,
            });
        });
        this.app.delete("/commands/repo/:id", async (req) => {
            const { id } = req.params;
            await (0, core_1.deleteRepo)(this.ctx, id);
            return { ok: true };
        });
    }
}
exports.RepoRoutes = RepoRoutes;
