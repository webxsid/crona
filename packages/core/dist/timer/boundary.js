"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.computeNextBoundary = computeNextBoundary;
/**
 * Decide the next boundary from current segment + settings
 */
function computeNextBoundary(current, settings, completedWorkCycles) {
    if (settings.timer_mode !== "structured")
        return null;
    if (!settings.breaks_enabled)
        return null;
    if (current === "work") {
        const isLongBreak = settings.long_break_enabled &&
            completedWorkCycles > 0 &&
            completedWorkCycles % settings.cycles_before_long_break === 0;
        return {
            nextSegment: isLongBreak ? "long_break" : "short_break",
            afterMinutes: settings.work_duration_minutes
        };
    }
    if (current === "short_break") {
        return {
            nextSegment: "work",
            afterMinutes: settings.short_break_minutes
        };
    }
    if (current === "long_break") {
        return {
            nextSegment: "work",
            afterMinutes: settings.long_break_minutes
        };
    }
    return null;
}
