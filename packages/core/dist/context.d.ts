import type { ICommandContext } from "./commands";
import type { EventBus } from "./events";
export declare function createCommandContext(input: {
    userId: string;
    deviceId: string;
    now: () => string;
    events: EventBus;
}): Promise<ICommandContext & {
    authToken: string;
}>;
