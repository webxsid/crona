import type { ICommandContext } from "../commands/context";
import type { TimerStatePayload } from "./timer.types";
import type { SessionSegmentType } from "../domain";
export declare class TimerService {
    private readonly ctx;
    constructor(ctx: ICommandContext);
    private boundaryTimer;
    private static instance;
    static getInstance(ctx: ICommandContext): TimerService;
    /**
     * Authoritative timer state
     * Derived ONLY from sessions + session_segments
     */
    getState(): Promise<TimerStatePayload>;
    /**
     * Start session (delegates to command)
     */
    start(issueId?: string): Promise<TimerStatePayload>;
    /**
     * Pause = work → rest (delegates to command)
     */
    pause(): Promise<TimerStatePayload>;
    /**
     * Resume = rest → work (delegates to command)
     */
    resume(): Promise<TimerStatePayload>;
    /**
     * End = close segment + stop session (delegates)
     */
    end(commitMessage?: string | undefined): Promise<TimerStatePayload>;
    restoreFromStash(input: {
        issueId: string;
        segmentType: SessionSegmentType;
        elapsedSeconds: number;
    }): Promise<void>;
    private scheduleBoundary;
    scheduleNextBoundary(): Promise<void>;
}
