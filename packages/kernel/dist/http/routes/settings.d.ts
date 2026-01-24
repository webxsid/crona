import type { FastifyInstance } from "fastify";
import type { ICommandContext } from "@crona/core";
/**
 * Core settings HTTP APIs
 * - Timer mode
 * - Break configuration
 * - Future kernel-level behavior
 */
export declare class SettingsRoutes {
    private readonly app;
    private readonly ctx;
    constructor(app: FastifyInstance, ctx: ICommandContext);
    register(): void;
    private registerQueries;
    private registerCommands;
}
