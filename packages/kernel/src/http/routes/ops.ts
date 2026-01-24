import type { FastifyInstance } from "fastify";
import type { ICommandContext, OpEntity, } from "@crona/core";

/**
 * Ops HTTP APIs
 * - Read-only operation log
 */
export class OpsRoutes {
  constructor(
    private readonly app: FastifyInstance,
    private readonly ctx: ICommandContext
  ) { }

  register() {
    this.registerQueries();
  }

  // ---------- Queries ----------

  private registerQueries() {
    /**
     * List ops with optional filters
     * GET /ops?entity=&entityId=&limit=
     */
    this.app.get("/ops", async (req) => {
      const {
        entity,
        entityId,
        limit
      } = req.query as {
        entity: string;
        entityId: string;
        limit: number
      };

      return this.ctx.ops.listByEntity(
        entity as OpEntity,
        entityId,
        this.ctx.userId,
        limit
      );
    });

    /**
     * Get latest ops
     * GET /ops/latest?limit=
     */
    this.app.get("/ops/latest", async (req) => {
      const { limit } = req.query as { limit?: string };

      return this.ctx.ops.latest(
        limit ? Number(limit) : 50
      );
    });

    /**
     * Get ops since a timestamp or cursor
     * GET /ops/since?since=
     */
    this.app.get("/ops/since", async (req) => {
      const { since } = req.query as { since: string };

      return this.ctx.ops.listSince(
        this.ctx.userId,
        since);
    });
  }
}
