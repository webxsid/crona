import type { ICommandContext } from "@crona/core";
import type { FastifyInstance } from "fastify";

export class HealthRoutes {
  constructor(
    private readonly app: FastifyInstance,
    private readonly ctx: ICommandContext
  ) { }

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

