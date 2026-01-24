import type { FastifyInstance } from "fastify";
import type { ICommandContext } from "@crona/core";
/**
 * Stash HTTP APIs
 * Stash = suspended active_context (+ session cleanup)
 */
export declare class StashRoutes {
    private readonly app;
    private readonly ctx;
    constructor(app: FastifyInstance, ctx: ICommandContext);
    register(): void;
    private registerQueries;
    private registerCommands;
}
