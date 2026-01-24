"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.HealthRoutes = void 0;
class HealthRoutes {
    app;
    ctx;
    constructor(app, ctx) {
        this.app = app;
        this.ctx = ctx;
    }
    register() {
        this.app.get("/health", async () => {
            const health = await this.ctx.health.check();
            return {
                status: health.db ? "ok" : "degraded",
                db: health.db,
                ok: health.db ? 1 : 0,
                uptime: process.uptime(),
            };
        });
    }
}
exports.HealthRoutes = HealthRoutes;
