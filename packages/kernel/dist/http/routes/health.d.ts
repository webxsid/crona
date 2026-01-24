import type { ICommandContext } from "@crona/core";
import type { FastifyInstance } from "fastify";
export declare class HealthRoutes {
    private readonly app;
    private readonly ctx;
    constructor(app: FastifyInstance, ctx: ICommandContext);
    register(): void;
}
