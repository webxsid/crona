"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.TimerService = void 0;
const time_1 = require("./time");
const commands_1 = require("../commands");
const boundary_1 = require("./boundary");
class TimerService {
    ctx;
    constructor(ctx) {
        this.ctx = ctx;
    }
    boundaryTimer = null;
    // singleton
    static instance = null;
    static getInstance(ctx) {
        if (!TimerService.instance) {
            TimerService.instance = new TimerService(ctx);
        }
        return TimerService.instance;
    }
    /**
     * Authoritative timer state
     * Derived ONLY from sessions + session_segments
     */
    async getState() {
        const now = this.ctx.now();
        const activeSession = await this.ctx.sessions.getActiveSession(this.ctx.userId);
        if (!activeSession) {
            return { state: "idle" };
        }
        const activeSegment = await this.ctx.sessionSegments.getActive(this.ctx.userId, this.ctx.deviceId, activeSession.id);
        if (!activeSegment) {
            // Should not normally happen, but safe fallback
            return {
                state: "running",
                sessionId: activeSession.id,
                issueId: activeSession.issueId,
                segmentType: "work",
                elapsedSeconds: (0, time_1.elapsedSeconds)(activeSession.startTime, now),
            };
        }
        const paused = activeSegment.segmentType !== "work";
        return {
            state: paused ? "paused" : "running",
            sessionId: activeSession.id,
            issueId: activeSession.issueId,
            segmentType: activeSegment.segmentType,
            elapsedSeconds: (0, time_1.elapsedSeconds)(activeSegment.startTime.toISOString(), now) + (activeSegment.elapsedOffsetSeconds || 0),
        };
    }
    /**
     * Start session (delegates to command)
     */
    async start(issueId) {
        const active = await this.ctx.sessions.getActiveSession(this.ctx.userId);
        if (active) {
            throw new Error("Cannot start a new session while another session is active. Please end the current session first.");
        }
        let resolvedIssueId = issueId;
        if (!resolvedIssueId) {
            const context = await this.ctx.activeContext.get(this.ctx.userId, this.ctx.deviceId);
            resolvedIssueId = context?.issueId;
        }
        if (!resolvedIssueId) {
            throw new Error("No issue specified and no active issue in context. Cannot start session.");
        }
        await (0, commands_1.startSession)(this.ctx, resolvedIssueId);
        await this.scheduleNextBoundary();
        if (issueId && issueId !== resolvedIssueId) {
            // Update active context to reflect the explicitly started issue
            await this.ctx.activeContext.set(this.ctx.userId, this.ctx.deviceId, {
                issueId: resolvedIssueId,
            });
            this.ctx.events.emit({
                type: "context.changed",
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
        await (0, commands_1.pauseSession)(this.ctx);
        await this.scheduleNextBoundary();
        const state = await this.getState();
        this.ctx.events.emit({ type: "timer.state", payload: state });
        return state;
    }
    /**
     * Resume = rest → work (delegates to command)
     */
    async resume() {
        await (0, commands_1.resumeSession)(this.ctx);
        await this.scheduleNextBoundary();
        const state = await this.getState();
        this.ctx.events.emit({ type: "timer.state", payload: state });
        return state;
    }
    /**
     * End = close segment + stop session (delegates)
     */
    async end(commitMessage) {
        await (0, commands_1.stopSession)(this.ctx, commitMessage);
        if (this.boundaryTimer) {
            clearTimeout(this.boundaryTimer);
            this.boundaryTimer = null;
        }
        const state = await this.getState();
        this.ctx.events.emit({ type: "timer.state", payload: state });
        return state;
    }
    async restoreFromStash(input) {
        // 1. Start a NEW session
        const session = await (0, commands_1.startSession)(this.ctx, input.issueId);
        // 2. Start a NEW segment (now-based)
        await this.ctx.sessionSegments.startSegment(this.ctx.userId, this.ctx.deviceId, session.id, input.segmentType);
        // 3. Adjust elapsed time (logical offset)
        await this.ctx.sessionSegments.applyElapsedOffset(session.id, input.elapsedSeconds);
        // 4. Reschedule boundaries (CRITICAL)
        await this.scheduleNextBoundary();
        const state = await this.getState();
        this.ctx.events.emit({ type: "timer.state", payload: state });
    }
    scheduleBoundary(delayMs, callback) {
        if (this.boundaryTimer) {
            clearTimeout(this.boundaryTimer);
        }
        this.boundaryTimer = setTimeout(() => {
            callback().catch(console.error);
        }, delayMs);
    }
    async scheduleNextBoundary() {
        const activeSession = await this.ctx.sessions.getActiveSession(this.ctx.userId);
        if (!activeSession)
            return;
        const activeSegment = await this.ctx.sessionSegments.getActive(this.ctx.userId, this.ctx.deviceId, activeSession.id);
        if (!activeSegment)
            return;
        // user-initiated rest overrides system boundaries
        if (activeSegment.segmentType === "rest")
            return;
        const settings = await this.ctx.coreSettings.getAllSettings();
        if (!settings)
            return;
        const completedCycles = await this.ctx.sessionSegments.countWorkSegments(activeSession.id);
        const boundary = (0, boundary_1.computeNextBoundary)(activeSegment.segmentType, settings[this.ctx.userId], completedCycles);
        if (!boundary)
            return;
        const delayMs = boundary.afterMinutes * 60 * 1000;
        this.scheduleBoundary(delayMs, async () => {
            // recheck state (crash / race safe)
            const current = await this.ctx.sessionSegments.getActive(this.ctx.userId, this.ctx.deviceId, activeSession.id);
            if (!current)
                return;
            if (current.segmentType !== activeSegment.segmentType)
                return;
            // perform transition via session command
            if (boundary.nextSegment === "work") {
                await (0, commands_1.resumeSession)(this.ctx);
            }
            else {
                await (0, commands_1.pauseSession)(this.ctx, boundary.nextSegment); // will become break via command
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
exports.TimerService = TimerService;
