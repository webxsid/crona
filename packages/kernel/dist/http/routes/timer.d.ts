import type { FastifyInstance } from "fastify";
import type { ICommandContext } from "@crona/core";
/**
 * Timer HTTP APIs
 * Read-only derived timer state
 */
export declare class TimerRoutes {
    private readonly app;
    private readonly _ctx;
    private readonly timer;
    constructor(app: FastifyInstance, _ctx: ICommandContext);
    register(): void;
    private registerQueries;
    private registerCommands;
}
