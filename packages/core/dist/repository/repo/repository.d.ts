import type { Repo } from "../../domain";
import type { IRepoRepository } from "./interface";
export declare class SqliteRepoRepository implements IRepoRepository {
    create(repo: Repo, meta: {
        userId: string;
        now: string;
    }): Promise<Repo>;
    getById(repoId: string, userId: string): Promise<Repo | null>;
    list(userId: string): Promise<Repo[]>;
    update(repoId: string, updates: {
        name?: string;
        color?: string;
    }, meta: {
        userId: string;
        deviceId: string;
        now: string;
    }): Promise<Repo>;
    softDelete(repoId: string, meta: {
        userId: string;
        now: string;
    }): Promise<void>;
}
