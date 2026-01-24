import type { Stream, StreamVisibility } from "../../domain";
import type { IStreamRepository } from "./interface";
export declare class SqliteStreamRepository implements IStreamRepository {
    create(stream: Stream, meta: {
        userId: string;
        now: string;
    }): Promise<Stream>;
    getById(streamId: string, userId: string): Promise<Stream | null>;
    listByRepo(repoId: string, userId: string): Promise<Stream[]>;
    update(streamId: string, updates: {
        name?: string;
        visibility?: StreamVisibility;
    }, meta: {
        userId: string;
        now: string;
    }): Promise<Stream>;
    softDelete(streamId: string, meta: {
        userId: string;
        now: string;
    }): Promise<void>;
}
