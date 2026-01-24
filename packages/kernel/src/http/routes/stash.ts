import type { FastifyInstance } from "fastify";
import type { ICommandContext } from "@crona/core";
import {
  stashPush,
  stashPop,
  stashDrop,
} from "@crona/core";

/**
 * Stash HTTP APIs
 * Stash = suspended active_context (+ session cleanup)
 */
export class StashRoutes {
  constructor(
    private readonly app: FastifyInstance,
    private readonly ctx: ICommandContext
  ) { }

  register() {
    this.registerQueries();
    this.registerCommands();
  }

  // ---------- Queries ----------

  private registerQueries() {
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
      const { id } = req.params as { id: string };
      return this.ctx.stash.get(id, this.ctx.userId);
    });
  }

  // ---------- Commands ----------

  private registerCommands() {
    /**
     * Create stash from current active_context
     * POST /stash
     */
    this.app.post("/stash", async (req) => {
      const body = req.body as { stashNote: string };
      return stashPush(this.ctx, body.stashNote);
    });

    /**
     * Apply stash (restore context, delete stash)
     * POST /stash/:id/apply
     */
    this.app.post("/stash/:id/apply", async (req) => {
      const { id } = req.params as { id: string };
      await stashPop(this.ctx, id);
      return { ok: true };
    });

    /**
     * Drop stash (delete without applying)
     * DELETE /stash/:id
     */
    this.app.delete("/stash/:id", async (req) => {
      const { id } = req.params as { id: string };
      await stashDrop(this.ctx, id);
      return { ok: true };
    });
  }
}
