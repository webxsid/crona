"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.IssueRoutes = void 0;
const core_1 = require("@crona/core");
/**
 * Issue HTTP APIs
 * - Queries
 * - Commands
 */
class IssueRoutes {
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
        this.app.get("/issues", async (req) => {
            const { streamId } = req.query;
            return this.ctx.issues.listByStream(streamId, this.ctx.userId);
        });
    }
    // ---------- Commands ----------
    registerCommands() {
        this.app.post("/issue", async (req) => {
            const { streamId, title, estimateMinutes } = req.body;
            return (0, core_1.createIssue)(this.ctx, {
                streamId,
                title,
                estimateMinutes: Number.isNaN(estimateMinutes) ? undefined : estimateMinutes
            });
        });
        this.app.put("/issue/:id", async (req) => {
            const { id } = req.params;
            const { title, estimateMinutes, notes } = req.body;
            return (0, core_1.updateIssue)(this.ctx, id, {
                title,
                estimateMinutes: Number.isNaN(estimateMinutes) ? undefined : estimateMinutes,
                notes
            });
        });
        this.app.delete("/issue/:id", async (req) => {
            const { id } = req.params;
            await (0, core_1.deleteIssue)(this.ctx, id);
            return { ok: true };
        });
        this.app.put("/issue/:id/status", async (req) => {
            const { id } = req.params;
            const { status } = req.body;
            return (0, core_1.changeIssueStatus)(this.ctx, id, status);
        });
    }
}
exports.IssueRoutes = IssueRoutes;
