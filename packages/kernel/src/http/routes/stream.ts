import type { FastifyInstance } from "fastify";
import type { ICommandContext } from "@crona/core";
import {
  createStream,
  updateStream,
  deleteStream,
} from "@crona/core";

/**
 * Stream HTTP APIs
 * - Queries
 * - Commands
 */
export class StreamRoutes {
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
    this.app.get("/streams", async (req) => {
      const { repoId } = req.query as { repoId: string };

      return this.ctx.streams.listByRepo(repoId, this.ctx.userId);
    });
  }

  // ---------- Commands ----------

  private registerCommands() {
    this.app.post("/stream", async (req) => {
      const { repoId, name } = req.body as {
        repoId: string;
        name: string;
      };

      return createStream(this.ctx, {
        repoId,
        name,
      });
    });

    this.app.put("/stream/:id", async (req) => {
      const { id } = req.params as { id: string };
      const { name } = req.body as {
        name: string;
      };

      return updateStream(this.ctx, id, {
        name,
      });
    });

    this.app.delete("/stream/:id", async (req) => {
      const { id } = req.params as { id: string };

      await deleteStream(this.ctx, id);

      return { ok: true };
    });
  }
}
