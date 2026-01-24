import type { FastifyInstance } from "fastify";
import type { ICommandContext } from "@crona/core";
/**
 * Stream HTTP APIs
 * - Queries
 * - Commands
 */
export declare class StreamRoutes {
    private readonly app;
    private readonly ctx;
    constructor(app: FastifyInstance, ctx: ICommandContext);
    register(): void;
    private registerQueries;
    private registerCommands;
}
