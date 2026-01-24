import type { FastifyInstance } from "fastify";
import type { ICommandContext } from "@crona/core";
/**
 * Ops HTTP APIs
 * - Read-only operation log
 */
export declare class OpsRoutes {
    private readonly app;
    private readonly ctx;
    constructor(app: FastifyInstance, ctx: ICommandContext);
    register(): void;
    private registerQueries;
}
