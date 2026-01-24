import type { ActiveContext, Issue, Repo, Session, SessionSegmentType, Stash, Stream } from "../domain";
import type { TimerStatePayload } from "../timer";
export type KernelEvent = {
    type: "repo.created";
    payload: Repo;
} | {
    type: "repo.updated";
    payload: Repo;
} | {
    type: "stream.created";
    payload: Stream;
} | {
    type: "issue.updated";
    payload: Issue;
} | {
    type: "session.started";
    payload: Session;
} | {
    type: "session.stopped";
    payload: Session;
} | {
    type: "timer.state";
    payload: TimerStatePayload;
} | {
    type: "context.changed";
    payload: Pick<ActiveContext, "deviceId" | "repoId" | "streamId" | "issueId">;
} | {
    type: "stash.created" | "stash.applied";
    payload: Pick<Stash, "id" | "deviceId" | "repoId" | "streamId" | "issueId">;
} | {
    type: "stash.dropped";
    payload: {
        id: string;
        deviceId: string;
    };
} | {
    type: "timer.boundary";
    payload: {
        from: SessionSegmentType;
        to: SessionSegmentType;
    };
};
