import type { FastifyInstance } from "fastify";
import { type ICommandContext } from "@crona/core";
export declare class ScratchRoutes {
    private readonly app;
    private readonly ctx;
    constructor(app: FastifyInstance, ctx: ICommandContext);
    register(): void;
    private registerQueries;
    private registerCommands;
}
