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
import { logInfo, logError } from "../../logger";

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
     * Get current active context with resolved titles
     * GET /context
     */
    this.app.get("/context", async () => {
      const context = await getActiveContext(this.ctx);
      if (!context) return null;

      const [repo, stream, issue] = await Promise.all([
        context.repoId ? this.ctx.repos.getById(context.repoId, this.ctx.userId) : Promise.resolve(null),
        context.streamId ? this.ctx.streams.getById(context.streamId, this.ctx.userId) : Promise.resolve(null),
        context.issueId ? this.ctx.issues.getById(context.issueId, this.ctx.userId) : Promise.resolve(null),
      ]);

      return {
        ...context,
        repoName: repo?.name,
        streamName: stream?.name,
        issueTitle: issue?.title,
      };
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
      const body = req.body as { repoId?: string; streamId?: string; issueId?: string };
      logInfo(`[context] PUT /context body=${JSON.stringify(body)}`);
      const { repoId, streamId, issueId } = body;

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
     * PUT /context/repo
     * Body: { repoId }
     * Clears stream + issue
     */
    this.app.put("/context/repo", async (req) => {
      const body = req.body as { repoId?: string };
      logInfo(`[context] PUT /context/repo body=${JSON.stringify(body)} headers=${JSON.stringify(req.headers)}`);
      const { repoId } = body;
      if (!repoId) {
        logError(`[context] PUT /context/repo missing repoId, body was: ${JSON.stringify(body)}`);
        throw new Error("repoId is required");
      }

      return switchRepo(this.ctx, repoId);
    });

    /**
     * Switch stream (within current repo)
     * PUT /context/stream
     * Body: {streamId}.
     * Clears issue
     */
    this.app.put("/context/stream", async (req) => {
      const body = req.body as { streamId?: string };
      logInfo(`[context] PUT /context/stream body=${JSON.stringify(body)} headers=${JSON.stringify(req.headers)}`);
      const { streamId } = body;
      if (!streamId) {
        logError(`[context] PUT /context/stream missing streamId, body was: ${JSON.stringify(body)}`);
        throw new Error("streamId is required");
      }

      return switchStream(this.ctx, streamId);
    });

    /**
     * Switch issue (within current stream)
     * PUT /context/issue
     * Body: {issueId}
     */
    this.app.put("/context/issue", async (req) => {
      const body = req.body as { issueId?: string };
      logInfo(`[context] PUT /context/issue body=${JSON.stringify(body)} headers=${JSON.stringify(req.headers)}`);
      const { issueId } = body;
      if (!issueId) {
        logError(`[context] PUT /context/issue missing issueId, body was: ${JSON.stringify(body)}`);
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
