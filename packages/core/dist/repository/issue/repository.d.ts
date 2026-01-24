import type { Issue, IssueStatus } from "../../domain";
import type { IIssueRepository } from "./interface";
export declare class SqliteIssueRepository implements IIssueRepository {
    create(issue: Issue, meta: {
        userId: string;
        now: string;
    }): Promise<Issue>;
    getById(issueId: string, userId: string): Promise<Issue | null>;
    listByStream(streamId: string, userId: string): Promise<Issue[]>;
    update(issueId: string, updates: {
        title?: string;
        status?: IssueStatus;
        estimateMinutes?: number | null;
        notes?: string | null;
    }, meta: {
        userId: string;
        now: string;
    }): Promise<Issue>;
    softDelete(issueId: string, meta: {
        userId: string;
        now: string;
    }): Promise<void>;
}
