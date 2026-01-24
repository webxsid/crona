import type { FastifyInstance } from "fastify";
import type { ICommandContext } from "@crona/core";
/**
 * Server-Sent Events (SSE)
 * - Realtime kernel → client events
 */
export declare class EventsRoutes {
    private readonly app;
    private readonly ctx;
    constructor(app: FastifyInstance, ctx: ICommandContext);
    register(): void;
}
