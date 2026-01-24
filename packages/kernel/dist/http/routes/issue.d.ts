import type { FastifyInstance } from "fastify";
import type { ICommandContext } from "@crona/core";
/**
 * Issue HTTP APIs
 * - Queries
 * - Commands
 */
export declare class IssueRoutes {
    private readonly app;
    private readonly ctx;
    constructor(app: FastifyInstance, ctx: ICommandContext);
    register(): void;
    private registerQueries;
    private registerCommands;
}
