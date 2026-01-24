import type { FastifyInstance } from "fastify";
import type { ICommandContext } from "@crona/core";
/**
 * Active context HTTP APIs
 * Git-like working context:
 * repo → stream → issue
 */
export declare class ContextRoutes {
    private readonly app;
    private readonly ctx;
    constructor(app: FastifyInstance, ctx: ICommandContext);
    register(): void;
    private registerQueries;
    private registerCommands;
}
