import type { Repo } from "../../domain";
export interface IRepoRepository {
    create(repo: Repo, meta: {
        userId: string;
        now: string;
    }): Promise<Repo>;
    getById(repoId: string, userId: string): Promise<Repo | null>;
    list(userId: string): Promise<Repo[]>;
    update(repoId: string, updates: Partial<Omit<Repo, "id">>, meta: {
        userId: string;
        now: string;
    }): Promise<Repo>;
    softDelete(repoId: string, meta: {
        userId: string;
        now: string;
    }): Promise<void>;
}
