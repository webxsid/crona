import type { FastifyInstance, FastifyRequest } from "fastify";
import type { ICommandContext, IssueStatus } from "@crona/core";
import {
  createIssue,
  updateIssue,
  deleteIssue,
  changeIssueStatus,
  markIssueTodoForDate,
  markIssueTodoForToday,
  clearIssueTodoForDate,
  computeDailyIssueSummaryForDate,
  computeDailyIssueSummaryForToday,
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
    this.app.get("/issues", async (req: FastifyRequest) => {
      const { streamId } = req.query as { streamId: string };

      return this.ctx.issues.listByStream(streamId, this.ctx.userId);
    });

    this.app.get("/issues/all", async () => {
      return this.ctx.issues.listAll(this.ctx.userId);
    });

    this.app.get("/issues/summary/today", async () => {
      const plan = await computeDailyIssueSummaryForToday(this.ctx);

      return plan;
    });

    this.app.get("/issues/summary/daily", async (req: FastifyRequest) => {
      const { date } = req.query as { date?: string };
      const requestedDate = date || this.ctx.now().split("T")[0] || "";
      return computeDailyIssueSummaryForDate(this.ctx, requestedDate);
    });
  }

  // ---------- Commands ----------

  private registerCommands() {
    this.app.post("/issue", async (req: FastifyRequest) => {
      const { streamId, title, estimateMinutes, todoForDate } = req.body as {
        streamId: string;
        title: string;
        estimateMinutes?: number;
        todoForDate?: string;
      };

      return createIssue(this.ctx, {
        streamId,
        title,
        estimateMinutes: Number.isNaN(estimateMinutes) ? undefined : estimateMinutes,
        todoForDate,
      });
    });

    this.app.put("/issue/:id", async (req: FastifyRequest) => {
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

    this.app.delete("/issue/:id", async (req: FastifyRequest) => {
      const { id } = req.params as { id: string };

      await deleteIssue(this.ctx, id);

      return { ok: true };
    });

    this.app.put("/issue/:id/status", async (req: FastifyRequest) => {
      const { id } = req.params as { id: string };
      const { status } = req.body as {
        status: IssueStatus
      };

      return changeIssueStatus(this.ctx, id, status);
    })

    this.app.put("/issue/:id/todo", async (req: FastifyRequest) => {
      const { id } = req.params as { id: string };
      const { date } = (req.body as { date?: string }) ?? {};

      const issue = date
        ? await markIssueTodoForDate(this.ctx, id, date)
        : await markIssueTodoForToday(this.ctx, id);

      return issue;
    });

    this.app.put("/issue/:id/todo/clear", async (req: FastifyRequest) => {
      const { id } = req.params as { id: string };

      const issue = await clearIssueTodoForDate(this.ctx, id);

      return issue;
    });


  }
}
