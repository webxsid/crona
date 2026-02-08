"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.ScratchRoutes = void 0;
const core_1 = require("@crona/core");
class ScratchRoutes {
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
    registerQueries() {
        // GET /scratchpads
        this.app.get("/scratchpads", async (req) => {
            const pinnedOnly = req.query.pinnedOnly === "1";
            return (0, core_1.listScratchpads)(this.ctx, {
                pinnedOnly
            });
        });
    }
    registerCommands() {
        // POST /scratchpads/register
        this.app.post("/scratchpads/register", async (req) => {
            const body = req.body;
            await (0, core_1.registerScratchpad)(this.ctx, body);
            return { ok: true };
        });
        // PUT /scratchpads/pin?path=...  body: { pinned: boolean }
        this.app.put("/scratchpads/pin", async (req) => {
            const { path } = req.query;
            const { pinned } = req.body;
            await (0, core_1.pinScratchpad)(this.ctx, path, pinned);
            return { ok: true };
        });
        // DELETE /scratchpads?path=...
        this.app.delete("/scratchpads", async (req) => {
            const { path } = req.query;
            await (0, core_1.removeScratchpad)(this.ctx, path);
            return { ok: true };
        });
    }
}
exports.ScratchRoutes = ScratchRoutes;
