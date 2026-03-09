import type { ActiveContext, Issue, Repo, ScratchPadMeta, Session, SessionSegmentType, Stash, Stream } from "../domain";
import type { TimerStatePayload } from "../timer";

export type KernelEvent =
  | { type: "repo.created"; payload: Repo }
  | { type: "repo.updated"; payload: Repo }
  | { type: "repo.deleted"; payload: { id: string } }
  | { type: "stream.created"; payload: Stream }
  | { type: "stream.updated"; payload: Stream }
  | { type: "stream.deleted"; payload: { id: string } }
  | { type: "issue.created"; payload: Issue }
  | { type: "issue.updated"; payload: Issue }
  | { type: "issue.deleted"; payload: { id: string } }
  | { type: "session.started"; payload: Session }
  | { type: "session.stopped"; payload: Session }
  | { type: "timer.state"; payload: TimerStatePayload }
  | { type: "context.repo.changed";   payload: { deviceId: string; repoId?: string | undefined; streamId?: string | undefined; issueId?: string | undefined } }
  | { type: "context.stream.changed"; payload: { deviceId: string; repoId?: string | undefined; streamId?: string | undefined; issueId?: string | undefined } }
  | { type: "context.issue.changed";  payload: { deviceId: string; repoId?: string | undefined; streamId?: string | undefined; issueId?: string | undefined } }
  | { type: "context.cleared";        payload: { deviceId: string } }
  | { type: "stash.created" | "stash.applied"; payload: Pick<Stash, "id" | "deviceId" | "repoId" | "streamId" | "issueId"> }
  | { type: "stash.dropped"; payload: { id: string; deviceId: string } }
  | { type: "timer.boundary"; payload: { from: SessionSegmentType; to: SessionSegmentType } }
  | { type: "timer.tick"; payload: { remainingSeconds: number } }
  | { type: "scratchpad.created"; payload: ScratchPadMeta }
  | { type: "scratchpad.updated"; payload: ScratchPadMeta }
  | { type: "scratchpad.deleted"; payload: { id: string } }
  ;
