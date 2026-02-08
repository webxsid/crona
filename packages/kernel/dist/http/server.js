"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.createHttpServer = createHttpServer;
const fastify_1 = __importDefault(require("fastify"));
const auth_1 = require("./auth");
const routes_1 = require("./routes");
function createHttpServer(ctx) {
    const app = (0, fastify_1.default)({
        logger: false,
    });
    // Auth (Bearer token)
    (0, auth_1.registerAuth)(app, ctx.authToken);
    // Routes
    new routes_1.EventsRoutes(app, ctx).register();
    new routes_1.HealthRoutes(app, ctx).register();
    new routes_1.RepoRoutes(app, ctx).register();
    new routes_1.StreamRoutes(app, ctx).register();
    new routes_1.IssueRoutes(app, ctx).register();
    new routes_1.SessionRoutes(app, ctx).register();
    new routes_1.StashRoutes(app, ctx).register();
    new routes_1.TimerRoutes(app, ctx).register();
    new routes_1.OpsRoutes(app, ctx).register();
    new routes_1.ContextRoutes(app, ctx).register();
    new routes_1.SettingsRoutes(app, ctx).register();
    new routes_1.ScratchRoutes(app, ctx).register();
    new routes_1.KernelRoutes(app, ctx).register();
    return app;
}
