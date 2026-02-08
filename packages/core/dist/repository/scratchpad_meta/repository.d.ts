import type { IScratchRepo } from "./interface";
import type { ScratchPadMeta } from "../../domain";
export declare class ScratchRepo implements IScratchRepo {
    upsert(meta: ScratchPadMeta, userMeta: {
        userId: string;
        deviceId: string;
    }): Promise<void>;
    list(userId: string, deviceId: string, options?: {
        pinnedOnly?: boolean;
    }): Promise<ScratchPadMeta[]>;
    get(path: string, meta: {
        userId: string;
        deviceId: string;
    }): Promise<ScratchPadMeta | null>;
    remove(path: string, meta: {
        userId: string;
        deviceId: string;
    }): Promise<void>;
}
