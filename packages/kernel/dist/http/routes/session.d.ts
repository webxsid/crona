import type { FastifyInstance } from "fastify";
import type { ICommandContext } from "@crona/core";
/**
 * Session HTTP APIs
 * Owns session lifecycle + segment transitions
 */
export declare class SessionRoutes {
    private readonly app;
    private readonly ctx;
    constructor(app: FastifyInstance, ctx: ICommandContext);
    register(): void;
    private registerQueries;
    private registerCommands;
}
