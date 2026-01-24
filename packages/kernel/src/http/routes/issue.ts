import type { FastifyInstance } from "fastify";
import type { ICommandContext, IssueStatus } from "@crona/core";
import {
  createIssue,
  updateIssue,
  deleteIssue,
  changeIssueStatus,
} from "@crona/core";

/**
 * Issue HTTP APIs
 * - Queries
 * - Commands
 */
export class IssueRoutes {
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
    this.app.get("/issues", async (req) => {
      const { streamId } = req.query as { streamId: string };

      return this.ctx.issues.listByStream(streamId, this.ctx.userId);
    });
  }

  // ---------- Commands ----------

  private registerCommands() {
    this.app.post("/issue", async (req) => {
      const { streamId, title, estimateMinutes } = req.body as {
        streamId: string;
        title: string;
        estimateMinutes?: number;
      };

      return createIssue(this.ctx, {
        streamId,
        title,
        estimateMinutes: Number.isNaN(estimateMinutes) ? undefined : estimateMinutes
      });
    });

    this.app.put("/issue/:id", async (req) => {
      const { id } = req.params as { id: string };
      const { title, estimateMinutes, notes } = req.body as {
        title?: string;
        estimateMinutes?: number;
        notes?: string;
      };

      return updateIssue(this.ctx, id, {
        title,
        estimateMinutes: Number.isNaN(estimateMinutes) ? undefined : estimateMinutes,
        notes
      });
    });

    this.app.delete("/issue/:id", async (req) => {
      const { id } = req.params as { id: string };

      await deleteIssue(this.ctx, id);

      return { ok: true };
    });

    this.app.put("/issue/:id/status", async (req) => {
      const { id } = req.params as { id: string };
      const { status } = req.body as {
        status: IssueStatus
      };

      return changeIssueStatus(this.ctx, id, status);
    })
  }
}
