import type { FastifyInstance } from "fastify";
import type { ICommandContext } from "@crona/core";
import {
  startSession,
  pauseSession,
  resumeSession,
  stopSession,
  ammendSessionNotes,
  listSessionHistory,
} from "@crona/core";

/**
 * Session HTTP APIs
 * Owns session lifecycle + segment transitions
 */
export class SessionRoutes {
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
     * List sessions for an issue
     * GET /sessions?issueId=...
     */
    this.app.get("/sessions", async (req) => {
      const { issueId } = req.query as { issueId?: string };
      if (!issueId) {
        throw new Error("issueId is required");
      }

      return this.ctx.sessions.listByIssue(
        issueId,
        this.ctx.userId
      );
    });

    /**
     * Get session by ID
     * GET /sessions/:sessionId
     */
    this.app.get("/sessions/:sessionId", async (req) => {
      const { sessionId } = req.params as { sessionId: string };

      return this.ctx.sessions.getSessiobById(
        sessionId,
        this.ctx.userId
      );
    })

    /**
     * Get active session (if any)
     * GET /sessions/active
     */
    this.app.get("/sessions/active", async () => {
      return this.ctx.sessions.getActiveSession(
        this.ctx.userId
      );
    });
  }

  // ---------- Commands ----------

  private registerCommands() {
    /**
     * Start a session for an issue
     * POST /sessions/start?issueId=...
     */
    this.app.post("/sessions/start", async (req) => {
      const { issueId } = req.query as { issueId?: string };
      if (!issueId) {
        throw new Error("issueId is required");
      }

      return startSession(this.ctx, issueId);
    });

    /**
     * Pause current session (work → rest segment)
     * POST /sessions/pause
     */
    this.app.post("/sessions/pause", async () => {
      const active = await this.ctx.sessions.getActiveSession(
        this.ctx.userId
      );
      if (!active) return { ok: true };

      await pauseSession(this.ctx);

      return { ok: true };
    });

    /**
     * Resume session (rest → work segment)
     * POST /sessions/resume
     */
    this.app.post("/sessions/resume", async () => {
      const active = await this.ctx.sessions.getActiveSession(
        this.ctx.userId
      );
      if (!active) return { ok: true };
      await resumeSession(this.ctx);
      return { ok: true };
    });

    /**
     * End the active session
     * POST /sessions/end
     */
    this.app.post("/sessions/end", async (req) => {
      const active = await this.ctx.sessions.getActiveSession(
        this.ctx.userId
      );
      if (!active) return { ok: true };

      const body = req.body as { commitMessage?: string };

      return stopSession(this.ctx, body.commitMessage);
    });

    /**
     * Ammend a session commit message
     * PATCH /sessions/note?id=...
     * body: { note: string }
     */
    this.app.patch("/sessions/note", async (req) => {
      const { id } = req.query as { id?: string };

      const body = req.body as { note: string };

      return ammendSessionNotes(this.ctx, body.note, id);
    });

    /**
     * List Session History
     * GET /sessions/history?repoId=...
     */
    this.app.get("/sessions/history", async (req) => {
      const {
        repoId,
        streamId,
        issueId,
        since,
        until,
        limit,
        offset,
        context,
      } = req.query as {
        repoId?: string;
        streamId?: string;
        issueId?: string;
        since?: string;
        until?: string;
        limit?: number;
        offset?: number;
        context?: "1" | "0";
      };

      return await listSessionHistory(
        this.ctx,
        {
          repoId,
          streamId,
          issueId,
          since,
          until,
          limit,
          offset,
        },
        context === "1"
      )
    })
  }
}
