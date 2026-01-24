import type { Stash } from "../../domain";
import type { IStashRepository } from "./interface";
export declare class StashRepository implements IStashRepository {
    list(userId: string): Promise<Stash[]>;
    get(id: string, userId: string): Promise<Stash | null>;
    save(stash: Stash): Promise<void>;
    delete(id: string, userId: string): Promise<void>;
    private map;
}
