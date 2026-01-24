"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.OpsRoutes = void 0;
/**
 * Ops HTTP APIs
 * - Read-only operation log
 */
class OpsRoutes {
    app;
    ctx;
    constructor(app, ctx) {
        this.app = app;
        this.ctx = ctx;
    }
    register() {
        this.registerQueries();
    }
    // ---------- Queries ----------
    registerQueries() {
        /**
         * List ops with optional filters
         * GET /ops?entity=&entityId=&limit=
         */
        this.app.get("/ops", async (req) => {
            const { entity, entityId, limit } = req.query;
            return this.ctx.ops.listByEntity(entity, entityId, this.ctx.userId, limit);
        });
        /**
         * Get latest ops
         * GET /ops/latest?limit=
         */
        this.app.get("/ops/latest", async (req) => {
            const { limit } = req.query;
            return this.ctx.ops.latest(limit ? Number(limit) : 50);
        });
        /**
         * Get ops since a timestamp or cursor
         * GET /ops/since?since=
         */
        this.app.get("/ops/since", async (req) => {
            const { since } = req.query;
            return this.ctx.ops.listSince(this.ctx.userId, since);
        });
    }
}
exports.OpsRoutes = OpsRoutes;
