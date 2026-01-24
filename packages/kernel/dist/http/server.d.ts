import type { FastifyInstance } from "fastify";
import type { ICommandContext } from "@crona/core";
export declare function createHttpServer(ctx: ICommandContext & {
    authToken: string;
}): FastifyInstance;
