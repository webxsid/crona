import type { FastifyInstance } from "fastify";
import type { ICommandContext } from "@crona/core";
import {
  createRepo,
  updateRepo,
  deleteRepo,
} from "@crona/core";

/**
 * Repo HTTP APIs
 * - Queries
 * - Commands
 */
export class RepoRoutes {
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
    this.app.get("/repos", async () => {
      return this.ctx.repos.list(this.ctx.userId);
    });
  }

  // ---------- Commands ----------

  private registerCommands() {
    this.app.post("/commands/repo", async (req) => {
      const { name } = req.body as {
        name: string;
      };

      return createRepo(this.ctx, { name });
    });

    this.app.put("/commands/repo/:id", async (req) => {
      const { id } = req.params as { id: string };
      const { name } = req.body as {
        id: string;
        name: string;
      };

      return updateRepo(this.ctx, id, {
        name,
      });
    });

    this.app.delete("/commands/repo/:id", async (req) => {
      const { id } = req.params as { id: string };

      await deleteRepo(this.ctx, id);

      return { ok: true };
    });
  }
}
