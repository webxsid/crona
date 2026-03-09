import type { ICommandContext } from "../commands/context";
import type { TimerStatePayload } from "./timer.types";
import { elapsedSeconds } from "./time";
import { startSession, stopSession, pauseSession, resumeSession } from "../commands";
import { computeNextBoundary } from "./boundary";
import type { CoreSettingsTable } from "../storage";
import type { SessionSegmentType } from "../domain";


export class TimerService {
  constructor(private readonly ctx: ICommandContext) { }
  private boundaryTimer: NodeJS.Timeout | null = null;

  // singleton
  private static instance: TimerService | null = null;
  static getInstance(ctx: ICommandContext) {
    if (!TimerService.instance) {
      TimerService.instance = new TimerService(ctx);
    }
    return TimerService.instance;
  }

  /**
   * Authoritative timer state
   * Derived ONLY from sessions + session_segments
   */
  async getState(): Promise<TimerStatePayload> {
    const now = this.ctx.now();

    const activeSession = await this.ctx.sessions.getActiveSession(
      this.ctx.userId
    );

    if (!activeSession) {
      return { state: "idle" };
    }

    const activeSegment = await this.ctx.sessionSegments.getActive(
      this.ctx.userId,
      this.ctx.deviceId,
      activeSession.id
    );

    if (!activeSegment) {
      // Should not normally happen, but safe fallback
      return {
        state: "running",
        sessionId: activeSession.id,
        issueId: activeSession.issueId,
        segmentType: "work",
        elapsedSeconds: elapsedSeconds(activeSession.startTime, now),
      };
    }

    const paused = activeSegment.segmentType !== "work";

    return {
      state: paused ? "paused" : "running",
      sessionId: activeSession.id,
      issueId: activeSession.issueId,
      segmentType: activeSegment.segmentType,
      elapsedSeconds: elapsedSeconds(
        activeSegment.startTime.toISOString(),
        now
      ) + (activeSegment.elapsedOffsetSeconds || 0),
    };
  }

  /**
   * Start session (delegates to command)
   */
  async start(issueId?: string) {
    const active = await this.ctx.sessions.getActiveSession(this.ctx.userId);
    if (active) {
      throw new Error(
        "Cannot start a new session while another session is active. Please end the current session first."
      );
    }
    let resolvedIssueId = issueId;
    if (!resolvedIssueId) {
      const context = await this.ctx.activeContext.get(
        this.ctx.userId,
        this.ctx.deviceId
      );
      resolvedIssueId = context?.issueId;
    }

    if (!resolvedIssueId) {
      throw new Error(
        "No issue specified and no active issue in context. Cannot start session."
      );
    }

    await startSession(this.ctx, resolvedIssueId);
    await this.scheduleNextBoundary();

    if (issueId && issueId !== resolvedIssueId) {
      // Update active context to reflect the explicitly started issue
      await this.ctx.activeContext.set(this.ctx.userId, this.ctx.deviceId, {
        issueId: resolvedIssueId,
      });
      this.ctx.events.emit({
        type: "context.issue.changed",
        payload: {
          deviceId: this.ctx.deviceId,
          issueId: resolvedIssueId,
        },
      });
    }

    const state = await this.getState();
    this.ctx.events.emit({ type: "timer.state", payload: state });
    return state;
  }

  /**
   * Pause = work → rest (delegates to command)
   */
  async pause() {
    await pauseSession(this.ctx);
    await this.scheduleNextBoundary();

    const state = await this.getState();
    this.ctx.events.emit({ type: "timer.state", payload: state });
    return state;
  }

  /**
   * Resume = rest → work (delegates to command)
   */
  async resume() {
    await resumeSession(this.ctx);
    await this.scheduleNextBoundary();

    const state = await this.getState();
    this.ctx.events.emit({ type: "timer.state", payload: state });
    return state;
  }

  /**
   * End = close segment + stop session (delegates)
   */
  async end(commitMessage?: string | undefined) {
    await stopSession(this.ctx, commitMessage);
    if (this.boundaryTimer) {
      clearTimeout(this.boundaryTimer);
      this.boundaryTimer = null;
    }

    const state = await this.getState();
    this.ctx.events.emit({ type: "timer.state", payload: state });
    return state;
  }

  async restoreFromStash(input: {
    issueId: string;
    segmentType: SessionSegmentType;
    elapsedSeconds: number;
  }) {
    // 1. Start a NEW session
    const session = await startSession(this.ctx, input.issueId);

    // 2. Start a NEW segment (now-based)
    await this.ctx.sessionSegments.startSegment(
      this.ctx.userId,
      this.ctx.deviceId,
      session.id,
      input.segmentType
    );

    // 3. Adjust elapsed time (logical offset)
    await this.ctx.sessionSegments.applyElapsedOffset(
      session.id,
      input.elapsedSeconds
    );

    // 4. Reschedule boundaries (CRITICAL)
    await this.scheduleNextBoundary();

    const state = await this.getState();
    this.ctx.events.emit({ type: "timer.state", payload: state });
  }


  private scheduleBoundary(
    delayMs: number,
    callback: () => Promise<void>
  ) {
    if (this.boundaryTimer) {
      clearTimeout(this.boundaryTimer);
    }

    this.boundaryTimer = setTimeout(() => {
      callback().catch(console.error);
    }, delayMs);
  }

  async scheduleNextBoundary() {
    const activeSession = await this.ctx.sessions.getActiveSession(
      this.ctx.userId
    );
    if (!activeSession) return;

    const activeSegment = await this.ctx.sessionSegments.getActive(
      this.ctx.userId,
      this.ctx.deviceId,
      activeSession.id
    );
    if (!activeSegment) return;

    // user-initiated rest overrides system boundaries
    if (activeSegment.segmentType === "rest") return;

    const settings = await this.ctx.coreSettings.getAllSettings();
    if (!settings) return;

    const completedCycles =
      await this.ctx.sessionSegments.countWorkSegments(
        activeSession.id
      );


    const boundary = computeNextBoundary(
      activeSegment.segmentType,
      settings[this.ctx.userId] as CoreSettingsTable,
      completedCycles
    );

    if (!boundary) return;

    const delayMs = boundary.afterMinutes * 60 * 1000;

    this.scheduleBoundary(delayMs, async () => {
      // recheck state (crash / race safe)
      const current = await this.ctx.sessionSegments.getActive(
        this.ctx.userId,
        this.ctx.deviceId,
        activeSession.id
      );

      if (!current) return;
      if (current.segmentType !== activeSegment.segmentType) return;

      // perform transition via session command
      if (boundary.nextSegment === "work") {
        await resumeSession(this.ctx);
      } else {
        await pauseSession(this.ctx, boundary.nextSegment); // will become break via command
      }

      this.ctx.events.emit({
        type: "timer.boundary",
        payload: {
          from: activeSegment.segmentType,
          to: boundary.nextSegment,
        },
      });

      await this.scheduleNextBoundary();
    });
  }
}

