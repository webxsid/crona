import type { FastifyInstance } from "fastify";
import type { ICommandContext } from "@crona/core";
export declare class KernelRoutes {
    private readonly app;
    private readonly _ctx;
    constructor(app: FastifyInstance, _ctx: ICommandContext);
    register(): void;
}
