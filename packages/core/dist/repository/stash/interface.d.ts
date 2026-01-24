import type { Stash } from '../../domain';
export interface IStashRepository {
    list(userId: string): Promise<Stash[]>;
    get(id: string, userId: string): Promise<Stash | null>;
    save(stash: Stash): Promise<void>;
    delete(id: string, userId: string): Promise<void>;
}
