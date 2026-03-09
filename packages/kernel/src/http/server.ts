import type { FastifyInstance, FastifyReply, FastifyRequest } from "fastify";
import Fastify from "fastify";
import { registerAuth } from "./auth";
import type { ICommandContext } from "@crona/core";
import { ContextRoutes, EventsRoutes, HealthRoutes, IssueRoutes, KernelRoutes, OpsRoutes, RepoRoutes, ScratchRoutes, SessionRoutes, SettingsRoutes, StashRoutes, StreamRoutes, TimerRoutes } from "./routes";
import { logError, logInfo } from "../logger";

export function createHttpServer(
  ctx: ICommandContext & { authToken: string }
): FastifyInstance {
  const app = Fastify({
    logger: false,
  });

  // Log every request + response
  app.addHook("onRequest", async (req: FastifyRequest) => {
    logInfo(`→ ${req.method} ${req.url}`);
  });

  app.addHook("onResponse", async (req: FastifyRequest, reply: FastifyReply) => {
    logInfo(`← ${req.method} ${req.url} ${reply.statusCode}`);
  });

  // Catch all unhandled route errors
  app.setErrorHandler((err: Error, req, reply) => {
    logError(`${req.method} ${req.url} failed`, err);
    console.error(`[kernel] ${req.method} ${req.url} error:`, err);
    void reply.status(500).send({ error: err.message });
  });

  // Auth (Bearer token)
  registerAuth(app, ctx.authToken);

  // Routes
  new EventsRoutes(app, ctx).register();
  new HealthRoutes(app, ctx).register();
  new RepoRoutes(app, ctx).register();
  new StreamRoutes(app, ctx).register();
  new IssueRoutes(app, ctx).register();
  new SessionRoutes(app, ctx).register();
  new StashRoutes(app, ctx).register();
  new TimerRoutes(app, ctx).register();
  new OpsRoutes(app, ctx).register();
  new ContextRoutes(app, ctx).register();
  new SettingsRoutes(app, ctx).register();
  new ScratchRoutes(app, ctx).register();
  new KernelRoutes(app, ctx).register();

  return app;
}
