import type { FastifyInstance } from "fastify";
import type { ICommandContext } from "@crona/core";
import {
  getActiveContext,
  switchRepo,
  switchStream,
  switchIssue,
  clearIssue,
  clearContext,
} from "@crona/core";

/**
 * Active context HTTP APIs
 * Git-like working context:
 * repo → stream → issue
 */
export class ContextRoutes {
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
     * Get current active context
     * GET /context
     */
    this.app.get("/context", async () => {
      return getActiveContext(this.ctx);
    });
  }

  // ---------- Commands ----------

  private registerCommands() {
    /**
     * Edit Active context object
     * PUT /context
     * Body: { repoId?, streamId?, issueId? }
     * If repoId changes, streamId and issueId are cleared
     * If streamId changes, issueId is cleared
     */
    this.app.put("/context", async (req) => {
      const { repoId, streamId, issueId } = req.body as {
        repoId?: string;
        streamId?: string;
        issueId?: string;
      };

      if (repoId) {
        await switchRepo(this.ctx, repoId);
      }

      if (streamId) {
        await switchStream(this.ctx, streamId);
      }

      if (issueId) {
        await switchIssue(this.ctx, issueId);
      }

      return getActiveContext(this.ctx);
    });

    /**
     * Switch repo
     * PUT /context/repo?repoId=...
     * Clears stream + issue
     */
    this.app.put("/context/repo", async (req) => {
      const { repoId } = req.query as { repoId?: string };
      if (!repoId) {
        throw new Error("repoId is required");
      }

      return switchRepo(this.ctx, repoId);
    });

    /**
     * Switch stream (within current repo)
     * PUT /context/stream?streamId=...
     * Clears issue
     */
    this.app.put("/context/stream", async (req) => {
      const { streamId } = req.query as { streamId?: string };
      if (!streamId) {
        throw new Error("streamId is required");
      }

      return switchStream(this.ctx, streamId);
    });

    /**
     * Switch issue (within current stream)
     * PUT /context/issue?issueId=...
     */
    this.app.put("/context/issue", async (req) => {
      const { issueId } = req.query as { issueId?: string };
      if (!issueId) {
        throw new Error("issueId is required");
      }

      return switchIssue(this.ctx, issueId);
    });

    /**
     * Clear active issue
     * DELETE /context/issue
     */
    this.app.delete("/context/issue", async () => {
      return clearIssue(this.ctx);
    });

    /**
     * Clear entire context
     * DELETE /context
     */
    this.app.delete("/context", async () => {
      await clearContext(this.ctx);
      return { ok: true };
    });
  }
}
