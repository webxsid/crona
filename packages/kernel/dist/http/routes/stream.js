"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.StreamRoutes = void 0;
const core_1 = require("@crona/core");
/**
 * Stream HTTP APIs
 * - Queries
 * - Commands
 */
class StreamRoutes {
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
        this.app.get("/streams", async (req) => {
            const { repoId } = req.query;
            return this.ctx.streams.listByRepo(repoId, this.ctx.userId);
        });
    }
    // ---------- Commands ----------
    registerCommands() {
        this.app.post("/stream", async (req) => {
            const { repoId, name } = req.body;
            return (0, core_1.createStream)(this.ctx, {
                repoId,
                name,
            });
        });
        this.app.put("/stream/:id", async (req) => {
            const { id } = req.params;
            const { name } = req.body;
            return (0, core_1.updateStream)(this.ctx, id, {
                name,
            });
        });
        this.app.delete("/stream/:id", async (req) => {
            const { id } = req.params;
            await (0, core_1.deleteStream)(this.ctx, id);
            return { ok: true };
        });
    }
}
exports.StreamRoutes = StreamRoutes;
