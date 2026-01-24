import type { FastifyInstance } from "fastify";
import type { ICommandContext } from "@crona/core";
import { TimerService } from "@crona/core";

/**
 * Timer HTTP APIs
 * Read-only derived timer state
 */
export class TimerRoutes {
  private readonly timer: TimerService;

  constructor(
    private readonly app: FastifyInstance,
    private readonly _ctx: ICommandContext
  ) {
    this.timer = new TimerService(_ctx);
  }

  register() {
    this.registerQueries();
    this.registerCommands();
  }

  // ---------- Queries ----------

  private registerQueries() {
    /**
     * Get current timer state
     * GET /timer/state
     */
    this.app.get("/timer/state", async () => {
      return this.timer.getState();
    });
  }

  // ---------- Commands ----------

  private registerCommands() {
    /**
     * Start timer
     * POST /timer/start?issueId=...
     * issueId optional (falls back to active context)
     */
    this.app.post("/timer/start", async (req) => {
      const { issueId } = req.query as { issueId?: string };
      return this.timer.start(issueId);
    });

    /**
     * Pause timer (work → rest)
     * POST /timer/pause
     */
    this.app.post("/timer/pause", async () => {
      return this.timer.pause();
    });

    /**
     * Resume timer (rest → work)
     * POST /timer/resume
     */
    this.app.post("/timer/resume", async () => {
      return this.timer.resume();
    });

    /**
     * End timer (stop session)
     * POST /timer/end
     */
    this.app.post("/timer/end", async (req) => {
      const body = req.body as { commitMessage?: string };
      return this.timer.end(body?.commitMessage);
    });
  }
}
