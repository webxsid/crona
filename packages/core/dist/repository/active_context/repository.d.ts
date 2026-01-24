import type { ActiveContext } from "../../domain";
import type { IActiveContextRepository } from "./interface";
export declare class ActiveContextRepository implements IActiveContextRepository {
    get(userId: string, deviceId: string): Promise<ActiveContext | null>;
    set(userId: string, deviceId: string, context: {
        repoId?: string;
        streamId?: string;
        issueId?: string;
    }): Promise<ActiveContext>;
    clear(userId: string, deviceId: string): Promise<void>;
    initializeDefaults(userId: string, deviceId: string): Promise<void>;
}
