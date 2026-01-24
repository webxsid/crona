import type { FastifyInstance } from "fastify";
import Fastify from "fastify";
import { registerAuth } from "./auth";
import type { ICommandContext } from "@crona/core";
import { ContextRoutes, EventsRoutes, HealthRoutes, IssueRoutes, OpsRoutes, RepoRoutes, SessionRoutes, SettingsRoutes, StashRoutes, StreamRoutes, TimerRoutes } from "./routes";

export function createHttpServer(
  ctx: ICommandContext & { authToken: string }
): FastifyInstance {
  const app = Fastify({
    logger: false,
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

  return app;
}
