import type { Op } from "../../domain";
import type { IOpRepository } from "./interface";
export declare class SqliteOpRepository implements IOpRepository {
    append(op: Op): Promise<void>;
    latest(limit: number): Promise<Op[]>;
    listSince(userId: string, sinceTimestamp: string): Promise<Op[]>;
    listByEntity(entity: Op["entity"], entityId: string, userId: string, limit: number | undefined): Promise<Op[]>;
    private mapRow;
}
